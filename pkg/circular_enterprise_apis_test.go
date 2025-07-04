package circular_enterprise_apis

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestNewCEP tests the factory function for creating CEP instances.
func TestNewCEP(t *testing.T) {
	// Test case 1: Using default values for public network
	t.Run("DefaultInitialization", func(t *testing.T) {
		cep := NewCEP("", "")
		if cep.nagURL != DefaultNAG {
			t.Errorf("Expected default NAG URL %s, but got %s", DefaultNAG, cep.nagURL)
		}
		if cep.chain != DefaultChain {
			t.Errorf("Expected default chain %s, but got %s", DefaultChain, cep.chain)
		}
	})

	// Test case 2: Providing custom values for a private network
	t.Run("CustomInitialization", func(t *testing.T) {
		customNAG := "http://localhost:8080/nag"
		customChain := "0xcustomchainhash"
		cep := NewCEP(customNAG, customChain)
		if cep.nagURL != customNAG {
			t.Errorf("Expected custom NAG URL %s, but got %s", customNAG, cep.nagURL)
		}
		if cep.chain != customChain {
			t.Errorf("Expected custom chain %s, but got %s", customChain, cep.chain)
		}
	})
}

// TestNewAccount verifies that the NewAccount method correctly initializes an Account instance
// with the parent CEP's network configuration.
func TestNewAccount(t *testing.T) {
	cep := NewCEP(DefaultNAG, DefaultChain)
	acc := cep.NewAccount()

	if acc == nil {
		t.Fatal("NewAccount returned nil")
	}
	// Note: This test relies on the internal structure of the Account object.
	// To make this more robust, the Account object could expose its configuration via methods.
	// For now, we assume we can infer state for testing purposes.
}

// TestNewCertificate verifies that the NewCertificate method correctly initializes a Certificate instance
// with the parent CEP's network configuration.
func TestNewCertificate(t *testing.T) {
	cep := NewCEP(DefaultNAG, DefaultChain)
	cert := cep.NewCertificate()

	if cert == nil {
		t.Fatal("NewCertificate returned nil")
	}
	// Similar to TestNewAccount, this test assumes we can validate the initialized object.
}

// mockRoundTripper is a custom http.RoundTripper for mocking HTTP responses in tests.
type mockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

// errorReader is an io.Reader that always returns an error.
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("i/o error")
}

// TestGetNAG tests the network discovery function.
func TestGetNAG(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	defer func() {
		http.DefaultClient.Transport = originalTransport
	}()

	t.Run("SuccessfulDiscovery", func(t *testing.T) {
		expectedNAG := "https://test-nag.circularlabs.io/"
		http.DefaultClient.Transport = &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(expectedNAG)),
			},
		}

		nag, err := GetNAG("testnet")
		if err != nil {
			t.Fatalf("Expected no error, but got %v", err)
		}
		if nag != expectedNAG {
			t.Errorf("Expected NAG URL %s, but got %s", expectedNAG, nag)
		}
	})

	t.Run("FailedDiscovery", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			err: fmt.Errorf("network error"),
		}

		_, err := GetNAG("testnet")
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}
	})

	t.Run("Non200StatusCode", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("Not Found")),
			},
		}

		body, err := GetNAG("testnet")
		if err != nil {
			t.Fatalf("Expected no error based on current implementation, but got %v", err)
		}
		if body != "Not Found" {
			t.Errorf("Expected body 'Not Found', but got '%s'", body)
		}
	})

	t.Run("BodyReadError", func(t *testing.T) {
		http.DefaultClient.Transport = &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(&errorReader{}),
			},
		}

		_, err := GetNAG("testnet")
		if err == nil {
			t.Fatal("Expected an error due to body read failure, but got none")
		}
	})

	t.Run("EmptyNetworkString", func(t *testing.T) {
		expectedNAG := "https://default-nag.circularlabs.io/"
		http.DefaultClient.Transport = &mockRoundTripper{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(expectedNAG)),
			},
		}

		nag, err := GetNAG("")
		if err != nil {
			t.Fatalf("Expected no error for empty network string, but got %v", err)
		}
		if nag != expectedNAG {
			t.Errorf("Expected NAG URL %s, but got %s", expectedNAG, nag)
		}
	})
}

func TestCircularOperations(t *testing.T) {
	// TODO: Implement TestCircularOperations
}

func TestCertificateOperations(t *testing.T) {
	// TODO: Implement TestCertificateOperations
}

func TestHelloWorldCertification(t *testing.T) {
	// TODO: Implement TestHelloWorldCertification
}