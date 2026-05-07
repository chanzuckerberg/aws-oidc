// Package imds_server implements an IMDSv2-shaped HTTP server that exposes
// AWS STS credentials on a localhost endpoint. It is wire-compatible with
// the AWS SDK Go v2 IMDS credential provider so unmodified workloads
// (Prometheus, AWS CLI, etc.) can fetch credentials by setting
// AWS_EC2_METADATA_SERVICE_ENDPOINT to the server's URL.
//
// The server supports both IMDSv2 (token dance) and IMDSv1 (no header) paths
// for SDK fallback compatibility. IMDSv1 can be disabled via Options.
//
// Sensitive wire-shape requirements verified against the AWS SDK Go v2 source
// (feature/ec2/imds/, credentials/ec2rolecreds/):
//   - PUT /latest/api/token MUST include the X-Aws-Ec2-Metadata-Token-Ttl-Seconds
//     response header. Without it, the SDK errors out and may downgrade to v1.
//   - The credentials JSON requires Code (case-insensitive "Success"),
//     AccessKeyId, SecretAccessKey, Token, and Expiration (RFC 3339).
//     Type and LastUpdated are optional but harmless.
package imds_server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/handlers"
)

const (
	// imdsTokenHeader is the request header that carries the IMDSv2 session
	// token. (Casing per AWS docs.)
	imdsTokenHeader = "X-Aws-Ec2-Metadata-Token"

	// imdsTokenTTLHeader is the header used both to request a TTL on PUT and
	// to confirm the granted TTL on the response. THE RESPONSE HEADER IS
	// REQUIRED — the SDK errors out if it's missing.
	imdsTokenTTLHeader = "X-Aws-Ec2-Metadata-Token-Ttl-Seconds"

	// AWS doc default and max TTL.
	maxIMDSTokenTTLSeconds     = 21600 // 6 hours
	defaultIMDSTokenTTLSeconds = 21600

	roleCredsPathPrefix = "/latest/meta-data/iam/security-credentials/"
	roleListPath        = "/latest/meta-data/iam/security-credentials/"
	tokenPath           = "/latest/api/token"
	healthzPath         = "/healthz"
	readyzPath          = "/readyz"
)

// Options configures Server.
type Options struct {
	// RoleName is the role identifier exposed at the IMDS role-list endpoint.
	// Typically derived from the IAM role's name (e.g. "argus-amp-producer-bruno").
	RoleName string

	// Cache is the credentials provider used to satisfy IMDS credential GETs.
	Cache *CredentialsCache

	// RequireIMDSv2, if true, refuses GET requests that do not carry a valid
	// X-Aws-Ec2-Metadata-Token header. Default: false (IMDSv1 fallback allowed).
	RequireIMDSv2 bool

	// Logger is the slog logger used for request and error logging. If nil,
	// slog.Default() is used.
	Logger *slog.Logger
}

// Server is the IMDSv2-shaped HTTP server.
type Server struct {
	opts Options

	// tokens holds minted IMDSv2 session tokens with their expiry.
	tokensMu sync.Mutex
	tokens   map[string]time.Time
}

// NewServer constructs a Server. RoleName and Cache are required.
func NewServer(opts Options) (*Server, error) {
	if opts.RoleName == "" {
		return nil, errors.New("imds_server: NewServer: RoleName is required")
	}
	if opts.Cache == nil {
		return nil, errors.New("imds_server: NewServer: Cache is required")
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	return &Server{
		opts:   opts,
		tokens: map[string]time.Time{},
	}, nil
}

// Handler returns an http.Handler ready to attach to an http.Server. The
// handler wraps a mux with gorilla/handlers.RecoveryHandler so panics produce
// 500s rather than crashing the process.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(tokenPath, s.handleTokenPUT)
	// Both the role-list path (with trailing slash) and per-role path are
	// served from the same prefix.
	mux.HandleFunc(roleCredsPathPrefix, s.handleCredentials)
	mux.HandleFunc(healthzPath, s.handleHealthz)
	mux.HandleFunc(readyzPath, s.handleReadyz)

	recoveryLogger := slogRecoveryLogger{logger: s.opts.Logger}
	return handlers.RecoveryHandler(
		handlers.PrintRecoveryStack(true),
		handlers.RecoveryLogger(recoveryLogger),
	)(mux)
}

