package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func Auth(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		tknStr := c.Request.Header.Get("token")
		if tknStr == "" {
			c.Status(http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				c.Status(http.StatusUnauthorized)
				return
			}
			c.Status(http.StatusBadRequest)
			return
		}
		if !tkn.Valid {
			c.Status(http.StatusUnauthorized)
			return
		}

		c.Set("claims", claims)

		c.Next()
	}
}
