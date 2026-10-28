package aws_config_server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
)

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

type EntitlementHandler struct {
	resolver *AccessResolver
}

func NewEntitlementHandler(resolver *AccessResolver) (*EntitlementHandler, error) {
	if resolver == nil {
		return nil, errors.New("access resolver is required")
	}
	return &EntitlementHandler{resolver: resolver}, nil
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
	principal, ok := principalFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	targetIssuer := normalizeIssuer(request.Issuer)
	targetUserID := request.UserID
	if targetIssuer == "" && targetUserID == "" {
		targetIssuer = principal.Issuer
		targetUserID = principal.Subject
	}
	if targetIssuer == "" || targetUserID == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	targetIsCaller := targetIssuer == principal.Issuer && targetUserID == principal.Subject
	if !targetIsCaller && !principal.Privileged {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	if targetIssuer != h.resolver.OIDCProvider() || !oktaUserIDPattern.MatchString(targetUserID) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	config, err := h.resolver.Resolve(r.Context(), targetUserID)
	if err != nil {
		slog.Error("resolving AWS entitlements", "request_id", requestID, "user_id", targetUserID, "error", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	items := entitlementsFromAWSConfig(config)
	slog.Info(
		"resolved AWS entitlements",
		"request_id", requestID,
		"principal_issuer", principal.Issuer,
		"principal_subject", principal.Subject,
		"target_user_id", targetUserID,
		"count", len(items),
	)
	err = json.NewEncoder(w).Encode(EntitlementList{Items: items})
	if err != nil {
		slog.Error("writing AWS entitlements", "request_id", requestID, "error", err)
	}
}

func entitlementsFromAWSConfig(config *AWSConfig) []Entitlement {
	seen := make(map[string]struct{})
	items := make([]Entitlement, 0)
	for _, profile := range config.Profiles {
		key := profile.AWSAccount.ID + "\x00" + profile.RoleARN
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, Entitlement{
			AccountID:    profile.AWSAccount.ID,
			AccountAlias: profile.AWSAccount.Alias,
			RoleARN:      profile.RoleARN,
			RoleName:     profile.RoleName,
		})
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
	return items
}
