package account

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
}

// nagResponse is the expected JSON response structure from the NAG server.
type nagResponse struct {
	S  int         `json:"s"`  // Status
	R  AccountInfo `json:"r"`  // Result
	NN string      `json:"nn"` // Network Node
}

// NewCEPAccount is a factory function that creates and initializes a new CEPAccount.
func NewCEPAccount() *CEPAccount {
	return &CEPAccount{
		CodeVersion: "1.0.13",
		NAGURL:      "https://nag.circularlabs.io/NAG.php?cep=",
		Blockchain:  "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2",
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
func (a *Account) UpdateAccount() (bool, error) {
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
	url := a.NAG_URL + "Circular_GetWalletNonce_" + a.NETWORK_NODE

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

// /*
//  * Returns the account data in JSON format and updates the Nonce field
//  */
//     async UpdateAccount() {
//         if (!this.address) {
//             throw new Error("Account is not open");
//         }

//         let data = {
//             "Blockchain": HexFix(this.blockchain),
//             "Address": HexFix(this.address),
//             "Version": this.codeVersion
//         };

//         try {
//             const response = await fetch(this.NAG_URL + 'Circular_GetWalletNonce_' + this.NETWORK_NODE, {
//                 method: 'POST',
//                 headers: { "Content-Type": "application/json" },
//                 body: JSON.stringify(data),
//             });

//             if (!response.ok) {
//                 throw new Error(`HTTP error! status: ${response.status}`);
//             }

//             const jsonResponse = await response.json();

//             if (jsonResponse.Result === 200 && jsonResponse.Response && jsonResponse.Response.Nonce !== undefined) {
//                 this.Nonce = jsonResponse.Response.Nonce + 1;
//                 return true;
//             } else {
//                 throw new Error('Invalid response format or missing Nonce field');
//             }
//         } catch (error) {
//             console.error('Error:', error);
//             return false;
//         }
//     }

// /*
//  *    selects the blockchain network
//  *    network : selected network and it can be 'devnet', 'testnet', 'mainnet' or a custom one
//  */
//   async  SetNetwork(network) {

//             try {
//                   // Construct the URL with the network parameter
//                   const nagUrl = NETWORK_URL + encodeURIComponent(network);

//                   // Make the fetch request
//                   const response = await fetch(nagUrl, {
//                       method: 'GET',
//                       headers: {
//                           'Accept': 'application/json'
//                       }
//                   });

//                   // Check if the request was successful
//                   if (!response.ok) {
//                       throw new Error(`HTTP error! status: ${response.status}`);
//                   }

//                   // Parse the JSON response
//                   const data = await response.json();

//                   // Check if the status is success and URL exists
//                   if (data.status === 'success' && data.url) {

//                       this.NAG_URL= data.url;

//                   } else {
//                       throw new Error(data.message || 'Failed to get URL');
//                   }

//               } catch (error) {
//                   console.log('Error fetching network URL:', error);
//                   throw error; // Re-throw the error so the caller can handle it
//               }

//     }

// /*
//  *    selects the blockchain
//  *    chain : blockchain address
//  */

// 	SetBlockchain(chain) {
//         this.blockchain = chain;
//     }

// /*
//  *    closes the account
//  */
//     Close() {
//         this.address = null;
//         this.publicKey = null;
//         this.info = null;
//         this.lastError='';
//         this.NAG_URL = DEFAULT_NAG;
//         this.NETWORK_NODE = '';
//         this.blockchain = DEFAULT_CHAIN;
//         this.LatestTxID = '';
//         this.data = {};
//         this.intervalSec = 2;
//     }

// /*
//  *    signs data
//  *          data : data that you wish to sign
//  *    provateKey : private key associated to the account
//  */
//     SignData(data, privateKey) {

//         if (!this.address) {
//             throw new Error("Account is not open");
//         }

//         const EC = elliptic.ec;
//         const ec = new EC('secp256k1');
//         const key = ec.keyFromPrivate(HexFix(privateKey), 'hex');
//         const msgHash = sha256(data);

//         // The signature is a DER-encoded hex string
//         const signature = key.sign(msgHash).toDER('hex');
//         return signature;
//     }

// /*
//  *   Searches a Transaction by its ID
//  *   The transaction will be searched initially between the pending transactions and then in the blockchain
//  *
//  *   blockNum: block where the transaction was saved
//  *   txID: transaction ID
//  */
// async GetTransaction(blockNum, txID) {
//     try {
//         let data = {
//             "Blockchain" : HexFix(this.blockchain),
//                     "ID" : HexFix(txID),
//                  "Start" : String(blockNum),
//                    "End" : String(blockNum),
//                "Version" : this.codeVersion
//         };

//         const response = await fetch(this.NAG_URL + 'Circular_GetTransactionbyID_' + this.NETWORK_NODE, {
//             method: 'POST',
//             headers: { 'Content-Type': 'application/json' },
//             body: JSON.stringify(data)
//         });

//         if (!response.ok) {
//             throw new Error('Network response was not ok');
//         }

//         return response.json();
//     } catch (error) {
//         console.error('Error:', error);
//         throw error;
//     }
// }

// /*
//  *   Searches a Transaction by its ID
//  *   The transaction will be searched initially between the pending transactions and then in the blockchain
//  *
//  *   TxID: transaction ID
//  *   Start: Starting block
//  *   End: End block
//  *
//  *   if End = 0 Start indicates the number of blocks starting from the last block minted
//  */
// async GetTransactionbyID(TxID, Start, End) {
//     try {
//         let data = {
//             "Blockchain" : HexFix(this.blockchain),
//                     "ID" : HexFix(TxID),
//                  "Start" : String(Start),
//                    "End" : String(End),
//                "Version" : this.codeVersion
//         };

//         const response = await fetch(this.NAG_URL + 'Circular_GetTransactionbyID_' + this.NETWORK_NODE, {
//             method: 'POST',
//             headers: { 'Content-Type': 'application/json' },
//             body: JSON.stringify(data)
//         });

//         if (!response.ok) {
//             throw new Error('Network response was not ok');
//         }

//         return response.json();
//     } catch (error) {
//         console.error('Error:', error);
//         throw error;
//     }
// }

// /*
//  *   Searches a Transaction by its ID
//  *   The transaction will be searched initially between the pending transactions and then in the blockchain
//  *
//  *   TxID: transaction ID
//  *   Start: Starting block
//  *   End: End block
//  *
//  *   if End = 0 Start indicates the number of blocks starting from the last block minted
//  */
// async GetTransaction(BlockID, TxID) {
//     try {
//         let data = {
//             "Blockchain" : HexFix(this.blockchain),
//                     "ID" : HexFix(TxID),
//                  "Start" : String(BlockID),
//                    "End" : String(BlockID),
//                "Version" : this.codeVersion
//         };

//         const response = await fetch(this.NAG_URL + 'Circular_GetTransactionbyID_' + this.NETWORK_NODE, {
//             method: 'POST',
//             headers: { 'Content-Type': 'application/json' },
//             body: JSON.stringify(data)
//         });

//         if (!response.ok) {
//             throw new Error('Network response was not ok');
//         }

//         return response.json();
//     } catch (error) {
//         console.error('Error:', error);
//         throw error;
//     }
// }

// /*
//  *    Submit data to the blockchain
//  *          data : data that you wish to sign
//  *    provateKey : private key associated to the account
//  */
//     async submitCertificate(pdata, privateKey) {
//         if (!this.address) {
//             throw new Error("Account is not open");
//         }

//         const PayloadObject = {
//             "Action": "CP_CERTIFICATE",
//             "Data": stringToHex(pdata)
//         };

//         const jsonstr = JSON.stringify(PayloadObject);
//         const Payload = stringToHex(jsonstr);
//         const Timestamp = getFormattedTimestamp();
//         const str = HexFix(this.blockchain) + HexFix(this.address) + HexFix(this.address) + Payload + this.Nonce + Timestamp;
//         const ID = sha256(str);
//         const Signature = this.signData(ID, privateKey);

//         let data = {
//             "ID": ID,
//             "From": HexFix(this.address),
//             "To": HexFix(this.address),
//             "Timestamp": Timestamp,
//             "Payload": String(Payload),
//             "Nonce": String(this.Nonce),
//             "Signature": Signature,
//             "Blockchain": HexFix(this.blockchain),
//             "Type": 'C_TYPE_CERTIFICATE',
//             "Version": this.codeVersion
//         };

//         try {
//             const response = await fetch(this.NAG_URL + 'Circular_AddTransaction_' + this.NETWORK_NODE, {
//                 method: 'POST',
//                 headers: { 'Content-Type': 'application/json' },
//                 body: JSON.stringify(data)
//             });

//             if (!response.ok) {
//                 throw new Error('Network response was not ok');
//             }

//             return await response.json();
//         } catch (error) {
//             console.error('Error:', error);
//             return { success: false, message: 'Server unreachable or request failed', error: error.toString() };
//         }
//     }

// /*
//  *    Recursive transaction finality polling
//  *    will search a transaction every  intervalSec seconds with a desired timeout.
//  *
//  *    Blockchain: blockchain where the transaction was submitted
//  *    TxID: Transaction ID
//  *    timeoutSec: Waiting timeout
//  *
//  */
// async GetTransactionOutcome(TxID, timeoutSec) {
//     return new Promise((resolve, reject) => {
//         const startTime = Date.now();
//         const interval = this.intervalSec * 1000;  // Convert seconds to milliseconds
//         const timeout = timeoutSec * 1000;    // Convert seconds to milliseconds

//         const checkTransaction = () => {
//             const elapsedTime = Date.now() - startTime;

//             console.log('Checking transaction...', { elapsedTime, timeout });

//             if (elapsedTime > timeout) {
//                 console.log('Timeout exceeded');
//                 reject(new Error('Timeout exceeded'));
//                 return;
//             }

//             this.getTransactionbyID(TxID, 0, 10).then(data => {
//                     console.log('Data received:', data);
//                     if (data.Result === 200 && data.Response !== 'Transaction Not Found' && data.Response.Status!=='Pending') {
//                         resolve(data.Response);  // Resolve if transaction is found and not 'Transaction Not Found'
//                     } else {
//                         console.log('Transaction not yet confirmed or not found, polling again...');
//                         setTimeout(checkTransaction, interval);  // Continue polling
//                     }
//                 })
//                 .catch(error => {
//                     console.log('Error fetching transaction:', error);
//                     reject(error);  // Reject on error
//                 });
//         };

//         setTimeout(checkTransaction, interval);  // Start polling
//     });
// }
