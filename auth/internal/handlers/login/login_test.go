package login

import (
	"auth/internal/domain/model"
	"auth/internal/domain/response"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"auth/internal/handlers/mocks"
	"auth/internal/lib/logger/slogdiscard"
	"auth/internal/storage/db"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginHandler(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctPassword"), bcrypt.DefaultCost)
	cases := []struct {
		name           string
		email          string
		password       string
		mockUser       *model.User
		mockError      error
		tokenError     error
		expectedStatus int
		respError      string
	}{
		{
			name:     "Success",
			email:    "valid@mail.com",
			password: "correctPassword",
			mockUser: &model.User{
				ID:       1,
				Email:    "valid@mail.com",
				Password: string(hashedPassword),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User not found",
			email:          "invalid@mail.com",
			password:       "correctPassword",
			mockUser:       nil,
			mockError:      db.ErrUserNotFound,
			expectedStatus: http.StatusUnauthorized,
			respError:      "status invalid credentials",
		},
		{
			name:     "Wrong password",
			email:    "valid@mail.com",
			password: "wrongPassword",
			mockUser: &model.User{
				Email:    "valid@mail.com",
				Password: string(hashedPassword),
			},
			expectedStatus: http.StatusUnauthorized,
			respError:      "invalid credentials",
		},
		{
			name:     "Token generate error",
			email:    "valid@mail.com",
			password: "correctPassword",
			mockUser: &model.User{
				ID:       1,
				Email:    "valid@mail.com",
				Password: string(hashedPassword),
			},
			tokenError:     errors.New("token error"),
			expectedStatus: http.StatusInternalServerError,
			respError:      "failed to generate access token",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			tokenMn := mocks.NewTokenMn(t)
			authMock := mocks.NewUserCreater(t)

			authMock.On("GetByEmail", tc.email).Return(tc.mockUser, tc.mockError)

			if tc.expectedStatus == http.StatusOK || tc.tokenError != nil {
				if tc.mockUser != nil {
					tokenMn.On("GenerateToken", tc.mockUser.ID,
						time.Duration(AccessTokenTTL)*time.Second, "access").
						Return("access_token", tc.tokenError)

					if tc.tokenError == nil {
						tokenMn.On("GenerateToken", tc.mockUser.ID, 
							time.Duration(RefreshTokenTTL)*time.Second, "refresh").
							Return("refresh_token", nil)
					}
				}
			}

			handler := Login(
				slogdiscard.NewDiscardLogger(),
				authMock,
				tokenMn,
			)

			body := fmt.Sprintf(
				`{"email": "%s", "password": "%s"}`,
				tc.email, tc.password,
			)

			req, err := http.NewRequest(
				http.MethodPost,
				"/auth/login",
				bytes.NewBufferString(body),
			)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)

			if tc.respError != "" {
				var resp response.Response
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				require.Equal(t, tc.respError, resp.Error)
			}

			if tc.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				require.Contains(t, resp, "access_token")
				require.Contains(t, resp, "refresh_token")
				require.Equal(t, tc.mockUser.Role, resp["role"])
			}

			authMock.AssertExpectations(t)
			tokenMn.AssertExpectations(t)

		})
	}
}
