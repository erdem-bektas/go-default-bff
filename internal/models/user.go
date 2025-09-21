package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role - Kullanıcı rolleri
type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"` // admin, user, moderator
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User - Kullanıcı modeli
type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Age       int       `json:"age"`
	Active    bool      `json:"active" gorm:"default:true"`
	RoleID    uuid.UUID `json:"role_id" gorm:"type:uuid;not null"`
	Role      Role      `json:"role" gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hook - ID oluştur
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// CreateUserRequest - User oluşturma isteği
type CreateUserRequest struct {
	Name   string    `json:"name" validate:"required,min=2,max=100"`
	Email  string    `json:"email" validate:"required,email"`
	Age    int       `json:"age" validate:"min=0,max=150"`
	Active *bool     `json:"active,omitempty"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// UpdateUserRequest - User güncelleme isteği
type UpdateUserRequest struct {
	Name   *string    `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email  *string    `json:"email,omitempty" validate:"omitempty,email"`
	Age    *int       `json:"age,omitempty" validate:"omitempty,min=0,max=150"`
	Active *bool      `json:"active,omitempty"`
	RoleID *uuid.UUID `json:"role_id,omitempty"`
}

// CreateRoleRequest - Role oluşturma isteği
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Description string `json:"description,omitempty"`
}

// UpdateRoleRequest - Role güncelleme isteği
type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Description *string `json:"description,omitempty"`
}
