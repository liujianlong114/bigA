package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lijianjun/bigA/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type AuthConfig struct {
	Secret           string
	Relax            bool
	DefaultAccountID int64
	InitialCash      float64
	TokenTTL         time.Duration
}

type authClaims struct {
	UserID    int64  `json:"uid"`
	AccountID int64  `json:"aid"`
	Username  string `json:"sub"`
	jwt.RegisteredClaims
}

func (c AuthConfig) issueToken(userID, accountID int64, username string) (string, time.Time, error) {
	exp := time.Now().Add(c.TokenTTL)
	claims := authClaims{
		UserID:    userID,
		AccountID: accountID,
		Username:  username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(c.Secret))
	return s, exp, err
}

func (c AuthConfig) parseToken(tokenStr string) (*authClaims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &authClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(c.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*authClaims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Username = strings.TrimSpace(body.Username)
	body.Nickname = strings.TrimSpace(body.Nickname)
	if utf8.RuneCountInString(body.Username) < 3 {
		writeErr(w, http.StatusBadRequest, "username at least 3 characters")
		return
	}
	if len(body.Password) < 6 {
		writeErr(w, http.StatusBadRequest, "password at least 6 characters")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash password failed")
		return
	}
	user, acct, err := h.Repo.RegisterUser(r.Context(), body.Username, string(hash), body.Nickname, h.Auth.InitialCash)
	if err != nil {
		if errors.Is(err, store.ErrUserExists) {
			writeErr(w, http.StatusConflict, "username already exists")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	token, exp, err := h.Auth.issueToken(user.ID, acct.ID, user.Username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "issue token failed")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"token":      token,
		"expires_at": exp,
		"user":       user,
		"account":    acct,
	})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	user, hash, err := h.Repo.GetUserByUsername(r.Context(), strings.TrimSpace(body.Username))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)) != nil {
		writeErr(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	acct, err := h.Repo.GetAccountByUserID(r.Context(), user.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if acct == nil {
		writeErr(w, http.StatusInternalServerError, "account not found")
		return
	}
	token, exp, err := h.Auth.issueToken(user.ID, acct.ID, user.Username)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "issue token failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": exp,
		"user":       user,
		"account":    acct,
	})
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "login required")
		return
	}
	user, err := h.Repo.GetUserByID(r.Context(), uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		writeErr(w, http.StatusUnauthorized, "user not found")
		return
	}
	acctID := h.accountID(r)
	acct, err := h.Repo.GetAccount(r.Context(), acctID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user":    user,
		"account": acct,
	})
}
