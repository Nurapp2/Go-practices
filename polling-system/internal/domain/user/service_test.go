package user_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"polling-system/internal/domain/user"

	"golang.org/x/crypto/bcrypt"
)


var ErrInputRequired = errors.New("email and password required")


var ErrEmailTaken = errors.New("email already taken")

// MockUserRepository -
type MockUserRepository struct {
	GetByEmailFunc func(email string) (*user.User, error)
	CreateFunc     func(u *user.User) error
	GetByIDFunc    func(id int64) (*user.User, error)
	ListFunc       func() ([]user.User, error)
	UpdateRoleFunc func(id int64, role string) error
}

func (m *MockUserRepository) GetByEmail(email string) (*user.User, error) {
	return m.GetByEmailFunc(email)
}

func (m *MockUserRepository) Create(u *user.User) error {
	return m.CreateFunc(u)
}

func (m *MockUserRepository) GetByID(id int64) (*user.User, error) {
	return m.GetByIDFunc(id)
}

func (m *MockUserRepository) List() ([]user.User, error) {
	return m.ListFunc()
}

func (m *MockUserRepository) UpdateRole(id int64, role string) error {
	return m.UpdateRoleFunc(id, role)
}

// --- Тестирование функции service.Register ---

func TestService_Register(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		password     string
		mockRepo     user.Repository
		wantErr      error // Ожидаемая ошибка
		checkCreated func(*user.User, error)
	}{
		{
			name:     "Success registration",
			email:    "new@example.com",
			password: "password123",
			mockRepo: &MockUserRepository{
				GetByEmailFunc: func(email string) (*user.User, error) {
					return nil, sql.ErrNoRows // Пользователь не существует
				},
				CreateFunc: func(u *user.User) error {
					u.ID = 1
					u.CreatedAt = time.Now()
					return nil
				},
			},
			wantErr: nil,
			checkCreated: func(u *user.User, err error) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				// Проверка хэша пароля
				if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("password123")); err != nil {
					t.Errorf("password hash verification failed: %v", err)
				}
			},
		},
		{
			name:     "Empty email or password",
			email:    "",
			password: "password123",
			mockRepo: &MockUserRepository{
				GetByEmailFunc: func(email string) (*user.User, error) {
					t.Fatal("GetByEmail should not be called")
					return nil, nil
				},
			},
			wantErr: ErrInputRequired, 
			checkCreated: func(u *user.User, err error) {
				
			},
		},
		{
			name:     "Email already taken",
			email:    "exists@example.com",
			password: "password123",
			mockRepo: &MockUserRepository{
				GetByEmailFunc: func(email string) (*user.User, error) {
					return &user.User{ID: 1, Email: email}, nil 
				},
				CreateFunc: func(u *user.User) error {
					t.Fatal("Create should not be called")
					return nil
				},
			},
			wantErr: ErrEmailTaken,
			checkCreated: func(u *user.User, err error) {
				
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := user.NewService(tt.mockRepo)
			u, err := s.Register(tt.email, tt.password)

			
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
					return
				}

				
				isSameError := errors.Is(err, tt.wantErr)
				isSameMessage := err.Error() == tt.wantErr.Error()

				if !isSameError && !isSameMessage {
					t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else if err != nil {
				
				t.Errorf("Register() unexpected error = %v", err)
				return
			}
			

			if tt.checkCreated != nil {
				tt.checkCreated(u, err)
			}
		})
	}
}
