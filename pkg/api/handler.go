package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/mathif92/prices-recommender/pkg/auth"
	"github.com/mathif92/prices-recommender/pkg/repositories"
)

type CollectorRunner interface {
	Run(ctx context.Context) error
}

type Handler struct {
	log        logrus.FieldLogger
	repo       repositories.DataRepository
	collector  CollectorRunner
	mux        *http.ServeMux
	jwtSecret  string
	googleAuth *auth.GoogleHandler
}

type Config struct {
	Log              logrus.FieldLogger
	Repo             repositories.DataRepository
	Collector        CollectorRunner
	JWTSecret        string
	GoogleClientID   string
	GoogleSecret     string
	GoogleRedirect   string
	BaseURL          string
}

func NewHandler(cfg Config) *Handler {
	h := &Handler{
		log:       cfg.Log,
		repo:      cfg.Repo,
		collector: cfg.Collector,
		jwtSecret: cfg.JWTSecret,
	}

	h.mux = http.NewServeMux()

	h.mux.HandleFunc("/api/auth/signup", h.handleSignup)
	h.mux.HandleFunc("/api/auth/login", h.handleLogin)
	h.mux.HandleFunc("/api/auth/logout", h.handleLogout)
	h.mux.Handle("/api/auth/me", h.withAuth(http.HandlerFunc(h.handleAuthMe)))

	if cfg.GoogleClientID != "" && cfg.GoogleSecret != "" {
		h.googleAuth = auth.NewGoogleHandler(auth.GoogleConfig{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleSecret,
			RedirectURL:  cfg.GoogleRedirect,
			JWTSecret:    cfg.JWTSecret,
			DataRepo:     cfg.Repo,
		})
		h.mux.HandleFunc("/api/auth/google/login", h.handleGoogleLogin)
		h.mux.HandleFunc("/api/auth/google/callback", h.handleGoogleCallback)
	}

	protected := http.NewServeMux()
	protected.HandleFunc("/api/settings", h.handleSettings)
	protected.HandleFunc("/api/settings/", h.handleSettingByKey)
	protected.HandleFunc("/api/vacations", h.handleVacations)
	protected.HandleFunc("/api/collect", h.handleCollect)

	h.mux.Handle("/api/settings", h.withAuth(protected))
	h.mux.Handle("/api/settings/", h.withAuth(protected))
	h.mux.Handle("/api/vacations", h.withAuth(protected))
	h.mux.Handle("/api/collect", h.withAuth(protected))

	h.mux.HandleFunc("/api/hotels", h.handleHotels)
	h.mux.HandleFunc("/api/hotels/", h.handleHotelByID)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) withAuth(next http.Handler) http.Handler {
	return auth.AuthMiddleware(h.jwtSecret, next)
}

func (h *Handler) userIDFromRequest(r *http.Request) int64 {
	return auth.UserIDFromContext(r.Context())
}

func parsePageParams(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			offset = (v - 1) * limit
		}
	}
	return
}

func (h *Handler) handleHotels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	location := strings.TrimSpace(r.URL.Query().Get("location"))
	withPrices := r.URL.Query().Get("with_prices") == "true"
	limit, offset := parsePageParams(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if withPrices {
		hotels, total, err := h.repo.ListHotelsWithPrices(ctx, limit, offset, location)
		if err != nil {
			h.log.Errorf("failed to list hotels with prices: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to list hotels")
			return
		}
		if hotels == nil {
			hotels = []repositories.HotelWithPrices{}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"hotels": hotels,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})
		return
	}

	var hotels []repositories.Hotel
	var total int
	var err error

	if location != "" {
		hotels, err = h.repo.SearchHotelsByLocationPaginated(ctx, location, limit, offset)
		if err == nil {
			total, err = h.repo.CountHotelsByLocation(ctx, location)
		}
	} else {
		hotels, err = h.repo.ListHotelsPaginated(ctx, limit, offset)
		if err == nil {
			total, err = h.repo.CountHotels(ctx)
		}
	}
	if err != nil {
		h.log.Errorf("failed to list hotels: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list hotels")
		return
	}
	if hotels == nil {
		hotels = []repositories.Hotel{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"hotels": hotels,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) handleHotelByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/hotels/")
	if idStr == "" || idStr == "/" {
		h.handleHotels(w, r)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid hotel id")
		return
	}

	sub := strings.TrimPrefix(r.URL.Path, "/api/hotels/"+idStr)
	sub = strings.TrimPrefix(sub, "/")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	switch sub {
	case "":
		h.getHotelDetail(ctx, w, id)
	case "reviews":
		h.getHotelReviews(ctx, w, id)
	case "prices":
		h.getHotelPrices(ctx, w, id)
	case "ratings":
		h.getHotelRatings(ctx, w, id)
	default:
		writeError(w, http.StatusNotFound, "unknown endpoint")
	}
}

