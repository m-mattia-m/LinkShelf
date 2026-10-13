package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"
	"backend/internal/infrastructure/api/model"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

// etagOf is a strong validator over the serialized page.
func etagOf(v any) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return `"` + hex.EncodeToString(sum[:16]) + `"`, nil
}

// etagMatches implements the weak comparison of If-None-Match, including "*" and lists.
func etagMatches(header, etag string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimPrefix(strings.TrimSpace(candidate), "W/")
		if candidate == "*" || candidate == etag {
			return true
		}
	}
	return false
}

// ListServiceDomains is the spec an external controller reconciles: every shelf domain.
func ListServiceDomains(svc *domain.Service) func(c context.Context, input *model.ServiceDomainListFilter) (*model.ServiceDomainListResponse, error) {
	return func(c context.Context, input *model.ServiceDomainListFilter) (*model.ServiceDomainListResponse, error) {
		items, next, err := svc.ShelfService.ListDomains(input.Cursor, input.Limit)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list domains", err)
		}

		page := model.ServiceDomainList{Items: items}
		if next != "" {
			page.NextCursor = &next
		}

		etag, err := etagOf(page)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to list domains", err)
		}
		if input.IfNoneMatch != "" && etagMatches(input.IfNoneMatch, etag) {
			return nil, huma.Status304NotModified()
		}
		return &model.ServiceDomainListResponse{ETag: etag, Body: page}, nil
	}
}

// SetServiceDomainStatus stores the status an external controller computed for a domain.
func SetServiceDomainStatus(svc *domain.Service) func(c context.Context, input *model.ServiceDomainStatusRequest) (*model.ServiceDomainResponse, error) {
	return func(c context.Context, input *model.ServiceDomainStatusRequest) (*model.ServiceDomainResponse, error) {
		status := model.DomainStatus{
			State:   input.Body.State,
			Reason:  input.Body.Reason,
			Message: input.Body.Message,
			Records: input.Body.Records,
		}

		item, err := svc.ShelfService.SetDomainStatus(input.Domain, input.Body.ShelfId, status)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrConflict):
				return nil, huma.Error409Conflict(err.Error())
			case errors.Is(err, domain.ErrInvalidInput):
				return nil, huma.Error400BadRequest(err.Error())
			}
			return nil, mapper.MapWriteError("failed to set the domain status", err)
		}

		zap.L().Info("service token set a domain status",
			zap.String("domain", item.Domain), zap.String("shelf_id", item.ShelfId), zap.String("state", item.Status.State), zap.String("reason", item.Status.Reason))
		return &model.ServiceDomainResponse{Body: *item}, nil
	}
}
