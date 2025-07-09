package services

import (
	"fmt"
	"log"
	"net/http"
	"webauthn-demo/database"
	"webauthn-demo/models"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

// WebAuthnService handles WebAuthn operations
type WebAuthnService struct {
	webauthn *webauthn.WebAuthn
	userRepo *database.UserRepository
	credRepo *database.CredentialRepository
	sessions map[string]*webauthn.SessionData // In-memory session store for demo
}

// NewWebAuthnService creates a new WebAuthn service
func NewWebAuthnService(db *gorm.DB, rpID, rpName, rpOrigin string) (*WebAuthnService, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: rpName,
		RPID:          rpID,
		RPOrigins:     []string{rpOrigin},
	}

	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		return nil, err
	}

	return &WebAuthnService{
		webauthn: webAuthn,
		userRepo: database.NewUserRepository(db),
		credRepo: database.NewCredentialRepository(db),
		sessions: make(map[string]*webauthn.SessionData),
	}, nil
}

// RegisterUser creates a new user and starts the registration process
func (s *WebAuthnService) RegisterUser(username, displayName string) (*models.User, error) {
	// Check if user already exists
	_, err := s.userRepo.GetUserByUsername(username)
	if err == nil {
		return nil, fmt.Errorf("user already exists")
	}

	// Create new user
	user := &models.User{
		Username:    username,
		DisplayName: displayName,
	}

	err = s.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// BeginRegistration starts the WebAuthn registration process
func (s *WebAuthnService) BeginRegistration(username string) (*protocol.CredentialCreation, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	credentialCreation, sessionData, err := s.webauthn.BeginRegistration(user)
	if err != nil {
		return nil, fmt.Errorf("failed to begin registration: %v", err)
	}

	// Store session data (in production, use a proper session store)
	s.sessions[username] = sessionData

	log.Printf("Started registration for user: %s", username)
	return credentialCreation, nil
}

// CompleteRegistration finishes the WebAuthn registration process
func (s *WebAuthnService) CompleteRegistration(username string, request *http.Request) (*webauthn.Credential, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	// Get session data
	sessionData, exists := s.sessions[username]
	if !exists {
		return nil, fmt.Errorf("no session found for user: %s", username)
	}

	credential, err := s.webauthn.FinishRegistration(user, *sessionData, request)
	if err != nil {
		return nil, fmt.Errorf("failed to finish registration: %v", err)
	}

	// Save credential to database
	dbCredential := &models.Credential{
		UserID: user.ID,
	}
	dbCredential.FromWebAuthnCredential(*credential)

	err = s.credRepo.CreateCredential(dbCredential)
	if err != nil {
		return nil, fmt.Errorf("failed to save credential: %v", err)
	}

	// Clean up session
	delete(s.sessions, username)

	log.Printf("Completed registration for user: %s", username)
	return credential, nil
}

// BeginAuthentication starts the WebAuthn authentication process
func (s *WebAuthnService) BeginAuthentication(username string) (*protocol.CredentialAssertion, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	credentialAssertion, sessionData, err := s.webauthn.BeginLogin(user)
	if err != nil {
		return nil, fmt.Errorf("failed to begin authentication: %v", err)
	}

	// Store session data
	s.sessions[username] = sessionData

	log.Printf("Started authentication for user: %s", username)
	return credentialAssertion, nil
}

// CompleteAuthentication finishes the WebAuthn authentication process
func (s *WebAuthnService) CompleteAuthentication(username string, request *http.Request) (*webauthn.Credential, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	// Get session data
	sessionData, exists := s.sessions[username]
	if !exists {
		return nil, fmt.Errorf("no session found for user: %s", username)
	}

	credential, err := s.webauthn.FinishLogin(user, *sessionData, request)
	if err != nil {
		return nil, fmt.Errorf("failed to finish authentication: %v", err)
	}

	// Update credential in database (sign count, etc.)
	dbCredential, err := s.credRepo.GetCredentialByID(credential.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %v", err)
	}

	dbCredential.FromWebAuthnCredential(*credential)
	err = s.credRepo.UpdateCredential(dbCredential)
	if err != nil {
		return nil, fmt.Errorf("failed to update credential: %v", err)
	}

	// Clean up session
	delete(s.sessions, username)

	log.Printf("Completed authentication for user: %s", username)
	return credential, nil
}

// GetUser retrieves a user by username
func (s *WebAuthnService) GetUser(username string) (*models.User, error) {
	return s.userRepo.GetUserByUsername(username)
}
