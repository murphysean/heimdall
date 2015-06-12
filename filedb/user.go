package filedb

import (
	giu "github.com/murphysean/gointerfaceutils"
)

type User map[string]interface{}

func (u User) GetId() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(u), "id")
}

func (u User) SetId(id string) {
	u["id"] = id
}

func (u User) GetName() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(u), "displayName")
}

func (u User) SetName(name string) {
	u["displayName"] = name
}

func (u User) GetConcents(clientId string) []string {
	return giu.MustGetStringArrayAtObjPath(map[string]interface{}(u), "clients."+clientId+".concents")
}

func (u User) SetConcents(clientId string, concents []string) {
	giu.SetValueAtObjPath(map[string]interface{}(u), "user.clients."+clientId+".concents", concents)
}

func (u User) GetRefreshTokens(clientId string) []string {
	return giu.MustGetStringArrayAtObjPath(map[string]interface{}(u), "clients."+clientId+".refresh_tokens")
}

func (u User) SetRefreshTokens(clientId string, refreshTokens []string) {
	giu.SetValueAtObjPath(map[string]interface{}(u), "user.clients."+clientId+".refresh_tokens", refreshTokens)
}
