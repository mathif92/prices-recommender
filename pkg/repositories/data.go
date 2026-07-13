package repositories

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type DataRepository interface {
	SaveHotels(ctx context.Context, hotels []Hotel) ([]Hotel, error)
	SaveHotelRatings(ctx context.Context, ratings []HotelRating) ([]HotelRating, error)
	SaveHotelReviews(ctx context.Context, reviews []HotelReview) ([]HotelReview, error)
	SavePrices(ctx context.Context, prices []Price) ([]Price, error)
	SaveUser(ctx context.Context, user User) (User, error)
	SaveUserSettings(ctx context.Context, settings UserSettings) (UserSettings, error)
	SaveVacations(ctx context.Context, vacations []Vacation) ([]Vacation, error)
	GetVacations(ctx context.Context, year int) ([]Vacation, error)
	DeleteVacationsByYear(ctx context.Context, year int) error
	ListHotels(ctx context.Context) ([]Hotel, error)
	GetHotelByID(ctx context.Context, id int64) (*Hotel, error)
	SearchHotelsByLocation(ctx context.Context, location string) ([]Hotel, error)
	ListHotelRatings(ctx context.Context, hotelID int64) ([]HotelRating, error)
	ListHotelReviews(ctx context.Context, hotelID int64) ([]HotelReview, error)
	ListPrices(ctx context.Context, hotelID int64) ([]Price, error)
	ListHotelsPaginated(ctx context.Context, limit, offset int) ([]Hotel, error)
	CountHotels(ctx context.Context) (int, error)
	SearchHotelsByLocationPaginated(ctx context.Context, location string, limit, offset int) ([]Hotel, error)
	CountHotelsByLocation(ctx context.Context, location string) (int, error)
	ListHotelsWithPrices(ctx context.Context, limit, offset int, location string) ([]HotelWithPrices, int, error)
	ListUserSettings(ctx context.Context, userID int64) ([]UserSettings, error)
	GetUserSetting(ctx context.Context, userID int64, key string) (*UserSettings, error)
	DeleteUserSetting(ctx context.Context, userID int64, key string) error
	GetCurrentPrice(ctx context.Context, hotelID int64, startDate, endDate time.Time) (*Price, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*User, error)
	UpdateUser(ctx context.Context, user User) error
}

type dataRepository struct {
	readDB  *sqlx.DB
	writeDB *sqlx.DB
}

func NewDataRepository(readDB, writeDB *sqlx.DB) DataRepository {
	return &dataRepository{
		readDB:  readDB,
		writeDB: writeDB,
	}
}

func (r *dataRepository) SaveHotels(ctx context.Context, hotels []Hotel) ([]Hotel, error) {
	if len(hotels) == 0 {
		return hotels, nil
	}

	query := dbQueries[insertHotels] + " RETURNING id"
	stmt, err := r.writeDB.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for i, hotel := range hotels {
		var id int64
		if err := stmt.QueryRowxContext(ctx, map[string]any{
			"name":             hotel.Name,
			"description":      hotel.Description,
			"link":             hotel.Link,
			"rating":           hotel.Rating,
			"location":         hotel.Location,
			"is_all_inclusive": hotel.IsAllInclusive,
			"created_at":       hotel.CreatedAt,
		}).Scan(&id); err != nil {
			return nil, err
		}
		hotels[i].ID = id
	}

	return hotels, nil
}

func (r *dataRepository) SaveHotelRatings(ctx context.Context, ratings []HotelRating) ([]HotelRating, error) {
	if len(ratings) == 0 {
		return ratings, nil
	}

	query := dbQueries[insertHotelRatings] + " RETURNING id"
	stmt, err := r.writeDB.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for i, rating := range ratings {
		var id int64
		if err := stmt.QueryRowxContext(ctx, map[string]any{
			"hotel_id":     rating.HotelID,
			"rating":       rating.Rating,
			"review_count": rating.ReviewCount,
			"created_at":   rating.CreatedAt,
		}).Scan(&id); err != nil {
			return nil, err
		}
		ratings[i].ID = id
	}

	return ratings, nil
}

func (r *dataRepository) SaveHotelReviews(ctx context.Context, reviews []HotelReview) ([]HotelReview, error) {
	if len(reviews) == 0 {
		return reviews, nil
	}

	query := dbQueries[insertHotelReviews] + " RETURNING id"
	stmt, err := r.writeDB.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for i, review := range reviews {
		var id int64
		if err := stmt.QueryRowxContext(ctx, map[string]any{
			"hotel_id":      review.HotelID,
			"name":          review.Name,
			"positive":      review.Positive,
			"negative":      review.Negative,
			"neutral":       review.Neutral,
			"external_link": review.ExternalLink,
			"created_at":    review.CreatedAt,
		}).Scan(&id); err != nil {
			return nil, err
		}
		reviews[i].ID = id
	}

	return reviews, nil
}

