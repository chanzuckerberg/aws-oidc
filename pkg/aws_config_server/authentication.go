package aws_config_server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	oidc "github.com/coreos/go-oidc"
)

const requestIDHeader = "X-Request-Id"

type oidcVerifier interface {
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type IdentityVerifier interface {
	Verify(context.Context, string) (Principal, error)
}

type Principal struct {
	Issuer     string
	Subject    string
	Email      string
	Privileged bool
}

type VerifierConfig struct {
	Issuer   string
	Verifier oidcVerifier
}

type PrivilegedIdentity struct {
	Issuer  string
	Subject string
}

type MultiVerifier struct {
	verifiers  map[string][]oidcVerifier
	privileged map[string]struct{}
}

func NewMultiVerifier(
	configs []VerifierConfig,
	privileged []PrivilegedIdentity,
) (*MultiVerifier, error) {
	if len(configs) == 0 {
		return nil, errors.New("at least one verifier is required")
	}
	verifiers := make(map[string][]oidcVerifier, len(configs))
	for _, config := range configs {
		issuer := normalizeIssuer(config.Issuer)
		if issuer == "" {
			return nil, errors.New("verifier issuer is required")
		}
		if config.Verifier == nil {
			return nil, fmt.Errorf("verifier for issuer %q is required", issuer)
		}
		verifiers[issuer] = append(verifiers[issuer], config.Verifier)
	}
	privilegedSet := make(map[string]struct{}, len(privileged))
	for _, identity := range privileged {
		issuer := normalizeIssuer(identity.Issuer)
		if _, ok := verifiers[issuer]; !ok {
			return nil, fmt.Errorf("privileged issuer %q has no verifier", issuer)
		}
		if strings.TrimSpace(identity.Subject) == "" {
			return nil, fmt.Errorf("privileged subject for issuer %q is required", issuer)
		}
		privilegedSet[identityKey(issuer, identity.Subject)] = struct{}{}
	}
	return &MultiVerifier{
		verifiers:  verifiers,
		privileged: privilegedSet,
	}, nil
}

func (v *MultiVerifier) Verify(ctx context.Context, rawToken string) (Principal, error) {
	claims, err := unverifiedTokenClaims(rawToken)
	if err != nil {
		return Principal{}, err
	}
	issuer := claims.Issuer
	verifiers := v.verifiers[normalizeIssuer(issuer)]
	if len(verifiers) == 0 {
		return Principal{}, fmt.Errorf("issuer %q is not trusted", issuer)
	}
	var verifyErr error
	for _, verifier := range verifiers {
		idToken, err := verifier.Verify(ctx, rawToken)
		if err != nil {
			verifyErr = err
			continue
		}
		if idToken == nil {
			verifyErr = errors.New("verifier returned no identity token")
			continue
		}
		verifiedIssuer := normalizeIssuer(idToken.Issuer)
		if verifiedIssuer != normalizeIssuer(issuer) {
			return Principal{}, errors.New("verified issuer does not match token issuer")
		}
		if strings.TrimSpace(idToken.Subject) == "" {
			return Principal{}, errors.New("verified token has no subject")
		}
		_, isPrivileged := v.privileged[identityKey(verifiedIssuer, idToken.Subject)]
		return Principal{
			Issuer:     verifiedIssuer,
			Subject:    idToken.Subject,
			Email:      claims.Email,
			Privileged: isPrivileged,
		}, nil
	}
	return Principal{}, fmt.Errorf("verify identity token: %w", verifyErr)
}

type principalContextKey struct{}

func principalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

type AuthMiddleware struct {
	handler  http.Handler
	verifier IdentityVerifier
}

func NewAuthMiddleware(handler http.Handler, verifier IdentityVerifier) *AuthMiddleware {
	return &AuthMiddleware{handler: handler, verifier: verifier}
}

func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get(requestIDHeader)
	if requestID != "" {
		w.Header().Set(requestIDHeader, requestID)
	}
	rawToken := stripBearerPrefixFromTokenString(r.Header.Get("Authorization"))
	if rawToken == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	principal, err := m.verifier.Verify(r.Context(), rawToken)
	if err != nil {
		slog.Warn("verifying identity token", "request_id", requestID, "error", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
	m.handler.ServeHTTP(w, r.WithContext(ctx))
}

type unverifiedClaims struct {
	Issuer string `json:"iss"`
	Email  string `json:"email"`
}

func unverifiedTokenClaims(rawToken string) (unverifiedClaims, error) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return unverifiedClaims{}, errors.New("identity token must contain three segments")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return unverifiedClaims{}, fmt.Errorf("decode identity token payload: %w", err)
	}
	var claims unverifiedClaims
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return unverifiedClaims{}, fmt.Errorf("parse identity token payload: %w", err)
	}
	if normalizeIssuer(claims.Issuer) == "" {
		return unverifiedClaims{}, errors.New("identity token has no issuer")
	}
	return claims, nil
}

func identityKey(issuer string, subject string) string {
	return normalizeIssuer(issuer) + "\x00" + subject
}

func normalizeIssuer(issuer string) string {
	return strings.TrimSuffix(strings.TrimSpace(issuer), "/")
}
