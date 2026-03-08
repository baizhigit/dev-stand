package convert

import (
	adV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/api_gateway/ad/v1"
	extV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/external/ad_service/ad/v1"
)

func ToExternalCreateAdRequest(req *adV1.CreateAdRequest) *extV1.CreateAdRequest {
	return &extV1.CreateAdRequest{
		Title:    req.GetTitle(),
		Category: req.GetCategory(),
		AuthorId: req.GetAuthorId(),
		Price:    req.GetPrice(),
	}
}

func ToCreateAdResponse(r *extV1.CreateAdResponse) *adV1.CreateAdResponse {
	return &adV1.CreateAdResponse{
		Ad: toAd(r.GetAd()),
	}
}

func ToExternalGetAdRequest(req *adV1.GetAdRequest) *extV1.GetAdRequest {
	return &extV1.GetAdRequest{
		AdId: req.GetAdId(),
	}
}

func ToGetAdResponse(r *extV1.GetAdResponse) *adV1.GetAdResponse {
	return &adV1.GetAdResponse{
		Ad: toAd(r.GetAd()),
	}
}

func ToExternalListAdsRequest(req *adV1.ListAdsRequest) *extV1.ListAdsRequest {
	return &extV1.ListAdsRequest{
		Category:  req.GetCategory(),
		Status:    extV1.AdStatus(req.GetStatus()), // adV1.AdStatus → extV1.AdStatus
		PageSize:  req.GetPageSize(),
		PageToken: req.GetPageToken(),
	}
}

func ToListAdsResponse(r *extV1.ListAdsResponse) *adV1.ListAdsResponse {
	ads := make([]*adV1.Ad, 0, len(r.GetAds()))
	for _, a := range r.GetAds() {
		ads = append(ads, toAd(a))
	}
	return &adV1.ListAdsResponse{
		Ads:           ads,
		NextPageToken: r.GetNextPageToken(),
	}
}

func ToExternalDeactivateRequest(req *adV1.DeactivateRequest) *extV1.DeactivateRequest {
	return &extV1.DeactivateRequest{
		Id: req.GetId(),
	}
}

func ToDeactivateResponse(r *extV1.DeactivateResponse) *adV1.DeactivateResponse {
	return &adV1.DeactivateResponse{
		Status: adV1.AdStatus(r.GetStatus()),
	}
}

// toAd is unexported — used internally by multiple response converters.
// This is the payoff of having a top-level Ad message:
// one mapping function shared by GetAd, ListAds, CreateAd responses.
func toAd(a *extV1.Ad) *adV1.Ad {
	if a == nil {
		return nil
	}
	return &adV1.Ad{
		Id:          a.GetId(),
		Title:       a.GetTitle(),
		Description: a.GetDescription(),
		Category:    a.GetCategory(),
		AuthorId:    a.GetAuthorId(),
		Price:       a.GetPrice(),
		Status:      adV1.AdStatus(a.GetStatus()), // cast: extV1.AdStatus → adV1.AdStatus
		CreatedAt:   a.GetCreatedAt(),
		Review:      toAdReview(a.GetReview()),
	}
}

func toAdReview(r *extV1.AdReview) *adV1.AdReview {
	if r == nil {
		return nil
	}
	return &adV1.AdReview{
		AvgReview: r.GetAvgReview(),
	}
}
