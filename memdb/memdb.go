package memdb

import (
	"github.com/murphysean/cache"
	"github.com/murphysean/heimdall"
	"time"
)

type login struct {
	id       string
	password string
}

type MemDB struct {
	loginMap   map[string]login
	clientMap  map[string]heimdall.Client
	tokenCache *cache.PowerCache
	tokenMap   map[string]heimdall.Token
	userMap    map[string]heimdall.User
}

func NewMemDB() *MemDB {
	db := new(MemDB)
	db.loginMap = make(map[string]login)
	db.clientMap = make(map[string]heimdall.Client)
	db.tokenCache = cache.NewPowerCache()
	db.tokenCache.ExpiresAfterWriteDuration = time.Minute * 60
	db.tokenCache.PeriodicMaintenance = time.Minute * 120
	db.tokenMap = make(map[string]heimdall.Token)
	db.userMap = make(map[string]heimdall.User)
	return db
}

func (db *MemDB) GetNumTokens() int {
	return db.tokenCache.Length()
}
