package serpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mathif92/prices-recommender/pkg/types"
)

const (
	googleHotelsEngine  = "google_hotels"
	googleFlightsEngine = "google_flights"
	googleTravelEngine  = "google_travel"
)

func encodeParams(params map[string]string) string {
	v := url.Values{}
	for k, val := range params {
		v.Set(k, val)
	}
	return v.Encode()
}

// Client defines the interface for interacting with the SerpAPI to retrieve travel recommendations.
type Client interface {
	GetFlights(ctx context.Context, origin string, destination string, startDate time.Time, endDate time.Time) (*GetFlightsResponse, error)
	GetHotels(ctx context.Context, params types.CollectParams) (*GetHotelsResponse, error)
	GetPackages(ctx context.Context, origin string, destination string, startDate time.Time, endDate time.Time) (*GetPackagesResponse, error)
}

type client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new instance of the SerpAPI client with the provided base URL and API key.
func NewClient(baseURL, apiKey string) Client {
	return &client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetFlights retrieves flight recommendations based on the provided parameters.
func (c *client) GetFlights(
	ctx context.Context,
	origin string,
	destination string,
	startDate time.Time,
	endDate time.Time,
) (*GetFlightsResponse, error) {
	url := fmt.Sprintf(
		"%s?engine=%s&departure_id=%s&arrival_id=%s&outbound_date=%s&return_date=%s&api_key=%s",
		c.baseURL,
		googleFlightsEngine,
		origin,
		destination,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		c.apiKey,
	)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	var response GetFlightsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetHotels retrieves hotel recommendations based on the provided parameters.
func (c *client) GetHotels(
	ctx context.Context,
	params types.CollectParams,
) (*GetHotelsResponse, error) {
	q := encodeParams(map[string]string{
		"engine":         googleHotelsEngine,
		"q":              params.Destination,
		"check_in_date":  params.CheckIn.Format("2006-01-02"),
		"check_out_date": params.CheckOut.Format("2006-01-02"),
		"adults":         fmt.Sprintf("%d", params.Adults),
		"children":       fmt.Sprintf("%d", params.Children),
		"children_ages":  params.ChildrenAges,
		"property_types": params.PropertyTypes,
		"amenities":      params.Amenities,
		"api_key":        c.apiKey,
	})
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"?"+q, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	var response GetHotelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetPackages retrieves package recommendations based on the provided parameters.
func (c *client) GetPackages(
	ctx context.Context,
	origin string,
	destination string,
	startDate time.Time,
	endDate time.Time,
) (*GetPackagesResponse, error) {
	url := fmt.Sprintf(
		"%s?engine=%s&q=%s&departure_id=%s&check_in_date=%s&check_out_date=%s&api_key=%s",
		c.baseURL,
		googleTravelEngine,
		destination,
		origin,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		c.apiKey,
	)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	var response GetPackagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
