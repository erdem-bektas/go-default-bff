package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/mux"
	"github.com/hashicorp/vault/api"
	"golang.org/x/oauth2"
)

// SaaSConfig holds configuration for a SaaS organization
type SaaSConfig struct {
	OrgID        string `json:"org_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	IssuerURL    string `json:"issuer_url"`
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	UserinfoURL  string `json:"userinfo_url"`
}

// SaaSManager manages multiple SaaS organizations
type SaaSManager struct {
	configs       map[string]*SaaSConfig
	vault         *api.Client
	verifiers     map[string]*oidc.IDTokenVerifier
	oauth2Configs map[string]*oauth2.Config
}

// NewSaaSManager creates a new SaaS manager
func NewSaaSManager() (*SaaSManager, error) {
	// Vault client setup
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = "http://localhost:8200"

	vault, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, fmt.Errorf("vault client error: %v", err)
	}

	vault.SetToken("dev-root")

	manager := &SaaSManager{
		configs:       make(map[string]*SaaSConfig),
		vault:         vault,
		verifiers:     make(map[string]*oidc.IDTokenVerifier),
		oauth2Configs: make(map[string]*oauth2.Config),
	}

	// Load SaaS configurations from Vault
	if err := manager.loadConfigurations(); err != nil {
		return nil, fmt.Errorf("failed to load configurations: %v", err)
	}

	return manager, nil
}

// loadConfigurations loads SaaS configurations from Vault
func (sm *SaaSManager) loadConfigurations() error {
	saasOrgs := []string{"sp1", "sp2"}

	for _, org := range saasOrgs {
		secret, err := sm.vault.Logical().Read(fmt.Sprintf("secret/data/saas/%s/oauth", org))
		if err != nil {
			log.Printf("Warning: Could not read config for %s: %v", org, err)
			continue
		}

		if secret == nil || secret.Data == nil {
			log.Printf("Warning: No config found for %s", org)
			continue
		}

		data := secret.Data["data"].(map[string]interface{})

		config := &SaaSConfig{
			OrgID:        data["org_id"].(string),
			ClientID:     data["client_id"].(string),
			ClientSecret: data["client_secret"].(string),
			IssuerURL:    data["issuer_url"].(string),
			AuthURL:      data["auth_url"].(string),
			TokenURL:     data["token_url"].(string),
			UserinfoURL:  data["userinfo_url"].(string),
		}

		sm.configs[org] = config

		// Setup OIDC verifier
		ctx := context.Background()
		provider, err := oidc.NewProvider(ctx, config.IssuerURL)
		if err != nil {
			log.Printf("Warning: Could not create OIDC provider for %s: %v", org, err)
			continue
		}

		sm.verifiers[org] = provider.Verifier(&oidc.Config{
			ClientID: config.ClientID,
		})

		// Setup OAuth2 config
		sm.oauth2Configs[org] = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  fmt.Sprintf("http://%s.localhost/auth/callback", org),
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthURL,
				TokenURL: config.TokenURL,
			},
			Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
		}

		log.Printf("Loaded configuration for SaaS: %s (OrgID: %s)", org, config.OrgID)
	}

	return nil
}

// GetConfig returns configuration for a SaaS organization
func (sm *SaaSManager) GetConfig(saasID string) (*SaaSConfig, bool) {
	config, exists := sm.configs[saasID]
	return config, exists
}

// AuthHandler handles authentication for a specific SaaS
func (sm *SaaSManager) AuthHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	saasID := vars["saas"]

	oauth2Config, exists := sm.oauth2Configs[saasID]
	if !exists {
		http.Error(w, fmt.Sprintf("SaaS %s not configured", saasID), http.StatusNotFound)
		return
	}

	// Generate state for CSRF protection (in production, use proper state management)
	state := "random-state-string"

	authURL := oauth2Config.AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// CallbackHandler handles OAuth callback for a specific SaaS
func (sm *SaaSManager) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	saasID := vars["saas"]

	oauth2Config, exists := sm.oauth2Configs[saasID]
	if !exists {
		http.Error(w, fmt.Sprintf("SaaS %s not configured", saasID), http.StatusNotFound)
		return
	}

	verifier, exists := sm.verifiers[saasID]
	if !exists {
		http.Error(w, fmt.Sprintf("Verifier for SaaS %s not configured", saasID), http.StatusNotFound)
		return
	}

	// Verify state (in production, implement proper state verification)
	state := r.URL.Query().Get("state")
	if state != "random-state-string" {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	ctx := context.Background()

	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Token exchange failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Verify ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token in response", http.StatusInternalServerError)
		return
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("ID token verification failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Extract user info
	var claims struct {
		Sub   string `json:"sub"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse claims: %v", err), http.StatusInternalServerError)
		return
	}

	// Return user info as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"saas":         saasID,
		"user_id":      claims.Sub,
		"name":         claims.Name,
		"email":        claims.Email,
		"access_token": token.AccessToken,
		"org_id":       sm.configs[saasID].OrgID,
	})
}

// StatusHandler shows status of all SaaS configurations
func (sm *SaaSManager) StatusHandler(w http.ResponseWriter, r *http.Request) {
	status := make(map[string]interface{})

	for saasID, config := range sm.configs {
		status[saasID] = map[string]interface{}{
			"org_id":     config.OrgID,
			"client_id":  config.ClientID,
			"issuer_url": config.IssuerURL,
			"configured": true,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func main() {
	manager, err := NewSaaSManager()
	if err != nil {
		log.Fatalf("Failed to create SaaS manager: %v", err)
	}

	r := mux.NewRouter()

	// SaaS specific routes
	r.HandleFunc("/auth/{saas}", manager.AuthHandler).Methods("GET")
	r.HandleFunc("/auth/{saas}/callback", manager.CallbackHandler).Methods("GET")

	// General routes
	r.HandleFunc("/status", manager.StatusHandler).Methods("GET")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Multi-SaaS Authentication Demo</title>
</head>
<body>
    <h1>Multi-SaaS Authentication Demo</h1>
    <h2>Available SaaS Organizations:</h2>
    <ul>
        <li><a href="/auth/sp1">Login to SP1 (SaaS Project 1)</a></li>
        <li><a href="/auth/sp2">Login to SP2 (SaaS Project 2)</a></li>
    </ul>
    <h2>Status:</h2>
    <p><a href="/status">View Configuration Status</a></p>
</body>
</html>
		`)
	}).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Visit http://localhost:%s to test multi-SaaS authentication", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
