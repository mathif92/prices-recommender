package serpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	serpapi "github.com/mathif92/prices-recommender/pkg/client/serpapi"
	"github.com/mathif92/prices-recommender/internal/dal/testhelpers"
	"github.com/mathif92/prices-recommender/pkg/repositories"
	"github.com/mathif92/prices-recommender/pkg/types"
)

var sharedSerpDB *sqlx.DB

func TestMain(m *testing.M) {
	db, cleanup, err := testhelpers.StartTestDB()
	if err != nil {
		panic("failed to start test database: " + err.Error())
	}
	sharedSerpDB = db
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func getTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	testhelpers.TruncateTables(t, sharedSerpDB)
	return sharedSerpDB
}

func sampleHotelsResponse() serpapi.GetHotelsResponse {
	return serpapi.GetHotelsResponse{
		SearchMetadata: serpapi.SearchMetadata{
			ID:     "test-id",
			Status: "Success",
		},
		SearchInformation: struct {
			TotalResults int `json:"total_results"`
		}{TotalResults: 2},
		Properties: []serpapi.Property{
			{
				Type:        "hotel",
				Name:        "Grand Resort Cancun",
				Link:        "https://www.google.com/travel/hotel/grand",
				Description: "Beachfront luxury resort",
				SerpapiPropertyDetailsLink: "https://serpapi.com/details/1",
				OverallRating: 4.5,
				Reviews:       350,
				TotalRate: serpapi.PriceValue{
					ExtractedLowest: 45000,
				},
				CheckInTime:  "15:00",
				CheckOutTime: "12:00",
				Amenities:   []string{"pool", "all inclusive", "wifi"},
				ReviewsBreakdown: []serpapi.ReviewBreakdown{
					{
						Name:         "Cleanliness",
						Positive:     150,
						Negative:     10,
						Neutral:      40,
						SerpapiLink: "https://serpapi.com/reviews/1/clean",
					},
					{
						Name:         "Location",
						Positive:     200,
						Negative:     5,
						Neutral:      45,
						SerpapiLink: "https://serpapi.com/reviews/1/loc",
					},
				},
			},
			{
				Type:        "hotel",
				Name:        "Beach Hotel Cancun",
				Link:        "https://www.google.com/travel/hotel/beach",
				SerpapiPropertyDetailsLink: "https://serpapi.com/details/2",
				OverallRating: 3.8,
				Reviews:       120,
				TotalRate: serpapi.PriceValue{
					ExtractedLowest: 25000,
				},
				Amenities: []string{"pool", "wifi"},
			},
		},
	}
}

func TestCollectHotels(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()

	mockResp := sampleHotelsResponse()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.String(), "engine=google_hotels")
		assert.Contains(t, r.URL.String(), "Cancun")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	apiClient := serpapi.NewClient(server.URL, "fake-api-key")
	dataRepo := repositories.NewDataRepository(db, db)
	collector := NewCollector(apiClient, dataRepo)

	params := types.CollectParams{
		Type:        types.RecommendationTypeHotel,
		Destination: "Cancun",
		CheckIn:     parseTime("2027-06-01"),
		CheckOut:    parseTime("2027-06-10"),
		Adults:      2,
		Children:    0,
	}

	results, err := collector.Collect(ctx, params)
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.Equal(t, "Grand Resort Cancun", results[0].Hotel.Name)
	assert.Equal(t, "Beach Hotel Cancun", results[1].Hotel.Name)
	assert.Equal(t, "Cancun", results[0].Hotel.Location)
	assert.True(t, results[0].Hotel.IsAllInclusive)
	assert.False(t, results[1].Hotel.IsAllInclusive)

	var hotelCount int
	require.NoError(t, db.GetContext(ctx, &hotelCount, "SELECT COUNT(*) FROM hotels"))
	assert.Equal(t, 2, hotelCount)

	var ratingCount int
	require.NoError(t, db.GetContext(ctx, &ratingCount, "SELECT COUNT(*) FROM hotel_ratings"))
	assert.Equal(t, 2, ratingCount)

	var reviewCount int
	require.NoError(t, db.GetContext(ctx, &reviewCount, "SELECT COUNT(*) FROM hotel_reviews"))
	assert.Equal(t, 2, reviewCount)

	var priceCount int
	require.NoError(t, db.GetContext(ctx, &priceCount, "SELECT COUNT(*) FROM prices"))
	assert.Equal(t, 2, priceCount)

	var hotelName string
	require.NoError(t, db.GetContext(ctx, &hotelName, "SELECT name FROM hotels ORDER BY name LIMIT 1"))
	assert.Equal(t, "Beach Hotel Cancun", hotelName)
}

