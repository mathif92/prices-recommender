package types

import (
	"time"
)

type RecommendationType string

const (
	RecommendationTypeFlight  RecommendationType = "flight"
	RecommendationTypeHotel   RecommendationType = "hotel"
	RecommendationTypePackage RecommendationType = "package"
)

type CollectParams struct {
	Type          RecommendationType
	Origin        string
	Destination   string
	CheckIn       time.Time
	CheckOut      time.Time
	Adults        int
	Children      int
	ChildrenAges  string
	PropertyTypes string
	Amenities     string
}
