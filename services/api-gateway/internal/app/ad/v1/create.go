package ad

import (
	"context"
	"strings"

	"github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/convert"
	"github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/grpcerr"

	adV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/api_gateway/ad/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Every handler follows this exact shape — no exceptions:
// 1. Validate input
// 2. Convert request to downstream type
// 3. Call downstream
// 4. Translate error
// 5. Convert response
func (i *Implementation) CreateAd(ctx context.Context, req *adV1.CreateAdRequest) (*adV1.CreateAdResponse, error) {
	if err := validateCreateAd(req); err != nil {
		return nil, err // already a gRPC status error, no further wrapping needed
	}

	resp, err := i.external.CreateAd(ctx, convert.ToExternalCreateAdRequest(req))
	if err != nil {
		return nil, grpcerr.Translate(err) // never return raw downstream errors
	}

	return convert.ToCreateAdResponse(resp), nil
}

// Validation in its own function: testable in isolation, handler stays readable
func validateCreateAd(req *adV1.CreateAdRequest) error {
	if strings.TrimSpace(req.GetTitle()) == "" {
		return status.Error(codes.InvalidArgument, "title is required")
	}
	if strings.TrimSpace(req.GetAuthorId()) == "" {
		return status.Error(codes.InvalidArgument, "author_id is required")
	}
	if strings.TrimSpace(req.GetCategory()) == "" {
		return status.Error(codes.InvalidArgument, "category is required")
	}
	if req.GetPrice() == nil {
		return status.Error(codes.InvalidArgument, "price is required")
	}
	if req.GetPrice().GetUnits() < 0 {
		return status.Error(codes.InvalidArgument, "price cannot be negative")
	}
	if strings.TrimSpace(req.GetPrice().GetCurrencyCode()) == "" {
		return status.Error(codes.InvalidArgument, "price.currency_code is required")
	}
	return nil
}
