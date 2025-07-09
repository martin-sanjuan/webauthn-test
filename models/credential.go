package models

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Credential represents a WebAuthn credential
type Credential struct {
	ID              uint                  `gorm:"primaryKey"`
	UserID          uint                  `gorm:"not null"`
	CredentialID    []byte                `gorm:"unique;not null"`
	PublicKey       []byte                `gorm:"not null"`
	AttestationType string                `gorm:"not null"`
	Transport       []string              `gorm:"serializer:json"`
	Flags           UserVerificationFlags `gorm:"embedded"`
	Authenticator   AuthenticatorData     `gorm:"embedded"`
	CreatedAt       int64                 `gorm:"autoCreateTime"`
	UpdatedAt       int64                 `gorm:"autoUpdateTime"`
}

// UserVerificationFlags represents the flags for user verification
type UserVerificationFlags struct {
	UserPresent    bool `gorm:"default:false"`
	UserVerified   bool `gorm:"default:false"`
	BackupEligible bool `gorm:"default:false"`
	BackupState    bool `gorm:"default:false"`
}

// AuthenticatorData represents the authenticator data
type AuthenticatorData struct {
	AAGUID       []byte `gorm:"size:16"`
	SignCount    uint32 `gorm:"default:0"`
	CloneWarning bool   `gorm:"default:false"`
}

// ToWebAuthnCredential converts the database credential to a WebAuthn credential
func (c *Credential) ToWebAuthnCredential() webauthn.Credential {
	// Convert []string to []protocol.AuthenticatorTransport
	transports := make([]protocol.AuthenticatorTransport, len(c.Transport))
	for i, t := range c.Transport {
		transports[i] = protocol.AuthenticatorTransport(t)
	}

	return webauthn.Credential{
		ID:              c.CredentialID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       transports,
		Flags: webauthn.CredentialFlags{
			UserPresent:    c.Flags.UserPresent,
			UserVerified:   c.Flags.UserVerified,
			BackupEligible: c.Flags.BackupEligible,
			BackupState:    c.Flags.BackupState,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:       c.Authenticator.AAGUID,
			SignCount:    c.Authenticator.SignCount,
			CloneWarning: c.Authenticator.CloneWarning,
		},
	}
}

// FromWebAuthnCredential creates a database credential from a WebAuthn credential
func (c *Credential) FromWebAuthnCredential(cred webauthn.Credential) {
	c.CredentialID = cred.ID
	c.PublicKey = cred.PublicKey
	c.AttestationType = cred.AttestationType

	// Convert []protocol.AuthenticatorTransport to []string
	c.Transport = make([]string, len(cred.Transport))
	for i, t := range cred.Transport {
		c.Transport[i] = string(t)
	}

	c.Flags = UserVerificationFlags{
		UserPresent:    cred.Flags.UserPresent,
		UserVerified:   cred.Flags.UserVerified,
		BackupEligible: cred.Flags.BackupEligible,
		BackupState:    cred.Flags.BackupState,
	}
	c.Authenticator = AuthenticatorData{
		AAGUID:       cred.Authenticator.AAGUID,
		SignCount:    cred.Authenticator.SignCount,
		CloneWarning: cred.Authenticator.CloneWarning,
	}
}
