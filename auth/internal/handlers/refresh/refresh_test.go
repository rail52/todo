package refresh

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

func TestRefreshHandler(t *testing.T) {
	var (
		validRefreshToken   = "valid.refresh.token"
		blacklistedToken    = "blacklisted.token"
		errorParseToken     = "error.parse.token"
		invalidTypeToken    = "invalid.type.token"
		failGenerateAcc     = "fail.generate.access"
		failGenerateRefresh = "fail.generate.refresh"
	)
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
				r.On("IsBlackListed", validRefreshToken).Return(false, nil)

				tm.On("ParseToken", validRefreshToken).Return(
					&token.Claims{
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						},
						TokenType: "refresh",
						UserID:    123,
						Role:      "user",
					},
					nil,
				)

				tm.On("GenerateToken", int64(123), "user", mock.AnythingOfType("time.Duration"), "access").
					Return("new_access_token", nil)
				tm.On("GenerateToken", int64(123), "user", mock.AnythingOfType("time.Duration"), "refresh").
					Return("new_refresh_token", nil)

				// ... если затем AddToBlackList, то тоже настраиваем:
				r.On("AddToBlackList", validRefreshToken, anyPositiveInt64()).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: map[string]interface{}{
				"access_token":  "new_access_token",
				"refresh_token": "new_refresh_token",
			},
		},
		{
			name: "Invalid JSON",
			request: requests.RFToken{
				RefreshToken: "not used, because we'll break JSON manually",
			},
			mockSetup:      func(r *mocks.RedisRepo, tm *mocks.TokenMn) {},
			expectedStatus: http.StatusBadRequest,
			expectedResp: map[string]interface{}{
				"error":  "invalid request",
				"status": "ERROR",
			},
		},
		{
			name: "Validation Error (empty token)",
			request: requests.RFToken{
				RefreshToken: "",
			},
			mockSetup:      func(r *mocks.RedisRepo, tm *mocks.TokenMn) {},
			expectedStatus: http.StatusBadRequest,
			expectedResp: map[string]interface{}{
				"error":  "invalid request",
				"status": "ERROR",
			},
		},
		{
			name: "Token Already Blacklisted",
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
			name: "Redis Check Error",
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
			name: "Parse Token Error",
			request: requests.RFToken{
				RefreshToken: errorParseToken,
			},
			mockSetup: func(r *mocks.RedisRepo, tm *mocks.TokenMn) {
				r.On("IsBlackListed", errorParseToken).Return(false, nil)
				tm.On("ParseToken", errorParseToken).Return(nil, errors.New("parse error"))
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
						TokenType: "access", // неправильный тип
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
			name: "Failed to Generate Access Token",
			request: requests.RFToken{
				RefreshToken: failGenerateAcc,
			},
			mockSetup: func(r *mocks.RedisRepo, tm *mocks.TokenMn) {
				r.On("IsBlackListed", failGenerateAcc).Return(false, nil)
				tm.On("ParseToken", failGenerateAcc).Return(
					&token.Claims{
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						},
						TokenType: "refresh",
						UserID:    123,
						Role:      "user",
					},
					nil,
				)
				// Ошибка при генерации access токена
				tm.On("GenerateToken", int64(123), "user", mock.AnythingOfType("time.Duration"), "access").
					Return("", errors.New("generate access token error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp: map[string]interface{}{
				"error":  "server error",
				"status": "ERROR",
			},
		},
		{
			name: "Failed to Generate Refresh Token",
			request: requests.RFToken{
				RefreshToken: failGenerateRefresh,
			},
			mockSetup: func(r *mocks.RedisRepo, tm *mocks.TokenMn) {
				r.On("IsBlackListed", failGenerateRefresh).Return(false, nil)
				tm.On("ParseToken", failGenerateRefresh).Return(
					&token.Claims{
						RegisteredClaims: jwt.RegisteredClaims{
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						},
						TokenType: "refresh",
						UserID:    123,
						Role:      "user",
					},
					nil,
				)
				// Сначала генерируем access, ок
				tm.On("GenerateToken", int64(123), "user", mock.AnythingOfType("time.Duration"), "access").
					Return("new_access_token", nil)
				// Ошибка при генерации refresh
				tm.On("GenerateToken", int64(123), "user", mock.AnythingOfType("time.Duration"), "refresh").
					Return("", errors.New("generate refresh token error"))
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

			handler := RefreshTokens(
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
				"/auth/refresh",
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
