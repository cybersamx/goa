package tests

import (
	"github.com/cybersamx/goa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/oauth2.v3/models"
	"testing"
	"time"
)

func TestClientStore_New_DefaultConfig(t *testing.T) {
	// Setup
	db, teardown := setupDB(t)
	defer teardown()

	// Run
	config := goa.ClientStoreConfig{}
	clientStore, err := goa.NewClientStore(db, config)

	// Assert
	defaultTableName := "ClientStore"
	defaultTTL := 5 * time.Second
	assert.NoError(t, err)
	assert.NotNil(t, clientStore)
	assert.Equal(t, defaultTableName, clientStore.Config.TableName)
	assert.Equal(t, defaultTTL, clientStore.Config.TTL)
	assert.NotNil(t, clientStore.Config.InfoLog)
	assert.NotNil(t, clientStore.Config.ErrLog)
	assert.True(t, db.HasTable(defaultTableName))
}

func TestClientStore_New_PassedConfig(t *testing.T) {
	// Setup
	db, teardown := setupDB(t)
	defer teardown()

	// Run
	tableName := "client_store"
	ttl := 6 * time.Second
	config := goa.ClientStoreConfig{
		TableName: tableName,
		TTL: ttl,
	}
	clientStore, err := goa.NewClientStore(db, config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, clientStore)
	assert.Equal(t, tableName, clientStore.Config.TableName)
	assert.Equal(t, ttl, clientStore.Config.TTL)
	assert.NotNil(t, clientStore.Config.InfoLog)
	assert.NotNil(t, clientStore.Config.ErrLog)
	assert.True(t, db.HasTable(tableName))
}

func TestClientStore_Create(t *testing.T) {
	// Setup
	db, teardown := setupDB(t)
	defer teardown()
	config := goa.ClientStoreConfig{}
	clientStore, err := goa.NewClientStore(db, config)

	// Run
	info := models.Client{
		ID: "abc",
		Domain: "example.com",
		Secret: "mysecret",
		UserID: "me",
	}
	clientStore.Create(&info)

	defaultTableName := "ClientStore"
	require.NoError(t, err)
	require.NotNil(t, clientStore)
	require.True(t, db.HasTable(defaultTableName))

	// Assert (model)
	t.Run("Model should be valid", func(t *testing.T) {
		clientInfo, err := clientStore.GetByID("abc")
		assert.NoError(t, err)
		assert.NotNil(t, clientInfo)
		assert.Equal(t, info.Secret, clientInfo.GetSecret())
		assert.Equal(t, info.Domain, clientInfo.GetDomain())
		assert.Equal(t, info.UserID, clientInfo.GetUserID())
	})
}
