package serpapi

import (
	"testing"
	"time"

	serpapi "github.com/mathif92/prices-recommender/pkg/client/serpapi"
	"github.com/mathif92/prices-recommender/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestMapPropertyToHotelData(t *testing.T) {
	params := types.CollectParams{
		Destination: "Cancun",
		CheckIn:     parseTime("2027-06-01"),
		CheckOut:    parseTime("2027-06-10"),
	}

	p := serpapi.Property{
		Name:        "Grand Resort Cancun",
		Description: "A luxurious beachfront resort",
		Link:        "https://example.com/grand",
		OverallRating: 4.5,
		Reviews:     200,
		TotalRate: serpapi.PriceValue{
			ExtractedLowest: 35000,
		},
		CheckInTime:  "15:00",
		CheckOutTime: "12:00",
		SerpapiPropertyDetailsLink: "https://serpapi.com/details",
		Amenities: []string{"pool", "all inclusive", "wifi"},
		ReviewsBreakdown: []serpapi.ReviewBreakdown{
			{Name: "Cleanliness", Positive: 80, Negative: 5, Neutral: 15, SerpapiLink: "https://serpapi.com/clean"},
		},
	}

	result := mapPropertyToHotelData(p, params)

	assert.Equal(t, "Grand Resort Cancun", result.Hotel.Name)
	assert.Equal(t, "Cancun", result.Hotel.Location)
	assert.Equal(t, "https://example.com/grand", result.Hotel.Link)
	assert.Equal(t, 4.5, result.Hotel.Rating)
	assert.True(t, result.Hotel.IsAllInclusive)
	assert.NotNil(t, result.Hotel.Description)
	assert.Equal(t, "A luxurious beachfront resort", *result.Hotel.Description)
	assert.NotNil(t, result.Hotel.CreatedAt)

	assert.Equal(t, 35000.0, result.Price.Price)
	assert.Equal(t, params.CheckIn, result.Price.StartDate)
	assert.Equal(t, params.CheckOut, result.Price.EndDate)
	assert.Equal(t, "USD", result.Price.Currency)
	assert.NotNil(t, result.Price.CheckinTime)
	assert.Equal(t, "15:00", *result.Price.CheckinTime)
	assert.Equal(t, "12:00", *result.Price.CheckoutTime)
	assert.NotNil(t, result.Price.PropertyDetailsLink)
	assert.NotNil(t, result.Price.CreatedAt)

	assert.Equal(t, 4.5, result.Rating.Rating)
	assert.Equal(t, 200, result.Rating.ReviewCount)

	assert.Len(t, result.Reviews, 1)
	assert.Equal(t, "Cleanliness", result.Reviews[0].Name)
	assert.Equal(t, 80, result.Reviews[0].Positive)
	assert.Equal(t, 5, result.Reviews[0].Negative)
	assert.Equal(t, 15, result.Reviews[0].Neutral)
}

func TestMapPropertyAllInclusiveDetection(t *testing.T) {
	tests := []struct {
		name     string
		amenities []string
		want     bool
	}{
		{"all inclusive keyword", []string{"pool", "all inclusive", "wifi"}, true},
		{"case insensitive", []string{"All Inclusive"}, true},
		{"no all inclusive", []string{"pool", "wifi", "breakfast"}, false},
		{"empty amenities", nil, false},
	}

	params := types.CollectParams{Destination: "Test"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := serpapi.Property{
				Name:       "Hotel",
				Link:       "https://example.com",
				Amenities:  tt.amenities,
				TotalRate:  serpapi.PriceValue{ExtractedLowest: 100},
			}
			result := mapPropertyToHotelData(p, params)
			assert.Equal(t, tt.want, result.Hotel.IsAllInclusive)
		})
	}
}

func TestMapPropertyEmptyDescription(t *testing.T) {
	params := types.CollectParams{Destination: "Test"}
	p := serpapi.Property{
		Name:      "No Desc Hotel",
		Link:      "https://example.com",
		TotalRate: serpapi.PriceValue{ExtractedLowest: 100},
	}

	result := mapPropertyToHotelData(p, params)
	assert.Nil(t, result.Hotel.Description)
}

func TestMapPropertyEmptyCheckinCheckout(t *testing.T) {
	params := types.CollectParams{Destination: "Test"}
	p := serpapi.Property{
		Name:      "Hotel",
		Link:      "https://example.com",
		TotalRate: serpapi.PriceValue{ExtractedLowest: 100},
	}

	result := mapPropertyToHotelData(p, params)
	assert.Nil(t, result.Price.CheckinTime)
	assert.Nil(t, result.Price.CheckoutTime)
	assert.Nil(t, result.Price.PropertyDetailsLink)
}

func TestMapPropertyNoReviewsBreakdown(t *testing.T) {
	params := types.CollectParams{Destination: "Test"}
	p := serpapi.Property{
		Name:      "Hotel",
		Link:      "https://example.com",
		TotalRate: serpapi.PriceValue{ExtractedLowest: 100},
	}

	result := mapPropertyToHotelData(p, params)
	assert.Empty(t, result.Reviews)
}

func TestMapPropertiesToHotelData(t *testing.T) {
	params := types.CollectParams{Destination: "Paris"}

	properties := []serpapi.Property{
		{
			Name:      "Hotel A",
			Link:      "https://example.com/a",
			TotalRate: serpapi.PriceValue{ExtractedLowest: 200},
		},
		{
			Name:      "Hotel B",
			Link:      "https://example.com/b",
			TotalRate: serpapi.PriceValue{ExtractedLowest: 300},
			Amenities: []string{"all inclusive"},
		},
		{
			Name:      "Hotel C",
			Link:      "https://example.com/c",
			TotalRate: serpapi.PriceValue{ExtractedLowest: 150},
		},
	}

	result := mapPropertiesToHotelData(properties, params)
	assert.Len(t, result, 3)
	assert.Equal(t, "Hotel A", result[0].Hotel.Name)
	assert.Equal(t, "Hotel B", result[1].Hotel.Name)
	assert.True(t, result[1].Hotel.IsAllInclusive)
	assert.Equal(t, "Hotel C", result[2].Hotel.Name)
}

func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}
