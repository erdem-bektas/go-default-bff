#!/bin/bash

# Test Services - Main Test Runner
# Usage: ./test-services.sh [vault|zitadel|functionality|all]

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test directory
TEST_DIR="tests"

# Function to print usage
print_usage() {
    echo "🧪 Test Services - Vault & Zitadel Health Checker"
    echo "================================================="
    echo ""
    echo "Usage: $0 [option]"
    echo ""
    echo "Options:"
    echo "  vault         - Test only Vault health and accessibility"
    echo "  zitadel       - Test only Zitadel basic health"
    echo "  functionality - Test Zitadel advanced functionality"
    echo "  all           - Run all tests (default)"
    echo "  help          - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0              # Run all tests"
    echo "  $0 vault       # Test only Vault"
    echo "  $0 zitadel     # Test only Zitadel basic"
    echo "  $0 all         # Run comprehensive tests"
}

# Function to run a test script
run_test() {
    local test_name="$1"
    local script_path="$TEST_DIR/$2"
    
    if [ -f "$script_path" ]; then
        echo -e "${BLUE}🔧 Running $test_name...${NC}"
        echo "=================================="
        chmod +x "$script_path"
        "$script_path"
        return $?
    else
        echo -e "${RED}❌ Test script not found: $script_path${NC}"
        return 1
    fi
}

# Main execution
case "${1:-all}" in
    "vault")
        run_test "Vault Health Tests" "test-vault-health.sh"
        ;;
    "zitadel")
        run_test "Zitadel Health Tests" "test-zitadel-health.sh"
        ;;
    "functionality")
        run_test "Zitadel Functionality Tests" "test-zitadel-functionality.sh"
        ;;
    "all")
        echo "🚀 Running Complete Service Health Tests..."
        echo "=========================================="
        echo ""
        
        # Run Vault tests
        run_test "Vault Health Tests" "test-vault-health.sh"
        VAULT_EXIT=$?
        
        echo ""
        echo ""
        
        # Run Zitadel tests
        run_test "Zitadel Health Tests" "test-zitadel-health.sh"
        ZITADEL_EXIT=$?
        
        echo ""
        echo ""
        
        # Summary
        echo "=========================================="
        echo "📊 Overall Test Results:"
        echo "=========================================="
        
        if [ $VAULT_EXIT -eq 0 ]; then
            echo -e "${GREEN}✅ Vault: All tests passed${NC}"
        else
            echo -e "${RED}❌ Vault: Some tests failed${NC}"
        fi
        
        if [ $ZITADEL_EXIT -eq 0 ]; then
            echo -e "${GREEN}✅ Zitadel: All tests passed${NC}"
        else
            echo -e "${RED}❌ Zitadel: Some tests failed${NC}"
        fi
        
        echo ""
        if [ $VAULT_EXIT -eq 0 ] && [ $ZITADEL_EXIT -eq 0 ]; then
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
            echo ""
            echo "🧪 Run specific tests:"
            echo "   ./test-services.sh vault        # Test only Vault"
            echo "   ./test-services.sh zitadel      # Test only Zitadel"
            echo "   ./test-services.sh functionality # Advanced Zitadel tests"
            exit 1
        fi
        ;;
    "help"|"-h"|"--help")
        print_usage
        ;;
    *)
        echo -e "${RED}❌ Unknown option: $1${NC}"
        echo ""
        print_usage
        exit 1
        ;;
esac