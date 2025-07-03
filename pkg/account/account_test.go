package account

import (
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