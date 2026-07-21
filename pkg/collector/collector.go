package collector

import (
	"context"

	"github.com/mathif92/prices-recommender/pkg/recommendations"
	"github.com/mathif92/prices-recommender/pkg/types"
	"github.com/sirupsen/logrus"
)

type Collector interface {
	Collect(ctx context.Context, params types.CollectParams) ([]types.HotelData, error)
	Name() string
	Drops() []recommendations.PriceDrop
	ResetDrops()
}

type collector struct {
	log        logrus.FieldLogger
	collectors []Collector
}

func NewCollector(log logrus.FieldLogger, collectors ...Collector) Collector {
	return &collector{
		log:        log,
		collectors: collectors,
	}
}

func (c *collector) Collect(ctx context.Context, params types.CollectParams) ([]types.HotelData, error) {
	var all []types.HotelData
	for _, collector := range c.collectors {
		data, err := collector.Collect(ctx, params)
		if err != nil {
			c.log.Infof("Error collecting data for collector %s: %v", collector.Name(), err)
			continue
		}
		all = append(all, data...)
	}

	return all, nil
}

func (c *collector) Name() string {
	return "Main Collector"
}

func (c *collector) Drops() []recommendations.PriceDrop {
	var all []recommendations.PriceDrop
	for _, col := range c.collectors {
		all = append(all, col.Drops()...)
	}
	return all
}

func (c *collector) ResetDrops() {
	for _, col := range c.collectors {
		col.ResetDrops()
	}
}
