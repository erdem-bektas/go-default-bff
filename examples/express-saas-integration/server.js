const express = require('express');
const session = require('express-session');
const passport = require('passport');
const OpenIDConnectStrategy = require('passport-openidconnect').Strategy;
const vault = require('node-vault');
const cors = require('cors');
const helmet = require('helmet');
const morgan = require('morgan');
const rateLimit = require('express-rate-limit');
require('dotenv').config();

const app = express();
const PORT = process.env.PORT || 3001;

// Security middleware
app.use(helmet({
    contentSecurityPolicy: false // Disable for demo purposes
}));

app.use(cors({
    origin: process.env.ALLOWED_ORIGINS?.split(',') || ['http://localhost:3001'],
    credentials: true
}));

// Rate limiting
const limiter = rateLimit({
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 100 // limit each IP to 100 requests per windowMs
});
app.use(limiter);

// Logging
app.use(morgan('combined'));

// Body parsing
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Session configuration
app.use(session({
    secret: process.env.SESSION_SECRET || 'your-secret-key-change-in-production',
    resave: false,
    saveUninitialized: false,
    cookie: {
        secure: process.env.NODE_ENV === 'production',
        httpOnly: true,
        maxAge: 24 * 60 * 60 * 1000 // 24 hours
    }
}));

// Passport initialization
app.use(passport.initialize());
app.use(passport.session());

// Vault client
const vaultClient = vault({
    apiVersion: 'v1',
    endpoint: process.env.VAULT_ADDR || 'http://localhost:8200',
    token: process.env.VAULT_TOKEN || 'dev-root'
});

// SaaS configurations storage
const saasConfigs = new Map();
const passportStrategies = new Map();

// Load SaaS configurations from Vault
async function loadSaaSConfigurations() {
    const saasOrgs = ['sp1', 'sp2'];

    for (const org of saasOrgs) {
        try {
            const result = await vaultClient.read(`secret/data/saas/${org}/oauth`);

            if (!result?.data?.data) {
                console.warn(`⚠️  No configuration found for ${org}`);
                continue;
            }

            const config = result.data.data;

            const saasConfig = {
                orgId: config.org_id,
                clientId: config.client_id,
                clientSecret: config.client_secret,
                issuerUrl: config.issuer_url,
                authUrl: config.auth_url,
                tokenUrl: config.token_url,
                userinfoUrl: config.userinfo_url,
                callbackUrl: `http://localhost:${PORT}/auth/${org}/callback`
            };

            saasConfigs.set(org, saasConfig);

            // Create Passport strategy for this SaaS
            const strategy = new OpenIDConnectStrategy({
                issuer: saasConfig.issuerUrl,
                authorizationURL: saasConfig.authUrl,
                tokenURL: saasConfig.tokenUrl,
                userInfoURL: saasConfig.userinfoUrl,
                clientID: saasConfig.clientId,
                clientSecret: saasConfig.clientSecret,
                callbackURL: saasConfig.callbackUrl,
                scope: ['openid', 'profile', 'email'],
                passReqToCallback: true
            }, async (req, issuer, profile, context, idToken, accessToken, refreshToken, done) => {
                // Extract SaaS ID from request
                const saasId = req.params.saas || req.session.currentSaas;

                const user = {
                    id: profile.id,
                    saasId: saasId,
                    orgId: saasConfigs.get(saasId)?.orgId,
                    name: profile.displayName || profile.name?.givenName + ' ' + profile.name?.familyName,
                    email: profile.emails?.[0]?.value || profile.email,
                    username: profile.username,
                    accessToken: accessToken,
                    refreshToken: refreshToken,
                    profile: profile._json
                };

                return done(null, user);
            });

            // Register strategy with dynamic name
            passport.use(`oidc-${org}`, strategy);
            passportStrategies.set(org, `oidc-${org}`);

            console.log(`✅ Loaded configuration for SaaS: ${org} (OrgID: ${saasConfig.orgId})`);

        } catch (error) {
            console.error(`❌ Failed to load configuration for ${org}:`, error.message);
        }
    }
}

// Passport serialization
passport.serializeUser((user, done) => {
    done(null, user);
});

passport.deserializeUser((user, done) => {
    done(null, user);
});

// Middleware to check if user is authenticated for specific SaaS
function requireAuth(saasId) {
    return (req, res, next) => {
        if (req.isAuthenticated() && req.user.saasId === saasId) {
            return next();
        }

        req.session.returnTo = req.originalUrl;
        res.redirect(`/auth/${saasId}`);
    };
}