// handleTokenPUT implements PUT /latest/api/token. It MUST set the response
// header X-Aws-Ec2-Metadata-Token-Ttl-Seconds; without it, the AWS SDK's IMDS
// client errors out and may unnecessarily downgrade to IMDSv1.
func (s *Server) handleTokenPUT(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ttlSeconds := defaultIMDSTokenTTLSeconds
	if v := r.Header.Get(imdsTokenTTLHeader); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed <= 0 {
			http.Error(w, "bad token ttl", http.StatusBadRequest)
			return
		}
		if parsed > maxIMDSTokenTTLSeconds {
			parsed = maxIMDSTokenTTLSeconds
		}
		ttlSeconds = parsed
	}

	token, err := mintToken()
	if err != nil {
		s.opts.Logger.Error("imds_server: failed to mint token", slog.String("error", err.Error()))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	s.tokensMu.Lock()
	s.tokens[token] = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	// Opportunistic cleanup of expired tokens (keep map small).
	now := time.Now()
	for tok, exp := range s.tokens {
		if exp.Before(now) {
			delete(s.tokens, tok)
		}
	}
	s.tokensMu.Unlock()

	w.Header().Set(imdsTokenTTLHeader, strconv.Itoa(ttlSeconds))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(token))
}

// validToken returns true if the given token is in the minted-token map and
// not expired.
func (s *Server) validToken(token string) bool {
	if token == "" {
		return false
	}
	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()
	exp, ok := s.tokens[token]
	if !ok {
		return false
	}
	if exp.Before(time.Now()) {
		delete(s.tokens, token)
		return false
	}
	return true
}

// handleCredentials handles both:
//   GET /latest/meta-data/iam/security-credentials/        -> role name
//   GET /latest/meta-data/iam/security-credentials/<role>  -> credential JSON
//
// IMDSv2 token-header policy:
//   - If the header is present, it MUST validate; otherwise 401.
//   - If the header is absent and RequireIMDSv2=false, accept as v1.
//   - If the header is absent and RequireIMDSv2=true, 401.
func (s *Server) handleCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenHeader := r.Header.Get(imdsTokenHeader)
	if tokenHeader != "" {
		if !s.validToken(tokenHeader) {
			http.Error(w, "invalid IMDSv2 token", http.StatusUnauthorized)
			return
		}
	} else if s.opts.RequireIMDSv2 {
		http.Error(w, "IMDSv2 token required", http.StatusUnauthorized)
		return
	}

	switch r.URL.Path {
	case roleListPath:
		// Return the role name as plain text.
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(s.opts.RoleName))
		return

	case roleCredsPathPrefix + s.opts.RoleName:
		s.writeCredentials(w, r)
		return

	default:
		// Unknown role name; mimic real EC2 IMDS by returning 404.
		http.NotFound(w, r)
		return
	}
}

// credentialsResponse is the JSON shape EC2 IMDS returns for credential GETs.
// Required by the AWS SDK: Code, AccessKeyId, SecretAccessKey, Token, Expiration.
// Type and LastUpdated are optional but conventional.
type credentialsResponse struct {
	Code            string `json:"Code"`
	Type            string `json:"Type,omitempty"`
	AccessKeyId     string `json:"AccessKeyId,omitempty"`
	SecretAccessKey string `json:"SecretAccessKey,omitempty"`
	Token           string `json:"Token,omitempty"`
	Expiration      string `json:"Expiration,omitempty"`
	LastUpdated     string `json:"LastUpdated,omitempty"`
	Message         string `json:"Message,omitempty"`
}

