package memdb

import (
	"code.google.com/p/go-uuid/uuid"
	"errors"
	"github.com/murphysean/heimdall"
)

func (db *MemDB) NewClient() heimdall.Client {
	c := make(Client)
	c["id"] = uuid.New()
	return c
}

func (db *MemDB) VerifyClient(clientId, clientSecret string) (heimdall.Client, error) {
	c, err := db.GetClient(clientId)
	if err != nil {
		return nil, err
	}
	if c.GetSecret() == clientSecret {
		return c, nil
	}
	return nil, heimdall.ErrInvalidCredentials
}

func (db *MemDB) CreateClient(client heimdall.Client) (heimdall.Client, error) {
	db.clientMap[client.GetId()] = client
	return client, nil
}

func (db *MemDB) GetClient(clientId string) (heimdall.Client, error) {
	client, ok := db.clientMap[clientId]
	if !ok {
		return client, errors.New("Not Found")
	}
	return client, nil
}

func (db *MemDB) UpdateClient(client heimdall.Client) (heimdall.Client, error) {
	return db.CreateClient(client)
}

func (db *MemDB) DeleteClient(clientId string) error {
	delete(db.clientMap, clientId)
	return nil
}
