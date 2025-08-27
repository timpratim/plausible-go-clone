package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/marcboeker/go-duckdb"
)

type Database struct {
	conn *sql.DB
}

func NewDatabase() (*Database, error) {
	// For MotherDuck, use the connection string format:
	// motherduck:?motherduck_token=your_token
	// For local DuckDB: just use a file path or ":memory:"
	
	dbPath := os.Getenv("MOTHERDUCK_TOKEN")
	var connStr string
	
	if dbPath != "" {
		// MotherDuck connection
		connStr = fmt.Sprintf("motherduck:?motherduck_token=%s", dbPath)
	} else {
		// Local DuckDB file
		connStr = "./analytics.duckdb"
	}

	db, err := sql.Open("duckdb", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{conn: db}
	
	// Initialize schema
	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

func (db *Database) initSchema() error {
	// Enable JSON extension
	if _, err := db.conn.Exec("INSTALL json"); err != nil {
		// Extension might already be installed, continue
	}
	if _, err := db.conn.Exec("LOAD json"); err != nil {
		return fmt.Errorf("failed to load JSON extension: %w", err)
	}

	// Create events table for page views and custom events
	createEventsTable := `
	CREATE TABLE IF NOT EXISTS events (
		id VARCHAR PRIMARY KEY,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		name VARCHAR NOT NULL,
		domain VARCHAR NOT NULL,
		url VARCHAR NOT NULL,
		referrer VARCHAR,
		user_agent VARCHAR,
		ip_address VARCHAR,
		country VARCHAR,
		region VARCHAR,
		city VARCHAR,
		browser VARCHAR,
		browser_version VARCHAR,
		os VARCHAR,
		os_version VARCHAR,
		device_type VARCHAR,
		screen_size VARCHAR,
		utm_source VARCHAR,
		utm_medium VARCHAR,
		utm_campaign VARCHAR,
		utm_term VARCHAR,
		utm_content VARCHAR,
		session_id VARCHAR,
		user_id VARCHAR,
		custom_properties JSON
	);`

	// Create sessions table for session tracking
	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR PRIMARY KEY,
		domain VARCHAR NOT NULL,
		user_id VARCHAR,
		start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		end_time TIMESTAMP,
		duration INTEGER,
		page_views INTEGER DEFAULT 0,
		bounce BOOLEAN DEFAULT true,
		entry_page VARCHAR,
		exit_page VARCHAR,
		referrer VARCHAR,
		utm_source VARCHAR,
		utm_medium VARCHAR,
		utm_campaign VARCHAR,
		country VARCHAR,
		region VARCHAR,
		city VARCHAR,
		browser VARCHAR,
		os VARCHAR,
		device_type VARCHAR
	);`

	// Create goals table for conversion tracking
	createGoalsTable := `
	CREATE TABLE IF NOT EXISTS goals (
		id VARCHAR PRIMARY KEY,
		domain VARCHAR NOT NULL,
		name VARCHAR NOT NULL,
		event_name VARCHAR,
		page_path VARCHAR,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(domain, name)
	);`

	tables := []string{createEventsTable, createSessionsTable, createGoalsTable}
	
	for _, query := range tables {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	log.Println("Database schema initialized successfully")
	return nil
}

func (db *Database) InsertEvent(event *AnalyticsEvent) error {
	query := `
	INSERT INTO events (
		id, name, domain, url, referrer, user_agent, ip_address,
		country, region, city, browser, browser_version, os, os_version,
		device_type, screen_size, utm_source, utm_medium, utm_campaign,
		utm_term, utm_content, session_id, user_id, custom_properties
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.conn.Exec(query,
		event.ID, event.Name, event.Domain, event.URL, event.Referrer,
		event.UserAgent, event.IPAddress, event.Country, event.Region,
		event.City, event.Browser, event.BrowserVersion, event.OS,
		event.OSVersion, event.DeviceType, event.ScreenSize,
		event.UTMSource, event.UTMMedium, event.UTMCampaign,
		event.UTMTerm, event.UTMContent, event.SessionID,
		event.UserID, event.CustomProperties,
	)

	return err
}

func (db *Database) GetPageViews(domain string, period string) ([]PageViewData, error) {
	var timeFilter string
	switch period {
	case "7d":
		timeFilter = "timestamp >= NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "timestamp >= NOW() - INTERVAL '30 days'"
	case "90d":
		timeFilter = "timestamp >= NOW() - INTERVAL '90 days'"
	default:
		timeFilter = "timestamp >= NOW() - INTERVAL '1 day'"
	}

	query := fmt.Sprintf(`
	SELECT 
		date_trunc('day', timestamp) as date,
		COUNT(*) as views,
		COUNT(DISTINCT session_id) as visitors
	FROM events 
	WHERE domain = ? AND name = 'pageview' AND %s
	GROUP BY date_trunc('day', timestamp)
	ORDER BY date`, timeFilter)

	rows, err := db.conn.Query(query, domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PageViewData
	for rows.Next() {
		var data PageViewData
		if err := rows.Scan(&data.Date, &data.Views, &data.Visitors); err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	return results, nil
}

func (db *Database) GetTopPages(domain string, period string, limit int) ([]TopPageData, error) {
	var timeFilter string
	switch period {
	case "7d":
		timeFilter = "timestamp >= NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "timestamp >= NOW() - INTERVAL '30 days'"
	case "90d":
		timeFilter = "timestamp >= NOW() - INTERVAL '90 days'"
	default:
		timeFilter = "timestamp >= NOW() - INTERVAL '1 day'"
	}

	query := fmt.Sprintf(`
	SELECT 
		url,
		COUNT(*) as views,
		COUNT(DISTINCT session_id) as visitors
	FROM events 
	WHERE domain = ? AND name = 'pageview' AND %s
	GROUP BY url
	ORDER BY views DESC
	LIMIT ?`, timeFilter)

	rows, err := db.conn.Query(query, domain, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TopPageData
	for rows.Next() {
		var data TopPageData
		if err := rows.Scan(&data.URL, &data.Views, &data.Visitors); err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	return results, nil
}

func (db *Database) GetSessionMetrics(domain string, period string) (float64, int, error) {
	var timeFilter string
	switch period {
	case "7d":
		timeFilter = "timestamp >= NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "timestamp >= NOW() - INTERVAL '30 days'"
	case "90d":
		timeFilter = "timestamp >= NOW() - INTERVAL '90 days'"
	default:
		timeFilter = "timestamp >= NOW() - INTERVAL '1 day'"
	}

	// Calculate bounce rate (sessions with only 1 page view)
	bounceQuery := fmt.Sprintf(`
	WITH session_views AS (
		SELECT session_id, COUNT(*) as page_views
		FROM events 
		WHERE domain = ? AND name = 'pageview' AND %s
		GROUP BY session_id
	)
	SELECT 
		COUNT(CASE WHEN page_views = 1 THEN 1 END) * 100.0 / COUNT(*) as bounce_rate
	FROM session_views`, timeFilter)

	var bounceRate float64
	err := db.conn.QueryRow(bounceQuery, domain).Scan(&bounceRate)
	if err != nil {
		bounceRate = 0
	}

	// Calculate average session duration (simplified)
	durationQuery := fmt.Sprintf(`
	WITH session_durations AS (
		SELECT 
			session_id,
			extract(epoch from (MAX(timestamp) - MIN(timestamp))) as duration_seconds
		FROM events 
		WHERE domain = ? AND name = 'pageview' AND %s
		GROUP BY session_id
		HAVING COUNT(*) > 1
	)
	SELECT AVG(duration_seconds) FROM session_durations`, timeFilter)

	var avgDuration float64
	err = db.conn.QueryRow(durationQuery, domain).Scan(&avgDuration)
	if err != nil {
		avgDuration = 0
	}

	return bounceRate, int(avgDuration), nil
}

func (db *Database) GetTopReferrers(domain string, period string, limit int) ([]ReferrerData, error) {
	var timeFilter string
	switch period {
	case "7d":
		timeFilter = "timestamp >= NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "timestamp >= NOW() - INTERVAL '30 days'"
	case "90d":
		timeFilter = "timestamp >= NOW() - INTERVAL '90 days'"
	default:
		timeFilter = "timestamp >= NOW() - INTERVAL '1 day'"
	}

	query := fmt.Sprintf(`
	SELECT 
		CASE 
			WHEN referrer = '' OR referrer IS NULL THEN 'Direct'
			ELSE referrer
		END as referrer,
		COUNT(*) as views,
		COUNT(DISTINCT session_id) as visitors
	FROM events 
	WHERE domain = ? AND name = 'pageview' AND %s
	GROUP BY referrer
	ORDER BY views DESC
	LIMIT ?`, timeFilter)

	rows, err := db.conn.Query(query, domain, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ReferrerData
	for rows.Next() {
		var data ReferrerData
		if err := rows.Scan(&data.Referrer, &data.Views, &data.Visitors); err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	return results, nil
}

func (db *Database) GetTopCountries(domain string, period string, limit int) ([]CountryData, error) {
	var timeFilter string
	switch period {
	case "7d":
		timeFilter = "timestamp >= NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "timestamp >= NOW() - INTERVAL '30 days'"
	case "90d":
		timeFilter = "timestamp >= NOW() - INTERVAL '90 days'"
	default:
		timeFilter = "timestamp >= NOW() - INTERVAL '1 day'"
	}

	query := fmt.Sprintf(`
	SELECT 
		COALESCE(country, 'Unknown') as country,
		COUNT(*) as views,
		COUNT(DISTINCT session_id) as visitors
	FROM events 
	WHERE domain = ? AND name = 'pageview' AND %s
	GROUP BY country
	ORDER BY views DESC
	LIMIT ?`, timeFilter)

	rows, err := db.conn.Query(query, domain, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CountryData
	for rows.Next() {
		var data CountryData
		if err := rows.Scan(&data.Country, &data.Views, &data.Visitors); err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	return results, nil
}

func (db *Database) GetTopBrowsers(domain string, period string, limit int) ([]BrowserData, error) {
	var timeFilter string
	switch period {
	case "7d":
		timeFilter = "timestamp >= NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "timestamp >= NOW() - INTERVAL '30 days'"
	case "90d":
		timeFilter = "timestamp >= NOW() - INTERVAL '90 days'"
	default:
		timeFilter = "timestamp >= NOW() - INTERVAL '1 day'"
	}

	query := fmt.Sprintf(`
	SELECT 
		COALESCE(browser, 'Unknown') as browser,
		COUNT(*) as views,
		COUNT(DISTINCT session_id) as visitors
	FROM events 
	WHERE domain = ? AND name = 'pageview' AND %s
	GROUP BY browser
	ORDER BY views DESC
	LIMIT ?`, timeFilter)

	rows, err := db.conn.Query(query, domain, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []BrowserData
	for rows.Next() {
		var data BrowserData
		if err := rows.Scan(&data.Browser, &data.Views, &data.Visitors); err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	return results, nil
}

func (db *Database) GetTopOS(domain string, period string, limit int) ([]OSData, error) {
	var timeFilter string
	switch period {
	case "7d":
		timeFilter = "timestamp >= NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "timestamp >= NOW() - INTERVAL '30 days'"
	case "90d":
		timeFilter = "timestamp >= NOW() - INTERVAL '90 days'"
	default:
		timeFilter = "timestamp >= NOW() - INTERVAL '1 day'"
	}

	query := fmt.Sprintf(`
	SELECT 
		COALESCE(os, 'Unknown') as os,
		COUNT(*) as views,
		COUNT(DISTINCT session_id) as visitors
	FROM events 
	WHERE domain = ? AND name = 'pageview' AND %s
	GROUP BY os
	ORDER BY views DESC
	LIMIT ?`, timeFilter)

	rows, err := db.conn.Query(query, domain, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []OSData
	for rows.Next() {
		var data OSData
		if err := rows.Scan(&data.OS, &data.Views, &data.Visitors); err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	return results, nil
}

func (db *Database) GetDeviceTypes(domain string, period string) ([]DeviceData, error) {
	var timeFilter string
	switch period {
	case "7d":
		timeFilter = "timestamp >= NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "timestamp >= NOW() - INTERVAL '30 days'"
	case "90d":
		timeFilter = "timestamp >= NOW() - INTERVAL '90 days'"
	default:
		timeFilter = "timestamp >= NOW() - INTERVAL '1 day'"
	}

	query := fmt.Sprintf(`
	SELECT 
		COALESCE(device_type, 'Unknown') as device_type,
		COUNT(*) as views,
		COUNT(DISTINCT session_id) as visitors
	FROM events 
	WHERE domain = ? AND name = 'pageview' AND %s
	GROUP BY device_type
	ORDER BY views DESC`, timeFilter)

	rows, err := db.conn.Query(query, domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []DeviceData
	for rows.Next() {
		var data DeviceData
		if err := rows.Scan(&data.DeviceType, &data.Views, &data.Visitors); err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	return results, nil
}

func (db *Database) GetRealtimeStats(domain string) (*RealtimeData, error) {
	// Get current visitors (last 5 minutes)
	currentVisitorsQuery := `
	SELECT COUNT(DISTINCT session_id) as current_visitors
	FROM events 
	WHERE domain = ? AND timestamp >= NOW() - INTERVAL '5 minutes'`

	var currentVisitors int
	err := db.conn.QueryRow(currentVisitorsQuery, domain).Scan(&currentVisitors)
	if err != nil {
		currentVisitors = 0
	}

	// Get top pages in last hour
	topPagesQuery := `
	SELECT 
		url,
		COUNT(*) as views,
		COUNT(DISTINCT session_id) as visitors
	FROM events 
	WHERE domain = ? AND name = 'pageview' AND timestamp >= NOW() - INTERVAL '1 hour'
	GROUP BY url
	ORDER BY views DESC
	LIMIT 10`

	rows, err := db.conn.Query(topPagesQuery, domain)
	if err != nil {
		return &RealtimeData{CurrentVisitors: currentVisitors}, nil
	}
	defer rows.Close()

	var topPages []TopPageData
	for rows.Next() {
		var data TopPageData
		if err := rows.Scan(&data.URL, &data.Views, &data.Visitors); err != nil {
			continue
		}
		topPages = append(topPages, data)
	}

	return &RealtimeData{
		CurrentVisitors: currentVisitors,
		TopPages:        topPages,
	}, nil
}

func (db *Database) CreateGoal(domain, name, eventName, pagePath string) (*Goal, error) {
	goal := &Goal{
		ID:        uuid.New().String(),
		Domain:    domain,
		Name:      name,
		EventName: eventName,
		PagePath:  pagePath,
		CreatedAt: time.Now(),
	}

	query := `
	INSERT INTO goals (id, domain, name, event_name, page_path, created_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	_, err := db.conn.Exec(query, goal.ID, goal.Domain, goal.Name, 
		goal.EventName, goal.PagePath, goal.CreatedAt)
	if err != nil {
		return nil, err
	}

	return goal, nil
}

func (db *Database) GetGoals(domain string) ([]Goal, error) {
	query := `
	SELECT id, domain, name, event_name, page_path, created_at
	FROM goals 
	WHERE domain = ?
	ORDER BY created_at DESC`

	rows, err := db.conn.Query(query, domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []Goal
	for rows.Next() {
		var goal Goal
		if err := rows.Scan(&goal.ID, &goal.Domain, &goal.Name, 
			&goal.EventName, &goal.PagePath, &goal.CreatedAt); err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}

	return goals, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}
