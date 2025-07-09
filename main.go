package main

import (
	"log"
	"net/http"
	"webauthn-demo/database"
	"webauthn-demo/services"

	"github.com/gin-gonic/gin"
)

var webauthnService *services.WebAuthnService

func main() {
	// Initialize database
	err := database.InitDB("webauthn.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize WebAuthn service
	webauthnService, err = services.NewWebAuthnService(
		database.GetDB(),
		"localhost",             // RP ID
		"WebAuthn Demo",         // RP Name
		"http://localhost:8080", // RP Origin
	)
	if err != nil {
		log.Fatal("Failed to initialize WebAuthn service:", err)
	}

	// Setup Gin router
	r := gin.Default()

	// Serve static files (for the HTML test page)
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// Home page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "WebAuthn Demo",
		})
	})

	// API routes
	api := r.Group("/api")
	{
		// User registration endpoint
		api.POST("/register", registerUser)

		// WebAuthn registration endpoints
		api.POST("/webauthn/register/begin", beginRegistration)
		api.POST("/webauthn/register/finish", finishRegistration)

		// WebAuthn authentication endpoints
		api.POST("/webauthn/login/begin", beginAuthentication)
		api.POST("/webauthn/login/finish", finishAuthentication)

		// User info endpoint
		api.GET("/user/:username", getUser)
	}

	log.Println("Server starting on :8080")
	r.Run(":8080")
}

// RegisterUserRequest represents the request to create a new user
type RegisterUserRequest struct {
	Username    string `json:"username" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
}

// BeginRegistrationRequest represents the request to start registration
type BeginRegistrationRequest struct {
	Username string `json:"username" binding:"required"`
}

// BeginAuthenticationRequest represents the request to start authentication
type BeginAuthenticationRequest struct {
	Username string `json:"username" binding:"required"`
}

// registerUser creates a new user account
func registerUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := webauthnService.RegisterUser(req.Username, req.DisplayName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"username":     user.Username,
			"display_name": user.DisplayName,
		},
	})
}

// beginRegistration starts the WebAuthn registration ceremony
func beginRegistration(c *gin.Context) {
	var req BeginRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	credentialCreation, err := webauthnService.BeginRegistration(req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, credentialCreation)
}

// finishRegistration completes the WebAuthn registration ceremony
func finishRegistration(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	credential, err := webauthnService.CompleteRegistration(username, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Registration completed successfully",
		"credential_id": credential.ID,
	})
}

// beginAuthentication starts the WebAuthn authentication ceremony
func beginAuthentication(c *gin.Context) {
	var req BeginAuthenticationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	credentialAssertion, err := webauthnService.BeginAuthentication(req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, credentialAssertion)
}

// finishAuthentication completes the WebAuthn authentication ceremony
func finishAuthentication(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	credential, err := webauthnService.CompleteAuthentication(username, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Authentication completed successfully",
		"credential_id": credential.ID,
	})
}

// getUser retrieves user information
func getUser(c *gin.Context) {
	username := c.Param("username")

	user, err := webauthnService.GetUser(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":     user.Username,
		"display_name": user.DisplayName,
		"credentials":  len(user.Credentials),
	})
}