func (r *dataRepository) SavePrices(ctx context.Context, prices []Price) ([]Price, error) {
	if len(prices) == 0 {
		return prices, nil
	}

	query := dbQueries[insertPrices] + " RETURNING id"
	stmt, err := r.writeDB.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for i, price := range prices {
		var id int64
		if err := stmt.QueryRowxContext(ctx, map[string]any{
			"hotel_id":               price.HotelID,
			"price":                  price.Price,
			"start_date":             price.StartDate,
			"end_date":               price.EndDate,
			"currency":               price.Currency,
			"checkin_time":           price.CheckinTime,
			"checkout_time":          price.CheckoutTime,
			"property_details_link":  price.PropertyDetailsLink,
			"created_at":             price.CreatedAt,
		}).Scan(&id); err != nil {
			return nil, err
		}
		prices[i].ID = id
	}

	return prices, nil
}

func (r *dataRepository) SaveUser(ctx context.Context, user User) (User, error) {
	query := dbQueries[insertUser] + " RETURNING id"
	stmt, err := r.writeDB.PrepareNamedContext(ctx, query)
	if err != nil {
		return user, err
	}
	defer stmt.Close()

	var id int64
	if err := stmt.QueryRowxContext(ctx, map[string]any{
		"email":         user.Email,
		"password_hash": user.PasswordHash,
		"password_salt": user.PasswordSalt,
		"google_id":     user.GoogleID,
		"display_name":  user.DisplayName,
		"avatar_url":    user.AvatarURL,
		"created_at":    user.CreatedAt,
	}).Scan(&id); err != nil {
		return user, err
	}
	user.ID = id
	return user, nil
}

func (r *dataRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.readDB.GetContext(ctx, &user, dbQueries[getUserByEmail], email); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *dataRepository) GetUserByID(ctx context.Context, id int64) (*User, error) {
	var user User
	if err := r.readDB.GetContext(ctx, &user, dbQueries[getUserByID], id); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *dataRepository) GetUserByGoogleID(ctx context.Context, googleID string) (*User, error) {
	var user User
	if err := r.readDB.GetContext(ctx, &user, dbQueries[getUserByGoogleID], googleID); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *dataRepository) UpdateUser(ctx context.Context, user User) error {
	_, err := r.writeDB.NamedExecContext(ctx, dbQueries[updateUser], user)
	return err
}

func (r *dataRepository) SaveUserSettings(ctx context.Context, settings UserSettings) (UserSettings, error) {
	query := dbQueries[insertUserSettings] + " RETURNING id"
	stmt, err := r.writeDB.PrepareNamedContext(ctx, query)
	if err != nil {
		return settings, err
	}
	defer stmt.Close()

	var id int64
	if err := stmt.QueryRowxContext(ctx, map[string]any{
		"user_id":       settings.UserID,
		"setting_key":   settings.SettingKey,
		"setting_value": settings.SettingValue,
		"created_at":    settings.CreatedAt,
	}).Scan(&id); err != nil {
		return settings, err
	}
	settings.ID = id
	return settings, nil
}

func (r *dataRepository) SaveVacations(ctx context.Context, vacations []Vacation) ([]Vacation, error) {
	if len(vacations) == 0 {
		return vacations, nil
	}

	query := dbQueries[insertVacations] + " RETURNING id"
	stmt, err := r.writeDB.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for i, v := range vacations {
		var id int64
		if err := stmt.QueryRowxContext(ctx, map[string]any{
			"name":            v.Name,
			"start_date":      v.StartDate,
			"end_date":        v.EndDate,
			"affected_levels": v.AffectedLevels,
			"year":            v.Year,
			"created_at":      v.CreatedAt,
		}).Scan(&id); err != nil {
			return nil, err
		}
		vacations[i].ID = id
	}

	return vacations, nil
}

func (r *dataRepository) GetVacations(ctx context.Context, year int) ([]Vacation, error) {
	var vacations []Vacation
	if err := r.readDB.SelectContext(ctx, &vacations, dbQueries[listVacations], year); err != nil {
		return nil, err
	}
	return vacations, nil
}

func (r *dataRepository) DeleteVacationsByYear(ctx context.Context, year int) error {
	_, err := r.writeDB.ExecContext(ctx, dbQueries[deleteVacationsByYear], year)
	return err
}

