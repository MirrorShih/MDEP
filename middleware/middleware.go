package middleware

import (
	"MDEP/controller"
	"MDEP/models"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func GitHubAPIMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, _ := controller.Store.Get(c.Request, "session1")
		tokenString, _ := session.Values["access_token"].(string)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is missing"})
			c.Abort()
			return
		}

		token := &oauth2.Token{
			AccessToken: tokenString,
		}

		client := controller.AuthController.Client(c, token)

		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"GitHub API request failed": err.Error()})
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
}
