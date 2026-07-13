package collector

import (
	"context"
	"errors"
	"testing"

	"github.com/mathif92/prices-recommender/pkg/recommendations"
	"github.com/mathif92/prices-recommender/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type fakeCollector struct {
	name    string
	data    []types.HotelData
	err     error
	called  bool
}

func (f *fakeCollector) Collect(_ context.Context, _ types.CollectParams) ([]types.HotelData, error) {
	f.called = true
	return f.data, f.err
}

func (f *fakeCollector) Name() string {
	return f.name
}

func (f *fakeCollector) Drops() []recommendations.PriceDrop {
	return nil
}

func discardLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(noopWriter{})
	return log
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestCompositeCollectorEmpty(t *testing.T) {
	log := discardLogger()

	c := NewCollector(log)
	result, err := c.Collect(context.Background(), types.CollectParams{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestCompositeCollectorSingle(t *testing.T) {
	log := discardLogger()

	data := []types.HotelData{
		{Hotel: types.Hotel{Name: "Test Hotel"}},
	}
	fake := &fakeCollector{name: "fake", data: data}
	c := NewCollector(log, fake)

	result, err := c.Collect(context.Background(), types.CollectParams{})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Test Hotel", result[0].Hotel.Name)
	assert.True(t, fake.called)
}

func TestCompositeCollectorAggregates(t *testing.T) {
	log := discardLogger()

	fake1 := &fakeCollector{
		name: "src1",
		data: []types.HotelData{
			{Hotel: types.Hotel{Name: "Hotel A"}},
		},
	}
	fake2 := &fakeCollector{
		name: "src2",
		data: []types.HotelData{
			{Hotel: types.Hotel{Name: "Hotel B"}},
			{Hotel: types.Hotel{Name: "Hotel C"}},
		},
	}
	c := NewCollector(log, fake1, fake2)

	result, err := c.Collect(context.Background(), types.CollectParams{})
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "Hotel A", result[0].Hotel.Name)
	assert.Equal(t, "Hotel B", result[1].Hotel.Name)
	assert.Equal(t, "Hotel C", result[2].Hotel.Name)
}

func TestCompositeCollectorSkipsError(t *testing.T) {
	log := discardLogger()

	fake1 := &fakeCollector{
		name: "failing",
		err:  errors.New("something went wrong"),
	}
	fake2 := &fakeCollector{
		name: "working",
		data: []types.HotelData{
			{Hotel: types.Hotel{Name: "Survivor Hotel"}},
		},
	}
	c := NewCollector(log, fake1, fake2)

	result, err := c.Collect(context.Background(), types.CollectParams{})
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Survivor Hotel", result[0].Hotel.Name)
	assert.True(t, fake1.called)
	assert.True(t, fake2.called)
}

func TestCompositeCollectorAllErrors(t *testing.T) {
	log := discardLogger()

	fake1 := &fakeCollector{
		name: "fail1",
		err:  errors.New("error 1"),
	}
	fake2 := &fakeCollector{
		name: "fail2",
		err:  errors.New("error 2"),
	}
	c := NewCollector(log, fake1, fake2)

	result, err := c.Collect(context.Background(), types.CollectParams{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}
