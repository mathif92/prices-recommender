package types

import "time"

type HotelData struct {
	Hotel   Hotel
	Price   HotelPrice
	Rating  HotelRating
	Reviews []HotelReview
}

type HotelPrice struct {
	ID                   int64      `json:"id"`
	Price                float64    `json:"price"`
	StartDate            time.Time  `json:"start_date"`
	EndDate              time.Time  `json:"end_date"`
	Currency             string     `json:"currency"`
	CheckinTime          *string    `json:"checkin_time"`
	CheckoutTime         *string    `json:"checkout_time"`
	PropertyDetailsLink  *string    `json:"property_details_link"`
	CreatedAt        *time.Time `json:"created_at"`
}

type Hotel struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Description    *string    `json:"description"`
	Link           string     `json:"link"`
	Rating         float64    `json:"rating"`
	Location       string     `json:"location"`
	IsAllInclusive bool       `json:"is_all_inclusive"`
	CreatedAt      *time.Time `json:"created_at"`
}

type HotelRating struct {
	ID          int64      `json:"id"`
	Rating      float64    `json:"rating"`
	ReviewCount int        `json:"review_count"`
	CreatedAt   *time.Time `json:"created_at"`
}

type HotelReview struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Positive     int        `json:"positive"`
	Negative     int        `json:"negative"`
	Neutral      int        `json:"neutral"`
	ExternalLink string     `json:"external_link"`
	CreatedAt    *time.Time `json:"created_at"`
}
