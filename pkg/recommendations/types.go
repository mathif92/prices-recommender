package recommendations

import (
	"time"
)

type PriceDrop struct {
	HotelID    int64
	HotelName  string
	Location   string
	StartDate  time.Time
	EndDate    time.Time
	OldPrice   float64
	NewPrice   float64
	Currency   string
	DropRatio  float64
}

type Config struct {
	SMTPServer string
	SMTPPort   string
	SMTPUser   string
	SMTPPass   string
	SMTPFrom   string
	DropThreshold float64
}

func DefaultConfig() Config {
	return Config{
		SMTPServer: "localhost",
		SMTPPort:   "587",
		DropThreshold: 0.10,
	}
}
