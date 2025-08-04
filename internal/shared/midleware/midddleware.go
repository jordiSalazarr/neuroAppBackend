package midleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jwtService "neuro.app.jordi/internal/shared/jwt"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		// Set example variable
		c.Set("example", "12345")

		// before request

		c.Next()

		// after request
		latency := time.Since(t)
		log.Print(latency)

		// access the status we are sending
		status := c.Writer.Status()
		log.Println(status)
	}
}

func ExtractJWTFromRequest(jwtSvc *jwtService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token JWT requerido"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtSvc.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token JWT inválido"})
			return
		}

		c.Set("id", claims.Id)

		c.Next()
	}
}
func GetUserIdFromRequest(c *gin.Context) (string, bool) {
	id, exists := c.Get("id")
	if !exists {
		return "", false
	}
	userId, ok := id.(string)
	if !ok {
		return "", false
	}
	return userId, true
}
