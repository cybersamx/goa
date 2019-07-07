package goa

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
	"log"
	"os"
	"time"
)

const (
	defaultTokenStoreTTL       = 5 * time.Second
	defaultTokenStoreTableName = "TokenStore"
)

type (
	TokenStoreConfig struct {
		ErrLog    *log.Logger
		InfoLog   *log.Logger
		TableName string
		TTL       time.Duration
	}

	TokenStoreItem struct {
		ID        string    `gorm:"primary_key"`
		CreatedAt time.Time
		UpdatedAt time.Time
		ExpiresAt time.Time
		Code      string    `gorm:"index:idx_code"`
		Access    string    `gorm:"index:idx_access"`
		Refresh   string    `gorm:"index:idx_refresh"`
		Data      string
	}

	TokenStore struct {
		Config TokenStoreConfig
		DB     *gorm.DB
		Ticker *time.Ticker
	}
)

// --- Private Functions ---

// Assign default values
func fillDefaultTokenStoreConfig(config TokenStoreConfig) TokenStoreConfig {
	if config.TTL <= 0 {
		config.TTL = defaultTokenStoreTTL
	}

	if config.TableName == "" {
		config.TableName = defaultTokenStoreTableName
	}

	if config.ErrLog == nil {
		config.ErrLog = log.New(os.Stderr, "[OAUTHGORM ERROR]:  ", log.Ldate|log.Ltime)
	}

	if config.InfoLog == nil {
		config.InfoLog = log.New(os.Stdout, "[OAUTHGORM INFO]:  ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	return config
}

// Remove expired tokens
func (ts *TokenStore) removeExpiredTokens() {
	for range ts.Ticker.C {
		now := time.Now().Unix()
		if err := ts.DB.Table(ts.Config.TableName).Where("expired_at > ?", now).Delete(&ClientStoreItem{}).Error; err != nil {
			logf(ts.Config.ErrLog, "problem removing expired items %v\n", err)
		}
	}
}

func (ts *TokenStore) getBy(field string, value string) (oauth2.TokenInfo, error) {
	if value == "" {
		return nil, nil
	}

	var fillItem TokenStoreItem
	if err := ts.DB.Table(ts.Config.TableName).First(&fillItem, fmt.Sprintf("%s = ?", field), value).Error; err != nil {
		return nil, err
	}

	var fillToken models.Token
	err := json.Unmarshal([]byte(fillItem.Data), &fillToken)
	if err != nil {
		return nil, err
	}

	return &fillToken, nil
}

func (ts *TokenStore) removeBy(field string, value string) error {
	if value == "" {
		return nil
	}

	// WARNING: When deleting a record, you need to ensure its primary field
	// has value, and GORM will use the primary key to delete the record, if the
	// primary key field is blank, GORM will delete all records for the model

	var foundItem TokenStoreItem
	if err := ts.DB.Table(ts.Config.TableName).First(&foundItem, fmt.Sprintf("%s = ?", field), value).Error; err != nil {
		return err
	}

	return ts.DB.Table(ts.Config.TableName).Delete(&foundItem).Error
}

// --- Public Functions ---

// Create a new token store
func NewTokenStore(db *gorm.DB, config TokenStoreConfig) (*TokenStore, error) {
	fullConfig := fillDefaultTokenStoreConfig(config)

	ts := TokenStore{
		DB:     db,
		Config: fullConfig,
		Ticker: time.NewTicker(defaultTokenStoreTTL),
	}

	model := TokenStoreItem{}
	if !db.HasTable(ts.Config.TableName) {
		if err := db.Table(ts.Config.TableName).CreateTable(&model).Error; err != nil {
			return nil, err
		}

		db.Table(ts.Config.TableName).AutoMigrate(&model)
	}

	go ts.removeExpiredTokens()

	return &ts, nil
}


func (ts *TokenStore) Close() {
	ts.Ticker.Stop()
}

// --- Implements oauth2.TokenStore ---

// Create and store new token info
func (ts *TokenStore) Create(info oauth2.TokenInfo) error {
	// Serialize TokenInfo object as a json
	infoData, err := json.Marshal(info)
	if err != nil {
		return err
	}
	newItem := TokenStoreItem{
		Data: string(infoData),
	}

	code := info.GetCode()
	if code != "" {
		newItem.Code = code
		newItem.ExpiresAt = info.GetCodeCreateAt().Add(info.GetCodeExpiresIn())
	} else {
		newItem.Access = info.GetAccess()
		newItem.ExpiresAt = info.GetAccessCreateAt().Add(info.GetAccessExpiresIn())

		refresh := info.GetRefresh()
		if refresh != "" {
			newItem.Refresh = info.GetRefresh()
			newItem.ExpiresAt = info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn())
		}
	}

	return ts.DB.Table(ts.Config.TableName).Create(&newItem).Error
}



// Get token info by Code
func (ts *TokenStore) GetByCode(code string) (oauth2.TokenInfo, error) {
	return ts.getBy("code", code)
}

// Get token info by Access
func (ts *TokenStore) GetByAccess(access string) (oauth2.TokenInfo, error) {
	return ts.getBy("access", access)
}

// Get token info by Refresh
func (ts *TokenStore) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	return ts.getBy("refresh", refresh)
}

// Remove token info by Code
func (ts *TokenStore) RmoveByCode(code string) error {
	return ts.removeBy("code", code)
}

// Remove token info by Access
func (ts *TokenStore) RemoveByAccess(access string) error {
	return ts.removeBy("access", access)
}

// Remove token info by Refresh
func (ts *TokenStore) RemoveByRefresh(refresh string) error {
	return ts.removeBy("refresh", refresh)
}
