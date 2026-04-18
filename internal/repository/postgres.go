package repository

import (
	"context"
	"wishlist-api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type Repo struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) InitSchema() {
	s := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS wishlists (
		id SERIAL PRIMARY KEY,
		user_id INT REFERENCES users(id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		description TEXT,
		event_date TIMESTAMP,
		short_token TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		wishlist_id INT REFERENCES wishlists(id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		description TEXT,
		link TEXT,
		priority INT DEFAULT 1,
		is_booked BOOLEAN DEFAULT FALSE
	);`
	r.db.MustExec(s)
}

func (r *Repo) CreateUser(ctx context.Context, u *domain.User) error {
	return r.db.QueryRowxContext(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", u.Email, u.PasswordHash).Scan(&u.ID)
}

func (r *Repo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, "SELECT * FROM users WHERE email = $1", email)
	return &u, err
}

func (r *Repo) CreateWishlist(ctx context.Context, w *domain.Wishlist) error {
	return r.db.QueryRowxContext(ctx, "INSERT INTO wishlists (user_id, title, description, event_date, short_token) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		w.UserID, w.Title, w.Description, w.EventDate, w.ShortToken).Scan(&w.ID)
}

func (r *Repo) GetWishlists(ctx context.Context, uid int) ([]domain.Wishlist, error) {
	var res []domain.Wishlist
	err := r.db.SelectContext(ctx, &res, "SELECT * FROM wishlists WHERE user_id = $1", uid)
	return res, err
}

func (r *Repo) DeleteWishlist(ctx context.Context, uid, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM wishlists WHERE id = $1 AND user_id = $2", id, uid)
	return err
}

func (r *Repo) AddItem(ctx context.Context, i *domain.Item) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO items (wishlist_id, title, description, link, priority) VALUES ($1, $2, $3, $4, $5)",
		i.WishlistID, i.Title, i.Description, i.Link, i.Priority)
	return err
}

func (r *Repo) DeleteItem(ctx context.Context, uid, itemId int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM items WHERE id = $1 AND wishlist_id IN (SELECT id FROM wishlists WHERE user_id = $2)", itemId, uid)
	return err
}

func (r *Repo) GetFullList(ctx context.Context, token string) (*domain.Wishlist, []domain.Item, error) {
	var w domain.Wishlist
	if err := r.db.GetContext(ctx, &w, "SELECT * FROM wishlists WHERE short_token = $1", token); err != nil {
		return nil, nil, err
	}
	var items []domain.Item
	err := r.db.SelectContext(ctx, &items, "SELECT * FROM items WHERE wishlist_id = $1", w.ID)
	return &w, items, err
}

func (r *Repo) AtomicBook(ctx context.Context, token string, itemID int) error {
	res, err := r.db.ExecContext(ctx, "UPDATE items SET is_booked = true WHERE id = $1 AND is_booked = false AND wishlist_id = (SELECT id FROM wishlists WHERE short_token = $2)", itemID, token)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return domain.ErrAlreadyBooked
	}
	return nil
}
