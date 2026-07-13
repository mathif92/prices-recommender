package serpapi

import (
	"strings"
	"time"

	serpapi "github.com/mathif92/prices-recommender/pkg/client/serpapi"
	"github.com/mathif92/prices-recommender/pkg/types"
)

func mapPropertyToHotelData(p serpapi.Property, params types.CollectParams) types.HotelData {
	now := time.Now()

	hotel := types.Hotel{
		Name:      p.Name,
		Link:      p.Link,
		Rating:    p.OverallRating,
		Location:  params.Destination,
		CreatedAt: &now,
	}
	if p.Description != "" {
		hotel.Description = &p.Description
	}
	for _, amenity := range p.Amenities {
		if strings.Contains(strings.ToLower(amenity), "all inclusive") {
			hotel.IsAllInclusive = true
			break
		}
	}

	price := types.HotelPrice{
		Price:               float64(p.TotalRate.ExtractedLowest),
		StartDate:           params.CheckIn,
		EndDate:             params.CheckOut,
		Currency:            "USD",
		CheckinTime:         strPtr(p.CheckInTime),
		CheckoutTime:        strPtr(p.CheckOutTime),
		PropertyDetailsLink: strPtr(p.SerpapiPropertyDetailsLink),
		CreatedAt:       &now,
	}

	rating := types.HotelRating{
		Rating:      p.OverallRating,
		ReviewCount: p.Reviews,
		CreatedAt:   &now,
	}

	reviews := make([]types.HotelReview, 0, len(p.ReviewsBreakdown))
	for _, rb := range p.ReviewsBreakdown {
		reviews = append(reviews, types.HotelReview{
			Name:         rb.Name,
			Positive:     rb.Positive,
			Negative:     rb.Negative,
			Neutral:      rb.Neutral,
			ExternalLink: rb.SerpapiLink,
			CreatedAt:    &now,
		})
	}

	return types.HotelData{
		Hotel:   hotel,
		Price:   price,
		Rating:  rating,
		Reviews: reviews,
	}
}

func mapPropertiesToHotelData(properties []serpapi.Property, params types.CollectParams) []types.HotelData {
	data := make([]types.HotelData, 0, len(properties))
	for _, p := range properties {
		data = append(data, mapPropertyToHotelData(p, params))
	}
	return data
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
