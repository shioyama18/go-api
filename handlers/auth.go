package handlers

import (
	"context"
	"crypto/sha256"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-api/recipes/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		collection: collection,
		ctx:        ctx,
	}
}

// swagger:operation POST /signin auth signIn
// Login with username and password
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'401':
//	    description: Invalid credentials
func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h := sha256.New()

	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"username": user.Username,
		"password": string(h.Sum([]byte(user.Password))),
	})
	if cur.Err() != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	sessionToken := xid.New().String()
	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("token", sessionToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "User signed in"})
}

// swagger:operation POST /signup auth signUp
// Create user with username and password
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: User creation failed
func (handler *AuthHandler) SignUpHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var result bson.M
	err := handler.collection.FindOne(handler.ctx, bson.M{
		"username": user.Username,
	}).Decode(&result)

	if err != nil && err != mongo.ErrNoDocuments {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	h := sha256.New()
	handler.collection.InsertOne(handler.ctx, bson.M{
		"username": user.Username,
		"password": string(h.Sum([]byte(user.Password))),
	})

	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

// swagger:operation POST /refresh auth refresh
// Refresh token
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'401':
//	    description: Invalid credentials
func (handler *AuthHandler) RefreshHandler(c *gin.Context) {
	session := sessions.Default(c)
	sessionToken := session.Get("token")
	sessionUser := session.Get("username")
	if sessionToken == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session cookie"})
		return
	}

	sessionToken = xid.New().String()
	session.Set("username", sessionUser.(string))
	session.Set("token", sessionToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "New session issued"})
}

// swagger:operation POST /signout auth signOut
// Signing out
// ---
// responses:
//
//	'200':
//	    description: Successful operation
func (handler *AuthHandler) SignOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Signed out..."})
}

func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionToken := session.Get("token")
		if sessionToken == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "Not logged in",
			})
			c.Abort()
		}
		c.Next()
	}
}
