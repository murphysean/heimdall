package filedb

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"errors"
	"github.com/murphysean/heimdall"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	CLIENTS_DIRECTORY = "clients"
)

func (db *FileDB) NewClient() heimdall.Client {
	c := make(Client)
	c["id"] = uuid.New()
	return c
}

func (db *FileDB) VerifyClient(clientId, clientSecret string) (heimdall.Client, error) {
	c, err := db.GetClient(clientId)
	if err != nil {
		return nil, err
	}
	if c.GetSecret() == clientSecret {
		return c, nil
	}
	return nil, heimdall.ErrInvalidCredentials
}

func (db *FileDB) CreateClient(client heimdall.Client) (heimdall.Client, error) {
	db.cache.Put(client.GetId(), client)
	b, err := json.Marshal(&client)
	if err != nil {
		return client, err
	}
	err = ioutil.WriteFile(filepath.Join(db.Directory, CLIENTS_DIRECTORY, client.GetId()+".json"), b, os.ModePerm)
	if err != nil {
		return client, err
	}
	return client, nil
}

func (db *FileDB) GetClient(clientId string) (heimdall.Client, error) {
	c, err := db.cache.GetWithValueLoader(clientId, db.getClient)
	if err != nil {
		return nil, err
	}
	if client, ok := c.(heimdall.Client); ok {
		return client, nil
	}
	return nil, errors.New("Unknown Value")
}

func (db *FileDB) getClient(clientId string) (interface{}, error) {
	b, err := ioutil.ReadFile(filepath.Join(db.Directory, CLIENTS_DIRECTORY, clientId+".json"))
	if err != nil {
		return nil, err
	}
	var client Client
	err = json.Unmarshal(b, &client)
	if err != nil {
		return client, err
	}
	return client, nil
}

func (db *FileDB) UpdateClient(client heimdall.Client) (heimdall.Client, error) {
	return db.CreateClient(client)
}

func (db *FileDB) DeleteClient(clientId string) error {
	db.cache.Invalidate(clientId)
	return os.Remove(filepath.Join(db.Directory, CLIENTS_DIRECTORY, clientId+".json"))
}
