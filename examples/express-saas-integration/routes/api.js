const express = require('express');
const { requireAuth, requireAdmin, validateSaaS, requireApiKey, addUserContext } = require('../middleware/auth');

function createApiRoutes(saasConfigs, vaultClient) {
  const router = express.Router();
  
  // Add user context to all API routes
  router.use(addUserContext());
  
  // Validate SaaS for all routes with :saas parameter
  router.use('/:saas/*', validateSaaS(saasConfigs));
  
  // Public API endpoints (no auth required)
  
  // Get available SaaS configurations
  router.get('/saas', (req, res) => {
    const configs = {};
    
    for (const [saasId, config] of saasConfigs) {
      configs[saasId] = {
        id: saasId,
        orgId: config.orgId,
        issuerUrl: config.issuerUrl,
        authUrl: `/auth/${saasId}`,
        dashboardUrl: `/dashboard/${saasId}`
      };
    }
    
    res.json({
      success: true,
      saas: configs,
      total: saasConfigs.size
    });
  });
  
  // Get SaaS configuration (public info only)
  router.get('/saas/:saas', (req, res) => {
    const config = req.saasConfig;
    
    res.json({
      success: true,
      saas: {
        id: req.saasId,
        orgId: config.orgId,
        issuerUrl: config.issuerUrl,
        authUrl: `/auth/${req.saasId}`,
        dashboardUrl: `/dashboard/${req.saasId}`,
        callbackUrl: config.callbackUrl
      }
    });
  });
  
  // Protected API endpoints (require authentication)
  
  // Get current user info
  router.get('/:saas/user', requireAuth(), (req, res) => {
    const { accessToken, refreshToken, ...userInfo } = req.user;
    
    res.json({
      success: true,
      user: {
        ...userInfo,
        hasAccessToken: !!accessToken,
        hasRefreshToken: !!refreshToken
      }
    });
  });
  
  // Get user's access token
  router.get('/:saas/token', requireAuth(), (req, res) => {
    if (!req.user.accessToken) {
      return res.status(404).json({
        success: false,
        error: 'No access token available'
      });
    }
    
    res.json({
      success: true,
      token: {
        accessToken: req.user.accessToken,
        tokenType: 'Bearer',
        saasId: req.user.saasId,
        orgId: req.user.orgId
      }
    });
  });
  
  // Get user profile from Zitadel
  router.get('/:saas/profile', requireAuth(), async (req, res) => {
    try {
      const axios = require('axios');
      const config = req.saasConfig;
      
      const response = await axios.get(config.userinfoUrl, {
        headers: {
          'Authorization': `Bearer ${req.user.accessToken}`
        }
      });
      
      res.json({
        success: true,
        profile: response.data
      });
      
    } catch (error) {
      res.status(500).json({
        success: false,
        error: 'Failed to fetch profile',
        message: error.message
      });
    }
  });
  
  // Admin-only endpoints
  
  // Get organization info (admin only)
  router.get('/:saas/admin/org', requireAuth(), requireAdmin(), async (req, res) => {
    try {
      // In a real implementation, you would call Zitadel Management API
      // For demo purposes, we'll return mock data
      
      res.json({
        success: true,
        organization: {
          id: req.user.orgId,
          saasId: req.user.saasId,
          name: `SaaS Project ${req.saasId.slice(-1)}`,
          domain: `${req.saasId}.localhost`,
          userCount: 1, // Mock data
          createdAt: '2024-01-01T00:00:00Z'
        }
      });
      
    } catch (error) {
      res.status(500).json({
        success: false,
        error: 'Failed to fetch organization info',
        message: error.message
      });
    }
  });
  
  // Get SaaS configuration (admin only, includes sensitive data)
  router.get('/:saas/admin/config', requireAuth(), requireAdmin(), (req, res) => {
    const config = req.saasConfig;
    
    res.json({
      success: true,
      config: {
        saasId: req.saasId,
        orgId: config.orgId,
        clientId: config.clientId,
        issuerUrl: config.issuerUrl,
        authUrl: config.authUrl,
        tokenUrl: config.tokenUrl,
        userinfoUrl: config.userinfoUrl,
        callbackUrl: config.callbackUrl
      }
    });
  });
  
  // API Key endpoints (alternative authentication)
  
  // Validate API key
  router.get('/:saas/validate-key', requireApiKey(saasConfigs), (req, res) => {
    res.json({
      success: true,
      valid: true,
      context: req.apiContext
    });
  });
  
  // Get SaaS info via API key
  router.get('/:saas/info', requireApiKey(saasConfigs), (req, res) => {
    const config = req.saasConfig;
    
    res.json({
      success: true,
      saas: {
        id: req.saasId,
        orgId: config.orgId,
        name: `SaaS Project ${req.saasId.slice(-1)}`,
        domain: `${req.saasId}.localhost`
      }
    });
  });
  
  // Webhook endpoints
  
  // Zitadel webhook receiver
  router.post('/:saas/webhooks/zitadel', validateSaaS(saasConfigs), (req, res) => {
    // In a real implementation, you would:
    // 1. Verify webhook signature
    // 2. Process the event
    // 3. Update your application state
    
    console.log(`📨 Received Zitadel webhook for ${req.saasId}:`, req.body);
    
    res.json({
      success: true,
      message: 'Webhook received',
      saasId: req.saasId
    });
  });
  
  // Health check endpoint
  router.get('/health', (req, res) => {
    res.json({
      success: true,
      status: 'healthy',
      timestamp: new Date().toISOString(),
      saasCount: saasConfigs.size,
      uptime: process.uptime()
    });
  });
  
  // Error handling for API routes
  router.use((err, req, res, next) => {
    console.error('API Error:', err);
    
    res.status(err.status || 500).json({
      success: false,
      error: err.message || 'Internal server error',
      ...(process.env.NODE_ENV === 'development' && { stack: err.stack })
    });
  });
  
  return router;
}

module.exports = createApiRoutes;