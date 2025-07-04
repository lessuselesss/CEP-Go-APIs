package account

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	decdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/lessuselesss/CEP-Go-APIs/internal/utils"
	. "github.com/lessuselesss/CEP-Go-APIs/pkg/certificate"
)

// CEPAccount holds the data for a Circular Enterprise Protocol account.
type CEPAccount struct {
	Address     string
	PublicKey   string
	Info        interface{}
	CodeVersion string
	LastError   string
	NAGURL      string
	NetworkNode string
	Blockchain  string
	LatestTxID  string
	Nonce       int
	Data        map[string]interface{}
	IntervalSec int
	NetworkURL  string
	PrivateKey  *secp256k1.PrivateKey
}

// NewCEPAccount is a factory function that creates and initializes a new CEPAccount.
func NewCEPAccount(nagURL, chain, version string) *CEPAccount {
	return &CEPAccount{
		CodeVersion: version,
		NAGURL:      nagURL,
		Blockchain:  chain,
		Nonce:       0,
		Data:        make(map[string]interface{}),
		IntervalSec: 2,
	}
}

// Open sets the account address. This is a prerequisite for many other
// account operations. It takes the account address as a string and
// returns an error if the address is invalid.
func (a *CEPAccount) Open(address string) error {
	if address == "" {
		return errors.New("Invalid address format")
	}
	a.Address = address
	return nil
}

// UpdateAccount fetches the latest account information from the blockchain
// via the NAG (Network Access Gateway). It updates the account's public key,
// nonce, and other network-related details.
func (a *CEPAccount) UpdateAccount() (bool, error) {
	if a.Address == "" {
		return false, errors.New("Account is not open")
	}

	// Prepare the request payload
	requestData := struct {
		Blockchain string `json:"Blockchain"`
		Address    string `json:"Address"`
		Version    string `json:"Version"`
	}{
		Blockchain: utils.HexFix(a.Blockchain),
		Address:    utils.HexFix(a.Address),
		Version:    a.CodeVersion,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Construct the full URL for the API endpoint
	url := a.NAGURL + "Circular_GetWalletNonce_" + a.NetworkNode

	// Make the HTTP POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("http post request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("network request failed with status: %s", resp.Status)
	}

	// Decode the JSON response
	var responseData struct {
		Result   int `json:"Result"`
		Response struct {
			Nonce int `json:"Nonce"`
		} `json:"Response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return false, fmt.Errorf("failed to decode response body: %w", err)
	}

	// Check for a successful result and update the nonce
	if responseData.Result == 200 {
		a.Nonce = responseData.Response.Nonce + 1
		return true, nil
	}

	return false, errors.New("failed to update account, invalid response from server")
}

// SetNetwork configures the account to use a specific blockchain network.
// It fetches the correct Network Access Gateway (NAG) URL for the given
// network identifier (e.g., "devnet", "testnet", "mainnet") and updates the
// NAG_URL field on the CEPAccount struct. A custom network URL can also be used.
func (a *CEPAccount) SetNetwork(network string) error {
	// Construct the full URL by appending the network identifier to the base network URL.
	nagURL, err := url.Parse(a.NetworkURL + network)
	if err != nil {
		return fmt.Errorf("invalid network URL: %w", err)
	}

	// Perform an HTTP GET request to retrieve network configuration details.
	resp, err := http.Get(nagURL.String())
	if err != nil {
		return fmt.Errorf("failed to fetch network URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("network request failed with status: %s", resp.Status)
	}

	// The response body is expected to be a JSON object containing the status
	// and the specific NAG URL for the requested network. We decode it into a
	// temporary struct.
	var result struct {
		Status  string `json:"status"`
		URL     string `json:"url"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode network response: %w", err)
	}

	// If the request was successful, update the account's NAG_URL.
	// Otherwise, return an error with the message from the provider.
	if result.Status == "success" && result.URL != "" {
		a.NAGURL = result.URL
	} else {
		// The 'message' field in the JSON response provides context for the failure.
		return fmt.Errorf("failed to set network: %s", result.Message)
	}

	return nil
}

// Close securely clears all sensitive credential data from the CEPAccount instance.
// It zeroes out the private key, public key, address, and permissions fields.
// It is a best practice to call this method when the account object is no longer
// needed to prevent sensitive data from lingering in the application's memory.
func (a *CEPAccount) Close() {
	// Setting the fields to their zero value effectively clears them.
	a.PrivateKey = nil
	a.PublicKey = ""
	a.Address = ""
}

// SignData creates a cryptographic signature for the given data using the
// account's private key. It operates by first hashing the input data with
// SHA-256 and then signing the resulting hash using ECDSA with the secp256k1 curve.
//
// The dataToSign parameter is the raw data to be signed.
//
// It returns the signature as a hex-encoded string in ASN.1 DER format.
// An error is returned if the private key is not available or if the
// signing process fails.
func (a *CEPAccount) SignData(dataToSign []byte) (string, error) {
	// A private key must be present in the account to sign data.
	if a.PrivateKey == nil {
		return "", fmt.Errorf("private key is not available for signing")
	}

	// Hash the input data using SHA-256. The signing algorithm operates on a
	// fixed-size hash of the data, not the raw data itself.
	hasher := sha256.New()
	hasher.Write(dataToSign)
	hashedData := hasher.Sum(nil)

	// Sign the hashed data with the private key using the secp256k1 library.
	// The Sign function from decred/dcrd/dcrec/secp256k1/v4/ecdsa is deterministic by default.
	signature := decdsa.Sign(a.PrivateKey, hashedData)

	// The signature is returned as a raw byte slice. We need to serialize it to DER format
	// for compatibility, as the original function returned ASN.1 DER.
	return hex.EncodeToString(signature.Serialize()), nil
}

