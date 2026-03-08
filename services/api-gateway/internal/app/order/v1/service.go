package order

import (
	orderV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/api_gateway/order/v1"
	extV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/external/order_service/order/v1"
)

// type Service interface {
// 	CreateOrder(context.Context, *orderV1.CreateOrderRequest) (*orderV1.CreateOrderResponse, error)
// }

type Implementation struct {
	orderV1.UnimplementedOrderServiceServer
	external extV1.OrderServiceClient
}

func New(external extV1.OrderServiceClient) *Implementation {
	return &Implementation{external: external}
}
