package filedb

import (
	"github.com/murphysean/cache"
	"time"
)

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
