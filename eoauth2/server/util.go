package server

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
)

// Parse basic authentication header
type BasicAuth struct {
	Username string
	Password string
}

// Parse bearer authentication header
type BearerAuth struct {
	Code string
}

// CheckClientSecret determines whether the given secret matches a secret held by the client.
// Public clients return true for a secret of ""
func CheckClientSecret(client Client, secret string) bool {
	switch client := client.(type) {
	case ClientSecretMatcher:
		// Prefer the more secure method of giving the secret to the client for comparison
		return client.ClientSecretMatches(secret)
	default:
		// Fallback to the less secure method of extracting the plain text secret from the client for comparison
		return client.GetSecret() == secret
	}
}

type BasicAuthParam struct {
	Authorization string
}

// Return authorization header data
func CheckBasicAuth(param BasicAuthParam) (*BasicAuth, error) {
	if param.Authorization == "" {
		return nil, nil
	}

	s := strings.SplitN(param.Authorization, " ", 2)
	if len(s) != 2 || s[0] != "Basic" {
		return nil, errors.New("Invalid authorization header")
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return nil, err
	}
	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return nil, errors.New("Invalid authorization message")
	}

	// Decode the client_id and client_secret pairs as per
	// https://tools.ietf.org/html/rfc6749#section-2.3.1

	username, err := url.QueryUnescape(pair[0])
	if err != nil {
		return nil, err
	}

	password, err := url.QueryUnescape(pair[1])
	if err != nil {
		return nil, err
	}

	return &BasicAuth{Username: username, Password: password}, nil
}
