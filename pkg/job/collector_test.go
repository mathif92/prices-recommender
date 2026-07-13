package job

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathif92/prices-recommender/internal/dal/testhelpers"
	"github.com/mathif92/prices-recommender/pkg/recommendations"
	"github.com/mathif92/prices-recommender/pkg/types"
)

type recordingCollector struct {
	mu    sync.Mutex
	calls []types.CollectParams
}

func (c *recordingCollector) Collect(_ context.Context, params types.CollectParams) ([]types.HotelData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, params)
	return nil, nil
}

func (c *recordingCollector) Name() string {
	return "recording"
}

func (c *recordingCollector) Drops() []recommendations.PriceDrop {
	return nil
}

func (c *recordingCollector) Params() []types.CollectParams {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]types.CollectParams, len(c.calls))
	copy(result, c.calls)
	return result
}

var sharedJobDB *sqlx.DB

func TestMain(m *testing.M) {
	db, cleanup, err := testhelpers.StartTestDB()
	if err != nil {
		panic("failed to start test database: " + err.Error())
	}
	sharedJobDB = db
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func getJobTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	testhelpers.TruncateTables(t, sharedJobDB)
	return sharedJobDB
}

func seedUserSettings(t *testing.T, db *sqlx.DB) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO users (id, email, password_hash, password_salt)
		VALUES (1, 'test@example.com', 'hash', 'salt')
		ON CONFLICT (email) DO NOTHING
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO user_settings (user_id, setting_key, setting_value)
		VALUES (1, 'collect_hotels_params', '{"locations":["Cancun","Riviera Maya"],"adults":2,"children":0}')
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO user_settings (user_id, setting_key, setting_value)
		VALUES (1, 'collect_dates', '{"dates":[{"name":"Summer","check_in":"2027-06-01","check_out":"2027-06-10"},{"name":"Winter","check_in":"2027-12-20","check_out":"2028-01-05"}]}')
	`)
	require.NoError(t, err)
}

func discardLogger() logrus.FieldLogger {
	log := logrus.New()
	log.SetOutput(noopWriter{})
	return log
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestRunSuccessfulCollection(t *testing.T) {
	db := getJobTestDB(t)
	ctx := context.Background()

	seedUserSettings(t, db)

	recorder := &recordingCollector{}
	log := discardLogger()

	jc := NewCollector(log, db, recorder, nil)
	err := jc.Run(ctx)
	require.NoError(t, err)

	params := recorder.Params()
	require.Len(t, params, 4)

	assert.Equal(t, "Cancun", params[0].Destination)
	assert.Equal(t, types.RecommendationTypeHotel, params[0].Type)
	assert.Equal(t, "2027-06-01", params[0].CheckIn.Format("2006-01-02"))
	assert.Equal(t, "2027-06-10", params[0].CheckOut.Format("2006-01-02"))
	assert.Equal(t, 2, params[0].Adults)

	assert.Equal(t, "Riviera Maya", params[1].Destination)
	assert.Equal(t, "2027-06-10", params[1].CheckOut.Format("2006-01-02"))

	assert.Equal(t, "Cancun", params[2].Destination)
	assert.Equal(t, "2027-12-20", params[2].CheckIn.Format("2006-01-02"))
	assert.Equal(t, "2028-01-05", params[2].CheckOut.Format("2006-01-02"))

	assert.Equal(t, "Riviera Maya", params[3].Destination)
	assert.Equal(t, "2027-12-20", params[3].CheckIn.Format("2006-01-02"))
	assert.Equal(t, "2028-01-05", params[3].CheckOut.Format("2006-01-02"))
}

func TestRunNoSettings(t *testing.T) {
	db := getJobTestDB(t)
	ctx := context.Background()

	recorder := &recordingCollector{}
	log := discardLogger()

	jc := NewCollector(log, db, recorder, nil)
	err := jc.Run(ctx)
	assert.Error(t, err)
	assert.Empty(t, recorder.Params())
}

func TestRunWithPartialSettings(t *testing.T) {
	db := getJobTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(`
		INSERT INTO users (id, email, password_hash, password_salt)
		VALUES (1, 'partial@example.com', 'hash', 'salt')
		ON CONFLICT (email) DO NOTHING
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO user_settings (user_id, setting_key, setting_value)
		VALUES (1, 'collect_hotels_params', '{"locations":["Paris"],"adults":1,"children":0}')
	`)
	require.NoError(t, err)

	recorder := &recordingCollector{}
	log := discardLogger()

	jc := NewCollector(log, db, recorder, nil)
	err = jc.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collect dates")
	assert.Empty(t, recorder.Params())
}
