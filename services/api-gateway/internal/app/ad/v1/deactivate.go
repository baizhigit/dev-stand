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

func (i *Implementation) Deactivate(ctx context.Context, req *adV1.DeactivateRequest) (*adV1.DeactivateResponse, error) {
	if err := validateDeactivate(req); err != nil {
		return nil, err
	}

	resp, err := i.external.Deactivate(ctx, convert.ToExternalDeactivateRequest(req))
	if err != nil {
		return nil, grpcerr.Translate(err)
	}

	return convert.ToDeactivateResponse(resp), nil
}

func validateDeactivate(req *adV1.DeactivateRequest) error {
	if strings.TrimSpace(req.GetId()) == "" {
		return status.Error(codes.InvalidArgument, "id is required")
	}
	return nil
}
