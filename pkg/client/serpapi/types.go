package serpapi

type GetHotelsResponse struct {
	SearchMetadata    SearchMetadata   `json:"search_metadata"`
	SearchParameters  SearchParameters `json:"search_parameters"`
	SearchInformation struct {
		TotalResults int `json:"total_results"`
	} `json:"search_information"`
	Brands            []Brand           `json:"brands"`
	Ads               []Ad              `json:"ads"`
	Properties        []Property        `json:"properties"`
	SerpapiPagination SerpapiPagination `json:"serpapi_pagination"`
}

type SearchMetadata struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	JSONEndpoint     string `json:"json_endpoint"`
	CreatedAt        string `json:"created_at"`
	ProcessedAt      string `json:"processed_at"`
	GoogleHotelsURL  string `json:"google_hotels_url"`
	RawHTMLFile      string `json:"raw_html_file"`
	PrettifyHTMLFile string `json:"prettify_html_file"`
	TotalTimeTaken   struct {
		Float float64 `json:"float"`
	} `json:"total_time_taken"`
}

type SearchParameters struct {
	Engine       string `json:"engine"`
	Q            string `json:"q"`
	GL           string `json:"gl"`
	HL           string `json:"hl"`
	Currency     string `json:"currency"`
	CheckInDate  string `json:"check_in_date"`
	CheckOutDate string `json:"check_out_date"`
	Adults       int    `json:"adults"`
	Children     int    `json:"children"`
}

type Brand struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Children []Brand `json:"children,omitempty"`
}

type GPSCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Ad struct {
	Name                       string         `json:"name"`
	Source                     string         `json:"source"`
	SourceIcon                 string         `json:"source_icon"`
	Link                       string         `json:"link"`
	PropertyToken              string         `json:"property_token"`
	SerpapiPropertyDetailsLink string         `json:"serpapi_property_details_link"`
	GPSCoordinates             GPSCoordinates `json:"gps_coordinates"`
	HotelClass                 int            `json:"hotel_class,omitempty"`
	Thumbnail                  string         `json:"thumbnail"`
	OverallRating              float64        `json:"overall_rating"`
	Reviews                    int            `json:"reviews"`
	Price                      string         `json:"price"`
	ExtractedPrice             int            `json:"extracted_price"`
	Amenities                  []string       `json:"amenities"`
}

type PriceValue struct {
	Lowest          string `json:"lowest"`
	ExtractedLowest int    `json:"extracted_lowest"`
}

type PriceOffer struct {
	Source       string     `json:"source"`
	Logo         string     `json:"logo"`
	NumGuests    int        `json:"num_guests,omitempty"`
	RatePerNight PriceValue `json:"rate_per_night"`
}

type Transportation struct {
	Type     string `json:"type"`
	Duration string `json:"duration"`
}

type NearbyPlace struct {
	Name            string           `json:"name"`
	Transportations []Transportation `json:"transportations,omitempty"`
}

type Image struct {
	Thumbnail     string `json:"thumbnail"`
	OriginalImage string `json:"original_image"`
}

type RatingCount struct {
	Stars int `json:"stars"`
	Count int `json:"count"`
}

type ReviewBreakdown struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	TotalMentioned int    `json:"total_mentioned"`
	Positive       int    `json:"positive"`
	Negative       int    `json:"negative"`
	Neutral        int    `json:"neutral"`
	CategoryToken  string `json:"category_token"`
	SerpapiLink    string `json:"serpapi_link"`
}

type HealthAndSafetyItem struct {
	Title     string `json:"title"`
	Available bool   `json:"available"`
}

type HealthAndSafetyGroup struct {
	Title string                `json:"title"`
	List  []HealthAndSafetyItem `json:"list"`
}

type HealthAndSafety struct {
	Groups []HealthAndSafetyGroup `json:"groups"`
}

