package main

import (
	"fmt"
	"log"
	"os"

	cep "github.com/circular-protocol/go-enterprise-apis/pkg"
	"github.com/joho/godotenv"
)

func main() {
	// This main function serves as an example of how to use the library.
	// The core logic is within the `pkg` directory.
	fmt.Println("Circular Enterprise APIs - Go Example")

	// Load environment variables from a .env file for easy configuration.
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, proceeding without it.")
	}

	// Retrieve private key and address from environment variables.
	privateKey := os.Getenv("CIRCULAR_PRIVATE_KEY")
	address := os.Getenv("CIRCULAR_ADDRESS")

	if privateKey == "" || address == "" {
		fmt.Println("Please set CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables.")
		return
	}

	// --- Example Workflow ---

	// 1. Initialize a new CEPAccount.
	account := cep.NewCEPAccount()

	// 2. Open the account with your address.
	if err := account.Open(address); err != nil {
		log.Fatalf("Failed to open account: %v", err)
	}
	fmt.Printf("Account opened for address: %s\n", account.Address)

	// 3. Set the desired network (e.g., "testnet").
	networkURL, err := account.SetNetwork("testnet")
	if err != nil {
		log.Fatalf("Failed to set network: %v", err)
	}
	fmt.Printf("Network set. NAG URL: %s\n", networkURL)

	// 4. Update account to get the latest nonce.
	if err := account.UpdateAccount(); err != nil {
		log.Fatalf("Failed to update account: %v", err)
	}
	fmt.Printf("Account updated. Current Nonce: %d\n", account.Nonce)

	// 5. Submit a data certificate to the blockchain.
	pdata := "Hello from the Go Enterprise API!"
	if err := account.SubmitCertificate(pdata, privateKey); err != nil {
		log.Fatalf("Failed to submit certificate: %v", err)
	}
	txID := account.LatestTxID
	if txID == "" {
		log.Fatalf("Could not extract TxID from submission response.")
	}
	fmt.Printf("Transaction submitted with TxID: %s\n", txID)

	// 6. Poll for the transaction outcome.
	fmt.Println("Polling for transaction outcome...")
	outcome, err := account.GetTransactionOutcome(txID, 60, 2) // Added intervalSec
	if err != nil {
		log.Fatalf("Failed to get transaction outcome: %v", err)
	}
	fmt.Printf("Transaction outcome received: %v\n", outcome)

	// 7. Fetch the final transaction details.
	if blockID, ok := outcome["BlockID"].(string); ok {
		fmt.Printf("Fetching full transaction from BlockID: %s\n", blockID)
		finalDetails, err := account.GetTransaction(blockID, txID)
		if err != nil {
			log.Fatalf("Failed to get final transaction details: %v", err)
		}
		fmt.Printf("Final transaction details: %v\n", finalDetails)
	} else {
		log.Println("Could not determine BlockID from outcome to fetch final details.")
	}

	// 8. Close the account to clear sensitive data.
	account.Close()
	fmt.Println("Account closed.")
}
