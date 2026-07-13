package recommendations

import (
	"context"
	"fmt"

	"github.com/mathif92/prices-recommender/pkg/repositories"
)

type DropDetector struct {
	repo      repositories.DataRepository
	threshold float64
}

func NewDropDetector(repo repositories.DataRepository, threshold float64) *DropDetector {
	return &DropDetector{repo: repo, threshold: threshold}
}

func (d *DropDetector) Detect(ctx context.Context, prices []repositories.Price) ([]PriceDrop, error) {
	if len(prices) == 0 {
		return nil, nil
	}

	var drops []PriceDrop
	seen := make(map[string]bool)

	for _, p := range prices {
		if p.Price <= 0 {
			continue
		}

		key := fmt.Sprintf("%d-%s-%s", p.HotelID, p.StartDate.Format("2006-01-02"), p.EndDate.Format("2006-01-02"))
		if seen[key] {
			continue
		}
		seen[key] = true

		prev, err := d.repo.GetCurrentPrice(ctx, p.HotelID, p.StartDate, p.EndDate)
		if err != nil {
			continue
		}
		if prev == nil || prev.Price <= 0 {
			continue
		}

		ratio := 1 - p.Price/prev.Price
		if ratio >= d.threshold {
			hotel, err := d.repo.GetHotelByID(ctx, p.HotelID)
			if err != nil {
				continue
			}

			drops = append(drops, PriceDrop{
				HotelID:   p.HotelID,
				HotelName: hotel.Name,
				Location:  hotel.Location,
				StartDate: p.StartDate,
				EndDate:   p.EndDate,
				OldPrice:  prev.Price,
				NewPrice:  p.Price,
				Currency:  p.Currency,
				DropRatio: ratio,
			})
		}
	}

	return drops, nil
}