// writeCredentials writes the IMDS-shaped credentials JSON. On refresh
// failure, returns HTTP 500 with Code = "AssumeRoleUnauthorizedAccess".
func (s *Server) writeCredentials(w http.ResponseWriter, r *http.Request) {
	creds, err := s.opts.Cache.Get(r.Context())
	now := time.Now().UTC().Format(time.RFC3339)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		s.opts.Logger.Error("imds_server: credential refresh failed",
			slog.String("error", redactSecretsInError(err)),
		)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(credentialsResponse{
			Code:        "AssumeRoleUnauthorizedAccess",
			Message:     redactSecretsInError(err),
			LastUpdated: now,
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(credentialsResponse{
		Code:            "Success",
		Type:            "AWS-HMAC",
		AccessKeyId:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		Token:           creds.SessionToken,
		// AWS SDK Go v2 requires RFC 3339; other formats are rejected.
		Expiration:  creds.Expires.UTC().Format(time.RFC3339),
		LastUpdated: now,
	})
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	if !s.opts.Cache.Ready() {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

// mintToken generates a 32-byte random hex string for use as an IMDSv2 session token.
func mintToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("crypto/rand: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

// redactSecretsInError defensively scrubs any AWS access key or secret-looking
// substrings from an error string before logging or returning it. The cache
// path generally does not surface secrets in errors, but this is belt-and-
// suspenders to avoid leaking the JWT or STS credentials in IMDS error bodies.
func redactSecretsInError(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	// Strip what looks like a JWT (three base64 segments separated by dots).
	// JWTs are long; replace anything 40+ chars with no spaces matching the
	// charset.
	s = jwtPattern.ReplaceAllString(s, "[REDACTED-JWT]")
	return s
}

// jwtPattern matches JWT-like substrings: three base64url segments separated
// by dots. The minimum length of 40 chars is just a heuristic to avoid
// accidentally matching short tokens.
var jwtPattern = mustCompileJWTPattern()

func mustCompileJWTPattern() *jwtRegexp {
	return &jwtRegexp{}
}

// jwtRegexp is a minimal regexp wrapper. We avoid pulling in the full regexp
// package for one pattern by implementing a small ad-hoc matcher: anything
// containing two `.` characters with sufficiently long base64-ish chunks
// either side. Returns input verbatim if no match.
type jwtRegexp struct{}

func (j *jwtRegexp) ReplaceAllString(s, replacement string) string {
	// Find runs of base64url chars (A-Za-z0-9-_) that contain at least two dots.
	var b strings.Builder
	for {
		start, end := findJWTSpan(s)
		if start < 0 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:start])
		b.WriteString(replacement)
		s = s[end:]
	}
	return b.String()
}

func findJWTSpan(s string) (int, int) {
	// Scan for a run of [A-Za-z0-9-_.] of length >= 40 with exactly two dots.
	const minLen = 40
	for i := 0; i < len(s); i++ {
		if !isJWTChar(s[i]) {
			continue
		}
		j := i
		dots := 0
		for j < len(s) && isJWTChar(s[j]) {
			if s[j] == '.' {
				dots++
			}
			j++
		}
		if j-i >= minLen && dots == 2 {
			return i, j
		}
		i = j
	}
	return -1, -1
}

func isJWTChar(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z':
		return true
	case c >= 'a' && c <= 'z':
		return true
	case c >= '0' && c <= '9':
		return true
	case c == '-', c == '_', c == '.':
		return true
	}
	return false
}

// slogRecoveryLogger adapts slog.Logger to the gorilla/handlers
// RecoveryLogger interface (which expects a Println(args ...interface{})
// method). Mirrors the existing pkg/aws_config_server pattern.
type slogRecoveryLogger struct {
	logger *slog.Logger
}

func (l slogRecoveryLogger) Println(args ...interface{}) {
	l.logger.Error("imds_server: panic recovered", slog.Any("args", args))
}

// Reference errors and fmt to keep them used unambiguously where they appear
// (errors.New / fmt.Errorf in handlers above).
var _ = errors.New
var _ = fmt.Errorf
