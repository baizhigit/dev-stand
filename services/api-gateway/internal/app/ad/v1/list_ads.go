package ad

import (
	"context"

	"github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/convert"
	"github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/grpcerr"
	adV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/api_gateway/ad/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxPageSize     = 100
	defaultPageSize = 20
)

func (i *Implementation) ListAds(ctx context.Context, req *adV1.ListAdsRequest) (*adV1.ListAdsResponse, error) {
	if err := validateListAds(req); err != nil {
		return nil, err
	}

	// Apply default page size if caller omitted it
	if req.GetPageSize() == 0 {
		req.PageSize = defaultPageSize
	}

	resp, err := i.external.ListAds(ctx, convert.ToExternalListAdsRequest(req))
	if err != nil {
		return nil, grpcerr.Translate(err)
	}

	return convert.ToListAdsResponse(resp), nil
}

func validateListAds(req *adV1.ListAdsRequest) error {
	if req.GetPageSize() < 0 {
		return status.Error(codes.InvalidArgument, "page_size cannot be negative")
	}
	if req.GetPageSize() > maxPageSize {
		return status.Errorf(codes.InvalidArgument, "page_size cannot exceed %d", maxPageSize)
	}
	return nil
}
