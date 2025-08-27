# Plausible Analytics Go Clone

A lightweight, privacy-focused web analytics server built in Go with MotherDuck (DuckDB) backend.

## Features

✅ **Implemented:**
- Event collection API with full analytics data model
- MotherDuck/DuckDB integration for high-performance analytics
- User agent parsing (browser, OS, device detection)
- Session tracking and IP geolocation
- UTM parameter tracking
- Custom event properties
- Basic dashboard stats API
- CORS support for web tracking
- CLI with serve and migrate commands

🚧 **In Progress:**
- Enhanced analytics queries (bounce rate, session duration, etc.)
- Realtime visitor tracking
- Goal conversion tracking

## Quick Start

### 1. Install Dependencies
```bash
go mod tidy
```

### 2. Setup Database
For **local DuckDB**:
```bash
go run . migrate
```

For **MotherDuck** (cloud):
```bash
export MOTHERDUCK_TOKEN=your_token_here
go run . migrate
```

### 3. Start Server
```bash
go run . serve
```

### 4. Track Events
Include the tracking script in your website:
```html
<script src="http://localhost:8080/js/script.js"></script>
```

Or send events manually:
```bash
./event.sh
```

## API Endpoints

- `POST /api/event` - Track events
- `GET /api/stats?domain=example.com&period=7d` - Get analytics stats
- `GET /health` - Health check

## Environment Variables

- `MOTHERDUCK_TOKEN` - MotherDuck authentication token (optional, uses local DuckDB if not set)

## Architecture

- **Database**: MotherDuck (DuckDB) for fast analytical queries
- **Backend**: Go with Gin framework
- **Tracking**: Lightweight JavaScript client
- **CLI**: urfave/cli for database management

## Development

Test the API:
```bash
chmod +x event.sh
./event.sh
```

## Deployment

Build for production:
```bash
go build -o plausible-analytics .
```
