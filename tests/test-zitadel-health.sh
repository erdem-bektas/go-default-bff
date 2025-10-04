#!/bin/bash

echo "🔍 Testing Zitadel Health and Functionality..."
echo "=============================================="

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

# Function to get access token
get_access_token() {
    local response=$(curl -s -X POST http://localhost:8080/oauth/v2/token \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=client_credentials&scope=openid profile urn:zitadel:iam:org:project:id:zitadel:aud" \
        --user "login-client:")
    
    echo "$response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4
}

echo ""
echo "📋 Test 1: Check if Zitadel containers are running"
ZITADEL_RUNNING=0
if docker ps | grep -q "zitadel-zitadel"; then
    ZITADEL_RUNNING=1
    print_result 0 "Zitadel container is running"
else
    print_result 1 "Zitadel container is not running"
fi

if docker ps | grep -q "zitadel-db"; then
    print_result 0 "Zitadel database container is running"
else
    print_result 1 "Zitadel database container is not running"
fi

echo ""
echo "📋 Test 2: Check Zitadel HTTP endpoints"
if curl -s -f http://localhost:8080/debug/healthz >/dev/null 2>&1; then
    print_result 0 "Zitadel health endpoint is accessible"
else
    print_result 1 "Zitadel health endpoint is not accessible"
fi

if curl -s -f http://localhost:8080/ui/console >/dev/null 2>&1; then
    print_result 0 "Zitadel Console UI is accessible"
else
    print_result 1 "Zitadel Console UI is not accessible"
fi

echo ""
echo "📋 Test 3: Check Zitadel OIDC Discovery"
DISCOVERY_RESPONSE=$(curl -s http://localhost:8080/.well-known/openid_configuration 2>/dev/null)
if echo "$DISCOVERY_RESPONSE" | grep -q '"issuer"'; then
    print_result 0 "OIDC Discovery endpoint works"
elif echo "$DISCOVERY_RESPONSE" | grep -q '"code":5'; then
    print_result 1 "OIDC Discovery endpoint not configured (Zitadel not fully initialized)"
else
    print_result 1 "OIDC Discovery endpoint failed"
fi

echo ""
echo "📋 Test 4: Test Zitadel API accessibility"
API_RESPONSE=$(curl -s http://localhost:8080/management/v1/orgs/me 2>/dev/null)
if [ $? -eq 0 ]; then
    print_result 0 "Zitadel Management API is accessible"
else
    print_result 1 "Zitadel Management API is not accessible"
fi

# Only run advanced tests if Zitadel is running
if [ $ZITADEL_RUNNING -eq 1 ]; then
    echo ""
    echo "📋 Test 5: Test admin login functionality"
    echo -e "${BLUE}ℹ️  Testing admin login with root@my-organization.localhost${NC}"
    
    # Get login page
    LOGIN_PAGE=$(curl -s http://localhost:8080/ui/login/loginname)
    if echo "$LOGIN_PAGE" | grep -q "loginname"; then
        print_result 0 "Login page is accessible"
    else
        print_result 1 "Login page is not accessible"
    fi

    echo ""
    echo "📋 Test 6: Test OAuth2 token endpoint"
    # Try to get a token (this might fail if client credentials are not set up properly)
    TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/oauth/v2/token \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=client_credentials&scope=openid" 2>/dev/null)
    
    if echo "$TOKEN_RESPONSE" | grep -q "access_token\|error"; then
        print_result 0 "OAuth2 token endpoint is responding"
    else
        print_result 1 "OAuth2 token endpoint is not responding"
    fi

    echo ""
    echo "📋 Test 7: Check database connectivity"
    # Try different possible container names
    DB_CONTAINER=$(docker ps --format "table {{.Names}}" | grep -E "(zitadel.*db|db.*zitadel)" | head -1)
    if [ -n "$DB_CONTAINER" ]; then
        DB_TEST=$(docker exec "$DB_CONTAINER" psql -U postgres -d zitadel -c "SELECT 1;" 2>/dev/null)
        if echo "$DB_TEST" | grep -q "1"; then
            print_result 0 "Database connectivity works"
        else
            print_result 1 "Database connectivity failed"
        fi
    else
        print_result 1 "Database container not found"
    fi

    echo ""
    echo "📋 Test 8: Check if Zitadel tables exist"
    if [ -n "$DB_CONTAINER" ]; then
        TABLES_TEST=$(docker exec "$DB_CONTAINER" psql -U postgres -d zitadel -c "\dt" 2>/dev/null)
        if echo "$TABLES_TEST" | grep -q "projections\|eventstore\|system"; then
            print_result 0 "Zitadel database tables exist"
        else
            print_result 1 "Zitadel database tables not found (may still be initializing)"
        fi
    else
        print_result 1 "Database container not found for table check"
    fi

    echo ""
    echo "📋 Test 9: Test Zitadel readiness"
    READY_RESPONSE=$(curl -s http://localhost:8080/debug/ready 2>/dev/null)
    if [ $? -eq 0 ]; then
        print_result 0 "Zitadel readiness check passed"
    else
        print_result 1 "Zitadel readiness check failed"
    fi

else
    echo -e "${YELLOW}⚠️  Skipping advanced tests - Zitadel container not running${NC}"
fi

echo ""
echo "=============================================="
echo "📊 Test Summary:"
echo -e "   ${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "   ${RED}Failed: $TESTS_FAILED${NC}"

echo ""
echo "🔗 Quick Access Links:"
echo "   - Zitadel Console: http://localhost:8080/ui/console"
echo "   - Zitadel Login: http://localhost:8080/ui/login"
echo "   - Admin Login: root@my-organization.localhost / RootPassword1!"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All Zitadel tests passed! Zitadel is healthy and functional.${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some Zitadel tests failed. Check the output above.${NC}"
    exit 1
fi