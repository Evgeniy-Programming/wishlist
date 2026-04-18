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

type RegisterRequest struct {
	Email    string `json:"email" example:"test@test.com"`
	Password string `json:"password" example:"123456"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"test@test.com"`
	Password string `json:"password" example:"123456"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type WishlistResponse struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	ShortToken string `json:"short_token"`
}

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

// @Summary Регистрация
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RegisterRequest true "Данные регистрации"
// @Success 201 {string} string "Created"
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var b RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	if err := h.svc.Register(r.Context(), b.Email, b.Password); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(201)
}

// @Summary Логин
// @Tags auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "Данные для входа"
// @Success 200 {object} LoginResponse
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var b LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	t, err := h.svc.Login(r.Context(), b.Email, b.Password)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	json.NewEncoder(w).Encode(LoginResponse{Token: t})
}

// @Summary Создать вишлист
// @Tags private
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body domain.Wishlist true "Параметры списка"
// @Success 200 {object} domain.Wishlist
// @Router /api/wishlists [post]
func (h *Handler) CreateWishlist(w http.ResponseWriter, r *http.Request) {
	var wl domain.Wishlist
	json.NewDecoder(r.Body).Decode(&wl)
	wl.UserID = r.Context().Value("uid").(int)
	h.svc.CreateWishlist(r.Context(), &wl)
	json.NewEncoder(w).Encode(wl)
}

// @Summary Мои вишлисты
// @Tags private
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} domain.Wishlist
// @Router /api/wishlists [get]
func (h *Handler) ListWishlists(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(int)
	res, _ := h.svc.GetMyLists(r.Context(), uid)
	json.NewEncoder(w).Encode(res)
}

// @Summary Удалить вишлист
// @Tags private
// @Security ApiKeyAuth
// @Param id path int true "ID списка"
// @Success 204 "No Content"
// @Router /api/wishlists/{id} [delete]
func (h *Handler) DeleteWishlist(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(int)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	h.svc.RemoveList(r.Context(), uid, id)
	w.WriteHeader(204)
}

// @Summary Добавить подарок
// @Tags private
// @Security ApiKeyAuth
// @Param id path int true "ID вишлиста"
// @Param input body domain.Item true "Данные подарка"
// @Success 201 "Created"
// @Router /api/wishlists/{id}/items [post]
func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	var i domain.Item
	json.NewDecoder(r.Body).Decode(&i)
	i.WishlistID, _ = strconv.Atoi(chi.URLParam(r, "id"))
	h.svc.AddGift(r.Context(), &i)
	w.WriteHeader(201)
}

// @Summary Удалить подарок
// @Tags private
// @Security ApiKeyAuth
// @Param itemId path int true "ID подарка"
// @Success 204 "No Content"
// @Router /api/wishlists/items/{itemId} [delete]
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid").(int)
	id, _ := strconv.Atoi(chi.URLParam(r, "itemId"))
	h.svc.RemoveGift(r.Context(), uid, id)
	w.WriteHeader(204)
}

// @Summary Публичный просмотр
// @Tags public
// @Param token path string true "Токен вишлиста"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /share/{token} [get]
func (h *Handler) GetPublicWishlist(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "token")
	res, err := h.svc.GetPublic(r.Context(), t)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	json.NewEncoder(w).Encode(res)
}

// @Summary Бронирование подарка
// @Tags public
// @Param token path string true "Токен вишлиста"
// @Param itemID path int true "ID подарка"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Already booked"
// @Router /share/{token}/book/{itemID} [post]
func (h *Handler) BookItem(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "token")
	id, _ := strconv.Atoi(chi.URLParam(r, "itemID"))
	if err := h.svc.Book(r.Context(), t, id); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(200)
}
