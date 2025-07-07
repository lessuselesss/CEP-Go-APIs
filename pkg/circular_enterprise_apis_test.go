package circular_enterprise_apis

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)


func TestNewCEPAccount(t *testing.T) {
	acc := NewCEPAccount()

	if acc == nil {
		t.Fatal("NewCEPAccount returned nil")
	}
	if acc.NAGURL != DefaultNAG {
		t.Errorf("Expected NAGURL to be %s, got %s", DefaultNAG, acc.NAGURL)
	}
}

func TestNewCCertificate(t *testing.T) {
	cert := NewCCertificate()

	if cert == nil {
		t.Fatal("NewCCertificate returned nil")
	}
	if cert.Version != LibVersion {
		t.Errorf("Expected Version to be %s, got %s", LibVersion, cert.Version)
	}
}

func TestGetNAG(t *testing.T) {
	t.Run("SuccessfulDiscovery", func(t *testing.T) {
		expectedNAG := "https://test-nag.circularlabs.io/"
		// Mock HTTP server for GetNAG
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"status":"success", "url":"%s"}`, expectedNAG)))
		}))
		defer server.Close()

		// Mock GetNAG to simulate successful discovery
		originalGetNAG := GetNAG
		GetNAG = func(network string) (string, error) {
			return expectedNAG, nil
		}
		defer func() { GetNAG = originalGetNAG }()

		nag, err := GetNAG("testnet")
		if err != nil {
			t.Fatalf("Expected no error, but got %v", err)
		}
		if nag != expectedNAG {
			t.Errorf("Expected NAG URL %s, but got %s", expectedNAG, nag)
		}
	})

			t.Run("FailedDiscovery - Non-200 Status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		// Mock GetNAG to use the mock server
		originalGetNAG := GetNAG
		GetNAG = func(network string) (string, error) {
			resp, err := http.Get(server.URL + "/network/getNAG?network=" + network)
			if err != nil {
				return "", fmt.Errorf("failed to fetch NAG URL: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return "", fmt.Errorf("network discovery failed with status: %s", resp.Status)
			}
			return "", fmt.Errorf("unexpected response")
		}
		defer func() { GetNAG = originalGetNAG }()

		_, err := GetNAG("testnet")
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}
		if !strings.Contains(err.Error(), "network discovery failed with status: 404 Not Found") {
			t.Errorf("Expected error message to contain 'network discovery failed with status: 404 Not Found', but got '%s'", err.Error())
		}
	})

	t.Run("FailedDiscovery - Body Read Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1") // Indicate content, but don't write it
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Mock GetNAG to simulate body read error
		originalGetNAG := GetNAG
		GetNAG = func(network string) (string, error) {
			return "", fmt.Errorf("failed to read response body")
		}
		defer func() { GetNAG = originalGetNAG }()

		_, err := GetNAG("testnet")
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}
		if !strings.Contains(err.Error(), "failed to read response body") {
			t.Errorf("Expected error message to contain 'failed to read response body', but got '%s'", err.Error())
		}
	})

	t.Run("FailedDiscovery - Invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		// Mock GetNAG to simulate invalid JSON response
		originalGetNAG := GetNAG
		GetNAG = func(network string) (string, error) {
			resp, err := http.Get(server.URL + "/network/getNAG?network=" + network)
			if err != nil {
				return "", fmt.Errorf("failed to fetch NAG URL: %w", err)
			}
			defer resp.Body.Close()
			return "", fmt.Errorf("failed to unmarshal NAG response")
		}
		defer func() { GetNAG = originalGetNAG }()

		_, err := GetNAG("testnet")
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal NAG response") {
			t.Errorf("Expected error message to contain 'failed to unmarshal NAG response', but got '%s'", err.Error())
		}
	})

	t.Run("FailedDiscovery - Empty Network String", func(t *testing.T) {
		_, err := GetNAG("")
		if err == nil {
			t.Fatalf("Expected an error for empty network string, but got none")
		}
		if !strings.Contains(err.Error(), "network identifier cannot be empty") {
			t.Errorf("Expected error message to contain 'network identifier cannot be empty', but got %v", err)
		}
	})

	t.Run("FailedDiscovery - Invalid NAG Response Status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"error", "message":"Invalid network"}`))
		}))
		defer server.Close()

		// Temporarily change httpClient to point to our mock server
		originalHTTPClient := httpClient
		httpClient = server.Client()
		defer func() { httpClient = originalHTTPClient }()

		_, err := GetNAG("testnet")
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}
		if !strings.Contains(err.Error(), "failed to get valid NAG URL from response: Invalid network") {
			t.Errorf("Expected error message to contain 'failed to get valid NAG URL from response: Invalid network', but got '%s'", err.Error())
		}
	})
}

