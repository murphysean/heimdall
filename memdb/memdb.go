package memdb

import (
	"crypto/rand"
	"fmt"
	"github.com/murphysean/cache"
	"github.com/murphysean/heimdall"
	"sync"
	"time"
)

func genUUIDv4() string {
	u := make([]byte, 16)
	rand.Read(u)
	//Set the version to 4
	u[6] = (u[6] | 0x40) & 0x4F
	u[8] = (u[8] | 0x80) & 0xBF
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

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

	m sync.RWMutex
}

func NewMemDB() *MemDB {
	db := new(MemDB)
	db.m.Lock()
	defer db.m.Unlock()
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
	db.m.RLock()
	defer db.m.RUnlock()
	return db.tokenCache.Length()
}
