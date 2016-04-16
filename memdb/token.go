package memdb

import (
	"sync"
	"time"
)

type Token struct {
	Id           string    `json:"id"`
	Type         string    `json:"type"`
	UserId       string    `json:"user_id"`
	ClientId     string    `json:"client_id"`
	Expires      time.Time `json:"expires"`
	Scope        []string  `json:"scope"`
	AccessType   string    `json:"access_type"`
	RefreshToken string    `json:"refresh_token"`

	sync.RWMutex
}

func (t Token) GetId() string {
	t.RLock()
	defer t.RUnlock()
	return t.Id
}

func (t Token) SetId(id string) {
	t.Lock()
	defer t.Unlock()
	t.Id = id
}

func (t Token) GetType() string {
	t.RLock()
	defer t.RUnlock()
	return t.Type
}

func (t Token) SetType(tp string) {
	t.Lock()
	defer t.Unlock()
	t.Type = tp
}

func (t Token) GetUserId() string {
	t.RLock()
	defer t.RUnlock()
	return t.UserId
}

func (t Token) SetUserId(userId string) {
	t.Lock()
	defer t.Unlock()
	t.UserId = userId
}

func (t Token) GetClientId() string {
	t.RLock()
	defer t.RUnlock()
	return t.ClientId
}

func (t Token) SetClientId(clientId string) {
	t.Lock()
	defer t.Unlock()
	t.ClientId = clientId
}

func (t Token) GetExpires() time.Time {
	t.RLock()
	defer t.RUnlock()
	return t.Expires
}

func (t Token) SetExpires(expires time.Time) {
	t.Lock()
	defer t.Unlock()
	t.Expires = expires
}

func (t Token) GetScope() []string {
	t.RLock()
	defer t.RUnlock()
	return t.Scope
}

func (t Token) SetScope(scope []string) {
	t.Lock()
	defer t.Unlock()
	t.Scope = scope
}

func (t Token) GetAccessType() string {
	t.RLock()
	defer t.RUnlock()
	return t.AccessType
}

func (t Token) SetAccessType(accessType string) {
	t.Lock()
	defer t.Unlock()
	t.AccessType = accessType
}

func (t Token) GetRefreshToken() string {
	t.RLock()
	defer t.RUnlock()
	return t.RefreshToken
}

func (t Token) SetRefreshToken(refreshToken string) {
	t.Lock()
	defer t.Unlock()
	t.RefreshToken = refreshToken
}