func (h *Handler) getHotelDetail(ctx context.Context, w http.ResponseWriter, id int64) {
	hotel, err := h.repo.GetHotelByID(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "hotel not found")
		return
	}

	ratings, _ := h.repo.ListHotelRatings(ctx, id)
	reviews, _ := h.repo.ListHotelReviews(ctx, id)
	prices, _ := h.repo.ListPrices(ctx, id)

	if ratings == nil {
		ratings = []repositories.HotelRating{}
	}
	if reviews == nil {
		reviews = []repositories.HotelReview{}
	}
	if prices == nil {
		prices = []repositories.Price{}
	}

	resp := map[string]any{
		"hotel":   hotel,
		"ratings": ratings,
		"reviews": reviews,
		"prices":  prices,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getHotelReviews(ctx context.Context, w http.ResponseWriter, id int64) {
	reviews, err := h.repo.ListHotelReviews(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "hotel not found")
		return
	}
	if reviews == nil {
		reviews = []repositories.HotelReview{}
	}
	writeJSON(w, http.StatusOK, reviews)
}

func (h *Handler) getHotelPrices(ctx context.Context, w http.ResponseWriter, id int64) {
	prices, err := h.repo.ListPrices(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "hotel not found")
		return
	}
	if prices == nil {
		prices = []repositories.Price{}
	}
	writeJSON(w, http.StatusOK, prices)
}

func (h *Handler) getHotelRatings(ctx context.Context, w http.ResponseWriter, id int64) {
	ratings, err := h.repo.ListHotelRatings(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "hotel not found")
		return
	}
	if ratings == nil {
		ratings = []repositories.HotelRating{}
	}
	writeJSON(w, http.StatusOK, ratings)
}

func (h *Handler) handleVacations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	yearStr := r.URL.Query().Get("year")
	year := time.Now().Year()
	if y, err := strconv.Atoi(yearStr); err == nil && y > 0 {
		year = y
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	vacations, err := h.repo.GetVacations(ctx, year)
	if err != nil {
		h.log.Errorf("failed to get vacations: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get vacations")
		return
	}
	if vacations == nil {
		vacations = []repositories.Vacation{}
	}
	writeJSON(w, http.StatusOK, vacations)
}

