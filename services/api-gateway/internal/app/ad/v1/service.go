package ad

import (
	adV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/api_gateway/ad/v1"
	extV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/external/ad_service/ad/v1"
)

// Service interface — callers and tests depend on this, not on Implementation
// type Service interface {
// 	CreateAd(context.Context, *adV1.CreateAdRequest) (*adV1.CreateAdResponse, error)
// 	GetAd(context.Context, *adV1.GetAdRequest) (*adV1.GetAdResponse, error)
// 	ListAds(context.Context, *adV1.ListAdsRequest) (*adV1.ListAdsResponse, error)
// 	Deactivate(context.Context, *adV1.DeactivateRequest) (*adV1.DeactivateResponse, error)
// }

type Implementation struct {
	adV1.UnimplementedAdServiceServer                       // forward-compatible: new proto methods don't break build
	external                          extV1.AdServiceClient // interface — mockable in tests
}

// New takes interface, returns interface — both sides are mockable
func New(external extV1.AdServiceClient) *Implementation {
	return &Implementation{external: external}
}
