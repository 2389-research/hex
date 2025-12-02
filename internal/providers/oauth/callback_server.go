// ABOUTME: OAuth callback server for handling OAuth2 redirects in CLI
// ABOUTME: Starts local HTTP server to receive authorization codes

// Package oauth provides OAuth2 authentication utilities for productivity providers
package oauth

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"time"
)

// CallbackServer handles OAuth2 callback redirects
type CallbackServer struct {
	server   *http.Server
	Port     int
	CodeChan chan string
	ErrChan  chan error
	State    string // CSRF protection state
}

// NewCallbackServer creates a new OAuth callback server on a random port
func NewCallbackServer(state string) (*CallbackServer, error) {
	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close() // Close immediately, we'll reopen in Start()

	cs := &CallbackServer{
		Port:     port,
		CodeChan: make(chan string, 1),
		ErrChan:  make(chan error, 1),
		State:    state,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", cs.handleCallback)

	cs.server = &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return cs, nil
}

// Start begins listening for OAuth callbacks
func (cs *CallbackServer) Start() error {
	go func() {
		if err := cs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cs.ErrChan <- err
		}
	}()
	return nil
}

// Stop gracefully shuts down the callback server
func (cs *CallbackServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return cs.server.Shutdown(ctx)
}

// RedirectURI returns the full redirect URI for this server
func (cs *CallbackServer) RedirectURI() string {
	return fmt.Sprintf("http://127.0.0.1:%d/callback", cs.Port)
}

// handleCallback processes the OAuth2 callback
func (cs *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state to prevent CSRF
	state := r.URL.Query().Get("state")
	if state != cs.State {
		cs.writeErrorPage(w, "Invalid state parameter. Possible CSRF attack.")
		cs.ErrChan <- fmt.Errorf("state mismatch: expected %s, got %s", cs.State, state)
		return
	}

	// Check for errors from OAuth provider
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		errDesc := r.URL.Query().Get("error_description")
		cs.writeErrorPage(w, fmt.Sprintf("OAuth error: %s - %s", errMsg, errDesc))
		cs.ErrChan <- fmt.Errorf("oauth error: %s - %s", errMsg, errDesc)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		cs.writeErrorPage(w, "No authorization code received")
		cs.ErrChan <- fmt.Errorf("no authorization code in callback")
		return
	}

	// Send code to channel
	cs.CodeChan <- code

	// Show success page
	cs.writeSuccessPage(w)
}

// writeSuccessPage renders a success page after OAuth callback
func (cs *CallbackServer) writeSuccessPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	tmpl := template.Must(template.New("success").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .container {
            background: white;
            padding: 3rem;
            border-radius: 10px;
            box-shadow: 0 10px 25px rgba(0,0,0,0.2);
            text-align: center;
            max-width: 400px;
        }
        h1 {
            color: #2d3748;
            margin-bottom: 1rem;
        }
        p {
            color: #4a5568;
            line-height: 1.6;
        }
        .success-icon {
            font-size: 4rem;
            color: #48bb78;
            margin-bottom: 1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">✓</div>
        <h1>Authentication Successful!</h1>
        <p>You can now close this window and return to your terminal.</p>
        <p style="font-size: 0.9em; color: #718096; margin-top: 2rem;">
            Pagen is now authorized to access your account.
        </p>
    </div>
</body>
</html>
`))

	_ = tmpl.Execute(w, nil)
}

// writeErrorPage renders an error page
func (cs *CallbackServer) writeErrorPage(w http.ResponseWriter, errMsg string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)

	tmpl := template.Must(template.New("error").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Error</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
        }
        .container {
            background: white;
            padding: 3rem;
            border-radius: 10px;
            box-shadow: 0 10px 25px rgba(0,0,0,0.2);
            text-align: center;
            max-width: 400px;
        }
        h1 {
            color: #2d3748;
            margin-bottom: 1rem;
        }
        p {
            color: #4a5568;
            line-height: 1.6;
        }
        .error-icon {
            font-size: 4rem;
            color: #f56565;
            margin-bottom: 1rem;
        }
        .error-message {
            background: #fed7d7;
            color: #c53030;
            padding: 1rem;
            border-radius: 5px;
            margin-top: 1rem;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="error-icon">✗</div>
        <h1>Authentication Failed</h1>
        <p>There was a problem authenticating your account.</p>
        <div class="error-message">{{.}}</div>
        <p style="font-size: 0.9em; color: #718096; margin-top: 2rem;">
            Please try again or contact support if the problem persists.
        </p>
    </div>
</body>
</html>
`))

	_ = tmpl.Execute(w, errMsg)
}
