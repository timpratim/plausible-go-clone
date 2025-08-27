package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AnalyticsService struct {
	db *Database
}

func NewAnalyticsService(db *Database) *AnalyticsService {
	return &AnalyticsService{db: db}
}

func (a *AnalyticsService) ProcessEvent(req *EventRequest, c *http.Request) (*AnalyticsEvent, error) {
	event := &AnalyticsEvent{
		ID:          uuid.New().String(),
		Timestamp:   time.Now(),
		Name:        req.Name,
		Domain:      req.Domain,
		URL:         req.URL,
		Referrer:    req.Referrer,
		UTMSource:   req.UTMSource,
		UTMMedium:   req.UTMMedium,
		UTMCampaign: req.UTMCampaign,
		UTMTerm:     req.UTMTerm,
		UTMContent:  req.UTMContent,
	}

	// Extract IP address
	event.IPAddress = a.getClientIP(c)
	
	// Extract User Agent
	event.UserAgent = c.Header.Get("User-Agent")
	
	// Parse User Agent for browser/OS info
	a.parseUserAgent(event)
	
	// Generate session ID (simplified - in production use proper session management)
	event.SessionID = a.generateSessionID(event.IPAddress, event.UserAgent)
	
	// Convert custom properties to JSON
	if req.CustomProperties != nil {
		customProps, err := json.Marshal(req.CustomProperties)
		if err == nil {
			event.CustomProperties = customProps
		}
	}

	// Store in database
	if err := a.db.InsertEvent(event); err != nil {
		return nil, err
	}

	return event, nil
}

func (a *AnalyticsService) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (a *AnalyticsService) parseUserAgent(event *AnalyticsEvent) {
	ua := event.UserAgent
	if ua == "" {
		return
	}

	// Simple user agent parsing (in production, use a proper UA parser library)
	ua = strings.ToLower(ua)

	// Browser detection
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		event.Browser = "Chrome"
	} else if strings.Contains(ua, "firefox") {
		event.Browser = "Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		event.Browser = "Safari"
	} else if strings.Contains(ua, "edg") {
		event.Browser = "Edge"
	} else {
		event.Browser = "Other"
	}

	// OS detection
	if strings.Contains(ua, "windows") {
		event.OS = "Windows"
	} else if strings.Contains(ua, "mac") {
		event.OS = "macOS"
	} else if strings.Contains(ua, "linux") {
		event.OS = "Linux"
	} else if strings.Contains(ua, "android") {
		event.OS = "Android"
	} else if strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		event.OS = "iOS"
	} else {
		event.OS = "Other"
	}

	// Device type detection
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		event.DeviceType = "Mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		event.DeviceType = "Tablet"
	} else {
		event.DeviceType = "Desktop"
	}
}

func (a *AnalyticsService) generateSessionID(ip, userAgent string) string {
	// Simple session ID generation based on IP and User Agent
	// In production, use proper session management with cookies
	data := ip + userAgent + time.Now().Format("2006-01-02-15")
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(data)).String()
}

func (a *AnalyticsService) GetDashboardStats(domain, period string) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Get page views data
	pageViews, err := a.db.GetPageViews(domain, period)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	for _, pv := range pageViews {
		stats.TotalViews += pv.Views
		stats.UniqueVisitors += pv.Visitors
	}

	// Get top pages
	topPages, err := a.db.GetTopPages(domain, period, 10)
	if err != nil {
		return nil, err
	}
	stats.TopPages = topPages

	// Get bounce rate and session time
	bounceRate, avgSessionTime, err := a.db.GetSessionMetrics(domain, period)
	if err == nil {
		stats.BounceRate = bounceRate
		stats.AvgSessionTime = avgSessionTime
	}

	// Get top referrers
	topReferrers, err := a.db.GetTopReferrers(domain, period, 10)
	if err == nil {
		stats.TopReferrers = topReferrers
	}

	// Get top countries
	topCountries, err := a.db.GetTopCountries(domain, period, 10)
	if err == nil {
		stats.TopCountries = topCountries
	}

	// Get top browsers
	topBrowsers, err := a.db.GetTopBrowsers(domain, period, 10)
	if err == nil {
		stats.TopBrowsers = topBrowsers
	}

	// Get top OS
	topOS, err := a.db.GetTopOS(domain, period, 10)
	if err == nil {
		stats.TopOS = topOS
	}

	// Get device types
	deviceTypes, err := a.db.GetDeviceTypes(domain, period)
	if err == nil {
		stats.DeviceTypes = deviceTypes
	}

	return stats, nil
}

func (a *AnalyticsService) GetRealtimeStats(domain string) (*RealtimeData, error) {
	return a.db.GetRealtimeStats(domain)
}

func (a *AnalyticsService) GetBreakdown(domain, period, breakdownType string) (interface{}, error) {
	switch breakdownType {
	case "pages":
		return a.db.GetTopPages(domain, period, 50)
	case "referrers":
		return a.db.GetTopReferrers(domain, period, 50)
	case "countries":
		return a.db.GetTopCountries(domain, period, 50)
	case "browsers":
		return a.db.GetTopBrowsers(domain, period, 50)
	case "os":
		return a.db.GetTopOS(domain, period, 50)
	case "devices":
		return a.db.GetDeviceTypes(domain, period)
	default:
		return nil, fmt.Errorf("unknown breakdown type: %s", breakdownType)
	}
}

func (a *AnalyticsService) CreateGoal(domain, name, eventName, pagePath string) (*Goal, error) {
	return a.db.CreateGoal(domain, name, eventName, pagePath)
}

func (a *AnalyticsService) GetGoals(domain string) ([]Goal, error) {
	return a.db.GetGoals(domain)
}
