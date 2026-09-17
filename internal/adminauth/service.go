package adminauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/nexusrun/nexus_aigateway/ext"
	"github.com/nexusrun/nexus_aigateway/internal/storage"
	"github.com/nexusrun/nexus_aigateway/internal/storage/sqlx"
)

const (
	cookieName   = "gomodel_admin_session"
	sessionTTL   = 12 * time.Hour
	minSecretLen = 32
)

type Service struct {
	store  Store
	secret []byte
}

func (s *Service) Name() string { return "admin-password" }

// New creates the database-backed admin authentication service. A bootstrap
// account is created once when both bootstrap values are supplied.
func New(ctx context.Context, shared storage.Storage, secret, bootstrapUsername, bootstrapPassword string) (*Service, error) {
	if len(secret) < minSecretLen {
		return nil, fmt.Errorf("admin session secret must be at least %d characters", minSecretLen)
	}
	store, err := storage.ResolveSQLBackend[Store](ctx, shared,
		func(db sqlx.DB) (Store, error) { return newSQLStore(ctx, db) },
		func(db *mongo.Database) (Store, error) { return newMongoStore(ctx, db) },
	)
	if err != nil {
		return nil, err
	}
	s := &Service{store: store, secret: []byte(secret)}
	if strings.TrimSpace(bootstrapUsername) != "" || bootstrapPassword != "" {
		if strings.TrimSpace(bootstrapUsername) == "" || bootstrapPassword == "" {
			return nil, fmt.Errorf("admin bootstrap username and password must be supplied together")
		}
		if err := s.bootstrap(ctx, bootstrapUsername, bootstrapPassword); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Service) bootstrap(ctx context.Context, username, password string) error {
	_, err := s.store.FindByUsername(ctx, strings.ToLower(strings.TrimSpace(username)))
	if err == nil {
		return nil
	}
	if err != ErrNotFound {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash bootstrap password: %w", err)
	}
	now := time.Now().UTC()
	return s.store.Create(ctx, Account{
		ID: uuid.NewString(), Username: strings.ToLower(strings.TrimSpace(username)),
		PasswordHash: string(hash), Role: "admin", Enabled: true, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *Service) Login(ctx context.Context, username, password string) (string, error) {
	a, err := s.store.FindByUsername(ctx, strings.ToLower(strings.TrimSpace(username)))
	if err != nil || !a.Enabled || bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidLogin
	}
	return s.sign(a.ID, time.Now().UTC().Add(sessionTTL))
}

func (s *Service) AuthenticateRequest(ctx context.Context, r *http.Request) (*ext.Authentication, error) {
	if !strings.HasPrefix(r.URL.Path, "/admin/") && r.URL.Path != "/admin" {
		return nil, nil
	}
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, nil
	}
	id, err := s.verify(cookie.Value)
	if err != nil {
		return nil, nil
	}
	a, err := s.store.FindByID(ctx, id)
	if err != nil || !a.Enabled {
		return nil, nil
	}
	return &ext.Authentication{PrincipalID: "admin:" + a.ID, DashboardAccess: true, Method: "password"}, nil
}

func (s *Service) SetSession(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: token, Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: int(sessionTTL.Seconds())})
}

func (s *Service) ClearSession(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}

type sessionPayload struct {
	ID      string `json:"id"`
	Expires int64  `json:"expires"`
}

func (s *Service) sign(id string, expires time.Time) (string, error) {
	payload, err := json.Marshal(sessionPayload{ID: id, Expires: expires.Unix()})
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (s *Service) verify(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", ErrInvalidLogin
	}
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(parts[0]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return "", ErrInvalidLogin
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", ErrInvalidLogin
	}
	var payload sessionPayload
	if json.Unmarshal(data, &payload) != nil || payload.ID == "" || time.Now().Unix() >= payload.Expires {
		return "", ErrInvalidLogin
	}
	return payload.ID, nil
}

func randomSecret() ([]byte, error) { b := make([]byte, 32); _, err := rand.Read(b); return b, err }
