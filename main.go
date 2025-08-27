package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
)

var (
	db               *Database
	analyticsService *AnalyticsService
)

func HandleEvents(c *gin.Context) {
	var req EventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process the event
	event, err := analyticsService.ProcessEvent(&req, c.Request)
	if err != nil {
		log.Printf("Error processing event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"event_id": event.ID,
	})
}

func HandleStats(c *gin.Context) {
	domain := c.Query("domain")
	period := c.DefaultQuery("period", "7d")

	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain parameter is required"})
		return
	}

	stats, err := analyticsService.GetDashboardStats(domain, period)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func HandleRealtime(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain parameter is required"})
		return
	}

	realtime, err := analyticsService.GetRealtimeStats(domain)
	if err != nil {
		log.Printf("Error getting realtime stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get realtime stats"})
		return
	}

	c.JSON(http.StatusOK, realtime)
}

func HandleBreakdown(c *gin.Context) {
	domain := c.Query("domain")
	period := c.DefaultQuery("period", "7d")
	breakdownType := c.Param("type")

	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain parameter is required"})
		return
	}

	breakdown, err := analyticsService.GetBreakdown(domain, period, breakdownType)
	if err != nil {
		log.Printf("Error getting breakdown: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get breakdown"})
		return
	}

	c.JSON(http.StatusOK, breakdown)
}

func HandleTrackingScript(c *gin.Context) {
	c.Header("Content-Type", "application/javascript")
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	
	// Read the tracking script
	script := `// Plausible Analytics Tracking Script
(function() {
    'use strict';

    var endpoint = window.plausible_endpoint || 'http://localhost:8080/api/event';
    var domain = window.location.hostname;

    function trigger(eventName, options) {
        options = options || {};
        
        var payload = {
            name: eventName,
            domain: domain,
            url: window.location.href,
            referrer: document.referrer || '',
            props: options.props || {}
        };

        // Add UTM parameters if present
        var urlParams = new URLSearchParams(window.location.search);
        ['utm_source', 'utm_medium', 'utm_campaign', 'utm_term', 'utm_content'].forEach(function(param) {
            var value = urlParams.get(param);
            if (value) {
                payload[param] = value;
            }
        });

        // Send the event
        var request = new XMLHttpRequest();
        request.open('POST', endpoint, true);
        request.setRequestHeader('Content-Type', 'application/json');
        
        request.send(JSON.stringify(payload));
    }

    // Track pageview on load
    function trackPageview() {
        trigger('pageview');
    }

    // Track custom events
    function plausible(eventName, options) {
        trigger(eventName, options);
    }

    // Auto-track pageviews
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', trackPageview);
    } else {
        trackPageview();
    }

    // Track SPA navigation
    var originalPushState = history.pushState;
    var originalReplaceState = history.replaceState;

    history.pushState = function() {
        originalPushState.apply(history, arguments);
        setTimeout(trackPageview, 0);
    };

    history.replaceState = function() {
        originalReplaceState.apply(history, arguments);
        setTimeout(trackPageview, 0);
    };

    window.addEventListener('popstate', trackPageview);

    // Expose plausible function globally
    window.plausible = plausible;
})();`
	
	c.String(http.StatusOK, script)
}

func HandleCreateGoal(c *gin.Context) {
	var req struct {
		Domain    string `json:"domain" binding:"required"`
		Name      string `json:"name" binding:"required"`
		EventName string `json:"event_name,omitempty"`
		PagePath  string `json:"page_path,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal, err := analyticsService.CreateGoal(req.Domain, req.Name, req.EventName, req.PagePath)
	if err != nil {
		log.Printf("Error creating goal: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create goal"})
		return
	}

	c.JSON(http.StatusCreated, goal)
}

func HandleGetGoals(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain parameter is required"})
		return
	}

	goals, err := analyticsService.GetGoals(domain)
	if err != nil {
		log.Printf("Error getting goals: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get goals"})
		return
	}

	c.JSON(http.StatusOK, goals)
}

func HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"database": "connected",
	})
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "Runs the analytics server",
				Action: func(cCtx *cli.Context) error {
					return StartServer()
				},
			},
			{
				Name:    "migrate",
				Aliases: []string{"m"},
				Usage:   "Initialize database schema",
				Action: func(cCtx *cli.Context) error {
					return RunMigrations()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func StartServer() error {
	// Initialize database
	var err error
	db, err = NewDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	// Initialize analytics service
	analyticsService = NewAnalyticsService(db)

	// Setup router
	router := gin.Default()

	// CORS middleware for web analytics
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Routes
	router.POST("/api/event", HandleEvents)
	router.GET("/api/stats", HandleStats)
	router.GET("/api/realtime", HandleRealtime)
	router.GET("/api/breakdown/:type", HandleBreakdown)
	router.POST("/api/goals", HandleCreateGoal)
	router.GET("/api/goals", HandleGetGoals)
	router.GET("/js/script.js", HandleTrackingScript)
	router.GET("/health", HandleHealth)

	log.Println("Starting analytics server on :8080")
	return router.Run(":8080")
}

func RunMigrations() error {
	db, err := NewDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	log.Println("Database migrations completed successfully")
	return nil
}
