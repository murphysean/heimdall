package memdb

import (
	"sync"
)

type User struct {
	Id      string `json:"id"`
	Name    string `json:"displayName"`
	Clients map[string]struct {
		Concents      []string `json:"concents"`
		RefreshTokens []string `json:"refresh_tokens"`
	} `json:"clients"`

	Username string `json:"username"`
	Password string `json:"password"`

	sync.RWMutex
}

func (u User) GetId() string {
	u.RLock()
	defer u.RUnlock()
	return u.Id
}

func (u User) SetId(id string) {
	u.Lock()
	defer u.Unlock()
	u.Id = id
}

func (u User) GetName() string {
	u.RLock()
	defer u.RUnlock()
	return u.Name
}

func (u User) SetName(name string) {
	u.Lock()
	defer u.Unlock()
	u.Name = name
}

func (u User) GetConcents(clientId string) []string {
	u.RLock()
	defer u.RUnlock()
	return u.Clients[clientId].Concents
}

func (u User) SetConcents(clientId string, concents []string) {
	u.Lock()
	defer u.Unlock()
	c := u.Clients[clientId]
	c.Concents = concents
	u.Clients[clientId] = c
}

func (u User) GetRefreshTokens(clientId string) []string {
	u.RLock()
	defer u.RUnlock()
	return u.Clients[clientId].RefreshTokens
}

func (u User) SetRefreshTokens(clientId string, refreshTokens []string) {
	u.Lock()
	defer u.Unlock()
	c := u.Clients[clientId]
	c.Concents = refreshTokens
	u.Clients[clientId] = c
}
