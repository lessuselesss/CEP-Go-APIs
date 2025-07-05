package circular_enterprise_apis

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)


// TestNewAccount verifies that the NewAccount method correctly initializes an Account instance
// with the parent CEP's network configuration.
func TestNewAccount(t *testing.T) {
	acc := NewCEPAccount(DefaultNAG, DefaultChain, LibVersion)

	if acc == nil {
		t.Fatal("NewAccount returned nil")
	}
	if acc.NAGURL != DefaultNAG {
		t.Errorf("Expected NAGURL to be %s, got %s", DefaultNAG, acc.NAGURL)
	}
}

// TestNewCertificate verifies that the NewCertificate method correctly initializes a Certificate instance
// with the parent CEP's network configuration.
func TestNewCertificate(t *testing.T) {
	cert := NewCertificate(LibVersion)

	if cert == nil {
		t.Fatal("NewCertificate returned nil")
	}
	if cert.Version != LibVersion {
		t.Errorf("Expected Version to be %s, got %s", LibVersion, cert.Version)
	}
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




func TestCertificateOperations(t *testing.T) {
	// TODO: Implement TestCertificateOperations
}

func TestHelloWorldCertification(t *testing.T) {
	// TODO: Implement TestHelloWorldCertification
}
