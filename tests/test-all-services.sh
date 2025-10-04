#!/bin/bash

echo "🚀 Running Complete Service Health Tests..."
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Make scripts executable
chmod +x test-vault-health.sh
chmod +x test-zitadel-health.sh

echo ""
echo -e "${BLUE}🔧 Step 1: Testing Vault Health...${NC}"
echo "=================================="
./test-vault-health.sh
VAULT_EXIT_CODE=$?

echo ""
echo ""
echo -e "${BLUE}🔧 Step 2: Testing Zitadel Health...${NC}"
echo "===================================="
./test-zitadel-health.sh
ZITADEL_EXIT_CODE=$?

echo ""
echo ""
echo "=========================================="
echo "📊 Overall Test Results:"
echo "=========================================="

if [ $VAULT_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Vault: All tests passed${NC}"
else
    echo -e "${RED}❌ Vault: Some tests failed${NC}"
fi

if [ $ZITADEL_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Zitadel: All tests passed${NC}"
else
    echo -e "${RED}❌ Zitadel: Some tests failed${NC}"
fi

echo ""
if [ $VAULT_EXIT_CODE -eq 0 ] && [ $ZITADEL_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}🎉 SUCCESS: All services are healthy and working properly!${NC}"
    echo ""
    echo "🌐 Your services are ready:"
    echo "   - Vault UI: http://localhost:8200 (token: dev-root)"
    echo "   - Zitadel Console: http://localhost:8080/ui/console"
    echo "   - Admin Login: root@my-organization.localhost / RootPassword1!"
    exit 0
else
    echo -e "${RED}⚠️  FAILURE: Some services have issues. Check the test output above.${NC}"
    echo ""
    echo "🔧 Troubleshooting tips:"
    echo "   1. Make sure services are started: ./start-vault-zitadel.sh"
    echo "   2. Wait a few minutes for services to fully initialize"
    echo "   3. Check Docker containers: docker ps"
    echo "   4. Check logs: docker logs <container-name>"
    exit 1
fi