// GetTransaction retrieves the details of a specific transaction from the blockchain
// using its unique transaction hash.
//
// The transactionHash parameter is the hex-encoded string identifying the transaction.
//
// It returns a map[string]interface{} containing the transaction details on success.
// An error is returned if the NAG_URL is not set, the request fails, or the
// response cannot be parsed.
func (a *CEPAccount) GetTransaction(transactionHash string) (map[string]interface{}, error) {
	// The Network Access Gateway URL must be set to know which network to query.
	if a.NAGURL == "" {
		return nil, fmt.Errorf("network is not set. Please call SetNetwork() first")
	}

	// Construct the full API endpoint URL for fetching a transaction.
	requestURL := fmt.Sprintf("%s/transaction/%s", a.NAGURL, transactionHash)

	// Perform the HTTP GET request.
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send get transaction request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("network returned an error: %s", resp.Status)
	}

	// Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the JSON response into a map. Using a map provides flexibility
	// as the transaction structure may vary.
	var transactionDetails map[string]interface{}
	if err := json.Unmarshal(body, &transactionDetails); err != nil {
		return nil, fmt.Errorf("failed to decode transaction details: %w", err)
	}

	return transactionDetails, nil
}

// GetTransactionByID retrieves the details of a specific transaction from the blockchain
// using its unique transaction ID (often the transaction hash).n//
// The transactionID parameter is the unique string identifying the transaction.
//
// On success, it returns a map[string]interface{} containing the transaction
// details. An error is returned if the NAG_URL is not set, the network request
// fails, or the response body cannot be properly parsed.
func (a *CEPAccount) GetTransactionByID(transactionID string) (map[string]interface{}, error) {
	// A Network Access Gateway URL must be configured to identify the target network.
	if a.NAGURL == "" {
		return nil, fmt.Errorf("network is not set. Please call SetNetwork() first")
	}

	// Construct the full API endpoint for fetching the transaction by its ID.
	requestURL := fmt.Sprintf("%s/transaction/%s", a.NAGURL, transactionID)

	// Execute the HTTP GET request to the network.
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send get transaction request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful HTTP response status.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("network returned a non-200 status code: %s", resp.Status)
	}

	// Read the entire body of the HTTP response.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the JSON response into a map. This provides a flexible structure
	// for accessing transaction data, which can have a variable schema.
	var transactionDetails map[string]interface{}
	if err := json.Unmarshal(body, &transactionDetails); err != nil {
		return nil, fmt.Errorf("failed to decode transaction JSON: %w", err)
	}

	return transactionDetails, nil
}

// SubmitCertificate sends a given certificate to the blockchain for processing
// and inclusion. It serializes the certificate object into a JSON payload and
// submits it to the account's configured Network Access Gateway (NAG) URL.
//
// The 'cert' parameter is a pointer to the Certificate object to be submitted.
//
// On success, it returns a map[string]interface{} containing the response from
// the network, which typically includes a transaction hash. An error is returned
// if the NAG_URL is not set, if the certificate cannot be serialized, or if the
// network request fails.
func (a *CEPAccount) SubmitCertificate(cert *Certificate) (map[string]interface{}, error) {
	// A Network Access Gateway URL must be configured to identify the target network.
	if a.NAGURL == "" {
		return nil, fmt.Errorf("network is not set. Please call SetNetwork() first")
	}

	// Marshal the Certificate struct into a JSON byte slice. This is the payload
	// that will be sent to the network.
	jsonPayload, err := json.Marshal(cert)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal certificate to JSON: %w", err)
	}

	// Create a new HTTP POST request. The body of the request is the JSON payload.
	req, err := http.NewRequest("POST", a.NAGURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request using a default client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit certificate: %w", err)
	}
	defer resp.Body.Close()

	// Read the response from the network.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-successful HTTP status codes.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("network returned an error - status: %s, body: %s", resp.Status, string(body))
	}

	// Unmarshal the JSON response into a map for flexible access to the result.
	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return responseMap, nil
}

// GetTransactionOutcome retrieves the final processing result of a transaction
// from the blockchain using its unique ID. This is often used to confirm
// whether a submitted transaction was successfully validated and included in a block.
//
// The 'transactionID' parameter is the unique string identifying the transaction.
//
// It returns a map[string]interface{} containing the outcome details on success.
// An error is returned if the NAG_URL is not configured, the network request fails,
// or the JSON response cannot be parsed.
func (a *CEPAccount) GetTransactionOutcome(TxID string, timeoutSec int) (map[string]interface{}, error) {
	startTime := time.Now()
	timeout := time.Duration(timeoutSec) * time.Second

	for {
		elapsedTime := time.Since(startTime)
		if elapsedTime > timeout {
			return nil, fmt.Errorf("timeout exceeded")
		}

		data, err := a.GetTransactionByID(TxID)
		if err != nil {
			// Continue polling even if there's an error, in case it's a temporary issue
			fmt.Printf("Error fetching transaction: %v, polling again...\n", err)
		} else {
			// Check for a definitive status
			if result, ok := data["Result"].(float64); ok && result == 200 {
				if response, ok := data["Response"].(map[string]interface{}); ok {
					if status, ok := response["Status"].(string); ok && status != "Pending" {
						return response, nil // Resolve if transaction is found and not pending
					}
				}
			}
		}

		fmt.Println("Transaction not yet confirmed or not found, polling again...")
		time.Sleep(time.Duration(a.IntervalSec) * time.Second) // Continue polling
	}
}
