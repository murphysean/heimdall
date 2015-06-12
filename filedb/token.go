package filedb

import (
	giu "github.com/murphysean/gointerfaceutils"
	"time"
)

type Token map[string]interface{}

func (t Token) GetId() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(t), "id")
}

func (t Token) SetId(id string) {
	t["id"] = id
}

func (t Token) GetType() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(t), "type")
}

func (t Token) SetType(tp string) {
	t["type"] = tp
}

func (t Token) GetUserId() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(t), "user_id")
}

func (t Token) SetUserId(userId string) {
	t["user_id"] = userId
}

func (t Token) GetClientId() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(t), "client_id")
}

func (t Token) SetClientId(clientId string) {
	t["client_id"] = clientId
}

func (t Token) GetExpires() time.Time {
	return giu.MustGetTimeAtObjPath(map[string]interface{}(t), "expires")
}

func (t Token) SetExpires(expires time.Time) {
	t["expires"] = expires.UTC().Format(time.RFC3339)
}

func (t Token) GetScope() []string {
	return giu.MustGetStringArrayAtObjPath(map[string]interface{}(t), "scope")
}

func (t Token) SetScope(scope []string) {
	t["scope"] = scope
}

func (t Token) GetAccessType() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(t), "access_type")
}

func (t Token) SetAccessType(accessType string) {
	t["access_type"] = accessType
}

func (t Token) GetRefreshToken() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(t), "refresh_token")
}

func (t Token) SetRefreshToken(refreshToken string) {
	t["refresh_token"] = refreshToken
}
