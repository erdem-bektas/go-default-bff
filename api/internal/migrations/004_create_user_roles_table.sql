-- Migration: Create user_roles table for normalized role relationship
-- Up

CREATE TABLE IF NOT EXISTS user_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    role VARCHAR(100) NOT NULL,
    org_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign key constraint with cascade delete
    CONSTRAINT fk_user_roles_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- Composite unique constraint for role uniqueness per user/org/project
    CONSTRAINT uk_user_roles_user_org_project_role UNIQUE (user_id, role, org_id, project_id)
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_org_id ON user_roles(org_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_project_id ON user_roles(project_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role);

-- Down (for rollback)
-- DROP TABLE IF EXISTS user_roles;