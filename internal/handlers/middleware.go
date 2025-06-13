package handlers

import (
    "net/http"
    
    "github.com/labstack/echo-contrib/session"
    "github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        sess, _ := session.Get("session", c)
        
        if sess.Values["user_id"] == nil {
            return c.Redirect(http.StatusSeeOther, "/login")
        }
        
        return next(c)
    }
}

func GuestMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        sess, _ := session.Get("session", c)
        
        if sess.Values["user_id"] != nil {
            return c.Redirect(http.StatusSeeOther, "/dashboard")
        }
        
        return next(c)
    }
}