package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"wishlist-api/internal/domain"
	"wishlist-api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Svc struct {
	repo   repository.Store
	jwtKey []byte
}

func New(r repository.Store, secret string) *Svc {
	return &Svc{repo: r, jwtKey: []byte(secret)}
}

func (s *Svc) Register(ctx context.Context, email, pass string) error {
	h, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return s.repo.CreateUser(ctx, &domain.User{Email: email, PasswordHash: string(h)})
}

func (s *Svc) Login(ctx context.Context, email, pass string) (string, error) {
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pass)); err != nil {
		return "", errors.New("auth failed")
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": u.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	return t.SignedString(s.jwtKey)
}

func (s *Svc) CreateWishlist(ctx context.Context, w *domain.Wishlist) error {
	b := make([]byte, 4)
	rand.Read(b)
	w.ShortToken = hex.EncodeToString(b)
	return s.repo.CreateWishlist(ctx, w)
}

func (s *Svc) GetMyLists(ctx context.Context, uid int) ([]domain.Wishlist, error) {
	return s.repo.GetWishlists(ctx, uid)
}

func (s *Svc) RemoveList(ctx context.Context, uid, id int) error {
	return s.repo.DeleteWishlist(ctx, uid, id)
}

func (s *Svc) AddGift(ctx context.Context, i *domain.Item) error {
	return s.repo.AddItem(ctx, i)
}

func (s *Svc) RemoveGift(ctx context.Context, uid, itemId int) error {
	return s.repo.DeleteItem(ctx, uid, itemId)
}

func (s *Svc) GetPublic(ctx context.Context, token string) (interface{}, error) {
	w, items, err := s.repo.GetFullList(ctx, token)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"info": w, "items": items}, nil
}

func (s *Svc) Book(ctx context.Context, token string, itemID int) error {
	return s.repo.AtomicBook(ctx, token, itemID)
}

func (s *Svc) GetSecret() []byte { return s.jwtKey }
