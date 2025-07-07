package circular_enterprise_apis

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"encoding/json"

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
			expectedError: "invalid address format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			acc := NewCEPAccount()
			err := acc.Open(tc.address)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error message '%s', but got '%s'", tc.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %s", err)
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
		expectError    bool
		expectedError  string
		expectedNonce  int64
	}{
		{
			name:           "Successful Update",
			accountAddress: "0x123",
			mockResponse:   `{"Result":200,"Response":{"Nonce":100}}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedNonce:  101,
		},
		{
			name:           "Account Not Open",
			accountAddress: "",
			mockResponse:   "",
			mockStatusCode: http.StatusOK,
			expectError:    true,
			expectedError:  "account is not open",
			expectedNonce:  0, // Nonce should not change from initial 0
		},
		{
			name:           "HTTP Error - Bad Status",
			accountAddress: "0x123",
			mockResponse:   "Internal Server Error",
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedError:  "network request failed with status: 500 Internal Server Error",
			expectedNonce:  0,
		},
		{
			name:           "Invalid JSON Response",
			accountAddress: "0x123",
			mockResponse:   `{"Result":200,"Response":{"Nonce":"invalid"}}`, // Nonce is not an int
			mockStatusCode: http.StatusOK,
			expectError:    true,
			expectedError:  "failed to decode response body",
			expectedNonce:  0,
		},
		{
			name:           "Server Response Result Not 200",
			accountAddress: "0x123",
			mockResponse:   `{"Result":400,"Response":{"Nonce":100}}`,
			mockStatusCode: http.StatusOK,
			expectError:    true,
			expectedError:  "failed to update account: invalid response from server",
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

			acc := NewCEPAccount()
			if tc.accountAddress != "" {
				acc.Open(tc.accountAddress)
			}

			// Override NAGURL and NetworkNode to point to the mock server
			// Note: In a real scenario, these might be configured via dependency injection
			// or environment variables for easier testing.
			acc.NAGURL = server.URL + "/" + "/NAG.php?cep=" // Adjust to match the expected URL structure
			acc.NetworkNode = "mock_node"             // This part of the URL is appended by UpdateAccount

			err := acc.UpdateAccount()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected failure but got success")
				}
				// Check LastError if an error is expected
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error message '%s', but got '%s'", tc.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got failure: %s", err)
				}
			}

			if acc.Nonce != tc.expectedNonce {
				t.Errorf("Expected Nonce to be %d, but got %d", tc.expectedNonce, acc.Nonce)
			}
		})
	}
}

func TestSetNetwork(t *testing.T) {
	testCases := []struct {
		name              string
		network           string
		mockNAGResponse   string
		mockNAGStatusCode int
		expectError       bool
		expectedErrorMsg  string
		expectedNagURL    string
	}{
		{
			name:              "Successful Network Set",
			network:           "mainnet",
			mockNAGResponse:   `{"status":"success", "url":"https://new.nag.url/"}`,
			mockNAGStatusCode: http.StatusOK,
			expectError:       false,
			expectedNagURL:    "https://new.nag.url/",
		},
		{
			name:              "Failed Network Set - HTTP Error",
			network:           "mainnet",
			mockNAGResponse:   "Internal Server Error",
			mockNAGStatusCode: http.StatusInternalServerError,
			expectError:       true,
			expectedErrorMsg:  "failed to get NAG URL",
			expectedNagURL:    DefaultNAG, // Should not change
		},
		{
			name:              "Failed Network Set - Invalid JSON",
			network:           "mainnet",
			mockNAGResponse:   `{"url":}`,
			mockNAGStatusCode: http.StatusOK,
			expectError:       true,
			expectedErrorMsg:  "failed to get NAG URL", // The error message will now come from GetNAG's internal error handling
			expectedNagURL:    DefaultNAG,
		},
		{
			name:              "Failed Network Set - Invalid NAG Response Status",
			network:           "mainnet",
			mockNAGResponse:   `{"status":"error", "message":"Invalid network"}`,
			mockNAGStatusCode: http.StatusOK,
			expectError:       true,
			expectedErrorMsg:  "failed to get NAG URL", // The error message will now come from GetNAG's internal error handling
			expectedNagURL:    DefaultNAG,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock HTTP server for NetworkURL endpoint
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.mockNAGStatusCode)
				w.Write([]byte(tc.mockNAGResponse))
			}))
			defer server.Close()

			// Temporarily change NetworkURL to point to our mock server
			originalNetworkURL := NetworkURL
			NetworkURL = server.URL + "/network/getNAG?network="
			defer func() { NetworkURL = originalNetworkURL }()

			acc := NewCEPAccount()
			returnedNagURL, err := acc.SetNetwork(tc.network)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				if !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %s", err)
				}
				if returnedNagURL != tc.expectedNagURL {
					t.Errorf("Expected returned NAGURL to be '%s', but got '%s'", tc.expectedNagURL, returnedNagURL)
				}
				if acc.NAGURL != tc.expectedNagURL {
					t.Errorf("Expected acc.NAGURL to be '%s', but got '%s'", tc.expectedNagURL, acc.NAGURL)
				}
			}
		})
	}
}

func TestSubmitCertificate(t *testing.T) {
	// Generate a new private key for testing
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	privateKeyHex := hex.EncodeToString(privateKey.Serialize())

	testCases := []struct {
		name             string
		pdata            string
		privateKey       string
		mockResponse     string
		mockStatusCode   int
		nagURL           string
		expectError      bool
		expectedErrorMsg string
		expectedTxID     string
		expectedNonce    int64
	}{
		{
			name:            "Successful Submission",
			pdata:           "test data",
			privateKey:      privateKeyHex,
			mockResponse:    `{"Result":200, "Response":{"TxID":"0x12345"}}`,
			mockStatusCode:  http.StatusOK,
			nagURL:          "http://localhost:8080/", // Added trailing slash
			expectError:     false,
			expectedTxID:    "0x12345",
			expectedNonce:   1, // Nonce should increment after successful submission
		},
		{
			name:             "Account Not Open",
			pdata:            "test data",
			privateKey:       privateKeyHex,
			mockResponse:     "",
			mockStatusCode:   0,
			nagURL:           "http://localhost:8080/",
			expectError:      true,
			expectedErrorMsg: "account is not open",
			expectedTxID:     "",
			expectedNonce:    0,
		},
		{
			name:             "NAGURL Not Set",
			pdata:            "test data",
			privateKey:       privateKeyHex,
			mockResponse:     "",
			mockStatusCode:   0,
			nagURL:           "",
			expectError:      true,
			expectedErrorMsg: "network is not set",
			expectedTxID:     "",
			expectedNonce:    0,
		},
		{
			name:             "Invalid Private Key",
			pdata:            "test data",
			privateKey:       "invalid",
			mockResponse:     "",
			mockStatusCode:   0,
			nagURL:           "http://localhost:8080/",
			expectError:      true,
			expectedErrorMsg: "failed to sign data: invalid private key hex string",
			expectedTxID:     "",
			expectedNonce:    0,
		},
		{
			name:             "HTTP Error - Non-200 Status",
			pdata:            "test data",
			privateKey:       privateKeyHex,
			mockResponse:     "Internal Server Error",
			mockStatusCode:   http.StatusInternalServerError,
			nagURL:           "http://localhost:8080/",
			expectError:      true,
			expectedErrorMsg: "network returned an error - status: 500 Internal Server Error",
			expectedTxID:     "",
			expectedNonce:    0,
		},
		{
			name:             "Invalid JSON Response",
			pdata:            "test data",
			privateKey:       privateKeyHex,
			mockResponse:     `invalid json`,
			mockStatusCode:   http.StatusOK,
			nagURL:           "http://localhost:8080/",
			expectError:      true,
			expectedErrorMsg: "failed to decode response JSON",
			expectedTxID:     "",
			expectedNonce:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var server *httptest.Server
			if tc.nagURL != "" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Verify the request body
					if r.Method != "POST" {
						t.Errorf("Expected POST request, got %s", r.Method)
					}
					if !strings.Contains(r.URL.Path, "Circular_AddTransaction_") {
						t.Errorf("Expected URL path to contain 'Circular_AddTransaction_', got %s", r.URL.Path)
					}

					bodyBytes, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatalf("Failed to read request body: %v", err)
					}
					var requestPayload map[string]interface{}
					err = json.Unmarshal(bodyBytes, &requestPayload)
					if err != nil {
						t.Fatalf("Failed to unmarshal request payload: %v", err)
					}

					// Basic checks for required fields
					if _, ok := requestPayload["ID"]; !ok {
						t.Error("Request payload missing 'ID'")
					}
					if _, ok := requestPayload["From"]; !ok {
						t.Error("Request payload missing 'From'")
					}
					if _, ok := requestPayload["Payload"]; !ok {
						t.Error("Request payload missing 'Payload'")
					}
					if _, ok := requestPayload["Signature"]; !ok {
						t.Error("Request payload missing 'Signature'")
					}
					if _, ok := requestPayload["Nonce"]; !ok {
						t.Error("Request payload missing 'Nonce'")
					}

					w.WriteHeader(tc.mockStatusCode)
						w.Write([]byte(tc.mockResponse))
				}))
				defer server.Close()
			}

			acc := NewCEPAccount()
			if tc.nagURL != "" {
				acc.NAGURL = server.URL + "/"
				acc.NetworkNode = "test_node" // Set a dummy network node for the URL construction
				acc.Open("0x123")             // Open the account for successful cases
			} else {
				acc.NAGURL = ""
			}

			initialNonce := acc.Nonce
			initialTxID := acc.LatestTxID

			err = acc.SubmitCertificate(tc.pdata, tc.privateKey)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected an error but got none")
				}
				if !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
				// Ensure nonce and TxID are not updated on error
				if acc.Nonce != initialNonce {
					t.Errorf("Nonce changed on error. Expected %d, got %d", initialNonce, acc.Nonce)
				}
				if acc.LatestTxID != initialTxID {
					t.Errorf("LatestTxID changed on error. Expected %s, got %s", initialTxID, acc.LatestTxID)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %s", err)
				}

				// Verify TxID and Nonce updates
				if acc.LatestTxID == "" {
					t.Errorf("Expected LatestTxID to be set, but it was empty")
				}
				if acc.Nonce != tc.expectedNonce {
					t.Errorf("Expected Nonce to be %d, but got %d", tc.expectedNonce, acc.Nonce)
				}
			}
		})
	}
}

func TestGetTransactionByID(t *testing.T) {
	testCases := []struct {
		name             string
		transactionID    string
		startBlock       int64
		endBlock         int64
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
			startBlock:     0,
			endBlock:       10,
			mockResponse:   `{"status":"success", "details":"transaction details by ID"}`,
			mockStatusCode: http.StatusOK,
			nagURL:         "http://localhost:8080/", // Added trailing slash
			expectError:    false,
			expectedResult: map[string]interface{}{"status": "success", "details": "transaction details by ID"},
		},
		{
			name:           "Successful GetTransactionByID with Blocks",
			transactionID:  "0xabcdef123456",
			startBlock:     100,
			endBlock:       200,
			mockResponse:   `{"status":"success", "details":"transaction details by ID and blocks"}`,
			mockStatusCode: http.StatusOK,
			nagURL:         "http://localhost:8080/", // Added trailing slash
			expectError:    false,
			expectedResult: map[string]interface{}{"status": "success", "details": "transaction details by ID and blocks"},
		},
		{
			name:             "NAGURL Not Set",
			transactionID:    "0xabcdef123456",
			startBlock:       0,
			endBlock:         0,
			mockResponse:     "",
			mockStatusCode:   0,
			nagURL:           "",
			expectError:      true,
			expectedErrorMsg: "network is not set",
		},
		{
			name:             "HTTP Error - Non-200 Status",
			transactionID:    "0xabcdef123456",
			startBlock:       0,
			endBlock:         0,
			mockResponse:     "Not Found",
			mockStatusCode:   http.StatusNotFound,
			nagURL:           "http://localhost:8080/",
			expectError:      true,
			expectedErrorMsg: "network request failed with status: 404 Not Found",
		},
		{
			name:             "Invalid JSON Response",
			transactionID:    "0xabcdef123456",
			startBlock:       0,
			endBlock:         0,
			mockResponse:     `{"status":"success", "details":}`, // Malformed JSON
			mockStatusCode:   http.StatusOK,
			nagURL:           "http://localhost:8080/",
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

			acc := NewCEPAccount()
			if tc.nagURL != "" {
				acc.NAGURL = server.URL + "/"
			} else {
				acc.NAGURL = ""
			}

			result, err := acc.getTransactionByID(tc.transactionID, tc.startBlock, tc.endBlock)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected an error but got none")
				}
				if !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %s", err)
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

func TestGetTransaction(t *testing.T) {
	testCases := []struct {
		name             string
		blockID          string
		transactionID    string
		mockResponse     string
		mockStatusCode   int
		nagURL           string
		expectError      bool
		expectedErrorMsg string
		expectedResult   map[string]interface{}
	}{
		{
			name:           "Successful GetTransaction",
			blockID:        "12345",
			transactionID:  "0xabcdef123456",
			mockResponse:   `{"status":"success", "details":"transaction details"}`,
			mockStatusCode: http.StatusOK,
			nagURL:         "http://localhost:8080/",
			expectError:    false,
			expectedResult: map[string]interface{}{"status": "success", "details": "transaction details"},
		},
		{
			name:             "Empty Block ID",
			blockID:          "",
			transactionID:    "0xabcdef123456",
			mockResponse:     "",
			mockStatusCode:   0,
			nagURL:           "http://localhost:8080/",
			expectError:      true,
			expectedErrorMsg: "blockID cannot be empty",
		},
		{
			name:             "Invalid Block ID",
			blockID:          "invalid",
			transactionID:    "0xabcdef123456",
			mockResponse:     "",
			mockStatusCode:   0,
			nagURL:           "http://localhost:8080/",
			expectError:      true,
			expectedErrorMsg: "invalid blockID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var server *httptest.Server
			if tc.nagURL != "" && tc.mockStatusCode > 0 {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockResponse))
				}))
				defer server.Close()
			}

			acc := NewCEPAccount()
			if tc.nagURL != "" {
				if server != nil {
					acc.NAGURL = server.URL + "/"
				} else {
					acc.NAGURL = tc.nagURL
				}
			}

			result, err := acc.GetTransaction(tc.blockID, tc.transactionID)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected an error but got none")
				}
				if !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %s", err)
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
		intervalSec      int
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
			intervalSec:     1,
			mockResponses:   []string{`{"Result":200, "Response":{"Status":"Pending"}}`, `{"Result":200, "Response":{"Status":"Confirmed", "Value":100}}`},
			mockStatusCodes: []int{http.StatusOK, http.StatusOK},
			nagURL:          "http://localhost:8080/", // Added trailing slash
			expectError:     false,
			expectedOutcome: map[string]interface{}{"Status": "Confirmed", "Value": float64(100)},
		},
		{
			name:             "Timeout Exceeded",
			TxID:             "0x456",
			timeoutSec:       1,
			intervalSec:      1,
			mockResponses:    []string{`{"Result":200, "Response":{"Status":"Pending"}}`, `{"Result":200, "Response":{"Status":"Pending"}}`},
			mockStatusCodes:  []int{http.StatusOK, http.StatusOK},
			nagURL:           "http://localhost:8080/",
			expectError:      true,
			expectedErrorMsg: "timeout exceeded while waiting for transaction outcome",
		},
		{
			name:             "NAGURL Not Set",
			TxID:             "0x789",
			timeoutSec:       1, // Short timeout since it should fail immediately
			intervalSec:      1,
			mockResponses:    []string{},
			mockStatusCodes:  []int{},
			nagURL:           "",
			expectError:      true,
			expectedErrorMsg: "network is not set",
		},
		{
			name:             "HTTP Error During Polling",
			TxID:             "0xabc",
			timeoutSec:       5,
			intervalSec:      1,
			mockResponses:    []string{`Internal Server Error`},
			mockStatusCodes:  []int{http.StatusInternalServerError},
			nagURL:           "http://localhost:8080/",
			expectError:      true,
			expectedErrorMsg: "timeout exceeded while waiting for transaction outcome", // Will eventually timeout if errors persist
		},
		{
			name:             "Invalid JSON Response During Polling",
			TxID:             "0xdef",
			timeoutSec:       5,
			intervalSec:      1,
			mockResponses:    []string{`{"Result":200, "Response":{"Status":"Pending"}}`, `invalid json`},
			mockStatusCodes:  []int{http.StatusOK, http.StatusOK},
			nagURL:           "http://localhost:8080/",
			expectError:      true,
			expectedErrorMsg: "timeout exceeded while waiting for transaction outcome", // Will eventually timeout if errors persist
		},
		{
			name:             "Transaction Not Found During Polling",
			TxID:             "0xghi",
			timeoutSec:       5,
			intervalSec:      1,
			mockResponses:    []string{`{"Result":200, "Response":"Transaction Not Found"}`, `{"Result":200, "Response":{"Status":"Confirmed", "Value":200}}`},
			mockStatusCodes:  []int{http.StatusOK, http.StatusOK},
			nagURL:           "http://localhost:8080/",
			expectError:      false,
			expectedOutcome: map[string]interface{}{"Status": "Confirmed", "Value": float64(200)},
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

			acc := NewCEPAccount()
			acc.IntervalSec = tc.intervalSec // Set interval from test case
			if tc.nagURL != "" {
				acc.NAGURL = server.URL + "/"
			} else {
				acc.NAGURL = ""
			}

			outcome, err := acc.GetTransactionOutcome(tc.TxID, tc.timeoutSec, tc.intervalSec)

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected an error but got none")
				}
				if !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %s", err)
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

func TestSignData(t *testing.T) {
	// Generate a new private key for testing
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	privateKeyHex := hex.EncodeToString(privateKey.Serialize())

	testCases := []struct {
		name          string
		dataToSign    string
		privateKeyHex string
		expectError   bool
	}{
		{
			name:          "Successful Signing",
			dataToSign:    "test data to be signed",
			privateKeyHex: privateKeyHex,
			expectError:   false,
		},
		{
			name:          "Invalid Private Key Hex",
			dataToSign:    "some data",
			privateKeyHex: "invalidhex",
			expectError:   true,
		},
		{
			name:          "Empty Data",
			dataToSign:    "",
			privateKeyHex: privateKeyHex,
			expectError:   false, // Signing empty data should still work
		},
		{
			name:          "Account Not Open",
			dataToSign:    "some data",
			privateKeyHex: privateKeyHex,
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			acc := NewCEPAccount()
			if tc.name != "Account Not Open" {
				acc.Open("0x123")
			}

			signature, err := acc.signData(tc.dataToSign, tc.privateKeyHex)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				if signature != "" {
					t.Errorf("Expected empty signature but got: %s", signature)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %s", err)
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
	privateKeyHex := hex.EncodeToString(privateKey.Serialize())

	acc := NewCEPAccount()

	message := "test message for RFC 6979"

	// Sign the same data twice.
	sig1, err := acc.signData(message, privateKeyHex)
	if err != nil {
		t.Fatalf("First signature generation failed: %s", err)
	}

	sig2, err := acc.signData(message, privateKeyHex)
	if err != nil {
		t.Fatalf("Second signature generation failed: %s", err)
	}

	// According to RFC 6979, signatures must be deterministic.
	if sig1 != sig2 {
		t.Errorf(`Signatures are not deterministic. RFC 6979 requires that signing the same data with the same key produces the same signature.\nSig1: %s\nSig2: %s`, sig1, sig2)
	}
}

func TestClose(t *testing.T) {
	acc := NewCEPAccount()

	// Populate fields with dummy values
	
	acc.PublicKey = "testPublicKey"
	acc.Address = "testAddress"

	// Call the Close method
	acc.Close()

	// Assert that fields are cleared
	if acc.PublicKey != "" {
		t.Errorf("Expected PublicKey to be empty, but got %s", acc.PublicKey)
	}
	if acc.Address != "" {
		t.Errorf("Expected Address to be empty, but got %s", acc.Address)
	}
}