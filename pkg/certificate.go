package circular_enterprise_apis

import (
	"encoding/json"
	"fmt"

	"github.com/circular-protocol/go-enterprise-apis/internal/utils"
)

// CCertificate represents the data structure for a Circular certificate.
type CCertificate struct {
	Data          string `json:"data"`
	PreviousTxID  string `json:"previousTxID"`
	PreviousBlock string `json:"previousBlock"`
	Version       string `json:"version"`
}

// NewCCertificate creates and initializes a new CCertificate instance.
func NewCCertificate() *CCertificate {
	return &CCertificate{
		Data:          "",
		PreviousTxID:  "",
		PreviousBlock: "",
		Version:       LibVersion,
	}
}

// SetData sets the data content of the certificate.
func (c *CCertificate) SetData(data string) {
	c.Data = utils.StringToHex(data)
}

// GetData retrieves the data content of the certificate.
func (c *CCertificate) GetData() (string, error) {
	data, err := utils.HexToString(c.Data)
	if err != nil {
		return "", fmt.Errorf("failed to convert hex to string: %w", err)
	}
	return data, nil
}

// GetJSONCertificate serializes the certificate into a JSON string.
func (c *CCertificate) GetJSONCertificate() (string, error) {
	certificateMap := map[string]interface{}{
		"data":          c.Data,
		"previousTxID":  c.PreviousTxID,
		"previousBlock": c.PreviousBlock,
		"version":       c.Version,
	}
	jsonBytes, err := json.Marshal(certificateMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal certificate to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// GetCertificateSize calculates the size of the JSON-serialized certificate in bytes.
func (c *CCertificate) GetCertificateSize() (int, error) {
	jsonString, err := c.GetJSONCertificate()
	if err != nil {
		return 0, fmt.Errorf("failed to get JSON certificate: %w", err)
	}
	return len(jsonString), nil
}

// SetPreviousTxID sets the previous transaction ID.
func (c *CCertificate) SetPreviousTxID(txID string) {
	c.PreviousTxID = txID
}

// SetPreviousBlock sets the previous block hash.
func (c *CCertificate) SetPreviousBlock(block string) {
	c.PreviousBlock = block
}

// GetPreviousTxID gets the previous transaction ID.
func (c *CCertificate) GetPreviousTxID() string {
	return c.PreviousTxID
}

// GetPreviousBlock gets the previous block hash.
func (c *CCertificate) GetPreviousBlock() string {
	return c.PreviousBlock
}
