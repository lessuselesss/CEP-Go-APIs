package circular_enterprise_apis

import (
	"io"
	"net/http"

	"github.com/lessuselesss/CEP-Go-APIs/pkg/account"
	"github.com/lessuselesss/CEP-Go-APIs/pkg/certificate"
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

// CEP is the central struct for the API, acting as a client instance.
// It holds shared configuration required for interacting with the blockchain network,
// mirroring the class-based design of the reference JS and Java implementations.
type CEP struct {
	nagURL string
	chain  string
}

// NewCEP is a factory function that creates and configures a CEP instance.
// This is the primary entry point for the library. It accepts network parameters
// to allow connection to private or custom enterprise networks. If empty strings
// are provided, it falls back to the default public network constants.
func NewCEP(nagURL, chain string) *CEP {
	c := &CEP{}
	if nagURL == "" {
		c.nagURL = DefaultNAG
	} else {
		c.nagURL = nagURL
	}
	if chain == "" {
		c.chain = DefaultChain
	} else {
		c.chain = chain
	}
	return c
}

// NewAccount creates a new Account instance, injecting the network configuration
// held by the parent CEP instance. This ensures the account object is correctly
// bound to the target network.
func (c *CEP) NewAccount() *account.CEPAccount {
	return account.NewCEPAccount(c.nagURL, c.chain, LibVersion)
}

// NewCertificate creates a new Certificate instance, injecting the network
// configuration held by the parent CEP instance.
func (c *CEP) NewCertificate() *certificate.Certificate {
	return certificate.NewCertificate(LibVersion)
}

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
