//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	cep "github.com/circular-protocol/go-enterprise-apis/pkg"
)

var (
	privateKeyHex string
	address      string
)

func TestMain(m *testing.M) {
	// Load .env.e2e file from the project root
	if err := godotenv.Load("../../.env.e2e"); err != nil {
		fmt.Println("No .env.e2e file found, falling back to .env")
		// If .env.e2e is not found, try to load .env from the project root
		if err := godotenv.Load("../../.env"); err != nil {
			fmt.Println("Error loading .env file, tests requiring env vars will be skipped.")
		}
	}

	privateKeyHex = os.Getenv("E2E_PRIVATE_KEY")
	address = os.Getenv("E2E_ADDRESS")

	// Run the tests
	os.Exit(m.Run())
}

func TestE2ECircularOperations(t *testing.T) {
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping E2E test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	acc := cep.NewCEPAccount()

	err := acc.Open(address)
	if err != nil {
		t.Fatalf("acc.Open() failed: %v", err)
	}

	err = acc.SetNetwork("testnet")
	if err != nil {
		t.Fatalf("acc.SetNetwork() failed: %v", err)
	}

	ok, err := acc.UpdateAccount()
	if !ok || err != nil {
		t.Fatalf("acc.UpdateAccount() failed: ok=%v, err=%v", ok, err)
	}

	resp, err := acc.SubmitCertificate("test message from Go E2E test", privateKeyHex)
	if err != nil {
		t.Fatalf("acc.SubmitCertificate() failed: %v", err)
	}

	txHash, ok := resp["txHash"].(string)
	if !ok {
		t.Fatal("txHash not found in response")
	}

	// Poll for transaction outcome
	var outcome map[string]interface{}
	for i := 0; i < 10; i++ {
		outcome, err = acc.GetTransactionOutcome(txHash, 10, acc.IntervalSec)
		if err == nil && outcome != nil {
			if status, ok := outcome["Status"].(string); ok && status == "Confirmed" {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", err)
	}

	if status, _ := outcome["Status"].(string); status != "Confirmed" {
		t.Errorf("Expected transaction status to be 'Confirmed', but got '%s'", status)
	}
}

func TestE2ECertificateOperations(t *testing.T) {
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping E2E test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	acc := cep.NewCEPAccount()

	err := acc.Open(address)
	if err != nil {
		t.Fatalf("acc.Open() failed: %v", err)
	}

	err = acc.SetNetwork("testnet")
	if err != nil {
		t.Fatalf("acc.SetNetwork() failed: %v", err)
	}

	_, err = acc.UpdateAccount()
	if err != nil {
		t.Fatalf("acc.UpdateAccount() failed: %v", err)
	}

	resp, err := acc.SubmitCertificate("{\"test\":\"data\"}", privateKeyHex)
	if err != nil {
		t.Fatalf("acc.SubmitCertificate() failed: %v", err)
	}

	txHash, ok := resp["txHash"].(string)
	if !ok {
		t.Fatal("txHash not found in response")
	}

	// Poll for transaction outcome
	var outcome map[string]interface{}
	for i := 0; i < 10; i++ {
		outcome, err = acc.GetTransactionOutcome(txHash, 10, acc.IntervalSec)
		if err == nil && outcome != nil {
			if status, ok := outcome["Status"].(string); ok && status == "Confirmed" {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", err)
	}

	if status, _ := outcome["Status"].(string); status != "Confirmed" {
		t.Errorf("Expected transaction status to be 'Confirmed', but got '%s'", status)
	}
}

func TestE2EHelloWorldCertification(t *testing.T) {
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping E2E test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	acc := cep.NewCEPAccount()

	err := acc.Open(address)
	if err != nil {
		t.Fatalf("acc.Open() failed: %v", err)
	}

	err = acc.SetNetwork("testnet")
	if err != nil {
		t.Fatalf("acc.SetNetwork() failed: %v", err)
	}

	_, err = acc.UpdateAccount()
	if err != nil {
		t.Fatalf("acc.UpdateAccount() failed: %v", err)
	}

	resp, err := acc.SubmitCertificate("Hello World", privateKeyHex)
	if err != nil {
		t.Fatalf("acc.SubmitCertificate() failed: %v", err)
	}

	txHash, ok := resp["txHash"].(string)
	if !ok {
		t.Fatal("txHash not found in response")
	}

	// Poll for transaction outcome
	var outcome map[string]interface{}
	for i := 0; i < 10; i++ {
		outcome, err = acc.GetTransactionOutcome(txHash, 10, acc.IntervalSec)
		if err == nil && outcome != nil {
			if status, ok := outcome["Status"].(string); ok && status == "Confirmed" {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", err)
	}

	if status, _ := outcome["Status"].(string); status != "Confirmed" {
		t.Errorf("Expected transaction status to be 'Confirmed', but got '%s'", status)
	}
}
