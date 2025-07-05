package circular_enterprise_apis

import (
	"io"
	"net/http"

	"github.com/Circular-Protocol/CEP-Go-APIs/internal/utils"
)

// Constants define default network parameters and library metadata.
const (
	// LibVersion specifies the current version of the library.
	LibVersion = "1.0.13"

	// NetworkURL is the base endpoint for discovering a Network Access Gateway (NAG).
	NetworkURL = "https://circularlabs.io/network/getNAG?network="

	// DefaultChain is the blockchain identifier for the default public network.
	DefaultChain = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"

	// DefaultNAG is the URL for the default public Network Access Gateway.
	DefaultNAG = "https://nag.circularlabs.io/NAG.php?cep="
)


// GetNAG is a standalone utility function for discovering the NAG URL for a
// given network identifier. It makes an HTTP request to the public NetworkURL endpoint.
func GetNAG(network string) (string, error) {
	resp, err := http.Get(NetworkURL + network)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
