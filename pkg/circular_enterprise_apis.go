package circular_enterprise_apis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var httpClient *http.Client = http.DefaultClient

// Constants define default network parameters and library metadata.
const (
	// LibVersion specifies the current version of the library.
	LibVersion = "1.0.13"

	// DefaultChain is the blockchain identifier for the default public network.
	DefaultChain = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"

	// DefaultNAG is the URL for the default public Network Access Gateway.
	DefaultNAG = "https://nag.circularlabs.io/NAG.php?cep="
)

// NetworkURL is the base endpoint for discovering a Network Access Gateway (NAG).
var NetworkURL = "https://circularlabs.io/network/getNAG?network="

// GetNAG is a standalone utility function for discovering the NAG URL for a
// given network identifier. It makes an HTTP request to the public NetworkURL endpoint.
func GetNAG(network string) (string, error) {
	if network == "" {
		return "", fmt.Errorf("network identifier cannot be empty")
	}

	resp, err := httpClient.Get(NetworkURL + network)
	if err != nil {
		return "", fmt.Errorf("failed to fetch NAG URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("network discovery failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// The response is expected to be a JSON object like {"status":"success", "url":"..."}
	var nagResponse struct {
		Status  string `json:"status"`
		URL     string `json:"url"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &nagResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal NAG response: %w", err)
	}

	if nagResponse.Status != "success" || nagResponse.URL == "" {
		return "", fmt.Errorf("failed to get valid NAG URL from response: %s", nagResponse.Message)
	}

	return nagResponse.URL, nil
}
