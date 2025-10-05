package models

import (
	"time"
)

// User - Updated user model for Zitadel integration with normalized role relationship
type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ZitadelID     string    `gorm:"uniqueIndex;not null" json:"zitadel_id"`
	Email         *string   `json:"email"` // Nullable for social logins, partial unique index
	EmailVerified bool      `gorm:"default:false" json:"email_verified"`
	Name          string    `json:"name"`
	GivenName     string    `json:"given_name"`
	FamilyName    string    `json:"family_name"`
	Username      string    `json:"username"`
	OrgID         string    `json:"org_id"`
	ProjectID     string    `json:"project_id"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Normalized role relationship (preferred over []string)
	Roles []UserRole `gorm:"foreignKey:UserID" json:"roles"`
}

// UserRole - Normalized user-role relationship with multi-tenant support
type UserRole struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;index"`
	Role      string `gorm:"not null"`
	OrgID     string `gorm:"not null"`
	ProjectID string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Composite unique index for role uniqueness per user/org/project
	User User `gorm:"constraint:OnDelete:CASCADE"`
}

// ZitadelUserInfo - User information from Zitadel OIDC
type ZitadelUserInfo struct {
	Sub               string              `json:"sub"`
	Name              string              `json:"name"`
	GivenName         string              `json:"given_name"`
	FamilyName        string              `json:"family_name"`
	PreferredUsername string              `json:"preferred_username"`
	Email             string              `json:"email"`
	EmailVerified     bool                `json:"email_verified"`
	Roles             []string            `json:"urn:zitadel:iam:org:project:roles"` // Configurable claim name
	OrgID             string              `json:"urn:zitadel:iam:org:id"`
	ProjectRoles      map[string][]string `json:"urn:zitadel:iam:org:project:roles:audience"` // Multi-project support
}

// CreateUserRequest - User creation request for Zitadel integration
type CreateUserRequest struct {
	ZitadelID     string  `json:"zitadel_id" validate:"required"`
	Email         *string `json:"email,omitempty" validate:"omitempty,email"`
	EmailVerified bool    `json:"email_verified"`
	Name          string  `json:"name" validate:"required,min=1,max=100"`
	GivenName     string  `json:"given_name"`
	FamilyName    string  `json:"family_name"`
	Username      string  `json:"username"`
	OrgID         string  `json:"org_id" validate:"required"`
	ProjectID     string  `json:"project_id" validate:"required"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// UpdateUserRequest - User update request
type UpdateUserRequest struct {
	Email         *string `json:"email,omitempty" validate:"omitempty,email"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
	Name          *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	GivenName     *string `json:"given_name,omitempty"`
	FamilyName    *string `json:"family_name,omitempty"`
	Username      *string `json:"username,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// UserRoleRequest - User role assignment request
type UserRoleRequest struct {
	UserID    uint   `json:"user_id" validate:"required"`
	Role      string `json:"role" validate:"required"`
	OrgID     string `json:"org_id" validate:"required"`
	ProjectID string `json:"project_id" validate:"required"`
}
