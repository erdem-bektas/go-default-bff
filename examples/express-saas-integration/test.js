const axios = require('axios');

const BASE_URL = 'http://localhost:3001';

async function testEndpoints() {
  console.log('🧪 Testing Express.js SaaS Integration...\n');
  
  const tests = [
    {
      name: 'Server Health Check',
      url: `${BASE_URL}/status`,
      method: 'GET'
    },
    {
      name: 'Home Page',
      url: `${BASE_URL}/`,
      method: 'GET'
    },
    {
      name: 'SaaS Configurations API',
      url: `${BASE_URL}/api/configs`,
      method: 'GET'
    },
    {
      name: 'SP1 Auth Redirect (should redirect)',
      url: `${BASE_URL}/auth/sp1`,
      method: 'GET',
      expectRedirect: true
    },
    {
      name: 'SP2 Auth Redirect (should redirect)',
      url: `${BASE_URL}/auth/sp2`,
      method: 'GET',
      expectRedirect: true
    },
    {
      name: 'Invalid SaaS (should return 404)',
      url: `${BASE_URL}/auth/invalid`,
      method: 'GET',
      expectError: true
    }
  ];
  
  for (const test of tests) {
    try {
      console.log(`🔍 Testing: ${test.name}`);
      
      const config = {
        method: test.method,
        url: test.url,
        timeout: 5000,
        validateStatus: () => true, // Don't throw on any status code
        maxRedirects: 0 // Don't follow redirects
      };
      
      const response = await axios(config);
      
      if (test.expectRedirect && (response.status === 302 || response.status === 301)) {
        console.log(`   ✅ PASS - Redirected to: ${response.headers.location}`);
      } else if (test.expectError && response.status >= 400) {
        console.log(`   ✅ PASS - Expected error: ${response.status}`);
      } else if (!test.expectRedirect && !test.expectError && response.status === 200) {
        console.log(`   ✅ PASS - Status: ${response.status}`);
        
        // Show some response data for API endpoints
        if (test.url.includes('/api/') || test.url.includes('/status')) {
          const data = response.data;
          if (data.configurations) {
            console.log(`   📋 SaaS Count: ${Object.keys(data.configurations).length}`);
          }
          if (data.success !== undefined) {
            console.log(`   📊 Success: ${data.success}`);
          }
        }
      } else {
        console.log(`   ⚠️  UNEXPECTED - Status: ${response.status}`);
      }
      
    } catch (error) {
      if (error.code === 'ECONNREFUSED') {
        console.log(`   ❌ FAIL - Server not running (${error.code})`);
      } else {
        console.log(`   ❌ FAIL - ${error.message}`);
      }
    }
    
    console.log('');
  }
  
  console.log('🎯 Test Summary:');
  console.log('   • Make sure server is running: npm start');
  console.log('   • Make sure Vault is running: ./start-vault-zitadel.sh');
  console.log('   • Make sure SaaS orgs are created: ./manage-saas-orgs.sh');
  console.log('   • Visit http://localhost:3001 to test authentication flow');
}

// Run tests
testEndpoints().catch(console.error);