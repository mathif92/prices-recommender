package serpapi

import (
	"context"
	"fmt"

	serpapi "github.com/mathif92/prices-recommender/pkg/client/serpapi"
	"github.com/mathif92/prices-recommender/pkg/collector"
	"github.com/mathif92/prices-recommender/pkg/recommendations"
	"github.com/mathif92/prices-recommender/pkg/repositories"
	"github.com/mathif92/prices-recommender/pkg/types"
)

const (
	collectorName     = "Serp API"
	dropThreshold     = 0.10
)

type serpApiCollector struct {
	apiClient serpapi.Client
	dataRepo  repositories.DataRepository
	drops     []recommendations.PriceDrop
}

func NewCollector(apiClient serpapi.Client, dataRepo repositories.DataRepository) collector.Collector {
	return &serpApiCollector{
		apiClient: apiClient,
		dataRepo:  dataRepo,
	}
}

func (d *serpApiCollector) Collect(ctx context.Context, params types.CollectParams) ([]types.HotelData, error) {
	switch params.Type {
	case types.RecommendationTypeHotel:
		return d.collectHotels(ctx, params)
	case types.RecommendationTypeFlight, types.RecommendationTypePackage:
		return nil, fmt.Errorf("%s collector does not support type: %s", collectorName, params.Type)
	default:
		return nil, nil
	}
}

func (d *serpApiCollector) Drops() []recommendations.PriceDrop {
	return d.drops
}

func (d *serpApiCollector) collectHotels(ctx context.Context, params types.CollectParams) ([]types.HotelData, error) {
	hotelsResponse, err := d.apiClient.GetHotels(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get hotels from SerpAPI: %w", err)
	}

	data := mapPropertiesToHotelData(hotelsResponse.Properties, params)

	hotelModels := make([]repositories.Hotel, 0, len(data))
	for _, hd := range data {
		hotelModels = append(hotelModels, toRepoHotel(hd))
	}
	saved, err := d.dataRepo.SaveHotels(ctx, hotelModels)
	if err != nil {
		return nil, fmt.Errorf("failed to save hotels: %w", err)
	}
	for i := range data {
		data[i].Hotel.ID = saved[i].ID
	}

	hotelRatingModels := make([]repositories.HotelRating, 0, len(data))
	for _, hd := range data {
		hotelRatingModels = append(hotelRatingModels, repositories.HotelRating{
			HotelID:     hd.Hotel.ID,
			Rating:      hd.Hotel.Rating,
			ReviewCount: hd.Rating.ReviewCount,
			CreatedAt:   hd.Rating.CreatedAt,
		})
	}
	if _, err := d.dataRepo.SaveHotelRatings(ctx, hotelRatingModels); err != nil {
		return nil, fmt.Errorf("failed to save hotel ratings: %w", err)
	}

	hotelReviewModels := make([]repositories.HotelReview, 0, len(data))
	for _, hd := range data {
		for _, review := range hd.Reviews {
			hotelReviewModels = append(hotelReviewModels, repositories.HotelReview{
				HotelID:      hd.Hotel.ID,
				Name:         review.Name,
				Positive:     review.Positive,
				Negative:     review.Negative,
				Neutral:      review.Neutral,
				ExternalLink: review.ExternalLink,
				CreatedAt:    review.CreatedAt,
			})
		}
	}
	if _, err := d.dataRepo.SaveHotelReviews(ctx, hotelReviewModels); err != nil {
		return nil, fmt.Errorf("failed to save hotel reviews: %w", err)
	}

	prices := make([]repositories.Price, 0, len(data))
	for _, hd := range data {
		if hd.Price.Price == 0 {
			continue
		}

		d.checkDrop(ctx, hd)

		prices = append(prices, repositories.Price{
			HotelID:       hd.Hotel.ID,
			Price:         hd.Price.Price,
			StartDate:     hd.Price.StartDate,
			EndDate:       hd.Price.EndDate,
			Currency:      hd.Price.Currency,
			CheckinTime:   hd.Price.CheckinTime,
			CheckoutTime:  hd.Price.CheckoutTime,
			CreatedAt: hd.Price.CreatedAt,
		})
	}
	if _, err := d.dataRepo.SavePrices(ctx, prices); err != nil {
		return nil, fmt.Errorf("failed to save prices: %w", err)
	}

	return data, nil
}

func (d *serpApiCollector) checkDrop(ctx context.Context, hd types.HotelData) {
	current, err := d.dataRepo.GetCurrentPrice(ctx, hd.Hotel.ID, hd.Price.StartDate, hd.Price.EndDate)
	if err != nil || current == nil || current.Price <= 0 || hd.Price.Price <= 0 {
		return
	}

	ratio := 1 - hd.Price.Price/current.Price
	if ratio >= dropThreshold {
		d.drops = append(d.drops, recommendations.PriceDrop{
			HotelID:   hd.Hotel.ID,
			HotelName: hd.Hotel.Name,
			Location:  hd.Hotel.Location,
			StartDate: hd.Price.StartDate,
			EndDate:   hd.Price.EndDate,
			OldPrice:  current.Price,
			NewPrice:  hd.Price.Price,
			Currency:  hd.Price.Currency,
			DropRatio: ratio,
		})
	}
}

func (d *serpApiCollector) Name() string {
	return collectorName
}

func toRepoHotel(hd types.HotelData) repositories.Hotel {
	return repositories.Hotel{
		Name:           hd.Hotel.Name,
		Description:    hd.Hotel.Description,
		Link:           hd.Hotel.Link,
		Rating:         hd.Hotel.Rating,
		Location:       hd.Hotel.Location,
		IsAllInclusive: hd.Hotel.IsAllInclusive,
		CreatedAt:      hd.Hotel.CreatedAt,
	}
}
