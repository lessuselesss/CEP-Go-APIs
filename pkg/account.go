package circular_enterprise_apis

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/circular-protocol/go-enterprise-apis/internal/utils"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// CEPAccount holds the data for a Circular Enterprise Protocol account.
type CEPAccount struct {
	Address     string
	PublicKey   string
	Info        interface{}
	CodeVersion string
	NAGURL      string
	NetworkNode string
	Blockchain  string
	LatestTxID  string
	Nonce       int64
	IntervalSec int
	NetworkURL  string
}

// NewCEPAccount is a factory function that creates and initializes a new CEPAccount.
func NewCEPAccount() *CEPAccount {
	return &CEPAccount{
		CodeVersion: LibVersion,
		NetworkURL:  NetworkURL,
		NAGURL:      DefaultNAG,
		Blockchain:  DefaultChain,
		Nonce:       0,
		IntervalSec: 2, // Default polling interval
	}
}

// Open sets the account address. It returns an error if the address is empty.
func (a *CEPAccount) Open(address string) error {
	if address == "" {
		return fmt.Errorf("invalid address format")
	}
	a.Address = address
	return nil
}

// Close securely clears sensitive data from the CEPAccount instance.
func (a *CEPAccount) Close() {
	a.Address = ""
	a.PublicKey = ""
	a.Info = nil
	a.NAGURL = ""
	a.NetworkNode = ""
	a.Blockchain = ""
	a.LatestTxID = ""
	a.Nonce = 0
	a.IntervalSec = 0
}

// SetNetwork configures the account to use a specific network by fetching its
// Network Access Gateway (NAG) URL.
func (a *CEPAccount) SetNetwork(network string) (string, error) {
	nagURL, err := GetNAG(network)
	if err != nil {
		return "", fmt.Errorf("failed to get NAG URL: %w", err)
	}
	a.NAGURL = nagURL
	a.NetworkNode = network // Store the network identifier
	return nagURL, nil
}

// SetBlockchain allows overriding the default blockchain address.
func (a *CEPAccount) SetBlockchain(chain string) {
	a.Blockchain = chain
}

