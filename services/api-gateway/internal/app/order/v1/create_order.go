package order

import (
	"context"
	"strings"

	"github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/convert"
	"github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/grpcerr"
	orderV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/api_gateway/order/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) CreateOrder(ctx context.Context, req *orderV1.CreateOrderRequest) (*orderV1.CreateOrderResponse, error) {
	if err := validateCreateOrder(req); err != nil {
		return nil, err
	}

	resp, err := i.external.CreateOrder(ctx, convert.ToExternalCreateOrderRequest(req))
	if err != nil {
		return nil, grpcerr.Translate(err)
	}

	return convert.ToCreateOrderResponse(resp), nil
}

func validateCreateOrder(req *orderV1.CreateOrderRequest) error {
	o := req.GetOrder() // get the nested message once — nil-safe via Get*
	if o == nil {
		return status.Error(codes.InvalidArgument, "order is required")
	}
	if strings.TrimSpace(o.GetAdId()) == "" {
		return status.Error(codes.InvalidArgument, "order.ad_id is required")
	}
	if strings.TrimSpace(o.GetClientId()) == "" {
		return status.Error(codes.InvalidArgument, "order.client_id is required")
	}
	if strings.TrimSpace(o.GetCategory()) == "" {
		return status.Error(codes.InvalidArgument, "order.category is required")
	}
	if o.GetPrice() == nil {
		return status.Error(codes.InvalidArgument, "order.price is required")
	}
	if o.GetPrice().GetUnits() < 0 {
		return status.Error(codes.InvalidArgument, "order.price cannot be negative")
	}
	if strings.TrimSpace(o.GetPrice().GetCurrencyCode()) == "" {
		return status.Error(codes.InvalidArgument, "order.price.currency_code is required")
	}
	return nil
}
