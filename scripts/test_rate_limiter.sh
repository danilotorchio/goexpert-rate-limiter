#!/bin/bash

# Test script for rate limiter demonstration
echo "🚀 Rate Limiter Test Script"
echo "=========================="

# Check if server is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Server is not running on port 8080"
    echo "Start the server with: make run or make docker-up"
    exit 1
fi

echo "✅ Server is running"
echo ""

# Test 1: IP-based rate limiting
echo "📋 Test 1: IP-based Rate Limiting"
echo "Making 15 requests without token..."
for i in {1..15}; do
    response=$(curl -s -w "%{http_code}" -o /tmp/response http://localhost:8080/)
    status_code=${response: -3}
    
    if [ $status_code -eq 200 ]; then
        echo "  Request $i: ✅ Allowed (200)"
    elif [ $status_code -eq 429 ]; then
        echo "  Request $i: ❌ Rate limited (429)"
        remaining=$(curl -s -I http://localhost:8080/ | grep -i "x-ratelimit-remaining" | cut -d' ' -f2 | tr -d '\r')
        echo "    Rate limit exceeded! Remaining: $remaining"
        break
    else
        echo "  Request $i: ⚠️  Unexpected status ($status_code)"
    fi
    
    sleep 0.1
done

echo ""
echo "⏱️  Waiting 2 seconds..."
sleep 2

# Test 2: Token-based rate limiting
echo ""
echo "📋 Test 2: Token-based Rate Limiting"
echo "Making requests with token 'abc123' (limit: 50 req/s)..."

for i in {1..60}; do
    response=$(curl -s -w "%{http_code}" -H "API_KEY: abc123" -o /tmp/response http://localhost:8080/)
    status_code=${response: -3}
    
    if [ $status_code -eq 200 ]; then
        echo "  Request $i: ✅ Allowed (200)"
    elif [ $status_code -eq 429 ]; then
        echo "  Request $i: ❌ Rate limited (429)"
        echo "    Token rate limit exceeded!"
        break
    else
        echo "  Request $i: ⚠️  Unexpected status ($status_code)"
    fi
    
    if [ $((i % 10)) -eq 0 ]; then
        echo "    ... ($i requests made)"
    fi
    
    sleep 0.02  # Faster requests to test rate limiting
done

echo ""
echo "⏱️  Waiting 2 seconds..."
sleep 2

# Test 3: Token priority over IP
echo ""
echo "📋 Test 3: Token Priority Over IP"
echo "First, exhaust IP limit..."

# Exhaust IP limit
for i in {1..12}; do
    curl -s http://localhost:8080/ > /dev/null
done

echo "IP should be blocked now. Testing with token..."

# Test with token (should work)
response=$(curl -s -w "%{http_code}" -H "API_KEY: test_token" -o /tmp/response http://localhost:8080/)
status_code=${response: -3}

if [ $status_code -eq 200 ]; then
    echo "  ✅ Token request allowed despite IP being blocked!"
    echo "  This demonstrates token priority over IP"
elif [ $status_code -eq 429 ]; then
    echo "  ❌ Token request blocked (unexpected)"
else
    echo "  ⚠️  Unexpected status ($status_code)"
fi

# Test 4: Different endpoints
echo ""
echo "📋 Test 4: Testing Different Endpoints"

endpoints=("/" "/health" "/api/data")

for endpoint in "${endpoints[@]}"; do
    response=$(curl -s -w "%{http_code}" -o /tmp/response http://localhost:8080$endpoint)
    status_code=${response: -3}
    
    if [ $status_code -eq 200 ]; then
        echo "  $endpoint: ✅ Accessible"
    elif [ $status_code -eq 429 ]; then
        echo "  $endpoint: ❌ Rate limited"
    else
        echo "  $endpoint: ⚠️  Status $status_code"
    fi
done

echo ""
echo "🎉 Test completed!"
echo "Check the application logs for detailed information"
echo ""
echo "💡 Pro tip: You can monitor rate limits using headers:"
echo "   curl -I http://localhost:8080/ | grep -i ratelimit"
echo ""
echo "📊 To see current Redis data:"
echo "   docker-compose exec redis redis-cli"
echo "   > KEYS *"
echo "   > GET rate_limit:ip:127.0.0.1" 