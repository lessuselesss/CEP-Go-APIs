package certificate

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Certificate represents a CIRCULAR certificate.
type Certificate struct {
	Data          string `json:"data"`
	PreviousTxID  string `json:"previousTxID"`
	PreviousBlock string `json:"previousBlock"`
	Version       string `json:"version"`
}

// NewCertificate creates and initializes a new Certificate instance.
func NewCertificate(version string) *Certificate {
	return &Certificate{
		Version: version,
	}
}

// SetData inserts application data into the certificate after converting it to a hexadecimal string.
// The `data` parameter is the string data to be stored.
func (c *Certificate) SetData(data string) {
	c.Data = hex.EncodeToString([]byte(data))
}

// GetData decodes the hexadecimal data from the certificate into a string.
// It returns the decoded string and an error if the data is not a valid hexadecimal format.
func (c *Certificate) GetData() (string, error) {
	decodedData, err := hex.DecodeString(c.Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode certificate data: %w", err)
	}
	return string(decodedData), nil
}

// GetJSONCertificate serializes the certificate into a JSON string.
// It returns the JSON string and an error if the serialization fails.
func (c *Certificate) GetJSONCertificate() (string, error) {
	// The json.Marshal function converts the struct into a JSON byte slice.
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		// Using fmt.Errorf to wrap the original error with more context.
		return "", fmt.Errorf("failed to marshal certificate to JSON: %w", err)
	}
	// Convert the byte slice to a string for the return value.
	return string(jsonBytes), nil
}

// GetCertificateSize calculates the size of the JSON-serialized certificate in bytes.
// It returns the size and an error if the serialization fails.
func (c *Certificate) GetCertificateSize() (int, error) {
	// The json.Marshal function converts the struct into a JSON byte slice.
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		// Using fmt.Errorf to wrap the original error with more context.
		return 0, fmt.Errorf("failed to marshal certificate to JSON: %w", err)
	}
	// The length of the byte slice is the size of the certificate in bytes.
	return len(jsonBytes), nil
}
