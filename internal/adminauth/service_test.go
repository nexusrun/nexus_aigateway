package adminauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type memoryStore struct{ accounts map[string]Account }

func (m *memoryStore) FindByUsername(_ context.Context, username string) (Account, error) {
	for _, account := range m.accounts {
		if account.Username == username { return account, nil }
	}
	return Account{}, ErrNotFound
}

func (m *memoryStore) FindByID(_ context.Context, id string) (Account, error) {
	account, ok := m.accounts[id]
	if !ok { return Account{}, ErrNotFound }
	return account, nil
}

func (m *memoryStore) Create(_ context.Context, account Account) error {
	m.accounts[account.ID] = account
	return nil
}

func (m *memoryStore) Close() error { return nil }

func TestLoginAndAuthenticateRequest(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct horse battery staple"), bcrypt.DefaultCost)
	require.NoError(t, err)
	store := &memoryStore{accounts: map[string]Account{
		"admin-id": {ID: "admin-id", Username: "admin", PasswordHash: string(hash), Enabled: true},
	}}
	service := &Service{store: store, secret: []byte("01234567890123456789012345678901")}

	token, err := service.Login(context.Background(), " ADMIN ", "correct horse battery staple")
	require.NoError(t, err)
	_, err = service.Login(context.Background(), "admin", "wrong")
	require.ErrorIs(t, err, ErrInvalidLogin)

	req := httptest.NewRequest("GET", "http://example.test/admin/runtime/config", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: token})
	authentication, err := service.AuthenticateRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "admin:admin-id", authentication.PrincipalID)
	require.True(t, authentication.DashboardAccess)
}

func TestExpiredSessionIsRejected(t *testing.T) {
	service := &Service{store: &memoryStore{accounts: map[string]Account{}}, secret: []byte("01234567890123456789012345678901")}
	token, err := service.sign("admin-id", time.Now().Add(-time.Minute))
	require.NoError(t, err)
	req := httptest.NewRequest("GET", "http://example.test/admin", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: token})
	identity, err := service.AuthenticateRequest(context.Background(), req)
	require.NoError(t, err)
	require.Nil(t, identity)
}
