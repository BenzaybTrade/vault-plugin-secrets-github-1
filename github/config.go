package github

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
)

const githubPublicAPI = "https://api.github.com"

const (
	errUnableToParsePrvKey  = Error("unable to parse private key")
	errUnableToParseBaseURL = Error("unable to parse base URL")
	errFieldDataNil         = Error("field data passed for updating was nil")
	errKeyNotPEMFormat      = Error("key is not a PEM formatted RSA private key")
)

// Config holds all configuration for the backend.
type Config struct {
	// AppID is the application identifier of the GitHub App.
	AppID int `json:"app_id"`

	// PrvKey is the private for signing GitHub access token requests (JWTs).
	// NOTE: Should be in a PEM PKCS#1 RSAPrivateKey format.
	PrvKey string `json:"prv_key"`

	// BaseURL is the base URL for API requests.
	// Defaults to GitHub's public API.
	BaseURL string `json:"base_url"`
}

// NewConfig returns a pre-configured Config struct.
func NewConfig() *Config {
	return &Config{BaseURL: githubPublicAPI}
}

// Update updates the configuration from the given field data only when the data
// is different.
func (c *Config) Update(d *framework.FieldData) (bool, error) {
	if d == nil {
		// NOTE: Use of the path framework ensures `d` is never nil.
		return false, errFieldDataNil
	}

	// Track changes to the configuration.
	var changed bool

	if baseURL, ok := d.GetOk(keyBaseURL); ok {
		nv, err := url.ParseRequestURI(baseURL.(string))
		if err != nil {
			return false, fmt.Errorf("%s: %w", errUnableToParseBaseURL, err)
		}

		if c.BaseURL != nv.String() {
			c.BaseURL = nv.String()
			changed = true
		}
	}

	if prvKey, ok := d.GetOk(keyPrvKey); ok {
		if nv := strings.TrimSpace(prvKey.(string)); c.PrvKey != nv {
			if err := validatePrvKeyStr(nv); err != nil {
				return false, err
			}

			c.PrvKey = nv
			changed = true
		}
	}

	if appID, ok := d.GetOk(keyAppID); ok {
		if nv := appID.(int); c.AppID != nv {
			c.AppID = nv
			changed = true
		}
	}

	return changed, nil
}

func validatePrvKeyStr(k string) error {
	pemKey, _ := pem.Decode([]byte(k))
	if pemKey == nil || pemKey.Type != "RSA PRIVATE KEY" {
		return errKeyNotPEMFormat
	}

	if _, err := x509.ParsePKCS1PrivateKey(pemKey.Bytes); err != nil {
		return fmt.Errorf("%s: %w", errUnableToParsePrvKey, err)
	}

	return nil
}
