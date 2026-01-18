// Package auth provides OAuth 2.0 authentication for Google Classroom API.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Configuration holds OAuth configuration settings.
type Configuration struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
}

// TokenInfo represents stored OAuth token information.
type TokenInfo struct {
	Email        string    `json:"email"`
	AccessToken  string    `json:"access_token"`
	Expiry       time.Time `json:"expiry"`
	NeedsRefresh bool      `json:"needs_refresh"`
}

// Authenticator handles OAuth 2.0 authentication flow.
type Authenticator struct {
	config     *oauth2.Config
	configPath string
	tokenPath  string
}

// NewAuthenticator creates a new Authenticator instance.
func NewAuthenticator(configPath string) (*Authenticator, error) {
	// Load configuration
	cfg, err := loadConfiguration(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Set up OAuth2 config
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/classroom.courses.readonly",
			"https://www.googleapis.com/auth/classroom.coursework.students",
			"https://www.googleapis.com/auth/classroom.rosters.readonly",
			"https://www.googleapis.com/auth/classroom.announcements.readonly",
			"https://www.googleapis.com/auth/classroom.profile.emails",
			"https://www.googleapis.com/auth/classroom.profile.photos",
		},
		Endpoint: google.Endpoint,
	}

	// Determine token storage path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	tokenPath := filepath.Join(homeDir, ".config", "google-classroom", "tokens.json")

	return &Authenticator{
		config:     oauthConfig,
		configPath: configPath,
		tokenPath:  tokenPath,
	}, nil
}

// loadConfiguration reads OAuth configuration from file.
func loadConfiguration(path string) (*Configuration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Return default config if file doesn't exist
		return &Configuration{
			ClientID:     "",
			ClientSecret: "",
			RedirectURI:  "http://localhost:8080/callback",
		}, nil
	}

	var cfg Configuration
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	return &cfg, nil
}

// TokenSource returns an OAuth2 token source for the stored token.
func (a *Authenticator) TokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	token, err := a.loadToken()
	if err != nil {
		return nil, err
	}
	return a.config.TokenSource(ctx, token), nil
}

// LoadToken loads the OAuth token from storage.
func (a *Authenticator) loadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(a.tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no stored token found")
		}
		return nil, fmt.Errorf("failed to read token: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return &token, nil
}

// SaveToken saves the OAuth token to storage with secure permissions.
func (a *Authenticator) SaveToken(token *oauth2.Token) error {
	// Ensure directory exists
	dir := filepath.Dir(a.tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	// Marshal token to JSON
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Write with secure permissions (owner read/write only)
	if err := os.WriteFile(a.tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token: %w", err)
	}

	return nil
}

// DeleteToken removes the stored OAuth token.
func (a *Authenticator) DeleteToken() error {
	if err := os.Remove(a.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

// IsAuthenticated checks if a valid token exists.
func (a *Authenticator) IsAuthenticated() bool {
	token, err := a.loadToken()
	if err != nil {
		return false
	}
	return token.Valid() || token.RefreshToken != ""
}

// GetAuthURL returns the OAuth consent URL.
func (a *Authenticator) GetAuthURL(state string) string {
	return a.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCode exchanges an authorization code for a token.
func (a *Authenticator) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := a.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

// RefreshToken refreshes the access token using the refresh token.
func (a *Authenticator) RefreshToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := a.loadToken()
	if err != nil {
		return nil, err
	}

	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	newToken, err := a.config.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Save the new token
	if err := a.SaveToken(newToken); err != nil {
		return nil, err
	}

	return newToken, nil
}

// OpenBrowser opens the default system browser to a URL.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// Login performs the full OAuth login flow.
func (a *Authenticator) Login(ctx context.Context) error {
	// Generate state for CSRF protection
	state := fmt.Sprintf("state_%d", time.Now().UnixNano())

	// Get auth URL
	authURL := a.GetAuthURL(state)

	// Open browser for consent
	fmt.Println("Opening browser for Google OAuth consent...")
	if err := OpenBrowser(authURL); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	// Start local server to receive callback
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server := &http.Server{Addr: ":8080"}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Verify state
		if r.URL.Query().Get("state") != state {
			errChan <- fmt.Errorf("state mismatch")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		// Get code
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no code in callback")
			http.Error(w, "No code", http.StatusBadRequest)
			return
		}

		codeChan <- code
		fmt.Fprintf(w, "<html><body><h1>Authentication successful!</h1><p>You can close this window.</p></body></html>")
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for code or error
	select {
	case code := <-codeChan:
		// Exchange code for token
		token, err := a.ExchangeCode(ctx, code)
		if err != nil {
			return fmt.Errorf("failed to exchange code: %w", err)
		}

		// Save token
		if err := a.SaveToken(token); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		// Shutdown server
		server.Shutdown(ctx)
		return nil

	case err := <-errChan:
		server.Shutdown(ctx)
		return err

	case <-time.After(5 * time.Minute):
		server.Shutdown(ctx)
		return fmt.Errorf("authentication timeout")
	}
}

// Status returns the current authentication status.
func (a *Authenticator) Status() (*TokenInfo, error) {
	token, err := a.loadToken()
	if err != nil {
		return &TokenInfo{
			NeedsRefresh: false,
		}, nil
	}

	info := &TokenInfo{
		AccessToken:  token.AccessToken,
		Expiry:       token.Expiry,
		NeedsRefresh: !token.Valid(),
	}

	return info, nil
}