// Middleware to extract SaaS ID from URL
function extractSaasId(req, res, next) {
    const saasId = req.params.saas;

    if (!saasConfigs.has(saasId)) {
        return res.status(404).json({
            error: `SaaS '${saasId}' not configured`,
            availableSaas: Array.from(saasConfigs.keys())
        });
    }

    req.saasId = saasId;
    req.session.currentSaas = saasId;
    next();
}

// API Routes
const createApiRoutes = require('./routes/api');
app.use('/api', createApiRoutes(saasConfigs, vaultClient));

// Routes

// Home page
app.get('/', (req, res) => {
    const availableSaas = Array.from(saasConfigs.keys());

    res.send(`
    <!DOCTYPE html>
    <html>
    <head>
        <title>Express.js Multi-SaaS Authentication</title>
        <style>
            body { font-family: Arial, sans-serif; margin: 40px; }
            .saas-card { 
                border: 1px solid #ddd; 
                padding: 20px; 
                margin: 10px 0; 
                border-radius: 8px;
                background: #f9f9f9;
            }
            .btn { 
                background: #007bff; 
                color: white; 
                padding: 10px 20px; 
                text-decoration: none; 
                border-radius: 4px;
                display: inline-block;
                margin: 5px;
            }
            .btn:hover { background: #0056b3; }
            .user-info { 
                background: #d4edda; 
                padding: 15px; 
                border-radius: 4px; 
                margin: 10px 0;
            }
        </style>
    </head>
    <body>
        <h1>🚀 Express.js Multi-SaaS Authentication Demo</h1>
        
        ${req.user ? `
          <div class="user-info">
            <h3>✅ Logged in to: ${req.user.saasId.toUpperCase()}</h3>
            <p><strong>Name:</strong> ${req.user.name}</p>
            <p><strong>Email:</strong> ${req.user.email}</p>
            <p><strong>Org ID:</strong> ${req.user.orgId}</p>
            <a href="/dashboard/${req.user.saasId}" class="btn">Go to Dashboard</a>
            <a href="/logout" class="btn" style="background: #dc3545;">Logout</a>
          </div>
        ` : ''}
        
        <h2>Available SaaS Organizations:</h2>
        
        ${availableSaas.map(saas => `
          <div class="saas-card">
            <h3>${saas.toUpperCase()} - SaaS Project ${saas.slice(-1)}</h3>
            <p>Domain: ${saas}.localhost</p>
            <a href="/auth/${saas}" class="btn">Login to ${saas.toUpperCase()}</a>
            <a href="/dashboard/${saas}" class="btn" style="background: #28a745;">Dashboard</a>
          </div>
        `).join('')}
        
        <h2>🔧 Management:</h2>
        <a href="/status" class="btn" style="background: #6c757d;">View Status</a>
        <a href="/api/configs" class="btn" style="background: #17a2b8;">API Configs</a>
    </body>
    </html>
  `);
});

// Authentication routes
app.get('/auth/:saas', extractSaasId, (req, res, next) => {
    const strategyName = passportStrategies.get(req.saasId);

    if (!strategyName) {
        return res.status(500).json({ error: 'Authentication strategy not configured' });
    }

    passport.authenticate(strategyName)(req, res, next);
});

app.get('/auth/:saas/callback', extractSaasId, (req, res, next) => {
    const strategyName = passportStrategies.get(req.saasId);

    if (!strategyName) {
        return res.status(500).json({ error: 'Authentication strategy not configured' });
    }

    passport.authenticate(strategyName, {
        successRedirect: req.session.returnTo || `/dashboard/${req.saasId}`,
        failureRedirect: `/?error=auth_failed&saas=${req.saasId}`
    })(req, res, next);
});

