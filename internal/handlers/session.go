package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const sessionCookieName = "razlom_session"
const sessionUserIDKey = "authenticatedUserID"
const sessionLifetime = 7 * 24 * time.Hour

type SessionManager struct {
	secret []byte
}

func NewSessionManager(secret string) (*SessionManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("session secret must be at least 32 bytes")
	}
	return &SessionManager{secret: []byte(secret)}, nil
}

func (s *SessionManager) SetCookie(c *gin.Context, userID uint) {
	expiresAt := time.Now().Add(sessionLifetime).Unix()
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%d:%d", userID, expiresAt)))
	signature := s.sign(payload)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    payload + "." + signature,
		Path:     "/",
		MaxAge:   int(sessionLifetime.Seconds()),
		HttpOnly: true,
		Secure:   c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *SessionManager) ClearCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *SessionManager) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie(sessionCookieName)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userID, err := s.userIDFromToken(cookie.Value)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
			c.Abort()
			return
		}
		c.Set(sessionUserIDKey, userID)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) uint {
	userID, _ := c.Get(sessionUserIDKey)
	id, _ := userID.(uint)
	return id
}

func (s *SessionManager) sign(value string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *SessionManager) userIDFromToken(token string) (uint, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || len(token) > 256 {
		return 0, errors.New("invalid session token")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, errors.New("invalid session signature")
	}
	expected, err := base64.RawURLEncoding.DecodeString(s.sign(parts[0]))
	if err != nil || !hmac.Equal(signature, expected) {
		return 0, errors.New("invalid session signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, errors.New("invalid session payload")
	}
	fields := strings.SplitN(string(payload), ":", 2)
	if len(fields) != 2 {
		return 0, errors.New("invalid session payload")
	}
	userID, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil || userID == 0 {
		return 0, errors.New("invalid session user")
	}
	expiresAt, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || expiresAt <= time.Now().Unix() {
		return 0, errors.New("expired session")
	}
	return uint(userID), nil
}
