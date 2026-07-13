package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/mathif92/prices-recommender/pkg/repositories"
)

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	JWTSecret    string
	DataRepo     repositories.DataRepository
}

type GoogleHandler struct {
	config *oauth2.Config
	cfg    GoogleConfig
	mu     sync.Mutex
	states map[string]bool
}

func NewGoogleHandler(cfg GoogleConfig) *GoogleHandler {
	return &GoogleHandler{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		},
		cfg:    cfg,
		states: make(map[string]bool),
	}
}

func (h *GoogleHandler) generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	state := hex.EncodeToString(b)
	h.mu.Lock()
	h.states[state] = true
	h.mu.Unlock()
	go func() {
		time.Sleep(10 * time.Minute)
		h.mu.Lock()
		delete(h.states, state)
		h.mu.Unlock()
	}()
	return state
}

func (h *GoogleHandler) verifyState(state string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.states[state] {
		delete(h.states, state)
		return true
	}
	return false
}

func (h *GoogleHandler) LoginURL() string {
	return h.config.AuthCodeURL(h.generateState(), oauth2.AccessTypeOnline)
}

func (h *GoogleHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		http.Error(w, "missing code or state", http.StatusBadRequest)
		return
	}

	if !h.verifyState(state) {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	token, err := h.config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("token exchange failed: %v", err), http.StatusInternalServerError)
		return
	}

	userInfo, err := h.fetchUserInfo(token.AccessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch user info: %v", err), http.StatusInternalServerError)
		return
	}

	email, _ := userInfo["email"].(string)
	googleID, _ := userInfo["id"].(string)
	name, _ := userInfo["name"].(string)
	picture, _ := userInfo["picture"].(string)

	if email == "" || googleID == "" {
		http.Error(w, "could not retrieve user email from Google", http.StatusInternalServerError)
		return
	}

	now := time.Now()

	existing, err := h.cfg.DataRepo.GetUserByEmail(r.Context(), email)
	if err != nil {
		user := repositories.User{
			Email:       email,
			GoogleID:    &googleID,
			DisplayName: &name,
			AvatarURL:   &picture,
			CreatedAt:   &now,
		}
		_, err = h.cfg.DataRepo.SaveUser(r.Context(), user)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create user: %v", err), http.StatusInternalServerError)
			return
		}
		existing, err = h.cfg.DataRepo.GetUserByEmail(r.Context(), email)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to retrieve created user: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		existing.GoogleID = &googleID
		existing.DisplayName = &name
		existing.AvatarURL = &picture
		if err := h.cfg.DataRepo.UpdateUser(r.Context(), *existing); err != nil {
			http.Error(w, fmt.Sprintf("failed to update user: %v", err), http.StatusInternalServerError)
			return
		}
	}

	jwt, err := GenerateToken(h.cfg.JWTSecret, existing.ID, existing.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate token: %v", err), http.StatusInternalServerError)
		return
	}

	frontendURL := r.URL.Query().Get("redirect")
	if frontendURL == "" {
		frontendURL = "/"
	}

	http.Redirect(w, r, fmt.Sprintf("%s#login?token=%s", frontendURL, jwt), http.StatusFound)
}

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func (h *GoogleHandler) fetchUserInfo(accessToken string) (map[string]interface{}, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info map[string]interface{}
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}

	return info, nil
}