func TestCollectHotelsAPIError(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{"error":"internal error"}`)
	}))
	defer server.Close()

	apiClient := serpapi.NewClient(server.URL, "fake-api-key")
	dataRepo := repositories.NewDataRepository(db, db)
	collector := NewCollector(apiClient, dataRepo)

	params := types.CollectParams{
		Type:        types.RecommendationTypeHotel,
		Destination: "Nowhere",
		CheckIn:     parseTime("2027-06-01"),
		CheckOut:    parseTime("2027-06-10"),
	}

	_, err := collector.Collect(ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code")
}

func TestCollectUnsupportedType(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not be called")
	}))
	defer server.Close()

	apiClient := serpapi.NewClient(server.URL, "fake-api-key")
	dataRepo := repositories.NewDataRepository(db, db)
	collector := NewCollector(apiClient, dataRepo)

	params := types.CollectParams{
		Type:        types.RecommendationTypeFlight,
		Destination: "Paris",
		CheckIn:     parseTime("2027-06-01"),
		CheckOut:    parseTime("2027-06-10"),
	}

	_, err := collector.Collect(ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support")
}

func TestCollectHotelsSkipsZeroPrice(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()

	mockResp := serpapi.GetHotelsResponse{
		SearchMetadata: serpapi.SearchMetadata{ID: "zero-price", Status: "Success"},
		Properties: []serpapi.Property{
			{
				Type:          "hotel",
				Name:          "Zero Price Hotel",
				Link:          "https://example.com/zero",
				OverallRating: 4.0,
				Reviews:       100,
				TotalRate:     serpapi.PriceValue{ExtractedLowest: 0},
				CheckInTime:   "15:00",
				CheckOutTime:  "12:00",
			},
			{
				Type:    "hotel",
				Name:    "Real Priced Hotel",
				Link:    "https://example.com/real",
				OverallRating: 3.5,
				Reviews: 50,
				TotalRate: serpapi.PriceValue{ExtractedLowest: 20000},
			},
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	apiClient := serpapi.NewClient(server.URL, "fake-api-key")
	dataRepo := repositories.NewDataRepository(db, db)
	collector := NewCollector(apiClient, dataRepo)

	params := types.CollectParams{
		Type:        types.RecommendationTypeHotel,
		Destination: "Test",
		CheckIn:     parseTime("2027-06-01"),
		CheckOut:    parseTime("2027-06-10"),
	}

	results, err := collector.Collect(ctx, params)
	require.NoError(t, err)
	require.Len(t, results, 2)

	var hotelCount int
	require.NoError(t, db.GetContext(ctx, &hotelCount, "SELECT COUNT(*) FROM hotels"))
	assert.Equal(t, 2, hotelCount)

	var priceCount int
	require.NoError(t, db.GetContext(ctx, &priceCount, "SELECT COUNT(*) FROM prices"))
	assert.Equal(t, 1, priceCount)

	var zeroCount int
	require.NoError(t, db.GetContext(ctx, &zeroCount, "SELECT COUNT(*) FROM prices WHERE price = 0"))
	assert.Equal(t, 0, zeroCount)
}

func TestCollectHotelsEmptyResponse(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serpapi.GetHotelsResponse{
			SearchMetadata: serpapi.SearchMetadata{ID: "empty", Status: "Success"},
			Properties:     []serpapi.Property{},
		})
	}))
	defer server.Close()

	apiClient := serpapi.NewClient(server.URL, "fake-api-key")
	dataRepo := repositories.NewDataRepository(db, db)
	collector := NewCollector(apiClient, dataRepo)

	params := types.CollectParams{
		Type:        types.RecommendationTypeHotel,
		Destination: "Empty",
		CheckIn:     parseTime("2027-06-01"),
		CheckOut:    parseTime("2027-06-10"),
	}

	results, err := collector.Collect(ctx, params)
	require.NoError(t, err)
	assert.Empty(t, results)

	var count int
	require.NoError(t, db.GetContext(ctx, &count, "SELECT COUNT(*) FROM hotels"))
	assert.Equal(t, 0, count)
}
