package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"t3sesame/internal/models"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthHandler struct {
	userService  *models.UserService
	googleConfig *oauth2.Config
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func NewOAuthHandler(userService *models.UserService) *OAuthHandler {
	googleConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("BASE_URL") + "/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &OAuthHandler{
		userService:  userService,
		googleConfig: googleConfig,
	}
}

func (h *OAuthHandler) GoogleLogin(c echo.Context) error {
	state := generateRandomState()

	// Store state in session for security
	sess, _ := session.Get("session", c)
	sess.Values["oauth_state"] = state
	sess.Save(c.Request(), c.Response())

	url := h.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *OAuthHandler) GoogleCallback(c echo.Context) error {
	// Verify state
	sess, _ := session.Get("session", c)
	storedState, ok := sess.Values["oauth_state"].(string)
	if !ok || storedState != c.QueryParam("state") {
		return c.String(http.StatusBadRequest, "Invalid state parameter")
	}

	// Exchange code for token
	code := c.QueryParam("code")
	token, err := h.googleConfig.Exchange(context.Background(), code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to exchange token")
	}

	// Get user info from Google
	client := h.googleConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to get user info")
	}
	defer resp.Body.Close()

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to decode user info")
	}

	// Check if user exists, create if not
	user, err := h.userService.GetUserByEmail(googleUser.Email)
	if err != nil {
		// User doesn't exist, create new user
		user, err = h.userService.CreateGoogleUser(googleUser.Name, googleUser.Email, googleUser.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to create user")
		}
	}

	// Set session
	sess.Values["user_id"] = user.ID
	sess.Values["username"] = user.Username
	delete(sess.Values, "oauth_state") // Clean up
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

func generateRandomState() string {
	// Simple state generation - in production, use crypto/rand
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
