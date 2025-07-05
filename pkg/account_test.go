package circular_enterprise_apis

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)


func TestOpen(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		expectError   bool
		expectedError string
	}{
		{
			name:          "Valid Address",
			address:       "0x1234567890abcdef",
			expectError:   false,
			expectedError: "",
		},
		{
			name:          "Empty Address",
			address:       "",
			expectError:   true,
			expectedError: "Invalid address format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)
			err := acc.Open(tc.address)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				}
				if err.Error() != tc.expectedError {
					t.Errorf("Expected error message '%s', but got '%s'", tc.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if acc.Address != tc.address {
					t.Errorf("Expected address to be '%s', but got '%s'", tc.address, acc.Address)
				}
			}
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	testCases := []struct {
		name           string
		accountAddress string
		mockResponse   string
		mockStatusCode int
		expectSuccess  bool
		expectError    bool
		expectedNonce  int
	}{
		{
			name:           "Successful Update",
			accountAddress: "0x123",
			mockResponse:   `{"Result":200,"Response":{"Nonce":100}}`,
			mockStatusCode: http.StatusOK,
			expectSuccess:  true,
			expectError:    false,
			expectedNonce:  101,
		},
		{
			name:           "Account Not Open",
			accountAddress: "",
			mockResponse:   "",
			mockStatusCode: http.StatusOK,
			expectSuccess:  false,
			expectError:    true,
			expectedNonce:  0, // Nonce should not change from initial 0
		},
		{
			name:           "HTTP Error - Bad Status",
			accountAddress: "0x123",
			mockResponse:   "Internal Server Error",
			mockStatusCode: http.StatusInternalServerError,
			expectSuccess:  false,
			expectError:    true,
			expectedNonce:  0,
		},
		{
			name:           "Invalid JSON Response",
			accountAddress: "0x123",
			mockResponse:   `{"Result":200,"Response":{"Nonce":"invalid"}}`, // Nonce is not an int
			mockStatusCode: http.StatusOK,
			expectSuccess:  false,
			expectError:    true,
			expectedNonce:  0,
		},
		{
			name:           "Server Response Result Not 200",
			accountAddress: "0x123",
			mockResponse:   `{"Result":400,"Response":{"Nonce":100}}`,
			mockStatusCode: http.StatusOK,
			expectSuccess:  false,
			expectError:    true,
			expectedNonce:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.mockStatusCode)
				w.Write([]byte(tc.mockResponse))
			}))
			defer server.Close()

			acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)
			if tc.accountAddress != "" {
				acc.Open(tc.accountAddress)
			}

			// Override NAGURL and NetworkNode to point to the mock server
			// Note: In a real scenario, these might be configured via dependency injection
			// or environment variables for easier testing.
			acc.NAGURL = server.URL + "/NAG.php?cep=" // Adjust to match the expected URL structure
			acc.NetworkNode = "mock_node"             // This part of the URL is appended by UpdateAccount

			success, err := acc.UpdateAccount()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}

			if success != tc.expectSuccess {
				t.Errorf("Expected success to be %v, but got %v", tc.expectSuccess, success)
			}

			if acc.Nonce != tc.expectedNonce {
				t.Errorf("Expected Nonce to be %d, but got %d", tc.expectedNonce, acc.Nonce)
			}
		})
	}
}

