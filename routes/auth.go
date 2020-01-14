package routes

import (
	"os"
	"time"
	"fmt"
	"net/http"
	"github.com/arifzaenal/gin-full-api/models"
	"github.com/arifzaenal/gin-full-api/config"
	"github.com/gin-gonic/gin"
	"github.com/danilopolani/gocialite/structs"
	"github.com/dgrijalva/jwt-go"
)

// tmp checktoken
func CheckToken(c *gin.Context) {
	c.JSON(200, gin.H{
		"msg" : "success login",
	})
}

// Redirect to correct oAuth URL
func RedirectHandler(c *gin.Context) {
	// Retrieve provider from route
	provider := c.Param("provider")

	// In this case we use a map to store our secrets, but you can use dotenv or your framework configuration
	// for example, in revel you could use revel.Config.StringDefault(provider + "_clientID", "") etc.
	providerSecrets := map[string]map[string]string{
		"github": {
			"clientID":     os.Getenv("CLIENT_ID_GH"),
			"clientSecret": os.Getenv("CLIENT_SECRET_GH"),
			"redirectURL":  os.Getenv("AUTH_REDIRECT_URL") + "/github/callback",
		},
		"google": {
			"clientID":     os.Getenv("CLIENT_ID_G"),
			"clientSecret": os.Getenv("CLIENT_SECRET_G"),
			"redirectURL":  os.Getenv("AUTH_REDIRECT_URL") + "/google/callback",
		},
	}

	providerScopes := map[string][]string{
		"github":   []string{"public_repo"},
		"google": []string{},
	}

	providerData := providerSecrets[provider]
	actualScopes := providerScopes[provider]
	authURL, err := config.Gocial.New().
		Driver(provider).
		Scopes(actualScopes).
		Redirect(
			providerData["clientID"],
			providerData["clientSecret"],
			providerData["redirectURL"],
		)

	// Check for errors (usually driver not valid)
	if err != nil {
		c.Writer.Write([]byte("Error: " + err.Error()))
		return
	}

	// Redirect with authURL
	c.Redirect(http.StatusFound, authURL)
}

// Handle callback of provider
func CallbackHandler(c *gin.Context) {
	// Retrieve query params for state and code
	state := c.Query("state")
	code := c.Query("code")
	provider := c.Param("provider")

	// Handle callback and check for errors
	user, _, err := config.Gocial.Handle(state, code)
	if err != nil {
		c.Writer.Write([]byte("Error: " + err.Error()))
		return
	}

	var newUser  = getOrRegisterUser(provider, user)
	var newToken = createToken(&newUser)

	c.JSON(200, gin.H{
		"data" 		: newUser,
		"token"		: newToken,
		"message" 	: "berhasil login",
	})
}

func getOrRegisterUser(provider string, user *structs.User) models.User {
	var userData models.User

	config.DB.Where("provider = ? AND social_id = ?", provider, user.ID).First(&userData)

	if userData.ID == 0 {
		newUser := models.User {
			FullName 	: user.FullName,
			Email		: user.Email,
			SocialId 	: user.ID,
			Provider 	: provider,
			Avatar 		: user.Avatar,
		}
		config.DB.Create(&newUser)
		return newUser
	}else{
		return userData
	}
}

func createToken(user *models.User) string {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id"   : user.ID,
		"user_role" : user.Role,
		"exp"		: time.Now().AddDate(0,0,7).Unix(),
		"iat": time.Now().Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		fmt.Println(err)
	}

	return tokenString
	// fmt.Println(tokenString, err)
}