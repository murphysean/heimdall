package filedb

import (
	"crypto/rand"
	"fmt"
	"github.com/murphysean/cache"
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

type FileDB struct {
	Directory string
	cache     *cache.PowerCache
}

func NewFileDB(dir string) *FileDB {
	db := new(FileDB)
	db.Directory = dir
	db.cache = cache.NewPowerCache()
	db.cache.ExpiresAfterWriteDuration = time.Minute * 60
	db.cache.PeriodicMaintenance = time.Minute * 120
	return db
}

func (db *FileDB) GetNumTokens() int {
	return db.cache.Length()
}
