-- Migration: Create roles table
-- Up
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default roles
INSERT INTO roles (name, description) VALUES 
    ('admin', 'System administrator with full access'),
    ('user', 'Regular user with limited access'),
    ('moderator', 'Moderator with content management access')
ON CONFLICT (name) DO NOTHING;

-- Down (for rollback)
-- DROP TABLE IF EXISTS roles;