package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type OIDCDiscoveryService struct {
	logger *zap.Logger
	client *http.Client
}

type OIDCConfiguration struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	UserinfoEndpoint                 string   `json:"userinfo_endpoint"`
	JWKSUri                          string   `json:"jwks_uri"`
	ScopesSupported                  []string `json:"scopes_supported"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	GrantTypesSupported              []string `json:"grant_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

func NewOIDCDiscoveryService(logger *zap.Logger) *OIDCDiscoveryService {
	return &OIDCDiscoveryService{
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *OIDCDiscoveryService) DiscoverConfiguration(ctx context.Context, issuer string) (*OIDCConfiguration, error) {
	discoveryURL := fmt.Sprintf("%s/.well-known/openid-configuration", issuer)

	s.logger.Info("Discovering OIDC configuration",
		zap.String("discovery_url", discoveryURL),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OIDC configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery failed with status: %d", resp.StatusCode)
	}

	var config OIDCConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode OIDC configuration: %w", err)
	}

	s.logger.Info("OIDC configuration discovered successfully",
		zap.String("issuer", config.Issuer),
		zap.String("auth_endpoint", config.AuthorizationEndpoint),
		zap.String("token_endpoint", config.TokenEndpoint),
		zap.String("jwks_uri", config.JWKSUri),
	)

	return &config, nil
}
