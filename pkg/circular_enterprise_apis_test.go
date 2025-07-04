package circular_enterprise_apis

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1"
	"github.com/lessuseless/Circular-Protocol/CEP-Go-APIs/pkg/certificate"
	"github.com/lessuselesss/CEP-Go-APIs/pkg/account"
)

// TestNewCEP tests the factory function for creating CEP instances.
func TestNewCEP(t *testing.T) {
	// Test case 1: Using default values for public network
	t.Run("DefaultInitialization", func(t *testing.T) {
		cep := NewCEP("", "")
		if cep.nagURL != DefaultNAG {
			t.Errorf("Expected default NAG URL %s, but got %s", DefaultNAG, cep.nagURL)
		}
		if cep.chain != DefaultChain {
			t.Errorf("Expected default chain %s, but got %s", DefaultChain, cep.chain)
		}
	})

	// Test case 2: Providing custom values for a private network
	t.Run("CustomInitialization", func(t *testing.T) {
		customNAG := "http://localhost:8080/nag"
		customChain := "0xcustomchainhash"
		cep := NewCEP(customNAG, customChain)
		if cep.nagURL != customNAG {
			t.Errorf("Expected custom NAG URL %s, but got %s", customNAG, cep.nagURL)
		}
		if cep.chain != customChain {
			t.Errorf("Expected custom chain %s, but got %s", customChain, cep.chain)
		}
	})
}

// TestNewAccount verifies that the NewAccount method correctly initializes an Account instance
// with the parent CEP's network configuration.
func TestNewAccount(t *testing.T) {
	cep := NewCEP(DefaultNAG, DefaultChain)
	acc := cep.NewAccount()

	if acc == nil {
		t.Fatal("NewAccount returned nil")
	}
	// Note: This test relies on the internal structure of the Account object.
	// To make this more robust, the Account object could expose its configuration via methods.
	// For now, we assume we can infer state for testing purposes.
}

// TestNewCertificate verifies that the NewCertificate method correctly initializes a Certificate instance
// with the parent CEP's network configuration.
func TestNewCertificate(t *testing.T) {
	cep := NewCEP(DefaultNAG, DefaultChain)
	cert := cep.NewCertificate()

	if cert == nil {
		t.Fatal("NewCertificate returned nil")
	}
	// Similar to TestNewAccount, this test assumes we can validate the initialized object.
}

// mockRoundTripper is a custom http.RoundTripper for mocking HTTP responses in tests.
type mockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

// errorReader is an io.Reader that always returns an error.
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("i/o error")
}

// TestGetNAG tests the network discovery function.
func TestGetNAG(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	defer func() {
		http.DefaultClient.Transport = originalTransport
	}()

	t.Run("SuccessfulDiscovery", func(t *testing.T) {
		expectedNAG := "https://test-nag.circularlabs.io/"
		http.DefaultClient.Transport = &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(expectedNAG)),
			},
		}

		nag, err := GetNAG("testnet")
		if err != nil {
			t.Fatalf("Expected no error, but got %v", err)
		}
		if nag != expectedNAG {
			t.Errorf("Expected NAG URL %s, but got %s", expectedNAG, nag)
		}
	})

	t.Run("FailedDiscovery", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			err: fmt.Errorf("network error"),
		}

		_, err := GetNAG("testnet")
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}
	})

	t.Run("Non200StatusCode", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("Not Found")),
			},
		}

		body, err := GetNAG("testnet")
		if err != nil {
			t.Fatalf("Expected no error based on current implementation, but got %v", err)
		}
		if body != "Not Found" {
			t.Errorf("Expected body 'Not Found', but got '%s'", body)
		}
	})

	t.Run("BodyReadError", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(&errorReader{}),
			},
		}

		_, err := GetNAG("testnet")
		if err == nil {
			t.Fatal("Expected an error due to body read failure, but got none")
		}
	})

	t.Run("EmptyNetworkString", func(t *testing.T) {
		expectedNAG := "https://default-nag.circularlabs.io/"
		http.DefaultClient.Transport = &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(expectedNAG)),
			},
		}

		nag, err := GetNAG("")
		if err != nil {
			t.Fatalf("Expected no error for empty network string, but got %v", err)
		}
		if nag != expectedNAG {
			t.Errorf("Expected NAG URL %s, but got %s", expectedNAG, nag)
		}
	})
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

func TestCircularOperations(t *testing.T) {
	privateKeyHex := os.Getenv("CIRCULAR_PRIVATE_KEY")
	address := os.Getenv("CIRCULAR_ADDRESS")
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	// Create a mock server to handle network requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "GetWalletNonce") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Result":200,"Response":{"Nonce":1}}`))
		} else if strings.Contains(r.URL.Path, "transaction") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Result":200,"Response":{"Status":"Confirmed"}}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"txHash":"0x12345","status":"success"}`))
		}
	}))
	defer server.Close()

	acc := account.NewCEPAccount(server.URL, "testnet", "1.0")

	// Decode the private key and set it on the account
	pkBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}
	acc.PrivateKey = secp256k1.PrivKeyFromBytes(pkBytes)

	err = acc.Open(address)
	if err != nil {
		t.Fatalf("acc.Open() failed: %v", err)
	}

	ok, err := acc.UpdateAccount()
	if !ok || err != nil {
		t.Fatalf("acc.UpdateAccount() failed: ok=%v, err=%v", ok, err)
	}

	cert := certificate.NewCertificate(acc.Blockchain, acc.CodeVersion)
	cert.SetData("test message")

	resp, err := acc.SubmitCertificate(cert)
	if err != nil {
		t.Fatalf("acc.SubmitCertificate() failed: %v", err)
	}

	txHash, ok := resp["txHash"].(string)
	if !ok {
		t.Fatal("txHash not found in response")
	}

	outcome, err := acc.GetTransactionOutcome(txHash, 10)
	if err != nil {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", err)
	}

	if status, _ := outcome["Status"].(string); status != "Confirmed" {
		t.Errorf("Expected transaction status to be 'Confirmed', but got '%s'", status)
	}

}

func TestCertificateOperations(t *testing.T) {
	// TODO: Implement TestCertificateOperations
}

func TestHelloWorldCertification(t *testing.T) {
	// TODO: Implement TestHelloWorldCertification
}
