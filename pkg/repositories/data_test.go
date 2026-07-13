package repositories

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathif92/prices-recommender/internal/dal/testhelpers"
)

var sharedDB *sqlx.DB

func TestMain(m *testing.M) {
	db, cleanup, err := testhelpers.StartTestDB()
	if err != nil {
		panic("failed to start test database: " + err.Error())
	}
	sharedDB = db
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func getTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	testhelpers.TruncateTables(t, sharedDB)
	return sharedDB
}

func TestSaveAndGetVacations(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	vacations := []Vacation{
		{Name: "Summer", StartDate: parseDate("2027-06-01"), EndDate: parseDate("2027-06-15"), AffectedLevels: "all", Year: 2027},
		{Name: "Winter", StartDate: parseDate("2027-12-20"), EndDate: parseDate("2028-01-05"), AffectedLevels: "all", Year: 2027},
	}

	_, err := repo.SaveVacations(ctx, vacations)
	require.NoError(t, err)

	got, err := repo.GetVacations(ctx, 2027)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "Summer", got[0].Name)
	assert.Equal(t, "Winter", got[1].Name)
	assert.Equal(t, 2027, got[0].Year)
	assert.Equal(t, 2027, got[1].Year)
}

func TestGetVacationsEmptyYear(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	got, err := repo.GetVacations(ctx, 2025)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestDeleteVacationsByYear(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	vacations := []Vacation{
		{Name: "Verano", StartDate: parseDate("2027-01-16"), EndDate: parseDate("2027-01-26"), AffectedLevels: "all", Year: 2027},
		{Name: "Invierno", StartDate: parseDate("2027-06-26"), EndDate: parseDate("2027-07-06"), AffectedLevels: "all", Year: 2027},
	}
	_, err := repo.SaveVacations(ctx, vacations)
	require.NoError(t, err)

	err = repo.DeleteVacationsByYear(ctx, 2027)
	require.NoError(t, err)

	got, err := repo.GetVacations(ctx, 2027)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestSaveVacationsUpsert(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	v1 := []Vacation{
		{Name: "Trip", StartDate: parseDate("2027-06-01"), EndDate: parseDate("2027-06-10"), AffectedLevels: "low", Year: 2027},
	}
	_, err := repo.SaveVacations(ctx, v1)
	require.NoError(t, err)

	v2 := []Vacation{
		{Name: "Trip", StartDate: parseDate("2027-07-01"), EndDate: parseDate("2027-07-10"), AffectedLevels: "high", Year: 2027},
	}
	_, err = repo.SaveVacations(ctx, v2)
	require.NoError(t, err)

	got, err := repo.GetVacations(ctx, 2027)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.True(t, parseDate("2027-07-01").Equal(got[0].StartDate))
	assert.Equal(t, "high", got[0].AffectedLevels)
}

func TestSaveHotels(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotels := []Hotel{
		{Name: "Grand Resort", Link: "https://example.com/grand", Rating: 4.5, Location: "Cancun", IsAllInclusive: true, CreatedAt: &now},
		{Name: "Beach Hotel", Link: "https://example.com/beach", Rating: 3.8, Location: "Cancun", IsAllInclusive: false, CreatedAt: &now},
	}

	_, err := repo.SaveHotels(ctx, hotels)
	require.NoError(t, err)

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM hotels")
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	var names []string
	err = db.SelectContext(ctx, &names, "SELECT name FROM hotels ORDER BY name")
	require.NoError(t, err)
	assert.Equal(t, []string{"Beach Hotel", "Grand Resort"}, names)
}

func TestSaveHotelsUpsert(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	var err error
	now := time.Now()
	h1 := []Hotel{
		{Name: "Grand Resort", Link: "https://example.com/grand", Rating: 4.5, Location: "Cancun", IsAllInclusive: true, CreatedAt: &now},
	}
	_, err = repo.SaveHotels(ctx, h1)
	require.NoError(t, err)

	later := time.Now().Add(time.Hour)
	h2 := []Hotel{
		{Name: "Grand Resort", Link: "https://example.com/grand-v2", Rating: 4.7, Location: "Cancun", IsAllInclusive: true, CreatedAt: &later},
	}
	_, err = repo.SaveHotels(ctx, h2)
	require.NoError(t, err)

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM hotels")
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var rating float64
	err = db.GetContext(ctx, &rating, "SELECT rating FROM hotels WHERE name = 'Grand Resort'")
	require.NoError(t, err)
	assert.Equal(t, 4.7, rating)
}

func strPtr(s string) *string {
	return &s
}

func TestSaveUserAndSettings(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	user := User{
		Email:        "test@example.com",
		PasswordHash: strPtr("hash123"),
		PasswordSalt: strPtr("salt123"),
		CreatedAt:    &now,
	}
	_, err := repo.SaveUser(ctx, user)
	require.NoError(t, err)

	var userID int64
	err = db.GetContext(ctx, &userID, "SELECT id FROM users WHERE email = $1", "test@example.com")
	require.NoError(t, err)

	settings := UserSettings{
		UserID:       userID,
		SettingKey:   "collect_hotels_params",
		SettingValue: `{"locations":["Cancun"],"adults":2}`,
		CreatedAt:    &now,
	}
	_, err = repo.SaveUserSettings(ctx, settings)
	require.NoError(t, err)

	var value string
	err = db.GetContext(ctx, &value, "SELECT setting_value FROM user_settings WHERE user_id = $1 AND setting_key = $2", userID, "collect_hotels_params")
	require.NoError(t, err)
	assert.JSONEq(t, `{"locations":["Cancun"],"adults":2}`, value)
}

func TestSaveUserSettingsUpsert(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	user := User{Email: "upsert@example.com", PasswordHash: strPtr("hash"), PasswordSalt: strPtr("salt"), CreatedAt: &now}
	_, err := repo.SaveUser(ctx, user)
	require.NoError(t, err)

	var userID int64
	err = db.GetContext(ctx, &userID, "SELECT id FROM users WHERE email = $1", "upsert@example.com")
	require.NoError(t, err)

	s1 := UserSettings{UserID: userID, SettingKey: "theme", SettingValue: "dark", CreatedAt: &now}
	_, err = repo.SaveUserSettings(ctx, s1)
	require.NoError(t, err)

	s2 := UserSettings{UserID: userID, SettingKey: "theme", SettingValue: "light", CreatedAt: &now}
	_, err = repo.SaveUserSettings(ctx, s2)
	require.NoError(t, err)

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM user_settings WHERE user_id = $1 AND setting_key = 'theme'", userID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var val string
	err = db.GetContext(ctx, &val, "SELECT setting_value FROM user_settings WHERE user_id = $1 AND setting_key = 'theme'", userID)
	require.NoError(t, err)
	assert.Equal(t, "light", val)
}

func TestSaveHotelRatings(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotel := Hotel{Name: "Test Hotel", Link: "https://example.com", Rating: 4.0, Location: "Paris", CreatedAt: &now}
	_, err := repo.SaveHotels(ctx, []Hotel{hotel})
	require.NoError(t, err)

	var hotelID int64
	err = db.GetContext(ctx, &hotelID, "SELECT id FROM hotels WHERE name = 'Test Hotel'")
	require.NoError(t, err)

	ratings := []HotelRating{
		{HotelID: hotelID, Rating: 4.0, ReviewCount: 100, CreatedAt: &now},
		{HotelID: hotelID, Rating: 3.0, ReviewCount: 50, CreatedAt: &now},
	}
	_, err = repo.SaveHotelRatings(ctx, ratings)
	require.NoError(t, err)

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM hotel_ratings WHERE hotel_id = $1", hotelID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSaveHotelReviews(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotel := Hotel{Name: "Review Hotel", Link: "https://example.com", Rating: 4.0, Location: "NYC", CreatedAt: &now}
	_, err := repo.SaveHotels(ctx, []Hotel{hotel})
	require.NoError(t, err)

	var hotelID int64
	err = db.GetContext(ctx, &hotelID, "SELECT id FROM hotels WHERE name = 'Review Hotel'")
	require.NoError(t, err)

	reviews := []HotelReview{
		{HotelID: hotelID, Name: "Cleanliness", Positive: 80, Negative: 5, Neutral: 15, ExternalLink: "https://example.com/clean", CreatedAt: &now},
		{HotelID: hotelID, Name: "Location", Positive: 60, Negative: 20, Neutral: 20, ExternalLink: "https://example.com/loc", CreatedAt: &now},
	}
	_, err = repo.SaveHotelReviews(ctx, reviews)
	require.NoError(t, err)

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM hotel_reviews WHERE hotel_id = $1", hotelID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSavePrices(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotel := Hotel{Name: "Price Hotel", Link: "https://example.com", Rating: 4.0, Location: "Tokyo", CreatedAt: &now}
	_, err := repo.SaveHotels(ctx, []Hotel{hotel})
	require.NoError(t, err)

	var hotelID int64
	err = db.GetContext(ctx, &hotelID, "SELECT id FROM hotels WHERE name = 'Price Hotel'")
	require.NoError(t, err)

	prices := []Price{
		{HotelID: hotelID, Price: 250.00, StartDate: parseDate("2027-06-01"), EndDate: parseDate("2027-06-05"), Currency: "USD", CreatedAt: &now},
		{HotelID: hotelID, Price: 300.00, StartDate: parseDate("2027-07-01"), EndDate: parseDate("2027-07-05"), Currency: "USD", CreatedAt: &now},
	}
	_, err = repo.SavePrices(ctx, prices)
	require.NoError(t, err)

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM prices WHERE hotel_id = $1", hotelID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSavePricesInsertsMultipleRows(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotel := Hotel{Name: "Insert Hotel", Link: "https://example.com", Rating: 3.0, Location: "Rome", CreatedAt: &now}
	_, err := repo.SaveHotels(ctx, []Hotel{hotel})
	require.NoError(t, err)

	var hotelID int64
	err = db.GetContext(ctx, &hotelID, "SELECT id FROM hotels WHERE name = 'Insert Hotel'")
	require.NoError(t, err)

	p1 := []Price{
		{HotelID: hotelID, Price: 150.00, StartDate: parseDate("2027-06-01"), EndDate: parseDate("2027-06-05"), Currency: "USD", CreatedAt: &now},
	}
	_, err = repo.SavePrices(ctx, p1)
	require.NoError(t, err)

	p2 := []Price{
		{HotelID: hotelID, Price: 180.00, StartDate: parseDate("2027-06-01"), EndDate: parseDate("2027-06-05"), Currency: "EUR", CreatedAt: &now},
	}
	_, err = repo.SavePrices(ctx, p2)
	require.NoError(t, err)

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM prices WHERE hotel_id = $1", hotelID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "upsert should keep only one row per (hotel_id, start_date, end_date)")

	var currency string
	err = db.GetContext(ctx, &currency, "SELECT currency FROM prices WHERE hotel_id = $1", hotelID)
	require.NoError(t, err)
	assert.Equal(t, "EUR", currency, "second upsert should overwrite the first")
}

func TestForeignKeyViolation(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	ratings := []HotelRating{
		{HotelID: 999, Rating: 4.0, ReviewCount: 10, CreatedAt: &now},
	}
	_, err := repo.SaveHotelRatings(ctx, ratings)
	assert.Error(t, err)
}

func TestInsertUserDuplicateEmail(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	u1 := User{Email: "dup@example.com", PasswordHash: strPtr("a"), PasswordSalt: strPtr("b"), CreatedAt: &now}
	_, err := repo.SaveUser(ctx, u1)
	require.NoError(t, err)

	u2 := User{Email: "dup@example.com", PasswordHash: strPtr("c"), PasswordSalt: strPtr("d"), CreatedAt: &now}
	_, err = repo.SaveUser(ctx, u2)
	require.NoError(t, err)

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE email = 'dup@example.com'")
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var hash string
	err = db.GetContext(ctx, &hash, "SELECT password_hash FROM users WHERE email = 'dup@example.com'")
	require.NoError(t, err)
	assert.Equal(t, "c", hash)
}

func TestSaveHotelsEmptySlice(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	_, err := repo.SaveHotels(ctx, []Hotel{})
	require.NoError(t, err)

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM hotels")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSaveHotelRatingsEmptySlice(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	_, err := repo.SaveHotelRatings(ctx, []HotelRating{})
	require.NoError(t, err)
}

func TestSaveHotelReviewsEmptySlice(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	_, err := repo.SaveHotelReviews(ctx, []HotelReview{})
	require.NoError(t, err)
}

func TestSavePricesEmptySlice(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	_, err := repo.SavePrices(ctx, []Price{})
	require.NoError(t, err)
}

func TestSaveVacationsEmptySlice(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	_, err := repo.SaveVacations(ctx, []Vacation{})
	require.NoError(t, err)
}

func TestDeleteVacationsByYearNonExistent(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	err := repo.DeleteVacationsByYear(ctx, 9999)
	require.NoError(t, err)
}

func TestSaveHotelReviewsForeignKeyViolation(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	reviews := []HotelReview{
		{HotelID: 999, Name: "Cleanliness", Positive: 80, Negative: 5, Neutral: 15, CreatedAt: &now},
	}
	_, err := repo.SaveHotelReviews(ctx, reviews)
	assert.Error(t, err)
}

func TestSaveHotelNilCreatedAt(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	hotels := []Hotel{
		{Name: "No CreatedAt", Link: "https://example.com", Rating: 3.0, Location: "Berlin"},
	}
	_, err := repo.SaveHotels(ctx, hotels)
	require.NoError(t, err)

	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM hotels WHERE name = 'No CreatedAt'")
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestListHotels(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotels := []Hotel{
		{Name: "Alpha", Link: "https://a.com", Rating: 4.0, Location: "Paris", CreatedAt: &now},
		{Name: "Beta", Link: "https://b.com", Rating: 3.0, Location: "London", CreatedAt: &now},
	}
	_, err := repo.SaveHotels(ctx, hotels)
	require.NoError(t, err)

	got, err := repo.ListHotels(ctx)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "Alpha", got[0].Name)
	assert.Equal(t, "Beta", got[1].Name)
}

func TestGetHotelByID(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotels := []Hotel{
		{Name: "Target Hotel", Link: "https://t.com", Rating: 4.5, Location: "Tokyo", CreatedAt: &now},
	}
	saved, err := repo.SaveHotels(ctx, hotels)
	require.NoError(t, err)

	got, err := repo.GetHotelByID(ctx, saved[0].ID)
	require.NoError(t, err)
	assert.Equal(t, "Target Hotel", got.Name)
	assert.Equal(t, "Tokyo", got.Location)
}

func TestGetHotelByIDNotFound(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	_, err := repo.GetHotelByID(ctx, 99999)
	assert.Error(t, err)
}

func TestSearchHotelsByLocation(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotels := []Hotel{
		{Name: "A", Link: "https://a.com", Rating: 4.0, Location: "Cancun, Mexico", CreatedAt: &now},
		{Name: "B", Link: "https://b.com", Rating: 3.0, Location: "Paris, France", CreatedAt: &now},
		{Name: "C", Link: "https://c.com", Rating: 5.0, Location: "Cancun, Mexico", CreatedAt: &now},
	}
	_, err := repo.SaveHotels(ctx, hotels)
	require.NoError(t, err)

	got, err := repo.SearchHotelsByLocation(ctx, "Cancun")
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = repo.SearchHotelsByLocation(ctx, "Paris")
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestListHotelsEmpty(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	got, err := repo.ListHotels(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestListHotelRatings(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotels := []Hotel{{Name: "Rated Hotel", Link: "https://r.com", Rating: 4.0, Location: "NYC", CreatedAt: &now}}
	saved, err := repo.SaveHotels(ctx, hotels)
	require.NoError(t, err)

	ratings := []HotelRating{{HotelID: saved[0].ID, Rating: 4.0, ReviewCount: 100, CreatedAt: &now}}
	_, err = repo.SaveHotelRatings(ctx, ratings)
	require.NoError(t, err)

	got, err := repo.ListHotelRatings(ctx, saved[0].ID)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, 4.0, got[0].Rating)
}

func TestListHotelReviews(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotels := []Hotel{{Name: "Reviewed Hotel", Link: "https://rv.com", Rating: 4.0, Location: "LA", CreatedAt: &now}}
	saved, err := repo.SaveHotels(ctx, hotels)
	require.NoError(t, err)

	reviews := []HotelReview{
		{HotelID: saved[0].ID, Name: "Cleanliness", Positive: 90, Negative: 5, Neutral: 5, CreatedAt: &now},
	}
	_, err = repo.SaveHotelReviews(ctx, reviews)
	require.NoError(t, err)

	got, err := repo.ListHotelReviews(ctx, saved[0].ID)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "Cleanliness", got[0].Name)
}

func TestListPrices(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := NewDataRepository(db, db)

	now := time.Now()
	hotels := []Hotel{{Name: "Priced Hotel", Link: "https://p.com", Rating: 3.0, Location: "Rome", CreatedAt: &now}}
	saved, err := repo.SaveHotels(ctx, hotels)
	require.NoError(t, err)

	prices := []Price{
		{HotelID: saved[0].ID, Price: 200, StartDate: parseDate("2027-06-01"), EndDate: parseDate("2027-06-05"), Currency: "USD", CreatedAt: &now},
	}
	_, err = repo.SavePrices(ctx, prices)
	require.NoError(t, err)

	got, err := repo.ListPrices(ctx, saved[0].ID)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, 200.00, got[0].Price)
}

func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}
