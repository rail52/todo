package logout

import (
	"auth/internal/domain/requests"
	"auth/internal/handlers/mocks"
	"auth/internal/lib/logger/slogdiscard"
	"auth/pkg/token"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLogoutHandler(t *testing.T) {
	validRefreshToken := "valid.refresh.token"
	expiredRefreshToken := "expired.refresh.token"
	invalidTypeToken := "invalid.type.token"
	blacklistedToken := "blacklisted.token"

	cases := []struct {
		name           string
		request        requests.RFToken
		mockSetup      func(*mocks.RedisRepo, *mocks.TokenMn)
		expectedStatus int
		expectedResp   map[string]interface{}
	}{
		{
			name: "Success",
			request: requests.RFToken{
				RefreshToken: validRefreshToken,
			},
			mockSetup: func(r *mocks.RedisRepo, tm *mocks.TokenMn) {
				r.On("IsBlackListed", validRefreshToken).Return(false, nil) // <-- ВАЖНО!
				// ... если вызывается ParseToken, то тоже настраиваем:
				tm.On("ParseToken", validRefreshToken).Return(
					&token.Claims{
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						},
						TokenType: "refresh",
					},
					nil,
				)
				// ... если затем AddToBlackList, то тоже настраиваем:
				r.On("AddToBlackList", validRefreshToken, anyPositiveInt64()).Return(nil)
			},
			expectedStatus: http.StatusOK,
			// При успехе у вас возвращается "status":"logged_out"
			expectedResp: map[string]interface{}{
				"status": "logged_out",
			},
		},
		{
			name: "Invalid JSON",
			request: requests.RFToken{
				RefreshToken: "invalid json",
			},
			mockSetup:      func(r *mocks.RedisRepo, tm *mocks.TokenMn) {},
			expectedStatus: http.StatusBadRequest,
			// Меняем на то, что реально возвращает ваш код
			expectedResp: map[string]interface{}{
				"error":  "invalid request",
				"status": "ERROR",
			},
		},
		{
			name: "Validation Error",
			request: requests.RFToken{
				RefreshToken: "", // пустая строка
			},
			mockSetup:      func(r *mocks.RedisRepo, tm *mocks.TokenMn) {},
			expectedStatus: http.StatusBadRequest,
			expectedResp: map[string]interface{}{
				"error":  "invalid request",
				"status": "ERROR",
			},
		},
		{
			name: "Already Blacklisted",
			request: requests.RFToken{
				RefreshToken: blacklistedToken,
			},
			mockSetup: func(r *mocks.RedisRepo, tm *mocks.TokenMn) {
				r.On("IsBlackListed", blacklistedToken).Return(true, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedResp: map[string]interface{}{
				"error":  "token revoked",
				"status": "ERROR",
			},
		},
		{
			name: "Redis Check Errr",
			request: requests.RFToken{
				RefreshToken: validRefreshToken,
			},
			mockSetup: func(r *mocks.RedisRepo, tm *mocks.TokenMn) {
				r.On("IsBlackListed", validRefreshToken).Return(false, errors.New("redis error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: map[string]interface{}{
				"error":  "server error",
				"status": "ERROR",
			},
		},
		{
			name: "Invalid Token",
			request: requests.RFToken{
				RefreshToken: expiredRefreshToken,
			},
			mockSetup: func(r *mocks.RedisRepo, tm *mocks.TokenMn) {
				r.On("IsBlackListed", expiredRefreshToken).Return(false, nil)
				tm.On("ParseToken", expiredRefreshToken).Return(nil, errors.New("parse error"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: map[string]interface{}{
				"error":  "invalid token",
				"status": "ERROR",
			},
		},
		{
			name: "Invalid Token Type",
			request: requests.RFToken{
				RefreshToken: invalidTypeToken,
			},
			mockSetup: func(r *mocks.RedisRepo, tm *mocks.TokenMn) {
				r.On("IsBlackListed", invalidTypeToken).Return(false, nil)
				tm.On("ParseToken", invalidTypeToken).Return(
					&token.Claims{
						TokenType: "access",
					},
					nil,
				)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp: map[string]interface{}{
				"error":  "invalid token type",
				"status": "ERROR",
			},
		},
		{
			name: "Redis Add Error",
			request: requests.RFToken{
				RefreshToken: validRefreshToken,
			},
			mockSetup: func(r *mocks.RedisRepo, tm *mocks.TokenMn) {
				r.On("IsBlackListed", validRefreshToken).Return(false, nil)
				tm.On("ParseToken", validRefreshToken).Return(
					&token.Claims{
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						},
						TokenType: "refresh",
					},
					nil,
				)
				r.On("AddToBlackList", validRefreshToken, anyPositiveInt64()).Return(errors.New("redis error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: map[string]interface{}{
				"error":  "server error",
				"status": "ERROR",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			redisMock := mocks.NewRedisRepo(t)
			tokenMock := mocks.NewTokenMn(t)
			tc.mockSetup(redisMock, tokenMock)

			handler := Logout(
				slogdiscard.NewDiscardLogger(),
				redisMock,
				tokenMock,
			)

			body, _ := json.Marshal(tc.request)

			if tc.name == "Invalid JSON" {
				body = []byte("{invalid json}")
			}

			req, err := http.NewRequest(
				http.MethodPost,
				"/auth/logout",
				bytes.NewBuffer(body),
			)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			require.Equal(t, tc.expectedResp, resp)

			redisMock.AssertExpectations(t)
			tokenMock.AssertExpectations(t)

		})
	}

}

func anyPositiveInt64() interface{} {
	return mock.MatchedBy(func(x int64) bool {
		return x > 0
	})
}
