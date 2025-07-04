package account

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
			acc := NewCEPAccount()
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

			acc := NewCEPAccount()
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
