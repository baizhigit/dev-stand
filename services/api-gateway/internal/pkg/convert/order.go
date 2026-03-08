package convert

import (
	orderV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/api_gateway/order/v1"
	extV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/external/order_service/order/v1"
)

func ToExternalCreateOrderRequest(req *orderV1.CreateOrderRequest) *extV1.CreateOrderRequest {
	return &extV1.CreateOrderRequest{
		Order: toExternalOrder(req.GetOrder()),
	}
}

func ToCreateOrderResponse(r *extV1.CreateOrderResponse) *orderV1.CreateOrderResponse {
	return &orderV1.CreateOrderResponse{
		Details: toDetails(r.GetDetails()),
	}
}

func toExternalOrder(o *orderV1.Order) *extV1.Order {
	if o == nil {
		return nil
	}
	return &extV1.Order{
		AdId:     o.GetAdId(),
		Category: o.GetCategory(),
		ClientId: o.GetClientId(),
		Price:    o.GetPrice(),
	}
}

func toDetails(d *extV1.Details) *orderV1.Details {
	if d == nil {
		return nil
	}
	return &orderV1.Details{
		OrderId:     d.GetOrderId(),
		OrderStatus: d.GetOrderStatus(),
	}
}
