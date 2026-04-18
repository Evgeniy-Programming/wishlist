package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"wishlist-api/internal/handler"
	"wishlist-api/internal/repository"
	"wishlist-api/internal/service"
)

func main() {
	db, err := sqlx.Connect("postgres", os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.New(db)
	repo.InitSchema() // Авто-миграция при старте

	svc := service.New(repo, os.Getenv("JWT_SECRET"))
	h := handler.New(svc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Get("/share/{token}", h.GetPublicWishlist)
	r.Post("/share/{token}/book/{itemID}", h.BookItem)

	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Route("/api/wishlists", func(r chi.Router) {
			r.Post("/", h.CreateWishlist)
			r.Get("/", h.ListWishlists)
			r.Delete("/{id}", h.DeleteWishlist)
			r.Post("/{id}/items", h.AddItem)
			r.Delete("/items/{itemId}", h.DeleteItem)
		})
	})

	r.Handle("/*", http.FileServer(http.Dir("./internal/web")))

	srv := &http.Server{Addr: ":8080", Handler: r}
	go func() {
		fmt.Println("Running on :8080")
		srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
}
