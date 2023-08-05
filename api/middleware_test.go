package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"simplebank/token"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
	t *testing.T,
	tokenMaker *token.PasetoMaker,
	req *http.Request,
	authType string,
	username string,
	duration time.Duration,
) {
	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	authorizationHeader := fmt.Sprintf("%s %s", authType, token)
	req.Header.Set("authorization", authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	testCase := []struct {
		name          string
		setupAuth     func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request)
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name: "UnsupposedAuthorization",
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, "unsupposed", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, tokenMaker *token.PasetoMaker, req *http.Request) {
				addAuthorization(t, tokenMaker, req, authTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)

			authPath := "/auth"
			server.router.GET(
				authPath,
				authMiddleware(server.tokenMaker),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, server.tokenMaker, req)

			server.router.ServeHTTP(w, req)
			tc.checkResponse(t, w)
		})
	}
}
