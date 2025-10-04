#!/bin/bash

echo "🔍 Testing Vault Health and Accessibility..."
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

echo ""
echo "📋 Test 1: Check if Vault container is running"
if docker ps | grep -q "vault-dev"; then
    print_result 0 "Vault container is running"
else
    print_result 1 "Vault container is not running"
fi

echo ""
echo "📋 Test 2: Check Vault HTTP endpoint accessibility"
if curl -s -f http://localhost:8200/v1/sys/health >/dev/null 2>&1; then
    print_result 0 "Vault HTTP endpoint is accessible"
else
    print_result 1 "Vault HTTP endpoint is not accessible"
fi

echo ""
echo "📋 Test 3: Check Vault authentication with dev token"
AUTH_RESPONSE=$(curl -s -H "X-Vault-Token: dev-root" http://localhost:8200/v1/auth/token/lookup-self 2>/dev/null)
if echo "$AUTH_RESPONSE" | grep -q '"display_name":"token"'; then
    print_result 0 "Vault authentication with dev-root token works"
else
    print_result 1 "Vault authentication failed"
fi

echo ""
echo "📋 Test 4: Check if Zitadel secrets exist in Vault"
SECRET_RESPONSE=$(curl -s -H "X-Vault-Token: dev-root" http://localhost:8200/v1/secret/data/zitadel/database 2>/dev/null)
if echo "$SECRET_RESPONSE" | grep -q '"postgres_host"'; then
    print_result 0 "Zitadel database secrets exist in Vault"
else
    print_result 1 "Zitadel database secrets not found in Vault"
fi

echo ""
echo "📋 Test 5: Check Zitadel config secrets in Vault"
CONFIG_RESPONSE=$(curl -s -H "X-Vault-Token: dev-root" http://localhost:8200/v1/secret/data/zitadel/config 2>/dev/null)
if echo "$CONFIG_RESPONSE" | grep -q '"masterkey"'; then
    print_result 0 "Zitadel config secrets exist in Vault"
else
    print_result 1 "Zitadel config secrets not found in Vault"
fi

echo ""
echo "📋 Test 6: Test external access to Vault (from outside container)"
EXTERNAL_TEST=$(curl -s -H "X-Vault-Token: dev-root" http://localhost:8200/v1/sys/seal-status 2>/dev/null)
if [ $? -eq 0 ] && echo "$EXTERNAL_TEST" | grep -q '"sealed":false'; then
    print_result 0 "Vault is accessible from external clients"
else
    print_result 1 "Vault external access failed"
fi

echo ""
echo "📋 Test 7: Check Vault UI accessibility"
if curl -s -f http://localhost:8200/ui/ >/dev/null 2>&1; then
    print_result 0 "Vault UI is accessible"
else
    print_result 1 "Vault UI is not accessible"
fi

echo ""
echo "📋 Test 8: Verify Vault is unsealed and ready"
STATUS_RESPONSE=$(curl -s http://localhost:8200/v1/sys/seal-status 2>/dev/null)
if echo "$STATUS_RESPONSE" | grep -q '"sealed":false'; then
    print_result 0 "Vault is unsealed and ready"
else
    print_result 1 "Vault is sealed or not ready"
fi

echo ""
echo "=============================================="
echo "📊 Test Summary:"
echo -e "   ${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "   ${RED}Failed: $TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All Vault tests passed! Vault is healthy and accessible.${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some Vault tests failed. Check the output above.${NC}"
    exit 1
fi