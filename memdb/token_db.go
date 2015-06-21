package memdb

import (
	"code.google.com/p/go-uuid/uuid"
	"errors"
	"github.com/murphysean/heimdall"
)

func (db *MemDB) NewToken() heimdall.Token {
	t := make(Token)
	t["id"] = uuid.New()
	return t
}

func (db *MemDB) CreateToken(token heimdall.Token) (heimdall.Token, error) {
	db.tokenCache.Put(token.GetId(), token)
	db.tokenCache.SetExpiresAt(token.GetId(), token.GetExpires())
	return token, nil
}

func (db *MemDB) GetToken(tokenId string) (heimdall.Token, error) {
	t, err := db.tokenCache.GetIfPresent(tokenId)
	if err != nil {
		return nil, errors.New("Not Found")
	}
	return t.(Token), nil
}

func (db *MemDB) UpdateToken(token heimdall.Token) (heimdall.Token, error) {
	return db.CreateToken(token)
}

func (db *MemDB) DeleteToken(tokenId string) error {
	db.tokenCache.Invalidate(tokenId)
	return nil
}
