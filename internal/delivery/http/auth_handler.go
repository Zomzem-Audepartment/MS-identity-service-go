package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zomzem/identity-service/internal/usecase"
)

type AuthHandler struct {
	authUsecase usecase.AuthUseCase
}

func NewAuthHandler(r chi.Router, auc usecase.AuthUseCase) {
	handler := &AuthHandler{authUsecase: auc}
	r.Post("/auth/login", handler.Login)
	r.Post("/auth/google", handler.LoginGoogle)
	r.Post("/auth/refresh", handler.Refresh)
	r.Post("/auth/logout", handler.Logout)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	type loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.authUsecase.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	h.setTokenCookie(w, resp.RefreshToken)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) LoginGoogle(w http.ResponseWriter, r *http.Request) {
	type googleReq struct {
		IDToken string `json:"idToken"`
		Token   string `json:"token"`
	}

	var req googleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token := req.IDToken
	if token == "" {
		token = req.Token
	}

	resp, err := h.authUsecase.LoginGoogle(r.Context(), token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	h.setTokenCookie(w, resp.RefreshToken)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var refreshToken string

	// 1. Try to get from cookie
	cookie, err := r.Cookie("refreshToken")
	if err == nil {
		refreshToken = cookie.Value
	}

	// 2. Fallback to body if not in cookie
	if refreshToken == "" {
		type refreshReq struct {
			RefreshToken string `json:"refreshToken"`
		}
		var req refreshReq
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Refresh token required"})
		return
	}

	resp, err := h.authUsecase.Refresh(r.Context(), refreshToken)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	h.setTokenCookie(w, resp.RefreshToken)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    token,
		Path:     "/",
		MaxAge:   604800, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}
