package filedb

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"github.com/murphysean/heimdall"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	TOKENS_DIRECTORY = "tokens"
)

func (db *FileDB) NewToken() heimdall.Token {
	t := make(Token)
	t["id"] = uuid.New()
	return t
}

func (db *FileDB) CreateToken(token heimdall.Token) (heimdall.Token, error) {
	b, err := json.Marshal(&token)
	if err != nil {
		return token, err
	}
	err = ioutil.WriteFile(filepath.Join(db.Directory, TOKENS_DIRECTORY, token.GetId()+".json"), b, os.ModePerm)
	if err != nil {
		return token, err
	}
	return token, nil
}

func (db *FileDB) GetToken(tokenId string) (heimdall.Token, error) {
	b, err := ioutil.ReadFile(filepath.Join(db.Directory, TOKENS_DIRECTORY, tokenId+".json"))
	if err != nil {
		return nil, err
	}
	var token Token
	err = json.Unmarshal(b, &token)
	if err != nil {
		return token, err
	}
	if token.GetExpires().Before(time.Now()) {
		return token, heimdall.ErrExpired
	}
	return token, nil
}

func (db *FileDB) UpdateToken(token heimdall.Token) (heimdall.Token, error) {
	return db.CreateToken(token)
}

func (db *FileDB) DeleteToken(tokenId string) error {
	return os.Remove(filepath.Join(db.Directory, TOKENS_DIRECTORY, tokenId+".json"))
}
