package models

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a WebAuthn user
type User struct {
	ID          uint         `gorm:"primaryKey"`
	UUID        string       `gorm:"unique;not null"`
	Username    string       `gorm:"unique;not null"`
	DisplayName string       `gorm:"not null"`
	Credentials []Credential `gorm:"foreignKey:UserID"`
	CreatedAt   int64        `gorm:"autoCreateTime"`
	UpdatedAt   int64        `gorm:"autoUpdateTime"`
}

// WebAuthnID returns the user's ID as required by the WebAuthn library
func (u *User) WebAuthnID() []byte {
	return []byte(u.UUID)
}

// WebAuthnName returns the user's username as required by the WebAuthn library
func (u *User) WebAuthnName() string {
	return u.Username
}

// WebAuthnDisplayName returns the user's display name as required by the WebAuthn library
func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnIcon returns the user's icon URL (not used in this demo)
func (u *User) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials returns the user's credentials as required by the WebAuthn library
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	credentials := make([]webauthn.Credential, len(u.Credentials))
	for i, cred := range u.Credentials {
		credentials[i] = cred.ToWebAuthnCredential()
	}
	return credentials
}

// BeforeCreate hook to generate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}
	return nil
}