func (r *dataRepository) ListHotels(ctx context.Context) ([]Hotel, error) {
	var hotels []Hotel
	if err := r.readDB.SelectContext(ctx, &hotels, dbQueries[listHotels]); err != nil {
		return nil, err
	}
	return hotels, nil
}

func (r *dataRepository) GetHotelByID(ctx context.Context, id int64) (*Hotel, error) {
	var hotel Hotel
	if err := r.readDB.GetContext(ctx, &hotel, dbQueries[getHotelByID], id); err != nil {
		return nil, err
	}
	return &hotel, nil
}

func (r *dataRepository) SearchHotelsByLocation(ctx context.Context, location string) ([]Hotel, error) {
	var hotels []Hotel
	if err := r.readDB.SelectContext(ctx, &hotels, dbQueries[listHotelsByLocation], location); err != nil {
		return nil, err
	}
	return hotels, nil
}

func (r *dataRepository) ListHotelRatings(ctx context.Context, hotelID int64) ([]HotelRating, error) {
	var ratings []HotelRating
	if err := r.readDB.SelectContext(ctx, &ratings, dbQueries[listHotelRatings], hotelID); err != nil {
		return nil, err
	}
	return ratings, nil
}

func (r *dataRepository) ListHotelReviews(ctx context.Context, hotelID int64) ([]HotelReview, error) {
	var reviews []HotelReview
	if err := r.readDB.SelectContext(ctx, &reviews, dbQueries[listHotelReviews], hotelID); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *dataRepository) ListPrices(ctx context.Context, hotelID int64) ([]Price, error) {
	var prices []Price
	if err := r.readDB.SelectContext(ctx, &prices, dbQueries[listPrices], hotelID); err != nil {
		return nil, err
	}
	return prices, nil
}

func (r *dataRepository) ListHotelsPaginated(ctx context.Context, limit, offset int) ([]Hotel, error) {
	var hotels []Hotel
	if err := r.readDB.SelectContext(ctx, &hotels, dbQueries[listHotelsPaginated], limit, offset); err != nil {
		return nil, err
	}
	return hotels, nil
}

func (r *dataRepository) CountHotels(ctx context.Context) (int, error) {
	var count int
	if err := r.readDB.GetContext(ctx, &count, dbQueries[countHotels]); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *dataRepository) SearchHotelsByLocationPaginated(ctx context.Context, location string, limit, offset int) ([]Hotel, error) {
	var hotels []Hotel
	if err := r.readDB.SelectContext(ctx, &hotels, dbQueries[listHotelsByLocationPaginated], location, limit, offset); err != nil {
		return nil, err
	}
	return hotels, nil
}

func (r *dataRepository) CountHotelsByLocation(ctx context.Context, location string) (int, error) {
	var count int
	if err := r.readDB.GetContext(ctx, &count, dbQueries[countHotelsByLocation], location); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *dataRepository) ListHotelsWithPrices(ctx context.Context, limit, offset int, location string) ([]HotelWithPrices, int, error) {
	var total int
	var hotels []HotelWithPrices

	if location != "" {
		if err := r.readDB.GetContext(ctx, &total, dbQueries[countHotelsWithPricesByLocation], location); err != nil {
			return nil, 0, err
		}
		if err := r.readDB.SelectContext(ctx, &hotels, dbQueries[listHotelsWithPricesByLocationPaginated], location, limit, offset); err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.readDB.GetContext(ctx, &total, dbQueries[countHotelsWithPrices]); err != nil {
			return nil, 0, err
		}
		if err := r.readDB.SelectContext(ctx, &hotels, dbQueries[listHotelsWithPricesPaginated], limit, offset); err != nil {
			return nil, 0, err
		}
	}

	return hotels, total, nil
}

func (r *dataRepository) ListUserSettings(ctx context.Context, userID int64) ([]UserSettings, error) {
	var settings []UserSettings
	if err := r.readDB.SelectContext(ctx, &settings, dbQueries[listUserSettings], userID); err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *dataRepository) GetUserSetting(ctx context.Context, userID int64, key string) (*UserSettings, error) {
	var setting UserSettings
	if err := r.readDB.GetContext(ctx, &setting, dbQueries[getUserSetting], userID, key); err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *dataRepository) DeleteUserSetting(ctx context.Context, userID int64, key string) error {
	_, err := r.writeDB.ExecContext(ctx, dbQueries[deleteUserSetting], userID, key)
	return err
}

func (r *dataRepository) GetCurrentPrice(ctx context.Context, hotelID int64, startDate, endDate time.Time) (*Price, error) {
	var price Price
	if err := r.readDB.GetContext(ctx, &price, dbQueries[getCurrentPrice], hotelID, startDate, endDate); err != nil {
		return nil, err
	}
	return &price, nil
}
