package middleware

import (
	"fmt"
	"time"

	"gym-api/m/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var cfg = config.Load()

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Printf("auth request required")
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header required"})
			return
		}
		tokenString := authHeader[len("Bearer "):]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return cfg.JWTKey, nil
		})
		if err != nil || !token.Valid {
			fmt.Printf("invalid token")
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			exp, ok := claims["exp"].(float64)
			if !ok || int64(exp) < jwt.NewNumericDate(time.Now()).Unix() {
				fmt.Printf("token expired")
				c.AbortWithStatusJSON(401, gin.H{"error": "Token expired"})
				return
			}
			c.Set("user_email", claims["email"])
		} else {
			fmt.Printf("invalid token claims")
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token claims"})
			return
		}
	}
}
