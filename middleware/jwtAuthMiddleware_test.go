package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func buildSignedToken(t *testing.T, claims jwt.MapClaims, secret string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenString
}

func TestJWTAuthMiddlewareSetsContextAndContinues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")

	tokenString := buildSignedToken(t, jwt.MapClaims{
		"email":   "member@example.com",
		"user_id": "507f1f77bcf86cd799439011",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}, "test-secret")

	router := gin.New()
	router.Use(JWTAuthMiddleware())
	router.GET("/secure", func(c *gin.Context) {
		email, exists := c.Get("user_email")
		if !exists {
			t.Fatalf("user_email missing from context")
		}
		if email != "member@example.com" {
			t.Fatalf("expected user_email member@example.com, got %v", email)
		}

		userID, exists := c.Get("user_id")
		if !exists {
			t.Fatalf("user_id missing from context")
		}
		if userID != "507f1f77bcf86cd799439011" {
			t.Fatalf("expected user_id 507f1f77bcf86cd799439011, got %v", userID)
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestJWTAuthMiddlewareRejectsMissingOrInvalidAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")

	t.Run("missing header", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTAuthMiddleware())
		router.GET("/secure", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d, got %d body=%s", http.StatusUnauthorized, resp.Code, resp.Body.String())
		}
	})

	t.Run("non bearer scheme", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTAuthMiddleware())
		router.GET("/secure", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		req.Header.Set("Authorization", "Token invalid")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d, got %d body=%s", http.StatusUnauthorized, resp.Code, resp.Body.String())
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTAuthMiddleware())
		router.GET("/secure", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d, got %d body=%s", http.StatusUnauthorized, resp.Code, resp.Body.String())
		}
	})
}

func TestJWTAuthMiddlewareRejectsExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")

	tokenString := buildSignedToken(t, jwt.MapClaims{
		"email":   "member@example.com",
		"user_id": "507f1f77bcf86cd799439011",
		"exp":     time.Now().Add(-time.Hour).Unix(),
	}, "test-secret")

	router := gin.New()
	router.Use(JWTAuthMiddleware())
	router.GET("/secure", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d body=%s", http.StatusUnauthorized, resp.Code, resp.Body.String())
	}
}

func TestInferObjectActionInfersObjectAndAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(InferObjectAction())
	router.GET("/exercises/:id", func(c *gin.Context) {
		obj, exists := c.Get("inferred_object")
		if !exists {
			t.Fatal("inferred_object missing")
		}
		if obj != "exercises" {
			t.Fatalf("expected inferred_object exercises, got %v", obj)
		}

		action, exists := c.Get("inferred_action")
		if !exists {
			t.Fatal("inferred_action missing")
		}
		if action != "read" {
			t.Fatalf("expected inferred_action read, got %v", action)
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/exercises/123", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAuthorizeAllowsMatchingSubjectAndRejectsMissingIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	enforcer, err := casbin.NewEnforcer("../config/rbac_model.conf")
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}
	if _, err := enforcer.AddPolicy("member@example.com", "exercises", "read"); err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	t.Run("allowed", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_email", "member@example.com")
			c.Set("inferred_object", "exercises")
			c.Set("inferred_action", "read")
			c.Next()
		})
		router.Use(Authorize(enforcer, nil))
		router.GET("/secure", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
		}
	})

	t.Run("missing identity", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("inferred_object", "exercises")
			c.Set("inferred_action", "read")
			c.Next()
		})
		router.Use(Authorize(enforcer, nil))
		router.GET("/secure", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d, got %d body=%s", http.StatusUnauthorized, resp.Code, resp.Body.String())
		}
	})
}

func TestAuthMiddlewareBehavesAsFailClosedWrapper(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing header aborts", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth(func(c *gin.Context) {
			c.Next()
		}))
		router.GET("/secure", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d, got %d body=%s", http.StatusUnauthorized, resp.Code, resp.Body.String())
		}
	})

	t.Run("valid auth continues", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth(func(c *gin.Context) {
			c.Set("auth_ok", true)
			c.Next()
		}))
		router.GET("/secure", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/secure", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
		}
	})
}
