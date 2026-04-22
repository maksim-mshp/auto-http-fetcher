package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT_GenerateAccessToken(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		accessTTL time.Duration
		userID    int
		wantErr   bool
	}{
		{
			name:      "success - generate token",
			secretKey: "my-secret-key",
			accessTTL: 24,
			userID:    123,
			wantErr:   false,
		},
		{
			name:      "success - with different user ID",
			secretKey: "another-secret",
			accessTTL: 1,
			userID:    456,
			wantErr:   false,
		},
		{
			name:      "success - empty secret key",
			secretKey: "",
			accessTTL: 1,
			userID:    1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtService := NewJWTService(tt.secretKey, tt.accessTTL)
			token, err := jwtService.GenerateAccessToken(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestJWT_ParseToken_ValidToken(t *testing.T) {
	secretKey := "test-secret-key"
	accessTTL := 24 * time.Hour
	userID := 123

	jwtService := NewJWTService(secretKey, accessTTL)

	token, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := jwtService.ParseToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
}

func TestJWT_ParseToken_InvalidToken(t *testing.T) {
	jwtService := NewJWTService("secret", 24*time.Hour)

	tests := []struct {
		name        string
		token       string
		expectedErr string
	}{
		{
			name:        "empty token",
			token:       "",
			expectedErr: "token is malformed",
		},
		{
			name:        "malformed token",
			token:       "invalid.token.string",
			expectedErr: "token is malformed",
		},
		{
			name:        "random string",
			token:       "not-a-jwt-token",
			expectedErr: "token is malformed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtService.ParseToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestJWT_ParseToken_ExpiredToken(t *testing.T) {
	secretKey := "test-secret"
	accessTTL := -1 * time.Hour

	jwtService := NewJWTService(secretKey, accessTTL)

	token, err := jwtService.GenerateAccessToken(123)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := jwtService.ParseToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "expired")
}

func TestJWT_ParseToken_WrongSecret(t *testing.T) {
	jwtService1 := NewJWTService("correct-secret", 24*time.Hour)

	token, err := jwtService1.GenerateAccessToken(123)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	jwtService2 := NewJWTService("wrong-secret", 24*time.Hour)
	claims, err := jwtService2.ParseToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "signature is invalid")
}

func TestJWT_ParseToken_TamperedToken(t *testing.T) {
	jwtService := NewJWTService("secret", 24*time.Hour)

	token, err := jwtService.GenerateAccessToken(123)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	tamperedToken := token + "x"

	claims, err := jwtService.ParseToken(tamperedToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWT_GenerateAndParse_Integration(t *testing.T) {
	secretKey := "integration-test-secret"
	accessTTL := 1 * time.Hour
	userID := 999

	jwtService := NewJWTService(secretKey, accessTTL)

	token, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := jwtService.ParseToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, userID, claims.UserID)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
	assert.True(t, claims.IssuedAt.Before(time.Now()))
}

func TestJWT_Claims_Structure(t *testing.T) {
	secretKey := "secret"
	accessTTL := 24 * time.Hour
	userID := 42

	jwtService := NewJWTService(secretKey, accessTTL)
	token, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)

	parser := jwt.Parser{}
	claims := &Claims{}
	_, _, err = parser.ParseUnverified(token, claims)
	require.NoError(t, err)

	assert.Equal(t, userID, claims.UserID)
	assert.NotZero(t, claims.ExpiresAt)
	assert.NotZero(t, claims.IssuedAt)
}
