package repositories

import "time"

type Hotel struct {
	ID             int64      `db:"id" json:"id"`
	Name           string     `db:"name" json:"name"`
	Description    *string    `db:"description" json:"description"`
	Link           string     `db:"link" json:"link"`
	Rating         float64    `db:"rating" json:"rating"`
	Location       string     `db:"location" json:"location"`
	IsAllInclusive bool       `db:"is_all_inclusive" json:"is_all_inclusive"`
	CreatedAt      *time.Time `db:"created_at" json:"created_at"`
}

type HotelRating struct {
	ID          int64      `db:"id" json:"id"`
	HotelID     int64      `db:"hotel_id" json:"hotel_id"`
	Rating      float64    `db:"rating" json:"rating"`
	ReviewCount int        `db:"review_count" json:"review_count"`
	CreatedAt   *time.Time `db:"created_at" json:"created_at"`
}

type HotelReview struct {
	ID           int64      `db:"id" json:"id"`
	HotelID      int64      `db:"hotel_id" json:"hotel_id"`
	Name         string     `db:"name" json:"name"`
	Positive     int        `db:"positive" json:"positive"`
	Negative     int        `db:"negative" json:"negative"`
	Neutral      int        `db:"neutral" json:"neutral"`
	ExternalLink string     `db:"external_link" json:"external_link"`
	CreatedAt    *time.Time `db:"created_at" json:"created_at"`
}

type Price struct {
	ID                  int64      `db:"id" json:"id"`
	HotelID             int64      `db:"hotel_id" json:"hotel_id"`
	Price               float64    `db:"price" json:"price"`
	StartDate           time.Time  `db:"start_date" json:"start_date"`
	EndDate             time.Time  `db:"end_date" json:"end_date"`
	Currency            string     `db:"currency" json:"currency"`
	CheckinTime         *string    `db:"checkin_time" json:"checkin_time"`
	CheckoutTime        *string    `db:"checkout_time" json:"checkout_time"`
	PropertyDetailsLink *string    `db:"property_details_link" json:"property_details_link"`
	CreatedAt           *time.Time `db:"created_at" json:"created_at"`
}

type User struct {
	ID           int64      `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash *string    `db:"password_hash" json:"-"`
	PasswordSalt *string    `db:"password_salt" json:"-"`
	GoogleID     *string    `db:"google_id" json:"-"`
	DisplayName  *string    `db:"display_name" json:"display_name"`
	AvatarURL    *string    `db:"avatar_url" json:"avatar_url"`
	CreatedAt    *time.Time `db:"created_at" json:"created_at"`
}

type Vacation struct {
	ID             int64      `db:"id" json:"id"`
	Name           string     `db:"name" json:"name"`
	StartDate      time.Time  `db:"start_date" json:"start_date"`
	EndDate        time.Time  `db:"end_date" json:"end_date"`
	AffectedLevels string     `db:"affected_levels" json:"affected_levels"`
	Year           int        `db:"year" json:"year"`
	CreatedAt      *time.Time `db:"created_at" json:"created_at"`
}

type UserSettings struct {
	ID           int64      `db:"id" json:"id"`
	UserID       int64      `db:"user_id" json:"user_id"`
	SettingKey   string     `db:"setting_key" json:"setting_key"`
	SettingValue string     `db:"setting_value" json:"setting_value"`
	CreatedAt    *time.Time `db:"created_at" json:"created_at"`
}

type HotelWithPrices struct {
	Hotel
	MinPrice      float64 `db:"min_price" json:"min_price"`
	MaxPrice      float64 `db:"max_price" json:"max_price"`
	PriceCurrency string  `db:"price_currency" json:"price_currency"`
	PriceCount    int     `db:"price_count" json:"price_count"`
}
