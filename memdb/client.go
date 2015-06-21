package memdb

import (
	giu "github.com/murphysean/gointerfaceutils"
)

type Client map[string]interface{}

func (c Client) GetId() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(c), "id")
}

func (c Client) SetId(id string) {
	c["id"] = id
}

func (c Client) GetSecret() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(c), "secret")
}

func (c Client) SetSecret(secret string) {
	c["secret"] = secret
}

func (c Client) GetName() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(c), "type")
}

func (c Client) SetName(name string) {
	c["name"] = name
}

func (c Client) GetType() string {
	return giu.MustGetStringAtObjPath(map[string]interface{}(c), "type")
}

func (c Client) SetType(t string) {
	c["type"] = t
}

func (c Client) GetInternal() bool {
	return giu.MustGetBooleanAtObjPath(map[string]interface{}(c), "internal")
}

func (c Client) SetInternal(internal bool) {
	c["internal"] = internal
}

func (c Client) GetRedirectURIs() []string {
	return giu.MustGetStringArrayAtObjPath(map[string]interface{}(c), "redirect_uris")
}

func (c Client) SetRedirectURIs(redirectURIs []string) {
	c["redirect_uris"] = redirectURIs
}
