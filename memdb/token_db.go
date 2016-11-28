package memdb

import (
	"errors"
	"github.com/murphysean/heimdall"
)

func (db *MemDB) NewToken() heimdall.Token {
	t := new(Token)
	t.Id = genUUIDv4()
	return t
}

func (db *MemDB) CreateToken(token heimdall.Token) (heimdall.Token, error) {
	db.m.Lock()
	defer db.m.Unlock()
	db.tokenCache.Put(token.GetId(), token)
	db.tokenCache.SetExpiresAt(token.GetId(), token.GetExpires())
	return token, nil
}

func (db *MemDB) GetToken(tokenId string) (heimdall.Token, error) {
	db.m.RLock()
	defer db.m.RUnlock()
	t, err := db.tokenCache.GetIfPresent(tokenId)
	if err != nil {
		return nil, errors.New("Not Found")
	}
	return t.(*Token), nil
}

func (db *MemDB) UpdateToken(token heimdall.Token) (heimdall.Token, error) {
	return db.CreateToken(token)
}

func (db *MemDB) DeleteToken(tokenId string) error {
	db.m.Lock()
	defer db.m.Unlock()
	db.tokenCache.Invalidate(tokenId)
	return nil
}
