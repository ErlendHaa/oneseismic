package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"
)

// OpenIDConfig is the expected return from the well-known endpoint
type OpenIDConfig struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported"`
	UserinfoEndpoint                           string   `json:"userinfo_endpoint"`
	CheckSessionIframe                         string   `json:"check_session_iframe"`
	EndSessionEndpoint                         string   `json:"end_session_endpoint"`
	JwksURI                                    string   `json:"jwks_uri"`
	RegistrationEndpoint                       string   `json:"registration_endpoint"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	AcrValuesSupported                         []string `json:"acr_values_supported"`
	SubjectTypesSupported                      []string `json:"subject_types_supported"`
	UserinfoSigningAlgValuesSupported          []string `json:"userinfo_signing_alg_values_supported"`
	UserinfoEncryptionAlgValuesSupported       []string `json:"userinfo_encryption_alg_values_supported"`
	UserinfoEncryptionEncValuesSupported       []string `json:"userinfo_encryption_enc_values_supported"`
	IDTokenSigningAlgValuesSupported           []string `json:"id_token_signing_alg_values_supported"`
	IDTokenEncryptionAlgValuesSupported        []string `json:"id_token_encryption_alg_values_supported"`
	IDTokenEncryptionEncValuesSupported        []string `json:"id_token_encryption_enc_values_supported"`
	RequestObjectSigningAlgValuesSupported     []string `json:"request_object_signing_alg_values_supported"`
	DisplayValuesSupported                     []string `json:"display_values_supported"`
	ClaimTypesSupported                        []string `json:"claim_types_supported"`
	ClaimsSupported                            []string `json:"claims_supported"`
	ClaimsParameterSupported                   bool     `json:"claims_parameter_supported"`
	ServiceDocumentation                       string   `json:"service_documentation"`
	UILocalesSupported                         []string `json:"ui_locales_supported"`
}

// JWK JSON Web Key
type JWK struct {
	Kty string   `json:"kty"`
	Use string   `json:"use"`
	Kid string   `json:"kid"`
	X5T string   `json:"x5t"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5C []string `json:"x5c"`
}

// JWKS keyset from openID
type JWKS struct {
	Keys []JWK `json:"keys"`
}

var configClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(url *url.URL, target interface{}) error {
	r, err := configClient.Get(url.String())
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return fmt.Errorf(
			"Json fetch error %s on %s",
			r.Status,
			url)

	}

	return json.NewDecoder(r.Body).Decode(target)
}

// GetKey gets the authservers signing key
func GetOIDCKeySet(authserver *url.URL) (map[string]interface{}, error) {
	if authserver == nil {
		return nil, fmt.Errorf("authserver is not found")
	}
	oidcConf := OpenIDConfig{}
	u, err := url.Parse(authserver.String() + "/.well-known/openid-configuration")
	if err != nil {
		return nil, fmt.Errorf("oidcConf url parse failed: %w", err)
	}
	err = getJSON(u, &oidcConf)
	if err != nil {
		return nil, fmt.Errorf("fetching oidc config failed: %w", err)
	}

	jwksURI := oidcConf.JwksURI
	u, err = url.Parse(jwksURI)
	if err != nil {
		return nil, fmt.Errorf("jwks url parse failed: %w", err)
	}
	return createWebKeySet(u)
}

func fromB64(b64 string) (big.Int, error) {
	b, err := base64.RawURLEncoding.DecodeString(b64)
	bi := big.Int{}
	if err != nil {
		return bi, fmt.Errorf("decoding B64 failed: %w", err)
	}

	bi.SetBytes(b)
	return bi, nil
}

func createWebKeySet(keysetURL *url.URL) (map[string]interface{}, error) {
	jwks := JWKS{}
	err := getJSON(keysetURL, &jwks)
	if err != nil {
		return nil, fmt.Errorf("fetching keyset failed: %w", err)
	}

	if len(jwks.Keys) == 0 {
		return nil, fmt.Errorf(
			"could not create keyset. No keys in key set")
	}
	jwksMap := make(map[string]interface{})

	for _, jwk := range jwks.Keys {

		if jwk.Kty == "RSA" {

			e, err := fromB64(jwk.E)
			if err != nil {
				return nil, fmt.Errorf("big int from  E: %w", err)
			}
			n, err := fromB64(jwk.N)
			if err != nil {
				return nil, fmt.Errorf("big int from  N: %w", err)
			}

			jwksMap[jwk.Kid] = &rsa.PublicKey{N: &n, E: int(e.Int64())}

		}
	}

	return jwksMap, nil

}