package aws_config_server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

const requestIDHeader = "X-Request-Id"

var oktaUserIDPattern = regexp.MustCompile(`^00u[A-Za-z0-9]{17}$`)

type EntitlementRequest struct {
	Issuer string `json:"issuer"`
	UserID string `json:"userId"`
}

type Entitlement struct {
	AccountID    string `json:"accountId"`
	AccountAlias string `json:"accountAlias,omitempty"`
	RoleARN      string `json:"roleArn"`
	RoleName     string `json:"roleName"`
}

type EntitlementList struct {
	Items []Entitlement `json:"items"`
}

type WorkloadAuthMiddleware struct {
	handler        http.Handler
	verifier       oidcVerifier
	trustedSubject string
}

func NewWorkloadAuthMiddleware(
	handler http.Handler,
	verifier oidcVerifier,
	trustedSubject string,
) (*WorkloadAuthMiddleware, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	if verifier == nil {
		return nil, errors.New("verifier is required")
	}
	if strings.TrimSpace(trustedSubject) == "" {
		return nil, errors.New("trusted subject is required")
	}
	return &WorkloadAuthMiddleware{
		handler:        handler,
		verifier:       verifier,
		trustedSubject: trustedSubject,
	}, nil
}

func (m *WorkloadAuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get(requestIDHeader)
	if requestID != "" {
		w.Header().Set(requestIDHeader, requestID)
	}
	rawToken := stripBearerPrefixFromTokenString(r.Header.Get("Authorization"))
	if rawToken == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	idToken, err := m.verifier.Verify(r.Context(), rawToken)
	if err != nil {
		slog.Warn("verifying entitlement caller", "request_id", requestID, "error", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if idToken.Subject != m.trustedSubject {
		slog.Warn(
			"rejecting entitlement caller",
			"request_id", requestID,
			"service_subject", idToken.Subject,
		)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	m.handler.ServeHTTP(w, r)
}

type EntitlementHandler struct {
	ownerIssuer      string
	oktaAppClient    okta.AppLister
	mappingsProvider MappingsProvider
}

func NewEntitlementHandler(
	ownerIssuer string,
	oktaAppClient okta.AppLister,
	mappingsProvider MappingsProvider,
) (*EntitlementHandler, error) {
	if strings.TrimSpace(ownerIssuer) == "" {
		return nil, errors.New("owner issuer is required")
	}
	if oktaAppClient == nil {
		return nil, errors.New("Okta application client is required")
	}
	if mappingsProvider == nil {
		return nil, errors.New("mappings provider is required")
	}
	return &EntitlementHandler{
		ownerIssuer:      strings.TrimSuffix(ownerIssuer, "/"),
		oktaAppClient:    oktaAppClient,
		mappingsProvider: mappingsProvider,
	}, nil
}

func (h *EntitlementHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get(requestIDHeader)
	if requestID != "" {
		w.Header().Set(requestIDHeader, requestID)
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request EntitlementRequest
	err := decoder.Decode(&request)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	err = decoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if strings.TrimSuffix(request.Issuer, "/") != h.ownerIssuer || !oktaUserIDPattern.MatchString(request.UserID) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	mappings, err := h.mappingsProvider(r.Context())
	if err != nil {
		slog.Error("reading entitlement rolemap", "request_id", requestID, "error", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	items, err := resolveEntitlements(r.Context(), request.UserID, h.oktaAppClient, mappings)
	if err != nil {
		slog.Error("resolving AWS entitlements", "request_id", requestID, "user_id", request.UserID, "error", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	slog.Info(
		"resolved AWS entitlements",
		"request_id", requestID,
		"user_id", request.UserID,
		"count", len(items),
	)
	err = json.NewEncoder(w).Encode(EntitlementList{Items: items})
	if err != nil {
		slog.Error("writing AWS entitlements", "request_id", requestID, "error", err)
	}
}

func resolveEntitlements(
	ctx context.Context,
	userID string,
	oktaAppClient okta.AppLister,
	mappings okta.OIDCRoleMappingsByKey,
) ([]Entitlement, error) {
	clientIDs, err := okta.GetClientIDs(ctx, userID, oktaAppClient)
	if err != nil {
		return nil, fmt.Errorf("get assigned applications: %w", err)
	}
	seen := make(map[string]struct{})
	items := make([]Entitlement, 0)
	for _, clientID := range clientIDs {
		for _, mapping := range mappings[clientID.String()] {
			roleARN, err := arn.Parse(mapping.AWSRoleARN)
			if err != nil {
				return nil, fmt.Errorf("parse role ARN: %w", err)
			}
			if roleARN.Service != "iam" || roleARN.AccountID != mapping.AWSAccountID || !strings.HasPrefix(roleARN.Resource, "role/") {
				return nil, fmt.Errorf("role ARN %q does not match account %q", mapping.AWSRoleARN, mapping.AWSAccountID)
			}
			key := mapping.AWSAccountID + "\x00" + mapping.AWSRoleARN
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			items = append(items, Entitlement{
				AccountID:    mapping.AWSAccountID,
				AccountAlias: mapping.AWSAccountAlias,
				RoleARN:      mapping.AWSRoleARN,
				RoleName:     strings.TrimPrefix(roleARN.Resource, "role/"),
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		leftAccount := items[i].AccountAlias
		if leftAccount == "" {
			leftAccount = items[i].AccountID
		}
		rightAccount := items[j].AccountAlias
		if rightAccount == "" {
			rightAccount = items[j].AccountID
		}
		if leftAccount != rightAccount {
			return leftAccount < rightAccount
		}
		if items[i].RoleName != items[j].RoleName {
			return items[i].RoleName < items[j].RoleName
		}
		return items[i].RoleARN < items[j].RoleARN
	})
	return items, nil
}
