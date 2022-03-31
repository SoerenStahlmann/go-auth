package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/soerenstahlmann/go-auth/ent/user"
	"github.com/soerenstahlmann/go-auth/middleware"
)

// Create a struct to read the username and password from the request body
type Credentials struct {
	Password string `json:"password" binding:"required"`
	Username string `json:"username" binding:"required"`
}

func (s *server) register(c *gin.Context) {

	var newUser Credentials

	err := c.BindJSON(&newUser)
	if err != nil {

		c.JSON(http.StatusBadRequest, Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}

	user, err := s.client.User.Create().
		SetUsername(newUser.Username).
		SetPassword(newUser.Password).
		Save(c.Request.Context())
	if err != nil {
		s.logger.Infof("could not create user %s: %v", newUser.Username, err)
		c.JSON(http.StatusInternalServerError, Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		})
		return
	}

	s.logger.Infow("register new user", "user", user.Username, "id", user.ID)

	c.JSON(http.StatusOK, Response{
		Status: http.StatusOK,
		Result: user,
	})
}

func (s *server) login(c *gin.Context) {

	var creds Credentials
	// Get the JSON body and decode into credentials
	err := c.BindJSON(&creds)
	if err != nil {
		s.logger.Infof("could not bind params to struct: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}

	user, err := s.client.User.Query().
		Where(user.Username(creds.Username)).
		Only(c.Request.Context())
	if err != nil {
		s.logger.Infof("could not find user %s in database: %v", creds.Username, err)
		c.JSON(http.StatusBadRequest, Response{
			Status: http.StatusBadRequest,
			Error:  err.Error(),
		})
		return
	}

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if user.Password != creds.Password {
		c.JSON(http.StatusUnauthorized, Response{
			Status: http.StatusUnauthorized,
			Error:  "wrong password for user",
		})
		return
	}

	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	expirationTime := time.Now().Add(30 * time.Second)
	// Create the JWT claims, which includes the username and expiry time
	claims := &middleware.Claims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		c.JSON(http.StatusInternalServerError, Response{
			Status: http.StatusInternalServerError,
			Error:  "wrong password for user",
		})
		return
	}

	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	c.Header("token", tokenString)
	c.Status(http.StatusOK)

}

func (s *server) refresh(c *gin.Context) {

	contextClaims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, Response{
			Status: http.StatusUnauthorized,
			Error:  "could not extract claims from context",
		})
		return
	}
	claims := contextClaims.(*middleware.Claims)

	// We ensure that a new token is not issued until enough time has elapsed
	// In this case, a new token will only be issued if the old token is within
	// 30 seconds of expiry. Otherwise, return a bad request status
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		c.JSON(http.StatusBadRequest, Response{
			Status: http.StatusBadRequest,
			Error:  "token stil valid for more than 30s",
		})
		return
	}

	// Now, create a new token for the current use, with a renewed expiration time
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		})
		return
	}

	// Set the new token as the users `token` cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	c.Header("token", tokenString)
	c.Status(http.StatusOK)
}