// UpdateAccount fetches the latest nonce for the account from the NAG.
func (a *CEPAccount) UpdateAccount() error {
	if a.Address == "" {
		return fmt.Errorf("account is not open")
	}

	requestData := map[string]string{
		"Blockchain": utils.HexFix(a.Blockchain),
		"Address":    utils.HexFix(a.Address),
		"Version":    a.CodeVersion,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	url := a.NAGURL + "Circular_GetWalletNonce_"
	if a.NetworkNode != "" {
		url += a.NetworkNode
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("http post request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("network request failed with status: %s", resp.Status)
	}

	var responseData struct {
		Result   int
		Response struct {
			Nonce int
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}

	if responseData.Result == 200 {
		a.Nonce = int64(responseData.Response.Nonce) + 1
		return nil
	}

	return fmt.Errorf("failed to update account: invalid response from server")
}

// signData creates a cryptographic signature for the given data hash.
func (a *CEPAccount) signData(message string, privateKeyHex string) (string, error) {
	if a.Address == "" {
		return "", fmt.Errorf("account is not open")
	}

	privateKeyBytes, err := hex.DecodeString(utils.HexFix(privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("invalid private key hex string: %w", err)
	}

	privateKey := secp256k1.PrivKeyFromBytes(privateKeyBytes)
	hash := sha256.Sum256([]byte(message))
	signature := ecdsa.Sign(privateKey, hash[:])

	return hex.EncodeToString(signature.Serialize()), nil
}

// SubmitCertificate creates, signs, and submits a data certificate to the blockchain.
func (a *CEPAccount) SubmitCertificate(pdata string, privateKeyHex string) error {
	if a.Address == "" {
		return fmt.Errorf("account is not open")
	}

	payloadObject := map[string]string{
		"Action": "CP_CERTIFICATE",
		"Data":   utils.StringToHex(pdata),
	}
	jsonStr, _ := json.Marshal(payloadObject)
	payload := utils.StringToHex(string(jsonStr))
	timestamp := utils.GetFormattedTimestamp()

	strToHash := utils.HexFix(a.Blockchain) + utils.HexFix(a.Address) + utils.HexFix(a.Address) + payload + fmt.Sprintf("%d", a.Nonce) + timestamp
	hash := sha256.Sum256([]byte(strToHash))
	id := hex.EncodeToString(hash[:])

	signature, err := a.signData(id, privateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to sign data: %w", err)
	}

	requestData := map[string]string{
		"ID":         id,
		"From":       utils.HexFix(a.Address),
		"To":         utils.HexFix(a.Address),
		"Timestamp":  timestamp,
		"Payload":    payload,
		"Nonce":      fmt.Sprintf("%d", a.Nonce),
		"Signature":  signature,
		"Blockchain": utils.HexFix(a.Blockchain),
		"Type":       "C_TYPE_CERTIFICATE",
		"Version":    a.CodeVersion,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	url := a.NAGURL + "Circular_AddTransaction_"
	if a.NetworkNode != "" {
		url += a.NetworkNode
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to submit certificate: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("network returned an error - status: %s, body: %s", resp.Status, string(body))
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err != nil {
		return fmt.Errorf("failed to decode response JSON: %w", err)
	}

	if result, ok := responseMap["Result"].(float64); ok && result == 200 {
		// Save our generated transaction ID
		a.LatestTxID = id
		a.Nonce++ // Increment nonce for the next transaction
	} else {
		// Extract the error message from the response if available
		if errMsg, ok := responseMap["Response"].(string); ok {
			return fmt.Errorf("certificate submission failed: %s", errMsg)
		} else {
			return fmt.Errorf("certificate submission failed with non-200 result code")
		}
	}

	return nil
}

// GetTransaction retrieves a transaction by its specific block and transaction ID.
func (a *CEPAccount) GetTransaction(blockID string, transactionID string) (map[string]interface{}, error) {
	if blockID == "" {
		return nil, fmt.Errorf("blockID cannot be empty")
	}
	// This function is a convenience wrapper around getTransactionByID,
	// searching within a single, specific block.
	startBlock, err := strconv.ParseInt(blockID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid blockID: %w", err)
	}
	result, err := a.getTransactionByID(transactionID, startBlock, startBlock)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// getTransactionByID retrieves details of a specific transaction by its ID within an optional block range.
func (a *CEPAccount) getTransactionByID(transactionID string, startBlock, endBlock int64) (map[string]interface{}, error) {
	if a.NAGURL == "" {
		return nil, fmt.Errorf("network is not set")
	}

	requestData := map[string]string{
		"Blockchain": utils.HexFix(a.Blockchain),
		"ID":         utils.HexFix(transactionID),
		"Start":      fmt.Sprintf("%d", startBlock),
		"End":        fmt.Sprintf("%d", endBlock),
		"Version":    a.CodeVersion,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	url := a.NAGURL + "Circular_GetTransactionbyID_"
	if a.NetworkNode != "" {
		url += a.NetworkNode
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("http post request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("network request failed with status: %s", resp.Status)
	}

	var transactionDetails map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&transactionDetails); err != nil {
		return nil, fmt.Errorf("failed to decode transaction JSON: %w", err)
	}

	return transactionDetails, nil
}

// GetTransactionOutcome polls for the final outcome of a transaction.
func (a *CEPAccount) GetTransactionOutcome(txID string, timeoutSec int, intervalSec int) (map[string]interface{}, error) {
	if a.NAGURL == "" {
		return nil, fmt.Errorf("network is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout exceeded while waiting for transaction outcome")
		case <-ticker.C:
			data, err := a.getTransactionByID(txID, 0, 10) // Search recent blocks
			if err != nil {
				// Log non-critical errors and continue polling
				fmt.Printf("Polling error: %s. Retrying.\n", err)
				continue
			}

			if result, ok := data["Result"].(float64); ok && result == 200 {
				if response, ok := data["Response"].(map[string]interface{}); ok {
					if status, ok := response["Status"].(string); ok && status != "Pending" {
						return response, nil // Transaction finalized
					}
				}
			}
		}
	}
}
