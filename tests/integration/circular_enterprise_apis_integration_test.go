package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	cep "github.com/lessuselesss/CEP-Go-APIs/pkg"
)

func TestMain(m *testing.M) {
	// Load .env file from the current directory (tests/integration)
	// This makes the environment variables available to the tests
	err := godotenv.Load()
	if err != nil {
		// If the .env file is not found, we don't fail the test,
		// as the tests are designed to skip if the env vars are not set.
		// We just log that it wasn't loaded.
		fmt.Println("Error loading .env file, tests requiring env vars will be skipped.")
	}
	// Run the tests
	os.Exit(m.Run())
}

func TestCircularOperations(t *testing.T) {
	privateKeyHex := os.Getenv("CIRCULAR_PRIVATE_KEY")
	address := os.Getenv("CIRCULAR_ADDRESS")
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	// Create a mock server to handle network requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "GetWalletNonce") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Result":200,"Response":{"Nonce":1}}`))
		} else if strings.Contains(r.URL.Path, "Circular_GetTransactionbyID_") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Result":200,"Response":{"Status":"Confirmed"}}`))
		} else if strings.Contains(r.URL.Path, "transaction") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"Result":200,"Response":{"Status":"Confirmed"}}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"txHash":"0x12345","status":"success"}`))
		}
	}))
	defer server.Close()

	acc := cep.NewCEPAccount(server.URL, "testnet", "1.0")

	// Decode the private key and set it on the account

	var err error
	err = acc.Open(address)
	if err != nil {
		t.Fatalf("acc.Open() failed: %v", err)
	}

	ok, err := acc.UpdateAccount()
	if !ok || err != nil {
		t.Fatalf("acc.UpdateAccount() failed: ok=%v, err=%v", ok, err)
	}

	var resp map[string]interface{}
	resp, err = acc.SubmitCertificate("test message", privateKeyHex)
	if err != nil {
		t.Fatalf("acc.SubmitCertificate() failed: %v", err)
	}

	txHash, ok := resp["txHash"].(string)
	if !ok {
		t.Fatal("txHash not found in response")
	}

	var outcome map[string]interface{}
	outcome, err = acc.GetTransactionOutcome(txHash, 10)
	if err != nil {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", err)
	}

	if status, _ := outcome["Status"].(string); status != "Confirmed" {
		t.Errorf("Expected transaction status to be 'Confirmed', but got '%s'", status)
	}

	// TODO: Implement TestCircularOperations
}

func TestCertificateOperations(t *testing.T) {
	// TODO: Implement TestCertificateOperations
}

func TestHelloWorldCertification(t *testing.T) {
	// TODO: Implement TestHelloWorldCertification
}
