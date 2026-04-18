package repository

import (
	"context"
	"wishlist-api/internal/domain"
)

type Store interface {
	CreateUser(ctx context.Context, u *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateWishlist(ctx context.Context, w *domain.Wishlist) error
	GetWishlists(ctx context.Context, uid int) ([]domain.Wishlist, error)
	DeleteWishlist(ctx context.Context, uid, id int) error // Проверь наличие этого метода
	AddItem(ctx context.Context, i *domain.Item) error
	DeleteItem(ctx context.Context, uid, itemId int) error
	GetFullList(ctx context.Context, token string) (*domain.Wishlist, []domain.Item, error)
	AtomicBook(ctx context.Context, token string, itemID int) error
}
