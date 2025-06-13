package handlers

import (
	"database/sql"
	"net/http"
	"t3sesame/internal/models"
	"t3sesame/internal/templates"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	userService *models.UserService
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		userService: models.NewUserService(db),
	}
}

func (h *AuthHandler) ShowLogin(c echo.Context) error {
	return templates.LoginPage().Render(c.Request().Context(), c.Response().Writer)
}

func (h *AuthHandler) ShowRegister(c echo.Context) error {
	return templates.RegisterPage().Render(c.Request().Context(), c.Response().Writer)
}

func (h *AuthHandler) Register(c echo.Context) error {
	username := c.FormValue("username")
	email := c.FormValue("email")
	password := c.FormValue("password")

	if username == "" || email == "" || password == "" {
		return templates.AuthError("All fields are required").
			Render(c.Request().Context(), c.Response().Writer)
	}

	user, err := h.userService.CreateUser(username, email, password)
	if err != nil {
		return templates.AuthError("Registration failed. Email or username may already exist").
			Render(c.Request().Context(), c.Response().Writer)
	}

	// Set session
	sess, _ := session.Get("session", c)
	sess.Values["user_id"] = user.ID
	sess.Values["username"] = user.Username
	sess.Save(c.Request(), c.Response())

	// Use HX-Redirect header instead
	c.Response().Header().Set("HX-Redirect", "/dashboard")
	return templates.AuthSuccessSimple("Registration successful! Redirecting...").
		Render(c.Request().Context(), c.Response().Writer)
}

func (h *AuthHandler) Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" {
		return templates.AuthError("Email and password are required").
			Render(c.Request().Context(), c.Response().Writer)
	}

	user, err := h.userService.GetUserByEmail(email)
	if err != nil {
		return templates.AuthError("Invalid email or password").
			Render(c.Request().Context(), c.Response().Writer)
	}

	if !h.userService.ValidatePassword(user, password) {
		return templates.AuthError("Invalid email or password").
			Render(c.Request().Context(), c.Response().Writer)
	}

	// Set session
	sess, _ := session.Get("session", c)
	sess.Values["user_id"] = user.ID
	sess.Values["username"] = user.Username
	sess.Save(c.Request(), c.Response())

	// Use HX-Redirect header instead
	c.Response().Header().Set("HX-Redirect", "/dashboard")
	return templates.AuthSuccessSimple("Login successful! Redirecting...").
		Render(c.Request().Context(), c.Response().Writer)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Values = make(map[interface{}]interface{})
	sess.Options = &sessions.Options{MaxAge: -1}
	sess.Save(c.Request(), c.Response())

	c.Response().Header().Set("HX-Redirect", "/login")
	return c.NoContent(http.StatusOK)
}

func (h *AuthHandler) Dashboard(c echo.Context) error {
	sess, _ := session.Get("session", c)
	username, ok := sess.Values["username"].(string)
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	return templates.Dashboard(username).Render(c.Request().Context(), c.Response().Writer)
}
