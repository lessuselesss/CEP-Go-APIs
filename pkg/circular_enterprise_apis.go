package circular_enterprise_apis

import (
	"github.com/lessuselesss/CEP-Go-APIs/pkg/account"
	"github.com/lessuselesss/CEP-Go-APIs/pkg/certificate"
)

const (
	// These constants are shared across the library.
	LibVersion   = "1.0.13"
	NetworkURL   = "https://circularlabs.io/network/getNAG?network="
	DefaultChain = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"
	DefaultNAG   = "https://nag.circularlabs.io/NAG.php?cep="
)

// You might have a client that uses the sub-packages.
type Client struct {
	// ... client configuration ...
}

// You can expose the sub-package constructors through the client.
func (c *Client) NewAccount() *account.Account {
	return account.NewAccount(LibVersion, DefaultNAG, DefaultChain)
}

func (c *Client) NewCertificate() *certificate.Certificate {
	return certificate.NewCertificate(LibVersion)
}
