package database

import (
	"log"
	"webauthn-demo/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the SQLite database connection
func InitDB(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	// Auto-migrate the schema
	err = DB.AutoMigrate(&models.User{}, &models.Credential{})
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// UserRepository provides user-related database operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByUsername retrieves a user by username
func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Credentials").Where("username = ?", username).First(&user).Error
	return &user, err
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Credentials").Where("id = ?", id).First(&user).Error
	return &user, err
}

// UpdateUser updates a user in the database
func (r *UserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// CredentialRepository provides credential-related database operations
type CredentialRepository struct {
	db *gorm.DB
}

// NewCredentialRepository creates a new credential repository
func NewCredentialRepository(db *gorm.DB) *CredentialRepository {
	return &CredentialRepository{db: db}
}

// CreateCredential creates a new credential in the database
func (r *CredentialRepository) CreateCredential(credential *models.Credential) error {
	return r.db.Create(credential).Error
}

// GetCredentialByID retrieves a credential by its ID
func (r *CredentialRepository) GetCredentialByID(credentialID []byte) (*models.Credential, error) {
	var credential models.Credential
	err := r.db.Where("credential_id = ?", credentialID).First(&credential).Error
	return &credential, err
}

// UpdateCredential updates a credential in the database
func (r *CredentialRepository) UpdateCredential(credential *models.Credential) error {
	return r.db.Save(credential).Error
}

// GetCredentialsByUserID retrieves all credentials for a user
func (r *CredentialRepository) GetCredentialsByUserID(userID uint) ([]models.Credential, error) {
	var credentials []models.Credential
	err := r.db.Where("user_id = ?", userID).Find(&credentials).Error
	return credentials, err
}
