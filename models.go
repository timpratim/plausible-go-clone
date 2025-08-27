package main

import (
	"encoding/json"
	"time"
)

type AnalyticsEvent struct {
	ID               string          `json:"id"`
	Timestamp        time.Time       `json:"timestamp"`
	Name             string          `json:"name"`
	Domain           string          `json:"domain"`
	URL              string          `json:"url"`
	Referrer         string          `json:"referrer,omitempty"`
	UserAgent        string          `json:"user_agent,omitempty"`
	IPAddress        string          `json:"ip_address,omitempty"`
	Country          string          `json:"country,omitempty"`
	Region           string          `json:"region,omitempty"`
	City             string          `json:"city,omitempty"`
	Browser          string          `json:"browser,omitempty"`
	BrowserVersion   string          `json:"browser_version,omitempty"`
	OS               string          `json:"os,omitempty"`
	OSVersion        string          `json:"os_version,omitempty"`
	DeviceType       string          `json:"device_type,omitempty"`
	ScreenSize       string          `json:"screen_size,omitempty"`
	UTMSource        string          `json:"utm_source,omitempty"`
	UTMMedium        string          `json:"utm_medium,omitempty"`
	UTMCampaign      string          `json:"utm_campaign,omitempty"`
	UTMTerm          string          `json:"utm_term,omitempty"`
	UTMContent       string          `json:"utm_content,omitempty"`
	SessionID        string          `json:"session_id,omitempty"`
	UserID           string          `json:"user_id,omitempty"`
	CustomProperties json.RawMessage `json:"custom_properties,omitempty"`
}

type PageViewData struct {
	Date     time.Time `json:"date"`
	Views    int       `json:"views"`
	Visitors int       `json:"visitors"`
}

type TopPageData struct {
	URL      string `json:"url"`
	Views    int    `json:"views"`
	Visitors int    `json:"visitors"`
}

type ReferrerData struct {
	Referrer string `json:"referrer"`
	Views    int    `json:"views"`
	Visitors int    `json:"visitors"`
}

type CountryData struct {
	Country  string `json:"country"`
	Views    int    `json:"views"`
	Visitors int    `json:"visitors"`
}

type BrowserData struct {
	Browser  string `json:"browser"`
	Views    int    `json:"views"`
	Visitors int    `json:"visitors"`
}

type OSData struct {
	OS       string `json:"os"`
	Views    int    `json:"views"`
	Visitors int    `json:"visitors"`
}

type DeviceData struct {
	DeviceType string `json:"device_type"`
	Views      int    `json:"views"`
	Visitors   int    `json:"visitors"`
}

type Goal struct {
	ID        string    `json:"id"`
	Domain    string    `json:"domain"`
	Name      string    `json:"name"`
	EventName string    `json:"event_name,omitempty"`
	PagePath  string    `json:"page_path,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ConversionData struct {
	Goal        string `json:"goal"`
	Conversions int    `json:"conversions"`
	Rate        float64 `json:"rate"`
}

type RealtimeData struct {
	CurrentVisitors int `json:"current_visitors"`
	TopPages        []TopPageData `json:"top_pages"`
}

type DashboardStats struct {
	TotalViews      int     `json:"total_views"`
	UniqueVisitors  int     `json:"unique_visitors"`
	BounceRate      float64 `json:"bounce_rate"`
	AvgSessionTime  int     `json:"avg_session_time"`
	TopPages        []TopPageData `json:"top_pages"`
	TopReferrers    []ReferrerData `json:"top_referrers"`
	TopCountries    []CountryData `json:"top_countries"`
	TopBrowsers     []BrowserData `json:"top_browsers"`
	TopOS           []OSData `json:"top_os"`
	DeviceTypes     []DeviceData `json:"device_types"`
	Conversions     []ConversionData `json:"conversions"`
}

// Event request payload for the API
type EventRequest struct {
	Name             string                 `json:"name" binding:"required"`
	Domain           string                 `json:"domain" binding:"required"`
	URL              string                 `json:"url" binding:"required"`
	Referrer         string                 `json:"referrer,omitempty"`
	UTMSource        string                 `json:"utm_source,omitempty"`
	UTMMedium        string                 `json:"utm_medium,omitempty"`
	UTMCampaign      string                 `json:"utm_campaign,omitempty"`
	UTMTerm          string                 `json:"utm_term,omitempty"`
	UTMContent       string                 `json:"utm_content,omitempty"`
	CustomProperties map[string]interface{} `json:"props,omitempty"`
}
