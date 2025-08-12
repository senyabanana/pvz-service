package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/senyabanana/pvz-service/internal/infrastructure/jwtutil"
)

const testSecret = "secret"

func generateToken(t *testing.T, userID, role, secret string) string {
	token, err := jwtutil.GenerateToken(userID, role, secret, time.Hour)
	assert.NoError(t, err)
	return token
}

func performRequest(t *testing.T, r http.Handler, token string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", "/", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		token        string
		allowedRoles []string
		wantStatus   int
	}{
		{
			name:         "valid token with allowed role",
			token:        generateToken(t, "123", "moderator", testSecret),
			allowedRoles: []string{"moderator"},
			wantStatus:   http.StatusOK,
		},
		{
			name:         "valid token with disallowed role",
			token:        generateToken(t, "123", "client", testSecret),
			allowedRoles: []string{"moderator"},
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "invalid token format",
			token:        "invalidtoken",
			allowedRoles: []string{"moderator"},
			wantStatus:   http.StatusUnauthorized,
		},
		{
			name:         "missing Authorization header",
			token:        "",
			allowedRoles: []string{"moderator"},
			wantStatus:   http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(RequireRole(testSecret, logrus.New(), tt.allowedRoles...))
			r.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			tokenStr := tt.token
			if strings.HasPrefix(tt.name, "invalid token format") {
				tokenStr = "invalid"
			}
			w := performRequest(t, r, tokenStr)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
