package certificate

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestSetData(t *testing.T) {
	testCases := []struct {
		name string
		data string
	}{
		{
			name: "Simple ASCII string",
			data: "hello world",
		},
		{
			name: "Empty string",
			data: "",
		},
		{
			name: "String with numbers and symbols",
			data: "123!@#$",
		},
		{
			name: "Unicode string",
			data: "你好, 世界",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cert := &Certificate{}
			cert.SetData(tc.data)

			expectedHex := hex.EncodeToString([]byte(tc.data))

			if cert.Data != expectedHex {
				t.Errorf("Expected Data to be '%s', but got '%s'", expectedHex, cert.Data)
			}
		})
	}
}

func TestGetData(t *testing.T) {
	testCases := []struct {
		name         string
		cert         Certificate
		expectedData string
		expectError  bool
	}{
		{
			name: "Simple ASCII string",
			cert: Certificate{
				Data: hex.EncodeToString([]byte("hello world")),
			},
			expectedData: "hello world",
			expectError:  false,
		},
		{
			name: "Unicode string",
			cert: Certificate{
				Data: hex.EncodeToString([]byte("你好, 世界")),
			},
			expectedData: "你好, 世界",
			expectError:  false,
		},
		{
			name: "Empty string",
			cert: Certificate{
				Data: "",
			},
			expectedData: "",
			expectError:  false,
		},
		{
			name: "Invalid hex data",
			cert: Certificate{
				Data: "this is not hex",
			},
			expectedData: "",
			expectError:  true,
		},
		{
			name: "Odd length hex string",
			cert: Certificate{
				Data: "123",
			},
			expectedData: "",
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.cert.GetData()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				if data != tc.expectedData {
					t.Errorf("Expected data to be '%s', but got '%s'", tc.expectedData, data)
				}
			}
		})
	}
}

func TestGetJSONCertificate(t *testing.T) {
	testCases := []struct {
		name         string
		cert         Certificate
		expectedJSON string
		expectError  bool
	}{
		{
			name: "Full certificate",
			cert: Certificate{
				Data:          hex.EncodeToString([]byte("hello world")),
				PreviousTxID:  "txid123",
				PreviousBlock: "block456",
				Version:       "1.0.0",
			},
			expectedJSON: `{"data":"68656c6c6f20776f726c64","previousTxID":"txid123","previousBlock":"block456","version":"1.0.0"}`,
			expectError:  false,
		},
		{
			name:         "Empty certificate",
			cert:         Certificate{},
			expectedJSON: `{"data":"","previousTxID":"","previousBlock":"","version":""}`,
			expectError:  false,
		},
		{
			name: "Certificate with some empty fields",
			cert: Certificate{
				Data:    hex.EncodeToString([]byte("some data")),
				Version: "1.1.0",
			},
			expectedJSON: `{"data":"736f6d652064617461","previousTxID":"","previousBlock":"","version":"1.1.0"}`,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonString, err := tc.cert.GetJSONCertificate()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Did not expect an error, but got: %v", err)
			}

			var resultData, expectedData map[string]interface{}

			err = json.Unmarshal([]byte(jsonString), &resultData)
			if err != nil {
				t.Fatalf("Failed to unmarshal actual JSON string: %v", err)
			}

			err = json.Unmarshal([]byte(tc.expectedJSON), &expectedData)
			if err != nil {
				t.Fatalf("Failed to unmarshal expected JSON string: %v", err)
			}

			if !reflect.DeepEqual(resultData, expectedData) {
				t.Errorf("Expected JSON \n%s\n, but got \n%s", tc.expectedJSON, jsonString)
			}
		})
	}
}

func TestGetCertificateSize(t *testing.T) {
	testCases := []struct {
		name         string
		cert         Certificate
		expectedSize int
		expectError  bool
	}{
		{
			name: "Full certificate",
			cert: Certificate{
				Data:          hex.EncodeToString([]byte("hello world")),
				PreviousTxID:  "txid123",
				PreviousBlock: "block456",
				Version:       "1.0.0",
			},
			expectedSize: 103,
			expectError:  false,
		},
		{
			name:         "Empty certificate",
			cert:         Certificate{},
			expectedSize: 61,
			expectError:  false,
		},
		{
			name: "Certificate with unicode data",
			cert: Certificate{
				Data: hex.EncodeToString([]byte("你好, 世界")),
			},
			expectedSize: 89,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			size, err := tc.cert.GetCertificateSize()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Did not expect an error, but got: %v", err)
			}

			if size != tc.expectedSize {
				t.Errorf("Expected size to be %d, but got %d", tc.expectedSize, size)
			}
		})
	}
}

const (
	DefaultNAG   = "http://localhost:8080"
	DefaultChain = "test-chain"
	LibVersion   = "1.0.0"
)

// Dummy CEPAccount and NewCEPAccount for testing purposes
type CEPAccount struct {
	NAGURL  string
	Chain   string
	Version string
}

func NewCEPAccount(nagURL, chain, version string) *CEPAccount {
	return &CEPAccount{
		NAGURL:  nagURL,
		Chain:   chain,
		Version: version,
	}
}

