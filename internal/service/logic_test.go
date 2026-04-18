package service

import (
	"context"
	"testing"
	"wishlist-api/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

// mockRepo реализует интерфейс repository.Store
type mockRepo struct {
	createUserFunc     func(u *domain.User) error
	getUserByEmailFunc func(email string) (*domain.User, error)
}

func (m *mockRepo) CreateUser(ctx context.Context, u *domain.User) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(u)
	}
	return nil
}

func (m *mockRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(email)
	}
	return nil, nil
}

// заглушки для repository.Store
func (m *mockRepo) CreateWishlist(ctx context.Context, w *domain.Wishlist) error { return nil }
func (m *mockRepo) GetWishlists(ctx context.Context, uid int) ([]domain.Wishlist, error) {
	return nil, nil
}
func (m *mockRepo) DeleteWishlist(ctx context.Context, uid, id int) error { return nil }
func (m *mockRepo) AddItem(ctx context.Context, i *domain.Item) error     { return nil }
func (m *mockRepo) DeleteItem(ctx context.Context, uid, itemId int) error { return nil }
func (m *mockRepo) GetFullList(ctx context.Context, t string) (*domain.Wishlist, []domain.Item, error) {
	return nil, nil, nil
}
func (m *mockRepo) AtomicBook(ctx context.Context, t string, id int) error { return nil }

// --- ТЕСТЫ ---

func TestRegister(t *testing.T) {
	repo := &mockRepo{}
	svc := New(repo, "secret-key")

	called := false
	repo.createUserFunc = func(u *domain.User) error {
		called = true
		if u.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", u.Email)
		}
		if len(u.PasswordHash) < 10 {
			t.Error("password was not hashed")
		}
		return nil
	}

	err := svc.Register(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !called {
		t.Error("repository.CreateUser was not called")
	}
}

func TestLogin(t *testing.T) {
	repo := &mockRepo{}
	svc := New(repo, "secret-key")

	password := "123456"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	repo.getUserByEmailFunc = func(email string) (*domain.User, error) {
		return &domain.User{
			ID:           1,
			Email:        email,
			PasswordHash: string(hash),
		}, nil
	}

	token, err := svc.Login(context.Background(), "test@example.com", password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &mockRepo{}
	svc := New(repo, "secret-key")

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-pass"), bcrypt.DefaultCost)

	repo.getUserByEmailFunc = func(email string) (*domain.User, error) {
		return &domain.User{
			ID:           1,
			Email:        email,
			PasswordHash: string(hash),
		}, nil
	}

	_, err := svc.Login(context.Background(), "test@example.com", "wrong-pass")
	if err == nil {
		t.Error("expected error for wrong password, but got nil")
	}
}
