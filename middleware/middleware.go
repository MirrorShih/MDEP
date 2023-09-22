package middleware

import (
	"MDEP/models"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

var ErrOAuthTokenNotFound = errors.New("OAuth token not found")

func GitHubAPIMiddleware(c *gin.Context) {
	ac := oauth2.Config{
		ClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRECT"),
		RedirectURL:  os.Getenv("GITHUB_OAUTH_REDIRECT_URL"),
		Scopes:       []string{"read:user", "repo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	token, err := getOAuthToken(c)
	// Use the token to make authenticated requests to GitHub API
	client := ac.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"GitHub API request failed": err.Error()})
		return
	}

	defer resp.Body.Close()
	var githubUser models.GitHubUser
	// parse the JSON response into the 'GitHubUser' struct.
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"parsing error": err.Error()})
		return
	}

	c.Set("github_user", githubUser)

	c.Next()
}

func getOAuthToken(c *gin.Context) (*oauth2.Token, error) {
	token, _ := c.Get("oauth_token")

	// change token type into *oauth2.Token
	if oauthToken, ok := token.(*oauth2.Token); ok {
		return oauthToken, nil
	}
	return nil, ErrOAuthTokenNotFound
}