// Dummy SubmitCertificate method for testing
func (acc *CEPAccount) SubmitCertificate(cert *Certificate) (map[string]interface{}, error) {
	if acc.NAGURL == "" {
		return nil, // Simulate error for missing network
			fmt.Errorf("network is not set. Please call SetNetwork() first")
	}
	// Simulate HTTP request/response for testing
	reqBody, _ := json.Marshal(cert)
	resp, err := http.Post(acc.NAGURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("network returned an error - status: %s, body: %s", resp.Status, string(body))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}
	return result, nil
}

func TestSubmitCertificate(t *testing.T) {
	testCases := []struct {
		name           string
		nagURL         string
		mockResponse   string
		mockStatusCode int
		cert           *Certificate
		expectError    bool
		expectedError  string
		expectedResult map[string]interface{}
	}{
		{
			name:           "Successful Certificate Submission",
			nagURL:         "http://localhost:8080",
			mockResponse:   `{"txHash": "0xabc123", "status": "success"}`,
			mockStatusCode: http.StatusOK,
			cert: &Certificate{
				Version: "1.0",
				Data:    "test data",
			},
			expectError:    false,
			expectedResult: map[string]interface{}{"txHash": "0xabc123", "status": "success"},
		},
		{
			name:           "NAGURL Not Set",
			nagURL:         "",
			mockResponse:   "",
			mockStatusCode: http.StatusOK,
			cert: &Certificate{
				Version: "1.0",
				Data:    "test data",
			},
			expectError:   true,
			expectedError: "network is not set. Please call SetNetwork() first",
		},
		{
			name:           "HTTP Error - Bad Status",
			nagURL:         "http://localhost:8080",
			mockResponse:   "Internal Server Error",
			mockStatusCode: http.StatusInternalServerError,
			cert: &Certificate{
				Version: "1.0",
				Data:    "test data",
			},
			expectError:   true,
			expectedError: "network returned an error - status: 500 Internal Server Error, body: Internal Server Error",
		},
		{
			name:           "Invalid JSON Response",
			nagURL:         "http://localhost:8080",
			mockResponse:   `{"txHash": "0xabc123", "status": "success"`, // Malformed JSON
			mockStatusCode: http.StatusOK,
			cert: &Certificate{
				Version: "1.0",
				Data:    "test data",
			},
			expectError:   true,
			expectedError: "failed to decode response JSON",
		},
		{
			name:           "Empty Certificate Data",
			nagURL:         "http://localhost:8080",
			mockResponse:   `{"txHash": "0xabc123", "status": "success"}`,
			mockStatusCode: http.StatusOK,
			cert:           &Certificate{}, // Empty certificate
			expectError:    false,
			expectedResult: map[string]interface{}{"txHash": "0xabc123", "status": "success"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock HTTP server if a NAGURL is provided
			var server *httptest.Server
			var capturedRequestBody []byte
			if tc.nagURL != "" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var err error
					capturedRequestBody, err = io.ReadAll(r.Body)
					if err != nil {
						t.Fatalf("Failed to read request body: %v", err)
					}
					w.WriteHeader(tc.mockStatusCode)
					w.Write([]byte(tc.mockResponse))
				}))
				defer server.Close()
			}

			// Import the necessary constants and NewCEPAccount from the appropriate package if available.
			// For example:
			// import "Circular-Protocol/CEP-Go-APIs/pkg/circular_enterprise_apis"
			// acc := circular_enterprise_apis.NewCEPAccount(circular_enterprise_apis.DefaultNAG, circular_enterprise_apis.DefaultChain, circular_enterprise_apis.LibVersion)

			// If the above import is set up, use:
			// acc := circular_enterprise_apis.NewCEPAccount(circular_enterprise_apis.DefaultNAG, circular_enterprise_apis.DefaultChain, circular_enterprise_apis.LibVersion)

			// Otherwise, fallback to the local definition (as before):
			acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)
			if tc.nagURL != "" {
				acc.NAGURL = server.URL
			} else {
				acc.NAGURL = tc.nagURL
			}

			result, err := acc.SubmitCertificate(tc.cert)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				}
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error message to contain '%s', but got '%s'", tc.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				// Compare the result map
				if len(result) != len(tc.expectedResult) {
					t.Errorf("Expected result map length %d, but got %d", len(tc.expectedResult), len(result))
				}
				for k, v := range tc.expectedResult {
					if result[k] != v {
						t.Errorf("Expected result[%s] to be %v, but got %v", k, v, result[k])
					}
				}

				// Verify the request body sent to the mock server
				expectedRequestBody, marshalErr := json.Marshal(tc.cert)
				if marshalErr != nil {
					t.Fatalf("Failed to marshal expected certificate: %v", marshalErr)
				}
				if !bytes.Equal(capturedRequestBody, expectedRequestBody) {
					t.Errorf("Request body mismatch. Expected %s, got %s", string(expectedRequestBody), string(capturedRequestBody))
				}
			}
		})
	}
}
