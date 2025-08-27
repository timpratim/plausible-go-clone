package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GeoLocation struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
}

type GeoService struct {
	client *http.Client
}

func NewGeoService() *GeoService {
	return &GeoService{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetLocationFromIP uses a free IP geolocation service
// In production, consider using MaxMind GeoIP2 or similar paid service
func (g *GeoService) GetLocationFromIP(ip string) (*GeoLocation, error) {
	// Skip private/local IPs
	if ip == "127.0.0.1" || ip == "::1" || ip == "" {
		return &GeoLocation{
			Country: "Unknown",
			Region:  "Unknown",
			City:    "Unknown",
		}, nil
	}

	// Using ipapi.co free service (100 requests/day limit)
	url := fmt.Sprintf("https://ipapi.co/%s/json/", ip)
	
	resp, err := g.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &GeoLocation{
			Country: "Unknown",
			Region:  "Unknown",
			City:    "Unknown",
		}, nil
	}

	var result struct {
		Country string `json:"country_name"`
		Region  string `json:"region"`
		City    string `json:"city"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &GeoLocation{
		Country: result.Country,
		Region:  result.Region,
		City:    result.City,
	}, nil
}
