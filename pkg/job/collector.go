package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/mathif92/prices-recommender/pkg/collector"
	"github.com/mathif92/prices-recommender/pkg/recommendations"
	"github.com/mathif92/prices-recommender/pkg/types"
)

const (
	HotelsParamsSettingKey = "collect_hotels_params"
	CollectDatesSettingKey = "collect_dates"
	NotificationEmailKey   = "notification_email"
)

type collectHotelsSettings struct {
	Locations     []string `json:"locations"`
	Adults        int      `json:"adults"`
	Children      int      `json:"children"`
	ChildrenAges  string   `json:"children_ages"`
	PropertyTypes string   `json:"property_types"`
	Amenities     string   `json:"amenities"`
}

type collectDateRange struct {
	Name     string `json:"name"`
	CheckIn  string `json:"check_in"`
	CheckOut string `json:"check_out"`
}

type collectDatesSettings struct {
	Dates []collectDateRange `json:"dates"`
}

type Collector struct {
	log       logrus.FieldLogger
	db        *sqlx.DB
	collector collector.Collector
	notifier  *recommendations.Notifier
}

func NewCollector(log logrus.FieldLogger, db *sqlx.DB, collector collector.Collector, notifier *recommendations.Notifier) *Collector {
	return &Collector{
		log:       log,
		db:        db,
		collector: collector,
		notifier:  notifier,
	}
}

func (c *Collector) Run(ctx context.Context, userID int64) error {
	var hotelsRaw string
	err := c.db.GetContext(ctx, &hotelsRaw, "SELECT setting_value FROM user_settings WHERE setting_key = $1 AND user_id = $2 LIMIT 1", HotelsParamsSettingKey, userID)
	if err != nil {
		return fmt.Errorf("failed to read hotels params setting: %w", err)
	}

	var hotelsSettings collectHotelsSettings
	if err := json.Unmarshal([]byte(hotelsRaw), &hotelsSettings); err != nil {
		return fmt.Errorf("failed to unmarshal hotels params: %w", err)
	}

	var datesRaw string
	err = c.db.GetContext(ctx, &datesRaw, "SELECT setting_value FROM user_settings WHERE setting_key = $1 AND user_id = $2 LIMIT 1", CollectDatesSettingKey, userID)
	if err != nil {
		return fmt.Errorf("failed to read collect dates setting: %w", err)
	}

	var datesSettings collectDatesSettings
	if err := json.Unmarshal([]byte(datesRaw), &datesSettings); err != nil {
		return fmt.Errorf("failed to unmarshal collect dates: %w", err)
	}

	for _, dr := range datesSettings.Dates {
		checkIn, err := time.Parse("2006-01-02", dr.CheckIn)
		if err != nil {
			c.log.Errorf("bad check_in date %q for %q: %v", dr.CheckIn, dr.Name, err)
			continue
		}
		checkOut, err := time.Parse("2006-01-02", dr.CheckOut)
		if err != nil {
			c.log.Errorf("bad check_out date %q for %q: %v", dr.CheckOut, dr.Name, err)
			continue
		}

		for _, loc := range hotelsSettings.Locations {
			if loc == "" {
				continue
			}

			params := types.CollectParams{
				Type:          types.RecommendationTypeHotel,
				Destination:   loc,
				CheckIn:       checkIn,
				CheckOut:      checkOut,
				Adults:        hotelsSettings.Adults,
				Children:      hotelsSettings.Children,
				ChildrenAges:  hotelsSettings.ChildrenAges,
				PropertyTypes: hotelsSettings.PropertyTypes,
				Amenities:     hotelsSettings.Amenities,
			}

			data, err := c.collector.Collect(ctx, params)
			if err != nil {
				c.log.Errorf("failed to collect for %q / %q: %v", dr.Name, loc, err)
				time.Sleep(2 * time.Second)
				continue
			}
			c.log.Infof("collected %d hotels for %q / %q (%s - %s)", len(data), dr.Name, loc, dr.CheckIn, dr.CheckOut)
			time.Sleep(2 * time.Second)
		}
	}

	drops := c.collector.Drops()
	if len(drops) > 0 {
		c.sendPriceDropAlert(ctx, userID, drops)
	}

	return nil
}

func (c *Collector) sendPriceDropAlert(ctx context.Context, userID int64, drops []recommendations.PriceDrop) {
	c.log.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	c.log.Infof("  PRICE DROPS DETECTED: %d hotel%s", len(drops), plural(len(drops)))
	c.log.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, d := range drops {
		c.log.Infof("  %d. %s", i+1, d.HotelName)
		c.log.Infof("     Location:  %s", d.Location)
		c.log.Infof("     Dates:     %s → %s", d.StartDate.Format("Jan 2, 2006"), d.EndDate.Format("Jan 2, 2006"))
		c.log.Infof("     Price:     %.2f %s → %.2f %s (−%.0f%%)", d.OldPrice, d.Currency, d.NewPrice, d.Currency, d.DropRatio*100)
		if i < len(drops)-1 {
			c.log.Info("     ─────────────────────────────────────────")
		}
	}

	c.log.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	var emailRaw string
	err := c.db.GetContext(ctx, &emailRaw, "SELECT setting_value FROM user_settings WHERE setting_key = $1 AND user_id = $2 LIMIT 1", NotificationEmailKey, userID)
	if err != nil || emailRaw == "" {
		c.log.Warn("no notification_email setting found, skipping email")
		return
	}

	if err := c.notifier.SendPriceDropAlertIfConfigured(emailRaw, drops); err != nil {
		c.log.Errorf("failed to send price drop alert: %v", err)
		return
	}

	c.log.Infof("sent price drop alert to %s", emailRaw)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
