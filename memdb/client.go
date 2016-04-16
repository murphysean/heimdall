package memdb

import (
	"sync"
)

type Client struct {
	Id           string   `json:"id"`
	Secret       string   `json:"secret"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Internal     bool     `json:"internal"`
	RedirectUris []string `json:"redirect_uris"`

	sync.RWMutex
}

func (c *Client) GetId() string {
	c.RLock()
	defer c.RUnlock()
	return c.Id
}

func (c *Client) SetId(id string) {
	c.Lock()
	defer c.Unlock()
	c.Id = id
}

func (c *Client) GetSecret() string {
	c.RLock()
	defer c.RUnlock()
	return c.Secret
}

func (c *Client) SetSecret(secret string) {
	c.Lock()
	defer c.Unlock()
	c.Secret = secret
}

func (c *Client) GetName() string {
	c.RLock()
	defer c.RUnlock()
	return c.Name
}

func (c *Client) SetName(name string) {
	c.Lock()
	defer c.Unlock()
	c.Name = name
}

func (c *Client) GetType() string {
	c.RLock()
	defer c.RUnlock()
	return c.Type
}

func (c *Client) SetType(t string) {
	c.Lock()
	defer c.Unlock()
	c.Type = t
}

func (c *Client) GetInternal() bool {
	c.RLock()
	defer c.RUnlock()
	return c.Internal
}

func (c *Client) SetInternal(internal bool) {
	c.Lock()
	defer c.Unlock()
	c.Internal = internal
}

func (c *Client) GetRedirectURIs() []string {
	c.RLock()
	defer c.RUnlock()
	return c.RedirectUris
}

func (c *Client) SetRedirectURIs(redirectURIs []string) {
	c.Lock()
	defer c.Unlock()
	c.RedirectUris = redirectURIs
}
