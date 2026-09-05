package handlers

import (
	"context"
	"errors"
	"gym-api/m/config"
	"gym-api/m/models"
	"net/http"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	return strings.Contains(email, "@") && strings.Count(email, "@") == 1
}

type AuthenticationHandler struct {
	DB             *mongo.Client
	Enforcer       *casbin.Enforcer
	GoogleVerifier  *oidc.IDTokenVerifier
	GoogleClientIDs []string
	GoogleIssuer    string
}

type googleTokenRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

func buildAuthPayload(tokenString string, roles []string) gin.H {
	if len(roles) == 0 {
		roles = []string{"member"}
	}

	roleList := make([]string, len(roles))
	copy(roleList, roles)

	isAdmin := false
	for _, role := range roleList {
		if strings.EqualFold(role, "admin") {
			isAdmin = true
			break
		}
	}

	return gin.H{
		"token":   tokenString,
		"roles":   roleList,
		"isAdmin": isAdmin,
	}
}

func buildUserProfileResponse(user models.User, roles []string) gin.H {
	if len(roles) == 0 {
		roles = []string{"member"}
	}

	roleList := make([]string, len(roles))
	copy(roleList, roles)

	isAdmin := false
	for _, role := range roleList {
		if strings.EqualFold(role, "admin") {
			isAdmin = true
			break
		}
	}

	profile := gin.H{
		"id":       user.ID.Hex(),
		"email":    user.Email,
		"provider": user.Provider,
		"roles":    roleList,
		"isAdmin":  isAdmin,
	}
	return profile
}

func (h *AuthenticationHandler) verifyGoogleToken(c *gin.Context, rawToken string) (GoogleTokenClaims, error) {
	if len(h.GoogleClientIDs) == 0 {
		return GoogleTokenClaims{}, errors.New("google client ID is not configured")
	}
	if h.GoogleIssuer == "" {
		return GoogleTokenClaims{}, errors.New("google issuer is not configured")
	}

	if h.GoogleVerifier == nil {
		return GoogleTokenClaims{}, errors.New("google token verifier is not configured")
	}
	return VerifyGoogleIDTokenWithVerifier(c.Request.Context(), h.GoogleVerifier, rawToken, h.GoogleClientIDs, h.GoogleIssuer)
}

func (h *AuthenticationHandler) upsertGoogleUser(c *gin.Context, claims GoogleTokenClaims) (models.User, error) {
	collection := h.DB.Database("gym-app").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	email := normalizeEmail(claims.Email)
	if !isValidEmail(email) {
		return models.User{}, errors.New("invalid email format from google identity")
	}

	userDoc := models.User{
		Email:           email,
		Provider:        "google",
		ProviderSubject: claims.Subject,
	}

	var existing models.User
	err := collection.FindOne(ctx, bson.M{"provider": "google", "provider_subject": claims.Subject}).Decode(&existing)
	if err == nil {
		if existing.Email != email {
			existing.Email = email
			if _, updateErr := collection.UpdateOne(ctx, bson.M{"_id": existing.ID}, bson.M{"$set": bson.M{"email": email}}); updateErr != nil {
				return models.User{}, updateErr
			}
		}
		return existing, nil
	}
	if err != mongo.ErrNoDocuments {
		return models.User{}, err
	}

	res, err := collection.InsertOne(ctx, userDoc)
	if err != nil {
		return models.User{}, err
	}
	userDoc.ID, _ = res.InsertedID.(primitive.ObjectID)
	if userDoc.ID.IsZero() {
		return models.User{}, errors.New("failed to read inserted google user id")
	}
	return userDoc, nil
}

func (h *AuthenticationHandler) Me(c *gin.Context) {
	emailValue, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User identity not found in request context"})
		return
	}

	email, ok := emailValue.(string)
	if !ok || strings.TrimSpace(email) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User email not found in request context"})
		return
	}

	collection := h.DB.Database("gym-app").Collection("users")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": normalizeEmail(email)}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	roles, err := h.Enforcer.GetRolesForUser(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	c.JSON(http.StatusOK, buildUserProfileResponse(user, roles))
}

func (h *AuthenticationHandler) Register(c *gin.Context) {
	var req googleTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims, err := h.verifyGoogleToken(c, req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Google authentication failed: " + err.Error()})
		return
	}

	user, err := h.upsertGoogleUser(c, claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = h.Enforcer.AddRoleForUser(user.Email, "member")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role to user"})
		return
	}

	roles, err := h.Enforcer.GetRolesForUser(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	payload := gin.H{"message": "Google user registered successfully"}
	for key, value := range buildAuthPayload("", roles) {
		if key == "token" {
			continue
		}
		payload[key] = value
	}

	c.JSON(http.StatusOK, payload)
}

func (h *AuthenticationHandler) Login(c *gin.Context) {
	var req googleTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims, err := h.verifyGoogleToken(c, req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Google authentication failed: " + err.Error()})
		return
	}

	user, err := h.upsertGoogleUser(c, claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	roles, err := h.Enforcer.GetRolesForUser(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	cfg := config.Load()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"email":   user.Email,
		"roles":   roles,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(cfg.JWTKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, buildAuthPayload(tokenString, roles))
}