func TestSetNetwork(t *testing.T) {
	// Define a default NAG URL to check if it gets updated correctly.
	const initialNagURL = "https://default.nag.url"

	testCases := []struct {
		name              string
		network           string
		mockResponse      string
		mockStatusCode    int
		initialNetworkURL string // Base URL for fetching network info
		expectError       bool
		expectedErrorMsg  string
		expectedNagURL    string
	}{
		{
			name:              "Successful Network Set",
			network:           "mainnet",
			mockResponse:      `{"status":"success", "url":"https://new.nag.url/"}`,
			mockStatusCode:    http.StatusOK,
			initialNetworkURL: "", // Will be replaced by mock server URL
			expectError:       false,
			expectedNagURL:    "https://new.nag.url/",
		},
		{
			name:              "Failed Network Set - API Error",
			network:           "invalidnet",
			mockResponse:      `{"status":"error", "message":"Invalid network specified"}`,
			mockStatusCode:    http.StatusOK,
			initialNetworkURL: "",
			expectError:       true,
			expectedErrorMsg:  "failed to set network: Invalid network specified",
			expectedNagURL:    initialNagURL, // Should not change
		},
		{
			name:              "Failed Network Set - HTTP Error",
			network:           "mainnet",
			mockResponse:      "Internal Server Error",
			mockStatusCode:    http.StatusInternalServerError,
			initialNetworkURL: "",
			expectError:       true,
			expectedErrorMsg:  "network request failed with status: 500 Internal Server Error",
			expectedNagURL:    initialNagURL, // Should not change
		},
		{
			name:              "Failed Network Set - Invalid JSON",
			network:           "mainnet",
			mockResponse:      `{"status":"success", "url":}`,
			mockStatusCode:    http.StatusOK,
			initialNetworkURL: "",
			expectError:       true,
			expectedErrorMsg:  "failed to decode network response", // Check for substring
			expectedNagURL:    initialNagURL,
		},
		{
			name:              "Failed Network Set - Invalid Base URL",
			network:           "mainnet",
			mockResponse:      "",
			mockStatusCode:    0,
			initialNetworkURL: "://invalid-url", // Malformed URL to trigger a parsing error
			expectError:       true,
			expectedErrorMsg:  "invalid network URL", // Check for substring
			expectedNagURL:    initialNagURL,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock server for tests that need it.
			var server *httptest.Server
			if tc.mockStatusCode > 0 {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockResponse))
				}))
				defer server.Close()
			}

			acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)
			acc.NAGURL = initialNagURL // Set initial NAG URL

			// Set the NetworkURL for the account.
			if tc.initialNetworkURL == "" && server != nil {
				// If no specific initialNetworkURL is provided, use the mock server's URL.
				acc.NetworkURL = server.URL + "/"
			} else {
				acc.NetworkURL = tc.initialNetworkURL
			}

			err := acc.SetNetwork(tc.network)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected an error but got nil")
				}
				if !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
			}

			if acc.NAGURL != tc.expectedNagURL {
				t.Errorf("Expected NAGURL to be '%s', but got '%s'", tc.expectedNagURL, acc.NAGURL)
			}
		})
	}
}

func TestSignData(t *testing.T) {
	// Generate a new private key for testing
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	testCases := []struct {
		name        string
		dataToSign  []byte
		privateKey  *secp256k1.PrivateKey
		expectError bool
	}{
		{
			name:        "Successful Signing",
			dataToSign:  []byte("test data to be signed"),
			privateKey:  privateKey,
			expectError: false,
		},
		{
			name:        "No Private Key",
			dataToSign:  []byte("some data"),
			privateKey:  nil, // Simulate no private key
			expectError: true,
		},
		{
			name:        "Empty Data",
			dataToSign:  []byte(""),
			privateKey:  privateKey,
			expectError: false, // Signing empty data should still work
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)
			acc.PrivateKey = tc.privateKey

			signature, err := acc.SignData(tc.dataToSign)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				}
				if signature != "" {
					t.Errorf("Expected empty signature but got: %s", signature)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if signature == "" {
					t.Errorf("Expected a non-empty signature but got empty")
				}
			}
		})
	}
}

func TestSignDataRFC6979(t *testing.T) {
	// Generate a new private key for testing.
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)
	acc.PrivateKey = privateKey

	data := []byte("test message for RFC 6979")

	// Sign the same data twice.
	sig1, err1 := acc.SignData(data)
	if err1 != nil {
		t.Fatalf("First signature generation failed: %v", err1)
	}

	sig2, err2 := acc.SignData(data)
	if err2 != nil {
		t.Fatalf("Second signature generation failed: %v", err2)
	}

	// According to RFC 6979, signatures must be deterministic.
	if sig1 != sig2 {
		t.Errorf(`Signatures are not deterministic. RFC 6979 requires that signing the same data with the same key produces the same signature.
Sig1: %s
Sig2: %s`, sig1, sig2)
	}
}

