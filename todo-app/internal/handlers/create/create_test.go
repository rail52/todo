package create

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"todo-app/internal/handlers/create/mocks"
	"todo-app/internal/lib/logger/slogdiscard"

	"todo-app/internal/domain/requests"

	"github.com/rail52/myprojects/dbpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateTask(t *testing.T) {
	// Создаем моки и логгер один раз для всех тестов
	mockCreator := mocks.NewCreator(t)
	mockProducer := mocks.NewKafkaProducer(t)
	log := slogdiscard.NewDiscardLogger()

	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func()
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:        "Empty request body",
			requestBody: "",
			mockSetup: func() {
				// Никаких вызовов не ожидаем
			},
			expectedStatus: http.StatusBadRequest,
			expectedJSON:  "request body is empty",
		},
		{
			name:        "Missing title",
			requestBody: `{"content": "test content"}`,
			mockSetup: func() {
				// Никаких вызовов не ожидаем
			},
			expectedStatus: http.StatusBadRequest,
			expectedJSON:  "Title or Content in request Body is empty or invalid",
		},
		{
			name:        "Valid request",
			requestBody: `{"title": "test", "content": "content"}`,
			mockSetup: func() {
				mockCreator.On("CreateTask", mock.Anything, &dbpb.CreateTaskRequest{
					Title:"test",
					Content: "content",
				}).Return(&dbpb.Task{
					Id:      int64(123),
					Title:   "test",
					Content: "content",
				}, nil)

				mockProducer.On("SendApiEvent", &requests.ApiRequest{Action: "created"}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedJSON:  `{"id":123,"title":"test","content":"content"}`,
		},
		{
			name:        "GRPC error",
			requestBody: `{"title": "test", "content": "content"}`,
			mockSetup: func() {
				mockCreator.On("CreateTask", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("grpc error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedJSON:  "grpc error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сбрасываем моки перед каждым тестом
			mockCreator.ExpectedCalls = nil
			mockProducer.ExpectedCalls = nil

			// Настраиваем моки
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// Создаем запрос
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tt.requestBody))
			require.NoError(t, err)

			// Создаем ResponseRecorder
			rr := httptest.NewRecorder()

			// Вызываем обработчик
			handler := CreateTask(log, mockCreator, mockProducer)
			handler.ServeHTTP(rr, req)

			// Проверяем статус код
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Проверяем тело ответа
			if tt.expectedJSON != "" {
				assert.Equal(t, tt.expectedJSON, strings.ReplaceAll(rr.Body.String(), "\n",""))
			}

			// Проверяем вызовы моков
			mockCreator.AssertExpectations(t)
			mockProducer.AssertExpectations(t)
		})
	}
}