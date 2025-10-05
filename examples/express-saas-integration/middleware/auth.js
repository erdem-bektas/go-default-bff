// Authentication middleware for Express.js SaaS integration

/**
 * Middleware to require authentication for specific SaaS
 * @param {string} saasId - The SaaS identifier
 * @returns {Function} Express middleware function
 */
function requireSaaSAuth(saasId) {
  return (req, res, next) => {
    // Check if user is authenticated
    if (!req.isAuthenticated()) {
      req.session.returnTo = req.originalUrl;
      return res.redirect(`/auth/${saasId}`);
    }
    
    // Check if user belongs to the correct SaaS
    if (req.user.saasId !== saasId) {
      return res.status(403).json({
        error: 'Access denied',
        message: `User is authenticated for ${req.user.saasId} but trying to access ${saasId}`,
        currentSaas: req.user.saasId,
        requestedSaas: saasId
      });
    }
    
    next();
  };
}

/**
 * Middleware to require authentication for any SaaS (extracted from URL)
 * @param {Function} saasExtractor - Function to extract SaaS ID from request
 * @returns {Function} Express middleware function
 */
function requireAuth(saasExtractor = (req) => req.params.saas) {
  return (req, res, next) => {
    const saasId = saasExtractor(req);
    
    if (!saasId) {
      return res.status(400).json({
        error: 'Bad request',
        message: 'SaaS identifier not found in request'
      });
    }
    
    return requireSaaSAuth(saasId)(req, res, next);
  };
}

/**
 * Middleware to check if user has specific role in their SaaS organization
 * @param {string|Array} roles - Required role(s)
 * @returns {Function} Express middleware function
 */
function requireRole(roles) {
  const requiredRoles = Array.isArray(roles) ? roles : [roles];
  
  return (req, res, next) => {
    if (!req.isAuthenticated()) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'Authentication required'
      });
    }
    
    const userRoles = req.user.profile?.roles || [];
    const hasRequiredRole = requiredRoles.some(role => userRoles.includes(role));
    
    if (!hasRequiredRole) {
      return res.status(403).json({
        error: 'Forbidden',
        message: `Required role(s): ${requiredRoles.join(', ')}`,
        userRoles: userRoles
      });
    }
    
    next();
  };
}

/**
 * Middleware to check if user is admin in their SaaS organization
 * @returns {Function} Express middleware function
 */
function requireAdmin() {
  return requireRole(['admin', 'ORG_OWNER']);
}

/**
 * Middleware to add user context to request
 * @returns {Function} Express middleware function
 */
function addUserContext() {
  return (req, res, next) => {
    if (req.isAuthenticated()) {
      req.userContext = {
        id: req.user.id,
        saasId: req.user.saasId,
        orgId: req.user.orgId,
        email: req.user.email,
        name: req.user.name,
        roles: req.user.profile?.roles || [],
        isAdmin: (req.user.profile?.roles || []).some(role => 
          ['admin', 'ORG_OWNER'].includes(role)
        )
      };
    }
    
    next();
  };
}

/**
 * Middleware to validate SaaS exists and is configured
 * @param {Map} saasConfigs - Map of SaaS configurations
 * @returns {Function} Express middleware function
 */
function validateSaaS(saasConfigs) {
  return (req, res, next) => {
    const saasId = req.params.saas;
    
    if (!saasId) {
      return res.status(400).json({
        error: 'Bad request',
        message: 'SaaS identifier required'
      });
    }
    
    if (!saasConfigs.has(saasId)) {
      return res.status(404).json({
        error: 'SaaS not found',
        message: `SaaS '${saasId}' is not configured`,
        availableSaas: Array.from(saasConfigs.keys())
      });
    }
    
    req.saasId = saasId;
    req.saasConfig = saasConfigs.get(saasId);
    req.session.currentSaas = saasId;
    
    next();
  };
}

/**
 * Middleware for API key authentication (alternative to session-based auth)
 * @param {Map} saasConfigs - Map of SaaS configurations
 * @returns {Function} Express middleware function
 */
function requireApiKey(saasConfigs) {
  return async (req, res, next) => {
    const apiKey = req.headers['x-api-key'] || req.query.api_key;
    
    if (!apiKey) {
      return res.status(401).json({
        error: 'Unauthorized',
        message: 'API key required'
      });
    }
    
    // In a real implementation, you would validate the API key against a database
    // For demo purposes, we'll just check if it matches a pattern
    const saasId = req.params.saas;
    const config = saasConfigs.get(saasId);
    
    if (!config) {
      return res.status(404).json({
        error: 'SaaS not found'
      });
    }
    
    // Simple API key validation (in production, use proper key management)
    const expectedKey = `${saasId}_${config.clientId.slice(0, 8)}`;
    
    if (apiKey !== expectedKey) {
      return res.status(401).json({
        error: 'Invalid API key'
      });
    }
    
    // Add API context to request
    req.apiContext = {
      saasId: saasId,
      orgId: config.orgId,
      keyType: 'api_key'
    };
    
    next();
  };
}

module.exports = {
  requireAuth,
  requireSaaSAuth,
  requireRole,
  requireAdmin,
  addUserContext,
  validateSaaS,
  requireApiKey
};