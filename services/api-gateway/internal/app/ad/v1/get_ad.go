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

func (i *Implementation) GetAd(ctx context.Context, req *adV1.GetAdRequest) (*adV1.GetAdResponse, error) {
	if err := validateGetAd(req); err != nil {
		return nil, err
	}

	resp, err := i.external.GetAd(ctx, convert.ToExternalGetAdRequest(req))
	if err != nil {
		return nil, grpcerr.Translate(err)
	}

	return convert.ToGetAdResponse(resp), nil
}

func validateGetAd(req *adV1.GetAdRequest) error {
	if strings.TrimSpace(req.GetAdId()) == "" {
		return status.Error(codes.InvalidArgument, "ad_id is required")
	}
	return nil
}
