package domain

import (
	"errors"
	"time"
)

var ErrAlreadyBooked = errors.New("item already booked")

type User struct {
	ID           int    `db:"id"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
}

type Wishlist struct {
	ID          int        `db:"id" json:"id"`
	UserID      int        `db:"user_id" json:"user_id"`
	Title       string     `db:"title" json:"title"`
	Description string     `db:"description" json:"description"`
	EventDate   *time.Time `db:"event_date" json:"event_date"`
	ShortToken  string     `db:"short_token" json:"short_token"`
}

type Item struct {
	ID          int    `db:"id" json:"id"`
	WishlistID  int    `db:"wishlist_id" json:"wishlist_id"`
	Title       string `db:"title" json:"title"`
	Description string `db:"description" json:"description"`
	Link        string `db:"link" json:"link"`
	Priority    int    `db:"priority" json:"priority"`
	IsBooked    bool   `db:"is_booked" json:"is_booked"`
}