func (h *Handler) handleCollect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	if err := h.collector.Run(ctx); err != nil {
		h.log.Errorf("collection failed: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "collection complete"})
}

func (h *Handler) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listSettings(w, r)
	case http.MethodPost:
		h.upsertSetting(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleSettingByKey(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/settings/")
	if key == "" || key == "/" {
		h.handleSettings(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getSetting(w, r, key)
	case http.MethodPut:
		h.upsertSettingByKey(w, r, key)
	case http.MethodDelete:
		h.deleteSetting(w, r, key)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) listSettings(w http.ResponseWriter, r *http.Request) {
	userID := h.userIDFromRequest(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	settings, err := h.repo.ListUserSettings(ctx, userID)
	if err != nil {
		h.log.Errorf("failed to list settings: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list settings")
		return
	}
	if settings == nil {
		settings = []repositories.UserSettings{}
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *Handler) getSetting(w http.ResponseWriter, r *http.Request, key string) {
	userID := h.userIDFromRequest(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	setting, err := h.repo.GetUserSetting(ctx, userID, key)
	if err != nil {
		writeError(w, http.StatusNotFound, "setting not found")
		return
	}
	writeJSON(w, http.StatusOK, setting)
}

func (h *Handler) upsertSetting(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Key == "" {
		writeError(w, http.StatusBadRequest, "key is required")
		return
	}
	h.doUpsertSetting(w, r, body.Key, body.Value)
}

func (h *Handler) upsertSettingByKey(w http.ResponseWriter, r *http.Request, key string) {
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	h.doUpsertSetting(w, r, key, body.Value)
}

func (h *Handler) doUpsertSetting(w http.ResponseWriter, r *http.Request, key, value string) {
	userID := h.userIDFromRequest(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	setting := repositories.UserSettings{
		UserID:       userID,
		SettingKey:   key,
		SettingValue: value,
	}
	saved, err := h.repo.SaveUserSettings(ctx, setting)
	if err != nil {
		h.log.Errorf("failed to save setting: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to save setting")
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (h *Handler) deleteSetting(w http.ResponseWriter, r *http.Request, key string) {
	userID := h.userIDFromRequest(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.repo.DeleteUserSetting(ctx, userID, key); err != nil {
		h.log.Errorf("failed to delete setting: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete setting")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	existing, err := h.repo.GetUserByEmail(r.Context(), body.Email)
	if err == nil && existing != nil {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		h.log.Errorf("failed to hash password: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	hashStr := string(hash)
	salt := "" // bcrypt includes the salt in the hash
	now := time.Now()
	name := body.DisplayName

	user := repositories.User{
		Email:        body.Email,
		PasswordHash: &hashStr,
		PasswordSalt: &salt,
		DisplayName:  &name,
		CreatedAt:    &now,
	}

	saved, err := h.repo.SaveUser(r.Context(), user)
	if err != nil {
		h.log.Errorf("failed to save user: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	if saved.ID == 1 {
		h.migrateSettingsToUser(r.Context(), saved.ID)
	}

	token, err := auth.GenerateToken(h.jwtSecret, saved.ID, saved.Email)
	if err != nil {
		h.log.Errorf("failed to generate token: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  saved,
	})
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if user.PasswordHash == nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(body.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.ID, user.Email)
	if err != nil {
		h.log.Errorf("failed to generate token: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}

func (h *Handler) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := h.userIDFromRequest(r)
	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (h *Handler) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.googleAuth == nil {
		writeError(w, http.StatusServiceUnavailable, "Google authentication is not configured")
		return
	}

	redirect := r.URL.Query().Get("redirect")
	loginURL := h.googleAuth.LoginURL()
	if redirect != "" {
		loginURL += "&redirect=" + redirect
	}

	http.Redirect(w, r, loginURL, http.StatusFound)
}

func (h *Handler) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.googleAuth == nil {
		writeError(w, http.StatusServiceUnavailable, "Google authentication is not configured")
		return
	}

	frontendURL := r.URL.Query().Get("redirect")
	if frontendURL == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		frontendURL = fmt.Sprintf("%s://%s", scheme, r.Host)
	}

	// Preserve redirect for the Google handler to use
	q := r.URL.Query()
	q.Set("redirect", frontendURL)
	r.URL.RawQuery = q.Encode()

	h.googleAuth.HandleCallback(w, r)
}

func (h *Handler) migrateSettingsToUser(ctx context.Context, userID int64) {
	settings, err := h.repo.ListUserSettings(ctx, 1)
	if err != nil {
		h.log.Warnf("failed to list settings for migration: %v", err)
		return
	}

	for _, s := range settings {
		s.UserID = userID
		s.ID = 0
		if _, err := h.repo.SaveUserSettings(ctx, s); err != nil {
			h.log.Warnf("failed to migrate setting %s: %v", s.SettingKey, err)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	fmt.Fprint(w, "prices-recommender API")
}