type Property struct {
	Type                           string            `json:"type"`
	Name                           string            `json:"name"`
	Description                    string            `json:"description,omitempty"`
	Link                           string            `json:"link,omitempty"`
	PropertyToken                  string            `json:"property_token"`
	SerpapiPropertyDetailsLink     string            `json:"serpapi_property_details_link"`
	GPSCoordinates                 GPSCoordinates    `json:"gps_coordinates"`
	CheckInTime                    string            `json:"check_in_time,omitempty"`
	CheckOutTime                   string            `json:"check_out_time,omitempty"`
	RatePerNight                   PriceValue        `json:"rate_per_night"`
	TotalRate                      PriceValue        `json:"total_rate"`
	Prices                         []PriceOffer      `json:"prices,omitempty"`
	NearbyPlaces                   []NearbyPlace     `json:"nearby_places,omitempty"`
	HotelClass                     string            `json:"hotel_class,omitempty"`
	ExtractedHotelClass            int               `json:"extracted_hotel_class,omitempty"`
	Images                         []Image           `json:"images,omitempty"`
	OverallRating                  float64           `json:"overall_rating,omitempty"`
	Reviews                        int               `json:"reviews,omitempty"`
	Ratings                        []RatingCount     `json:"ratings,omitempty"`
	LocationRating                 float64           `json:"location_rating,omitempty"`
	ReviewsBreakdown               []ReviewBreakdown `json:"reviews_breakdown,omitempty"`
	Amenities                      []string          `json:"amenities,omitempty"`
	ExcludedAmenities              []string          `json:"excluded_amenities,omitempty"`
	EssentialInfo                  []string          `json:"essential_info,omitempty"`
	HealthAndSafety                HealthAndSafety   `json:"health_and_safety,omitempty"`
	EcoCertified                   bool              `json:"eco_certified,omitempty"`
	SerpapiGoogleHotelsReviewsLink string            `json:"serpapi_google_hotels_reviews_link,omitempty"`
	SerpapiGoogleHotelsPhotosLink  string            `json:"serpapi_google_hotels_photos_link,omitempty"`
}

type SerpapiPagination struct {
	CurrentFrom   int    `json:"current_from"`
	CurrentTo     int    `json:"current_to"`
	NextPageToken string `json:"next_page_token"`
	Next          string `json:"next"`
}

type GetFlightsResponse struct {
	SearchMetadata   SearchMetadata    `json:"search_metadata"`
	SearchParameters SearchParameters  `json:"search_parameters"`
	BestFlights      []FlightOffer     `json:"best_flights"`
	OtherFlights     []FlightOffer     `json:"other_flights"`
	PriceInsights    *PriceInsights    `json:"price_insights,omitempty"`
}

type FlightOffer struct {
	Flights        []Flight          `json:"flights"`
	Layovers       []Layover         `json:"layovers,omitempty"`
	TotalDuration  string            `json:"total_duration"`
	Price          int               `json:"price"`
	Type           string            `json:"type"`
	AirlineLogo    string            `json:"airline_logo"`
	BookingMetadata map[string]string `json:"booking_metadata,omitempty"`
	TravelClass    string            `json:"travel_class,omitempty"`
	Extensions     []string          `json:"extensions,omitempty"`
}

type Flight struct {
	DepartureAirport  Airport   `json:"departure_airport"`
	ArrivalAirport    Airport   `json:"arrival_airport"`
	DepartureTime     string    `json:"departure_time"`
	ArrivalTime       string    `json:"arrival_time"`
	Duration          string    `json:"duration"`
	Airplane          string    `json:"airplane,omitempty"`
	Airline           string    `json:"airline"`
	AirlineLogo       string    `json:"airline_logo"`
	TravelClass       string    `json:"travel_class,omitempty"`
	FlightNumber      int       `json:"flight_number"`
	Extensions        []string  `json:"extensions,omitempty"`
	TicketAlsoSoldBy  []string  `json:"ticket_also_sold_by,omitempty"`
}

type Airport struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Time string `json:"time"`
}

type Layover struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	Duration  string `json:"duration"`
	Overnight bool   `json:"overnight"`
}

type PriceInsights struct {
	LowestPrice      int    `json:"lowest_price"`
	PriceLevel       string `json:"price_level"`
	TypicalPriceRange []int `json:"typical_price_range,omitempty"`
}

type GetPackagesResponse struct {
	SearchMetadata   SearchMetadata     `json:"search_metadata"`
	SearchParameters SearchParameters   `json:"search_parameters"`
	Packages         []PackageOffer     `json:"packages,omitempty"`
}

type PackageOffer struct {
	HotelName       string            `json:"hotel_name,omitempty"`
	HotelRating     float64           `json:"hotel_rating,omitempty"`
	HotelStars      int               `json:"hotel_stars,omitempty"`
	Price           int               `json:"price"`
	PricePerPerson  int               `json:"price_per_person,omitempty"`
	TotalPrice      int               `json:"total_price,omitempty"`
	Flights         []Flight          `json:"flights,omitempty"`
	HotelLink       string            `json:"hotel_link,omitempty"`
	HotelImage      string            `json:"hotel_image,omitempty"`
	DetailsLink     string            `json:"details_link,omitempty"`
	Extensions      []string          `json:"extensions,omitempty"`
	BookingMetadata map[string]string `json:"booking_metadata,omitempty"`
}
