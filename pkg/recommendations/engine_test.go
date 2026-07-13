package recommendations

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathif92/prices-recommender/internal/dal/testhelpers"
	"github.com/mathif92/prices-recommender/pkg/repositories"
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

func pricePtr(p float64) *float64 {
	return &p
}

func TestDetectNoPreviousPrice(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := repositories.NewDataRepository(db, db)

	detector := NewDropDetector(repo, 0.10)

	now := time.Now()
	hotels, err := repo.SaveHotels(ctx, []repositories.Hotel{
		{Name: "Test Hotel", Link: "https://example.com", Location: "Test", CreatedAt: &now},
	})
	require.NoError(t, err)

	prices := []repositories.Price{
		{HotelID: hotels[0].ID, Price: 200, StartDate: now, EndDate: now.Add(7 * 24 * time.Hour), Currency: "USD", CreatedAt: &now},
	}

	drops, err := detector.Detect(ctx, prices)
	require.NoError(t, err)
	assert.Empty(t, drops)
}

func TestDetectPriceDropAboveThreshold(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := repositories.NewDataRepository(db, db)

	detector := NewDropDetector(repo, 0.10)

	now := time.Now()
	checkIn := time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC)
	checkOut := time.Date(2027, 6, 10, 0, 0, 0, 0, time.UTC)

	hotels, err := repo.SaveHotels(ctx, []repositories.Hotel{
		{Name: "Test Hotel", Link: "https://example.com", Location: "Test", CreatedAt: &now},
	})
	require.NoError(t, err)

	prevTime := now.Add(-24 * time.Hour)
	_, err = repo.SavePrices(ctx, []repositories.Price{
		{HotelID: hotels[0].ID, Price: 500, StartDate: checkIn, EndDate: checkOut, Currency: "USD", CreatedAt: &prevTime},
	})
	require.NoError(t, err)

	// Detect drop before saving the new price
	drops, err := detector.Detect(ctx, []repositories.Price{
		{HotelID: hotels[0].ID, Price: 400, StartDate: checkIn, EndDate: checkOut, Currency: "USD", CreatedAt: &now},
	})
	require.NoError(t, err)
	require.Len(t, drops, 1)
	assert.Equal(t, hotels[0].ID, drops[0].HotelID)
	assert.Equal(t, "Test Hotel", drops[0].HotelName)
	assert.Equal(t, 500.0, drops[0].OldPrice)
	assert.Equal(t, 400.0, drops[0].NewPrice)
	assert.Equal(t, "USD", drops[0].Currency)
	assert.InDelta(t, 0.20, drops[0].DropRatio, 0.01)
}

func TestDetectPriceDropBelowThreshold(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := repositories.NewDataRepository(db, db)

	detector := NewDropDetector(repo, 0.10)

	now := time.Now()
	checkIn := time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC)
	checkOut := time.Date(2027, 6, 10, 0, 0, 0, 0, time.UTC)

	hotels, err := repo.SaveHotels(ctx, []repositories.Hotel{
		{Name: "Test Hotel", Link: "https://example.com", Location: "Test", CreatedAt: &now},
	})
	require.NoError(t, err)

	prevTime := now.Add(-24 * time.Hour)
	_, err = repo.SavePrices(ctx, []repositories.Price{
		{HotelID: hotels[0].ID, Price: 500, StartDate: checkIn, EndDate: checkOut, Currency: "USD", CreatedAt: &prevTime},
	})
	require.NoError(t, err)

	prices := []repositories.Price{
		{HotelID: hotels[0].ID, Price: 470, StartDate: checkIn, EndDate: checkOut, Currency: "USD", CreatedAt: &now},
	}

	drops, err := detector.Detect(ctx, prices)
	require.NoError(t, err)
	assert.Empty(t, drops)
}

func TestDetectSkipsZeroPrice(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := repositories.NewDataRepository(db, db)

	detector := NewDropDetector(repo, 0.10)

	now := time.Now()
	checkIn := time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC)
	checkOut := time.Date(2027, 6, 10, 0, 0, 0, 0, time.UTC)

	hotels, err := repo.SaveHotels(ctx, []repositories.Hotel{
		{Name: "Test Hotel", Link: "https://example.com", Location: "Test", CreatedAt: &now},
	})
	require.NoError(t, err)

	prevTime := now.Add(-24 * time.Hour)
	_, err = repo.SavePrices(ctx, []repositories.Price{
		{HotelID: hotels[0].ID, Price: 500, StartDate: checkIn, EndDate: checkOut, Currency: "USD", CreatedAt: &prevTime},
	})
	require.NoError(t, err)

	prices := []repositories.Price{
		{HotelID: hotels[0].ID, Price: 0, StartDate: checkIn, EndDate: checkOut, Currency: "USD", CreatedAt: &now},
	}

	drops, err := detector.Detect(ctx, prices)
	require.NoError(t, err)
	assert.Empty(t, drops)
}

func TestDetectDeduplicatesByHotelAndDates(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()
	repo := repositories.NewDataRepository(db, db)

	detector := NewDropDetector(repo, 0.10)

	now := time.Now()
	checkIn := time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC)
	checkOut := time.Date(2027, 6, 10, 0, 0, 0, 0, time.UTC)

	hotels, err := repo.SaveHotels(ctx, []repositories.Hotel{
		{Name: "Test Hotel", Link: "https://example.com", Location: "Test", CreatedAt: &now},
	})
	require.NoError(t, err)

	prevTime := now.Add(-24 * time.Hour)
	_, err = repo.SavePrices(ctx, []repositories.Price{
		{HotelID: hotels[0].ID, Price: 500, StartDate: checkIn, EndDate: checkOut, Currency: "USD", CreatedAt: &prevTime},
	})
	require.NoError(t, err)

	drops, err := detector.Detect(ctx, []repositories.Price{
		{HotelID: hotels[0].ID, Price: 400, StartDate: checkIn, EndDate: checkOut, Currency: "USD", CreatedAt: &now},
		{HotelID: hotels[0].ID, Price: 400, StartDate: checkIn, EndDate: checkOut, Currency: "USD", CreatedAt: &now},
	})
	require.NoError(t, err)
	require.Len(t, drops, 1)
	assert.Equal(t, 400.0, drops[0].NewPrice)
}

func TestNotifierSkipsWhenNotConfigured(t *testing.T) {
	cfg := Config{SMTPServer: "", SMTPFrom: "test@example.com"}
	n := NewNotifier(cfg)

	err := n.SendPriceDropAlertIfConfigured("", []PriceDrop{
		{HotelName: "Test Hotel", Location: "Cancun", OldPrice: 500, NewPrice: 400, Currency: "USD", DropRatio: 0.20},
	})
	assert.NoError(t, err)

	err = n.SendPriceDropAlertIfConfigured("user@example.com", []PriceDrop{
		{HotelName: "Test Hotel", Location: "Cancun", OldPrice: 500, NewPrice: 400, Currency: "USD", DropRatio: 0.20},
	})
	assert.NoError(t, err)
}

func TestNotifierEmptyDrops(t *testing.T) {
	cfg := Config{SMTPServer: "smtp.example.com"}
	n := NewNotifier(cfg)

	err := n.SendPriceDropAlertIfConfigured("user@example.com", nil)
	assert.NoError(t, err)

	err = n.SendPriceDropAlertIfConfigured("user@example.com", []PriceDrop{})
	assert.NoError(t, err)
}
