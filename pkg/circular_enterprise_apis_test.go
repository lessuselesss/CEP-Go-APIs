package circular_enterprise_apis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)


// TestNewAccount verifies that the NewAccount method correctly initializes an Account instance
// with the parent CEP's network configuration.
func TestNewAccount(t *testing.T) {
	acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)

	if acc == nil {
		t.Fatal("NewAccount returned nil")
	}
	if acc.NAGURL != DefaultNAG {
		t.Errorf("Expected NAGURL to be %s, got %s", DefaultNAG, acc.NAGURL)
	}
}

// TestNewCertificate verifies that the NewCertificate method correctly initializes a Certificate instance
// with the parent CEP's network configuration.
func TestNewCertificate(t *testing.T) {
	cert := NewCertificate(LibVersion)

	if cert == nil {
		t.Fatal("NewCertificate returned nil")
	}
	if cert.Version != LibVersion {
		t.Errorf("Expected Version to be %s, got %s", LibVersion, cert.Version)
	}
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


func TestCertificateOperations(t *testing.T) {
	// TODO: Implement TestCertificateOperations
}

func TestHelloWorldCertification(t *testing.T) {
	// TODO: Implement TestHelloWorldCertification
}
