package circular_enterprise_apis

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

	"github.com/lessuselesss/CEP-Go-APIs/internal/utils"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	decdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
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
		Blockchain: a.Blockchain,
		Address:    a.Address,
		Version:    a.CodeVersion,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Construct the full URL for the API endpoint
	url := fmt.Sprintf("%s/Circular_GetWalletNonce_%s", a.NAGURL, a.NetworkNode)

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
	a.PublicKey = ""
	a.Address = ""
}

// SignData creates a cryptographic signature for the given data using the
// provided private key. It operates by first hashing the input data with
// SHA-256 and then signing the resulting hash using ECDSA with the secp256k1 curve.
//
// The dataToSign parameter is the raw data to be signed.
// The privateKeyHex parameter is the hex-encoded private key string.
//
// It returns the signature as a hex-encoded string in ASN.1 DER format.
// An error is returned if the private key is invalid or if the
// signing process fails.
func (a *CEPAccount) SignData(dataToSign []byte, privateKeyHex string) (string, error) {
	// Decode the hex-encoded private key string into a byte slice.
	privateKeyBytes, err := hex.DecodeString(utils.HexFix(privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("invalid private key hex string: %w", err)
	}

	// Parse the private key bytes into a secp256k1.PrivateKey object.
	privateKey := secp256k1.PrivKeyFromBytes(privateKeyBytes)
	if privateKey == nil {
		return "", fmt.Errorf("failed to parse private key from bytes")
	}

	// Hash the input data using SHA-256. The signing algorithm operates on a
	// fixed-size hash of the data, not the raw data itself.
	hasher := sha256.New()
	hasher.Write(dataToSign)
	hashedData := hasher.Sum(nil)

	// Sign the hashed data with the private key using the secp256k1 library.
	// The Sign function from decred/dcrd/dcrec/secp256k1/v4/ecdsa is deterministic by default.
	signature := decdsa.Sign(privateKey, hashedData)

	// The signature is returned as a raw byte slice. We need to serialize it to DER format
	// for compatibility, as the original function returned ASN.1 DER.
	return hex.EncodeToString(signature.Serialize()), nil
}



// GetTransactionByID retrieves the details of a specific transaction from the blockchain
// using its unique transaction ID, and optionally a start and end block.
//
// The transactionID parameter is the unique string identifying the transaction.
// The startBlock and endBlock parameters are optional and can be empty strings.
//
// On success, it returns a map[string]interface{} containing the transaction
// details. An error is returned if the NAG_URL is not set, the network request
// fails, or the response body cannot be properly parsed.
func (a *CEPAccount) GetTransactionByID(transactionID, startBlock, endBlock string) (map[string]interface{}, error) {
	// A Network Access Gateway URL must be configured to identify the target network.
	if a.NAGURL == "" {
		return nil, fmt.Errorf("network is not set. Please call SetNetwork() first")
	}

	// Prepare the request payload
	requestData := struct {
		TxID  string `json:"TxID"`
		Start string `json:"Start"`
		End   string `json:"End"`
	}{
		TxID:  transactionID,
		Start: startBlock,
		End:   endBlock,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Construct the full URL for the API endpoint
	requestURL := fmt.Sprintf("%s/Circular_GetTransactionbyID_%s", a.NAGURL, a.NetworkNode)

	// Make the HTTP POST request
	resp, err := http.Post(requestURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("http post request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("network request failed with status: %s", resp.Status)
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
// The 'pdata' parameter is the raw data to be included in the certificate.
// The 'privateKey' parameter is the private key used for signing.
//
// On success, it returns a map[string]interface{} containing the response from
// the network, which typically includes a transaction hash. An error is returned
// if the NAG_URL is not set, if the certificate cannot be serialized, or if the
// network request fails.
func (a *CEPAccount) SubmitCertificate(pdata string, privateKey string) (map[string]interface{}, error) {
	// A Network Access Gateway URL must be configured to identify the target network.
	if a.NAGURL == "" {
		return nil, fmt.Errorf("network is not set. Please call SetNetwork() first")
	}

	// Create the PayloadObject
	payloadObject := map[string]interface{}{
		"data": pdata,
	}

	// Marshal PayloadObject to JSON string
	payloadObjectBytes, err := json.Marshal(payloadObject)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload object: %w", err)
	}
	payload := hex.EncodeToString(payloadObjectBytes)

	// Generate Timestamp
	timestamp := utils.GetFormattedTimestamp()

	// Construct the string for hashing
	str := fmt.Sprintf("%s%s%s%s", a.Address, a.Blockchain, payload, timestamp)

	// Generate ID using SHA-256
	hasher := sha256.New()
	hasher.Write([]byte(str))
	id := hex.EncodeToString(hasher.Sum(nil))

	// Call SignData to get the Signature
	signature, err := a.SignData([]byte(str), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Construct the final data payload for the HTTP request
	requestData := map[string]interface{}{
		"ID":         id,
		"Address":    a.Address,
		"Blockchain": a.Blockchain,
		"Payload":    payload,
		"Timestamp":  timestamp,
		"Signature":  signature,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Create a new HTTP POST request. The body of the request is the JSON payload.
	req, err := http.NewRequest("POST", a.NAGURL, bytes.NewBuffer(jsonData))
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
	if a.NAGURL == "" {
		return nil, fmt.Errorf("network is not set. Please call SetNetwork() first")
	}
	startTime := time.Now()
	timeout := time.Duration(timeoutSec) * time.Second

	for {
		elapsedTime := time.Since(startTime)
		if elapsedTime > timeout {
			return nil, fmt.Errorf("timeout exceeded")
		}

		data, err := a.GetTransactionByID(TxID, "", "")
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
