-- Migration: Update users table for Zitadel integration
-- Up

-- Drop existing foreign key constraint and role_id column
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_id_fkey;
ALTER TABLE users DROP COLUMN IF EXISTS role_id;
ALTER TABLE users DROP COLUMN IF EXISTS age;

-- Add new columns for Zitadel integration
ALTER TABLE users ADD COLUMN IF NOT EXISTS zitadel_id VARCHAR(255) UNIQUE NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS given_name VARCHAR(100) DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS family_name VARCHAR(100) DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(100) DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS org_id VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS project_id VARCHAR(255) NOT NULL DEFAULT '';

-- Change ID column to use integer instead of UUID for better performance
-- Note: This is a breaking change - in production, you'd need a more careful migration strategy
-- ALTER TABLE users ALTER COLUMN id TYPE INTEGER USING (ROW_NUMBER() OVER ());

-- Make email nullable for social logins
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- Rename active to is_active for consistency
ALTER TABLE users RENAME COLUMN active TO is_active;

-- Create partial unique index for email (allows multiple NULL emails but enforces uniqueness for non-NULL values)
DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX idx_users_email_not_null ON users(email) WHERE email IS NOT NULL;

-- Create indexes for new columns
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_zitadel_id ON users(zitadel_id);
CREATE INDEX IF NOT EXISTS idx_users_org_id ON users(org_id);
CREATE INDEX IF NOT EXISTS idx_users_project_id ON users(project_id);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Down (for rollback)
-- DROP INDEX IF EXISTS idx_users_email_not_null;
-- DROP INDEX IF EXISTS idx_users_zitadel_id;
-- DROP INDEX IF EXISTS idx_users_org_id;
-- DROP INDEX IF EXISTS idx_users_project_id;
-- ALTER TABLE users DROP COLUMN IF EXISTS zitadel_id;
-- ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
-- ALTER TABLE users DROP COLUMN IF EXISTS given_name;
-- ALTER TABLE users DROP COLUMN IF EXISTS family_name;
-- ALTER TABLE users DROP COLUMN IF EXISTS username;
-- ALTER TABLE users DROP COLUMN IF EXISTS org_id;
-- ALTER TABLE users DROP COLUMN IF EXISTS project_id;
-- ALTER TABLE users RENAME COLUMN is_active TO active;
-- ALTER TABLE users ALTER COLUMN email SET NOT NULL;