func TestClose(t *testing.T) {
	acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)

	// Populate fields with dummy values
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	acc.PrivateKey = privateKey
	acc.PublicKey = "testPublicKey"
	acc.Address = "testAddress"

	// Call the Close method
	acc.Close()

	// Assert that fields are cleared
	if acc.PrivateKey != nil {
		t.Errorf("Expected PrivateKey to be nil, but got %v", acc.PrivateKey)
	}
	if acc.PublicKey != "" {
		t.Errorf("Expected PublicKey to be empty, but got %s", acc.PublicKey)
	}
	if acc.Address != "" {
		t.Errorf("Expected Address to be empty, but got %s", acc.Address)
	}
}
func TestGetTransaction(t *testing.T) {
	testCases := []struct {
		name             string
		transactionHash  string
		mockResponse     string
		mockStatusCode   int
		nagURL           string
		expectError      bool
		expectedErrorMsg string
		expectedResult   map[string]interface{}
	}{
		{
			name:            "Successful GetTransaction",
			transactionHash: "0xabcdef123456",
			mockResponse:    `{"status":"success", "details":"transaction details"}`,
			mockStatusCode:  http.StatusOK,
			nagURL:          "http://localhost:8080",
			expectError:     false,
			expectedResult:  map[string]interface{}{"status": "success", "details": "transaction details"},
		},
		{
			name:             "NAGURL Not Set",
			transactionHash:  "0xabcdef123456",
			mockResponse:     "",
			mockStatusCode:   0,
			nagURL:           "",
			expectError:      true,
			expectedErrorMsg: "network is not set. Please call SetNetwork() first",
		},
		{
			name:             "HTTP Error - Non-200 Status",
			transactionHash:  "0xabcdef123456",
			mockResponse:     "Not Found",
			mockStatusCode:   http.StatusNotFound,
			nagURL:           "http://localhost:8080",
			expectError:      true,
			expectedErrorMsg: "network returned an error: 404 Not Found",
		},
		{
			name:             "Invalid JSON Response",
			transactionHash:  "0xabcdef123456",
			mockResponse:     `{"status":"success", "details":}`, // Malformed JSON
			mockStatusCode:   http.StatusOK,
			nagURL:           "http://localhost:8080",
			expectError:      true,
			expectedErrorMsg: "failed to decode transaction details",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var server *httptest.Server
			if tc.nagURL != "" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockResponse))
				}))
				defer server.Close()
			}

			acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)
			if tc.nagURL != "" {
				acc.NAGURL = server.URL
			} else {
				acc.NAGURL = ""
			}

			result, err := acc.GetTransaction(tc.transactionHash)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected an error but got nil")
				}
				if !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if len(result) != len(tc.expectedResult) {
					t.Errorf("Expected result map length %d, but got %d", len(tc.expectedResult), len(result))
				}
				for k, v := range tc.expectedResult {
					if result[k] != v {
						t.Errorf("Expected result[%s] to be %v, but got %v", k, v, result[k])
					}
				}
			}
		})
	}
}

func TestGetTransactionByID(t *testing.T) {
	testCases := []struct {
		name             string
		transactionID    string
		mockResponse     string
		mockStatusCode   int
		nagURL           string
		expectError      bool
		expectedErrorMsg string
		expectedResult   map[string]interface{}
	}{
		{
			name:           "Successful GetTransactionByID",
			transactionID:  "0xabcdef123456",
			mockResponse:   `{"status":"success", "details":"transaction details by ID"}`,
			mockStatusCode: http.StatusOK,
			nagURL:         "http://localhost:8080",
			expectError:    false,
			expectedResult: map[string]interface{}{"status": "success", "details": "transaction details by ID"},
		},
		{
			name:             "NAGURL Not Set",
			transactionID:    "0xabcdef123456",
			mockResponse:     "",
			mockStatusCode:   0,
			nagURL:           "",
			expectError:      true,
			expectedErrorMsg: "network is not set. Please call SetNetwork() first",
		},
		{
			name:             "HTTP Error - Non-200 Status",
			transactionID:    "0xabcdef123456",
			mockResponse:     "Not Found",
			mockStatusCode:   http.StatusNotFound,
			nagURL:           "http://localhost:8080",
			expectError:      true,
			expectedErrorMsg: "network returned a non-200 status code: 404 Not Found",
		},
		{
			name:             "Invalid JSON Response",
			transactionID:    "0xabcdef123456",
			mockResponse:     `{"status":"success", "details":}`, // Malformed JSON
			mockStatusCode:   http.StatusOK,
			nagURL:           "http://localhost:8080",
			expectError:      true,
			expectedErrorMsg: "failed to decode transaction JSON",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var server *httptest.Server
			if tc.nagURL != "" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockResponse))
				}))
				defer server.Close()
			}

			acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)
			if tc.nagURL != "" {
				acc.NAGURL = server.URL
			} else {
				acc.NAGURL = ""
			}

			result, err := acc.GetTransactionByID(tc.transactionID)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected an error but got nil")
				}
				if !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if len(result) != len(tc.expectedResult) {
					t.Errorf("Expected result map length %d, but got %d", len(tc.expectedResult), len(result))
				}
				for k, v := range tc.expectedResult {
					if result[k] != v {
						t.Errorf("Expected result[%s] to be %v, but got %v", k, v, result[k])
					}
				}
			}
		})
	}
}

