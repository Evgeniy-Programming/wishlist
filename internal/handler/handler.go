package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"wishlist-api/internal/domain"
	"wishlist-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	svc *service.Svc
}

func New(s *service.Svc) *Handler { return &Handler{svc: s} }

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return h.svc.GetSecret(), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "unauthorized", 401)
			return
		}
		uid := int(token.Claims.(jwt.MapClaims)["uid"].(float64))
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "uid", uid)))
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var b struct{ Email, Password string }
	json.NewDecoder(r.Body).Decode(&b)
	if err := h.svc.Register(r.Context(), b.Email, b.Password); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(201)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var b struct{ Email, Password string }
	json.NewDecoder(r.Body).Decode(&b)
	t, err := h.svc.Login(r.Context(), b.Email, b.Password)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": t})
}

func (h *Handler) CreateWishlist(w http.ResponseWriter, r *http.Request) {
	var wl domain.Wishlist
	json.NewDecoder(r.Body).Decode(&wl)
	wl.UserID = r.Context().Value("uid").(int)
	h.svc.CreateWishlist(r.Context(), &wl)
	json.NewEncoder(w).Encode(wl)
}

func (h *Handler) ListWishlists(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(int)
	res, _ := h.svc.GetMyLists(r.Context(), uid)
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) DeleteWishlist(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(int)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	h.svc.RemoveList(r.Context(), uid, id)
	w.WriteHeader(204)
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	var i domain.Item
	json.NewDecoder(r.Body).Decode(&i)
	i.WishlistID, _ = strconv.Atoi(chi.URLParam(r, "id"))
	h.svc.AddGift(r.Context(), &i)
	w.WriteHeader(201)
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(int)
	id, _ := strconv.Atoi(chi.URLParam(r, "itemId"))
	h.svc.RemoveGift(r.Context(), uid, id)
	w.WriteHeader(204)
}

func (h *Handler) GetPublicWishlist(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "token")
	res, err := h.svc.GetPublic(r.Context(), t)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) BookItem(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "token")
	id, _ := strconv.Atoi(chi.URLParam(r, "itemID"))
	if err := h.svc.Book(r.Context(), t, id); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(200)
}
