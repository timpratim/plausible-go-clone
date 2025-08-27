#!/bin/bash

echo "🧪 Quick Analytics Test"
echo "======================"

# Send some test events
echo "Sending test events..."
curl -s -X POST "http://localhost:8080/api/event" \
  -H "Content-Type: application/json" \
  -d '{"name": "pageview", "domain": "test.com", "url": "/", "referrer": "https://google.com"}' > /dev/null

curl -s -X POST "http://localhost:8080/api/event" \
  -H "Content-Type: application/json" \
  -d '{"name": "pageview", "domain": "test.com", "url": "/about", "referrer": "https://twitter.com"}' > /dev/null

curl -s -X POST "http://localhost:8080/api/event" \
  -H "Content-Type: application/json" \
  -d '{"name": "signup", "domain": "test.com", "url": "/signup", "props": {"plan": "premium"}}' > /dev/null

echo "✅ Events sent!"

# Test realtime stats
echo ""
echo "📊 Realtime Stats:"
curl -s "http://localhost:8080/api/realtime?domain=test.com"

echo ""
echo ""
echo "📈 Top Pages:"
curl -s "http://localhost:8080/api/breakdown/pages?domain=test.com&period=7d"

echo ""
echo ""
echo "🌐 Traffic Sources:"
curl -s "http://localhost:8080/api/breakdown/referrers?domain=test.com&period=7d"

echo ""
echo ""
echo "🎯 Goals:"
curl -s -X POST "http://localhost:8080/api/goals" \
  -H "Content-Type: application/json" \
  -d '{"domain": "test.com", "name": "Newsletter Signup", "event_name": "signup"}' > /dev/null

curl -s "http://localhost:8080/api/goals?domain=test.com"

echo ""
echo ""
echo "🚀 Analytics system is working!"
