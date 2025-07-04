package certificate

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
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
