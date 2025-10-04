#!/bin/bash

echo "🔍 Testing Zitadel Advanced Functionality..."
echo "============================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print test result
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ PASS${NC}: $2"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAIL${NC}: $2"
        ((TESTS_FAILED++))
    fi
}

# Function to print info
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Wait for Zitadel to be fully ready
echo ""
echo "📋 Test 1: Wait for Zitadel full initialization"
print_info "Checking if Zitadel is ready..."

HEALTH_CHECK=$(curl -s http://localhost:8080/debug/ready 2>/dev/null)
if [ $? -eq 0 ]; then
    print_result 0 "Zitadel is ready and initialized"
else
    print_result 1 "Zitadel is not ready yet"
fi

echo ""
echo "📋 Test 2: Test Zitadel Console Access"
CONSOLE_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/ui/console)
if [ "$CONSOLE_RESPONSE" = "200" ]; then
    print_result 0 "Zitadel Console is accessible"
else
    print_result 1 "Zitadel Console access failed (HTTP: $CONSOLE_RESPONSE)"
fi

echo ""
echo "📋 Test 3: Test Management API"
# Test without authentication first
MGMT_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/management/v1/orgs/me)
if [ "$MGMT_RESPONSE" = "401" ] || [ "$MGMT_RESPONSE" = "403" ]; then
    print_result 0 "Management API is responding (authentication required as expected)"
elif [ "$MGMT_RESPONSE" = "200" ]; then
    print_result 0 "Management API is accessible"
else
    print_result 1 "Management API unexpected response (HTTP: $MGMT_RESPONSE)"
fi

echo ""
echo "📋 Test 4: Test OIDC Configuration Endpoint"
OIDC_CONFIG=$(curl -s http://localhost:8080/.well-known/openid_configuration 2>/dev/null)
if echo "$OIDC_CONFIG" | grep -q '"issuer"'; then
    print_result 0 "OIDC Configuration is available"
    ISSUER=$(echo "$OIDC_CONFIG" | grep -o '"issuer":"[^"]*"' | cut -d'"' -f4)
    print_info "OIDC Issuer: $ISSUER"
else
    # Check if it's a "not found" error which might mean Zitadel is still initializing
    if echo "$OIDC_CONFIG" | grep -q '"code":5'; then
        print_result 1 "OIDC Configuration not ready (Zitadel may still be initializing)"
    else
        print_result 1 "OIDC Configuration endpoint failed"
    fi
fi

echo ""
echo "📋 Test 5: Test Login Page"
LOGIN_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/ui/login)
if [ "$LOGIN_RESPONSE" = "200" ]; then
    print_result 0 "Login page is accessible"
else
    print_result 1 "Login page access failed (HTTP: $LOGIN_RESPONSE)"
fi

echo ""
echo "📋 Test 6: Test Database Schema"
DB_CONTAINER=$(docker ps --format "table {{.Names}}" | grep -E "(zitadel.*db|db.*zitadel)" | head -1)
if [ -n "$DB_CONTAINER" ]; then
    print_info "Using database container: $DB_CONTAINER"
    
    # Check if database exists
    DB_EXISTS=$(docker exec "$DB_CONTAINER" psql -U postgres -lqt | cut -d \| -f 1 | grep -w zitadel)
    if [ -n "$DB_EXISTS" ]; then
        print_result 0 "Zitadel database exists"
        
        # Check for some key tables
        TABLES=$(docker exec "$DB_CONTAINER" psql -U postgres -d zitadel -c "\dt" 2>/dev/null)
        if echo "$TABLES" | grep -q "eventstore\|projections\|system"; then
            print_result 0 "Zitadel core tables are present"
        else
            print_result 1 "Zitadel tables not found or still creating"
        fi
    else
        print_result 1 "Zitadel database not found"
    fi
else
    print_result 1 "Database container not found"
fi

echo ""
echo "📋 Test 7: Test Service Health Endpoints"
HEALTH_ENDPOINTS=("/debug/healthz" "/debug/ready")
for endpoint in "${HEALTH_ENDPOINTS[@]}"; do
    HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080$endpoint")
    if [ "$HEALTH_RESPONSE" = "200" ]; then
        print_result 0 "Health endpoint $endpoint is working"
    else
        print_result 1 "Health endpoint $endpoint failed (HTTP: $HEALTH_RESPONSE)"
    fi
done

echo ""
echo "📋 Test 8: Test OAuth2 Endpoints"
# Test token endpoint (should return error without proper credentials)
TOKEN_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/oauth/v2/token)
if [ "$TOKEN_RESPONSE" = "400" ] || [ "$TOKEN_RESPONSE" = "401" ]; then
    print_result 0 "OAuth2 token endpoint is responding (error expected without credentials)"
else
    print_result 1 "OAuth2 token endpoint unexpected response (HTTP: $TOKEN_RESPONSE)"
fi

# Test authorization endpoint
AUTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/oauth/v2/authorize")
if [ "$AUTH_RESPONSE" = "400" ] || [ "$AUTH_RESPONSE" = "302" ]; then
    print_result 0 "OAuth2 authorization endpoint is responding"
else
    print_result 1 "OAuth2 authorization endpoint failed (HTTP: $AUTH_RESPONSE)"
fi

echo ""
echo "============================================"
echo "📊 Advanced Test Summary:"
echo -e "   ${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "   ${RED}Failed: $TESTS_FAILED${NC}"

echo ""
echo "🔗 Service URLs:"
echo "   - Zitadel Console: http://localhost:8080/ui/console"
echo "   - Zitadel Login: http://localhost:8080/ui/login"
echo "   - Health Check: http://localhost:8080/debug/healthz"
echo "   - OIDC Config: http://localhost:8080/.well-known/openid_configuration"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All advanced Zitadel tests passed! Service is fully functional.${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some advanced tests failed. This might be normal if Zitadel is still initializing.${NC}"
    echo -e "${BLUE}💡 Try running the test again in a few minutes if initialization is still in progress.${NC}"
    exit 1
fi