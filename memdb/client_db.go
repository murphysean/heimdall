package memdb

import (
	"errors"
	"github.com/murphysean/heimdall"
)

func (db *MemDB) NewClient() heimdall.Client {
	c := new(Client)
	c.Id = genUUIDv4()
	return c
}

func (db *MemDB) VerifyClient(clientId, clientSecret string) (heimdall.Client, error) {
	db.m.RLock()
	defer db.m.RUnlock()
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
	db.m.Lock()
	defer db.m.Unlock()
	db.clientMap[client.GetId()] = client
	return client, nil
}

func (db *MemDB) GetClient(clientId string) (heimdall.Client, error) {
	db.m.RLock()
	defer db.m.RUnlock()
	client, ok := db.clientMap[clientId]
	if !ok {
		return client, errors.New("Not Found")
	}
	return client, nil
}

func (db *MemDB) UpdateClient(client heimdall.Client) (heimdall.Client, error) {
	db.m.Lock()
	defer db.m.Unlock()
	return db.CreateClient(client)
}

func (db *MemDB) DeleteClient(clientId string) error {
	db.m.Lock()
	defer db.m.Unlock()
	delete(db.clientMap, clientId)
	return nil
}
