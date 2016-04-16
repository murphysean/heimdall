package sqldb

import (
	"github.com/murphysean/heimdall"
	"strings"
)

func (db *SqlDB) NewClient() heimdall.Client {
	c := new(Client)
	c.Id = genUUIDv4()
	return c
}

func (db *SqlDB) VerifyClient(clientId, clientSecret string) (heimdall.Client, error) {
	c, err := db.GetClient(clientId)
	if err != nil {
		return nil, err
	}
	if c.GetSecret() == clientSecret {
		return c, nil
	}
	return nil, heimdall.ErrInvalidCredentials
}

func (db *SqlDB) CreateClient(client heimdall.Client) (heimdall.Client, error) {
	_, err := db.Db.Exec("INSERT OR REPLACE INTO clients (id,name,secret,type,internal,redirecturis) VALUES (?,?,?,?,?,?)", client.GetId(), client.GetName(), client.GetSecret(), client.GetType(), client.GetInternal(), strings.Join(client.GetRedirectURIs(), ","))
	if err != nil {
		return client, err
	}
	return client, nil
}

func (db *SqlDB) GetClient(clientId string) (heimdall.Client, error) {
	c := new(Client)
	c.Id = clientId
	var redirectUris string
	err := db.Db.QueryRow("SELECT name,secret,type,internal,redirecturis FROM clients WHERE id = ?", clientId).Scan(&c.Name, &c.Secret, &c.Type, &c.Internal, &redirectUris)
	c.RedirectUris = strings.Split(redirectUris, ",")
	if err != nil {
		return c, err
	}
	return c, nil
}

func (db *SqlDB) UpdateClient(client heimdall.Client) (heimdall.Client, error) {
	return db.CreateClient(client)
}

func (db *SqlDB) DeleteClient(clientId string) error {
	_, err := db.Db.Exec("DELETE FROM clients WHERE id = ?", clientId)
	return err
}