func TestGetTransactionOutcome(t *testing.T) {
	testCases := []struct {
		name             string
		TxID             string
		timeoutSec       int
		mockResponses    []string
		mockStatusCodes  []int
		nagURL           string
		expectError      bool
		expectedErrorMsg string
		expectedOutcome  map[string]interface{}
	}{
		{
			name:            "Successful Outcome - Not Pending",
			TxID:            "0x123",
			timeoutSec:      5,
			mockResponses:   []string{`{"Result":200, "Response":{"Status":"Pending"}}`, `{"Result":200, "Response":{"Status":"Confirmed", "Value":100}}`},
			mockStatusCodes: []int{http.StatusOK, http.StatusOK},
			nagURL:          "http://localhost:8080",
			expectError:     false,
			expectedOutcome: map[string]interface{}{"Status": "Confirmed", "Value": float64(100)},
		},
		{
			name:             "Timeout Exceeded",
			TxID:             "0x456",
			timeoutSec:       1,
			mockResponses:    []string{`{"Result":200, "Response":{"Status":"Pending"}}`, `{"Result":200, "Response":{"Status":"Pending"}}`},
			mockStatusCodes:  []int{http.StatusOK, http.StatusOK},
			nagURL:           "http://localhost:8080",
			expectError:      true,
			expectedErrorMsg: "timeout exceeded",
		},
		{
			name:             "NAGURL Not Set",
			TxID:             "0x789",
			timeoutSec:       1, // Short timeout since it should fail immediately
			mockResponses:    []string{},
			mockStatusCodes:  []int{},
			nagURL:           "",
			expectError:      true,
			expectedErrorMsg: "network is not set. Please call SetNetwork() first",
		},
		{
			name:             "HTTP Error During Polling",
			TxID:             "0xabc",
			timeoutSec:       5,
			mockResponses:    []string{`Internal Server Error`},
			mockStatusCodes:  []int{http.StatusInternalServerError},
			nagURL:           "http://localhost:8080",
			expectError:      true,
			expectedErrorMsg: "timeout exceeded", // Will eventually timeout if errors persist
		},
		{
			name:             "Invalid JSON Response During Polling",
			TxID:             "0xdef",
			timeoutSec:       5,
			mockResponses:    []string{`{"Result":200, "Response":{"Status":"Pending"}}`, `invalid json`},
			mockStatusCodes:  []int{http.StatusOK, http.StatusOK},
			nagURL:           "http://localhost:8080",
			expectError:      true,
			expectedErrorMsg: "timeout exceeded", // Will eventually timeout if errors persist
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestCount := 0
			var server *httptest.Server
			if tc.nagURL != "" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if requestCount < len(tc.mockResponses) {
						w.WriteHeader(tc.mockStatusCodes[requestCount])
						w.Write([]byte(tc.mockResponses[requestCount]))
						requestCount++
					} else {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"Result":200, "Response":{"Status":"Pending"}}`)) // Default to pending if no more mock responses
					}
				}))
				defer server.Close()
			}

			acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)
			acc.IntervalSec = 1 // Set a short interval for faster test execution
			if tc.nagURL != "" {
				acc.NAGURL = server.URL
			} else {
				acc.NAGURL = ""
			}

			outcome, err := acc.GetTransactionOutcome(tc.TxID, tc.timeoutSec)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected an error but got nil")
				}
				if !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if len(outcome) != len(tc.expectedOutcome) {
					t.Errorf("Expected outcome map length %d, but got %d", len(tc.expectedOutcome), len(outcome))
				}
				for k, v := range tc.expectedOutcome {
					if outcome[k] != v {
						t.Errorf("Expected outcome[%s] to be %v, but got %v", k, v, outcome[k])
					}
				}
			}
		})
	}
}
