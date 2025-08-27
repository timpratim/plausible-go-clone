#!/bin/bash

# Test pageview event
curl http://localhost:8080/api/event \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{
        "name": "pageview",
        "domain": "example.com",
        "url": "/",
        "referrer": "https://google.com",
        "utm_source": "google",
        "utm_medium": "organic"
    }'

echo -e "\n\n"

# Test custom event
curl http://localhost:8080/api/event \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{
        "name": "signup",
        "domain": "example.com",
        "url": "/signup",
        "props": {
            "plan": "premium",
            "source": "landing_page"
        }
    }'

echo -e "\n\n"

# Test stats endpoint
curl "http://localhost:8080/api/stats?domain=example.com&period=7d" \
    --include \
    --header "Content-Type: application/json"