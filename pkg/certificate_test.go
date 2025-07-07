package circular_enterprise_apis

import (
	"encoding/hex"
	"encoding/json"
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
			cert := &CCertificate{}
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
		cert         CCertificate
		expectedData string
		expectError  bool
	}{
		{
			name: "Simple ASCII string",
			cert: CCertificate{
				Data: hex.EncodeToString([]byte("hello world")),
			},
			expectedData: "hello world",
			expectError:  false,
		},
		{
			name: "Unicode string",
			cert: CCertificate{
				Data: hex.EncodeToString([]byte("你好, 世界")),
			},
			expectedData: "你好, 世界",
			expectError:  false,
		},
		{
			name: "Empty string",
			cert: CCertificate{
				Data: "",
			},
			expectedData: "",
			expectError:  false,
		},
		{
			name: "Invalid hex data",
			cert: CCertificate{
				Data: "this is not hex",
			},
			expectedData: "",
			expectError:  true,
		},
		{
			name: "Odd length hex string",
			cert: CCertificate{
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
					t.Errorf("Did not expect an error, but got: %s", err)
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
		cert         CCertificate
		expectedJSON string
		expectError  bool
	}{
		{
			name: "Full certificate",
			cert: CCertificate{
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
			cert:         CCertificate{},
			expectedJSON: `{"data":"","previousTxID":"","previousBlock":"","version":""}`,
			expectError:  false,
		},
		{
			name: "Certificate with some empty fields",
			cert: CCertificate{
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
				t.Errorf("Did not expect an error, but got: %s", err)
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
		cert         CCertificate
		expectedSize int
		expectError  bool
	}{
		{
			name: "Full certificate",
			cert: CCertificate{
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
			cert:         CCertificate{},
			expectedSize: 61,
			expectError:  false,
		},
		{
			name: "Certificate with unicode data",
			cert: CCertificate{
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
				t.Errorf("Did not expect an error, but got: %s", err)
			}

			if size != tc.expectedSize {
				t.Errorf("Expected size to be %d, but got %d", tc.expectedSize, size)
			}
		})
	}
}


