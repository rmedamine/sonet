package utils

import (
	"net/http"
	"strings"
)

func GeTCookie(name string, r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	} else if authHeader != "" {
		return authHeader
	}
	session, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return session.Value
}

func GetSessionCookie(r *http.Request) string {
	return GeTCookie("session", r)
}
