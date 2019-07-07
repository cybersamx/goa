package goa

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
	"log"
	"os"
	"time"
)

const (
	defaultClientStoreTTL       = 5 * time.Second
	defaultClientStoreTableName = "ClientStore"
)

type (
	ClientStoreConfig struct {
		ErrLog    *log.Logger
		InfoLog   *log.Logger
		TableName string
		TTL       time.Duration
	}

	ClientStoreItem struct {
		ID        string `gorm:"primary_key"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Domain    string
		Secret    string
		UserID    string
	}

	ClientStore struct {
		Config ClientStoreConfig
		DB     *gorm.DB
		Ticker *time.Ticker
	}
)

// --- Private Functions ---

// Assign default values
func fillDefaultClientStoreConfig(config ClientStoreConfig) ClientStoreConfig {
	if config.TTL <= 0 {
		config.TTL = defaultClientStoreTTL
	}

	if config.TableName == "" {
		config.TableName = defaultClientStoreTableName
	}

	if config.ErrLog == nil {
		config.ErrLog = log.New(os.Stderr, "[GOA ERROR]:  ", log.Ldate|log.Ltime)
	}

	if config.InfoLog == nil {
		config.InfoLog = log.New(os.Stdout, "[GOA INFO]:  ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	return config
}

// Remove expired tokens
func (cs *ClientStore) removeExpiredClientStoreItems() {
	for range cs.Ticker.C {
		now := time.Now().Unix()
		if err := cs.DB.Table(cs.Config.TableName).Where("expired_at > ?", now).Delete(&ClientStoreItem{}).Error; err != nil {
			logf(cs.Config.ErrLog, "problem removing expired items %v\n", err)
		}
	}
}

// --- Public Functions ---

// Create a new client store
func NewClientStore(db *gorm.DB, config ClientStoreConfig) (*ClientStore, error) {
	fullConfig := fillDefaultClientStoreConfig(config)

	cs := ClientStore{
		DB:     db,
		Config: fullConfig,
		Ticker: time.NewTicker(defaultClientStoreTTL),
	}

	csModel := ClientStoreItem{}
	if !db.HasTable(cs.Config.TableName) {
		if err := db.Table(cs.Config.TableName).CreateTable(&csModel).Error; err != nil {
			return nil, err
		}

		db.Table(cs.Config.TableName).AutoMigrate(&csModel)
	}

	go cs.removeExpiredClientStoreItems()

	return &cs, nil
}


func (cs *ClientStore) Close() {
	cs.Ticker.Stop()
}

// --- Implements oauth2.ClientStore ---

// Create and store new client info
func (cs *ClientStore) Create(info oauth2.ClientInfo) error {
	newItem := ClientStoreItem{
		ID: info.GetID(),
		Secret: info.GetSecret(),
		Domain: info.GetDomain(),
		UserID: info.GetUserID(),
	}

	return cs.DB.Table(cs.Config.TableName).Create(newItem).Error
}

// Get client  info by ID
func (cs *ClientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	if id == "" {
		return nil, nil
	}

	var fillItem ClientStoreItem
	if err := cs.DB.Table(cs.Config.TableName).First(&fillItem, "id = ?", id).Error; err != nil {
		return nil, err
	}

	info := models.Client{
		ID:fillItem.ID,
		Secret: fillItem.Secret,
		Domain: fillItem.Domain,
		UserID: fillItem.UserID,
	}

	return &info, nil
}