// Dashboard routes (protected)
app.get('/dashboard/:saas', extractSaasId, requireAuth(req => req.params.saas), (req, res) => {
    const config = saasConfigs.get(req.saasId);

    res.send(`
    <!DOCTYPE html>
    <html>
    <head>
        <title>${req.saasId.toUpperCase()} Dashboard</title>
        <style>
            body { font-family: Arial, sans-serif; margin: 40px; }
            .dashboard { background: #f8f9fa; padding: 20px; border-radius: 8px; }
            .user-card { background: white; padding: 15px; margin: 10px 0; border-radius: 4px; }
            .btn { 
                background: #007bff; 
                color: white; 
                padding: 10px 20px; 
                text-decoration: none; 
                border-radius: 4px;
                display: inline-block;
                margin: 5px;
            }
        </style>
    </head>
    <body>
        <div class="dashboard">
            <h1>🏢 ${req.saasId.toUpperCase()} Dashboard</h1>
            
            <div class="user-card">
                <h3>👤 User Information</h3>
                <p><strong>ID:</strong> ${req.user.id}</p>
                <p><strong>Name:</strong> ${req.user.name}</p>
                <p><strong>Email:</strong> ${req.user.email}</p>
                <p><strong>SaaS:</strong> ${req.user.saasId}</p>
                <p><strong>Organization ID:</strong> ${req.user.orgId}</p>
            </div>
            
            <div class="user-card">
                <h3>🔧 SaaS Configuration</h3>
                <p><strong>Client ID:</strong> ${config.clientId}</p>
                <p><strong>Issuer:</strong> ${config.issuerUrl}</p>
                <p><strong>Callback URL:</strong> ${config.callbackUrl}</p>
            </div>
            
            <div class="user-card">
                <h3>🔑 Actions</h3>
                <a href="/api/user/${req.saasId}" class="btn">Get User JSON</a>
                <a href="/api/token/${req.saasId}" class="btn">Get Access Token</a>
                <a href="/" class="btn" style="background: #6c757d;">Home</a>
                <a href="/logout" class="btn" style="background: #dc3545;">Logout</a>
            </div>
        </div>
    </body>
    </html>
  `);
});

// API routes
app.get('/api/user/:saas', extractSaasId, requireAuth(req => req.params.saas), (req, res) => {
    const { accessToken, refreshToken, ...userInfo } = req.user;
    res.json({
        success: true,
        saas: req.saasId,
        user: userInfo,
        hasAccessToken: !!accessToken
    });
});

app.get('/api/token/:saas', extractSaasId, requireAuth(req => req.params.saas), (req, res) => {
    res.json({
        success: true,
        saas: req.saasId,
        accessToken: req.user.accessToken,
        tokenType: 'Bearer',
        expiresIn: 3600 // This should come from the actual token
    });
});

app.get('/api/configs', (req, res) => {
    const configs = {};

    for (const [saasId, config] of saasConfigs) {
        configs[saasId] = {
            orgId: config.orgId,
            clientId: config.clientId,
            issuerUrl: config.issuerUrl,
            callbackUrl: config.callbackUrl,
            hasStrategy: passportStrategies.has(saasId)
        };
    }

    res.json({
        success: true,
        configurations: configs,
        totalSaas: saasConfigs.size
    });
});

// Status page
app.get('/status', (req, res) => {
    const status = {
        server: 'running',
        timestamp: new Date().toISOString(),
        saasCount: saasConfigs.size,
        configurations: {}
    };

    for (const [saasId, config] of saasConfigs) {
        status.configurations[saasId] = {
            configured: true,
            hasStrategy: passportStrategies.has(saasId),
            orgId: config.orgId
        };
    }

    res.json(status);
});

// Logout
app.get('/logout', (req, res) => {
    req.logout((err) => {
        if (err) {
            console.error('Logout error:', err);
        }
        req.session.destroy((err) => {
            if (err) {
                console.error('Session destroy error:', err);
            }
            res.redirect('/');
        });
    });
});

// Error handling
app.use((err, req, res, next) => {
    console.error('Error:', err);
    res.status(500).json({
        error: 'Internal server error',
        message: process.env.NODE_ENV === 'development' ? err.message : 'Something went wrong'
    });
});

// 404 handler
app.use((req, res) => {
    res.status(404).json({
        error: 'Not found',
        path: req.path,
        availableEndpoints: [
            'GET /',
            'GET /auth/:saas',
            'GET /dashboard/:saas',
            'GET /api/configs',
            'GET /status'
        ]
    });
});

// Initialize and start server
async function startServer() {
    try {
        console.log('🔄 Loading SaaS configurations from Vault...');
        await loadSaaSConfigurations();

        if (saasConfigs.size === 0) {
            console.warn('⚠️  No SaaS configurations loaded. Make sure to run ./manage-saas-orgs.sh first');
        }

        app.listen(PORT, () => {
            console.log(`🚀 Express.js Multi-SaaS server running on http://localhost:${PORT}`);
            console.log(`📋 Loaded ${saasConfigs.size} SaaS configurations`);
            console.log(`🔗 Available SaaS: ${Array.from(saasConfigs.keys()).join(', ')}`);
            console.log(`💡 Visit http://localhost:${PORT} to test authentication`);
        });

    } catch (error) {
        console.error('❌ Failed to start server:', error);
        process.exit(1);
    }
}

// Graceful shutdown
process.on('SIGTERM', () => {
    console.log('🛑 Received SIGTERM, shutting down gracefully');
    process.exit(0);
});

process.on('SIGINT', () => {
    console.log('🛑 Received SIGINT, shutting down gracefully');
    process.exit(0);
});

startServer();