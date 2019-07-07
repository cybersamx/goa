package tests

import (
	"fmt"
	"github.com/cybersamx/goa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
	"testing"
	"time"
)

type tokenTestEntry struct{
	tokenType string
	token models.Token
}


func getTokenTestEntries() []tokenTestEntry {
	return []tokenTestEntry{
		{
			tokenType:"code",
			token: models.Token{
				ClientID: "abc",
				Code: "code_123",
				CodeCreateAt: time.Now(),
				CodeExpiresIn: 5 * time.Minute,
				RedirectURI: "http://localhost",
				Scope: "profile",
				UserID: "me",
			},
		},
		{
			tokenType:"access",
			token: models.Token{
				ClientID: "abc",
				Access: "access_123",
				AccessCreateAt: time.Now(),
				AccessExpiresIn: 5 * time.Minute,
				RedirectURI: "http://localhost",
				Scope: "profile",
				UserID: "me",
			},
		},
		{
			tokenType:"refresh",
			token: models.Token{
				ClientID: "abc",
				Refresh: "refresh_123",
				RefreshCreateAt: time.Now(),
				RefreshExpiresIn: 5 * time.Minute,
				RedirectURI: "http://localhost",
				Scope: "profile",
				UserID: "me",
			},
		},
	}
}

func TestTokenStore_New_DefaultConfig(t *testing.T) {
	// Setup
	db, teardown := setupDB(t)
	defer teardown()

	// Run
	config := goa.TokenStoreConfig{}
	tokenStore, err := goa.NewTokenStore(db, config)

	// Assert
	defaultTableName := "TokenStore"
	defaultTTL := 5 * time.Second
	assert.NoError(t, err)
	assert.NotNil(t, tokenStore)
	assert.Equal(t, defaultTableName, tokenStore.Config.TableName)
	assert.Equal(t, defaultTTL, tokenStore.Config.TTL)
	assert.NotNil(t, tokenStore.Config.InfoLog)
	assert.NotNil(t, tokenStore.Config.ErrLog)
	assert.True(t, db.HasTable(defaultTableName))
}

func TestTokenStore_New_PassedConfig(t *testing.T) {
	// Setup
	db, teardown := setupDB(t)
	defer teardown()

	// Run
	tableName := "token_store"
	ttl := 6 * time.Second
	config := goa.TokenStoreConfig{
		TableName: tableName,
		TTL: ttl,
	}
	tokenStore, err := goa.NewTokenStore(db, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, tokenStore)
	assert.Equal(t, tableName, tokenStore.Config.TableName)
	assert.Equal(t, ttl, tokenStore.Config.TTL)
	assert.NotNil(t, tokenStore.Config.InfoLog)
	assert.NotNil(t, tokenStore.Config.ErrLog)
	assert.True(t, db.HasTable(tableName))
}

func TestTokenStore_Create(t *testing.T) {
	var testEntries = getTokenTestEntries()

	t.Run("Different token inputs", func(t *testing.T) {
		// Setup
		db, teardown := setupDB(t)
		defer teardown()
		config := goa.TokenStoreConfig{}
		tokenStore, err := goa.NewTokenStore(db, config)

		// Assert (table)
		defaultTableName := "TokenStore"
		require.NoError(t, err)
		require.NotNil(t, tokenStore)
		require.True(t, db.HasTable(defaultTableName))

		for _, entry := range testEntries {
			// Run
			tokenStore.Create(&entry.token)

			// Assert (model)
			t.Run(fmt.Sprintf("Token %s should be valid", entry.tokenType), func(t *testing.T) {
				var tokenInfo oauth2.TokenInfo

				if entry.tokenType == "code" {
					tokenInfo, err = tokenStore.GetByCode(entry.token.Code)
				} else if entry.tokenType == "access" {
					tokenInfo, err = tokenStore.GetByAccess(entry.token.Access)
				} else if entry.tokenType == "refresh" {
					tokenInfo, err = tokenStore.GetByRefresh(entry.token.Refresh)
				} else {
					t.Fatalf("unknown token type %s", entry.tokenType)
				}

				assert.NoError(t, err)
				assert.NotNil(t, tokenInfo)
				assert.Equal(t, entry.token.ClientID, tokenInfo.GetClientID())
				assert.Equal(t, entry.token.RedirectURI, tokenInfo.GetRedirectURI())
				assert.Equal(t, entry.token.Scope, tokenInfo.GetScope())
				assert.Equal(t, entry.token.UserID, tokenInfo.GetUserID())
			})
		}
	})
}
