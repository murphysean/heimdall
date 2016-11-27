package heimdall

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound           = errors.New("Not Found")
	ErrExpired            = errors.New("Expired")
	ErrInvalidCredentials = errors.New("Invalid Credentials")
)

const (
	AuthorizationResponseTypeToken  = "token"
	AuthorizationResponseTypeCode   = "code"
	TokenGrantTypeAuthCode          = "authorization_code"
	TokenGrantTypeClientCredentials = "client_credentials"
	TokenGrantTypeRefreshToken      = "refresh_token"
	TokenGrantTypePassword          = "password"
	TokenTypeBasic                  = "Basic"
	TokenTypeSession                = "Session"
	TokenTypeBearer                 = "Bearer"
	TokenTypeRefresh                = "Refresh"
	TokenTypeCode                   = "AuthorizationCode"
	TokenTypeConcent                = "UserConcent"
	TokenAccessTypeOffline          = "offline"
	TokenAccessTypeOnline           = "online"
)

type HeimdallDB interface {
	CreateObj
	TokenDB
	UserDB
	ClientDB
}

type CreateObj interface {
	NewToken() Token
	NewUser() User
	NewClient() Client
}

type TokenDB interface {
	CreateToken(token Token) (Token, error)
	GetToken(tokenId string) (Token, error)
	UpdateToken(token Token) (Token, error)
	DeleteToken(tokenId string) error
}

type UserDB interface {
	VerifyUser(username, password string) (User, error)
	CreateUser(user User) (User, error)
	GetUser(userId string) (User, error)
	UpdateUser(user User) (User, error)
	DeleteUser(userId string) error
}

type ClientDB interface {
	VerifyClient(clientId, clientSecret string) (Client, error)
	CreateClient(client Client) (Client, error)
	GetClient(clientId string) (Client, error)
	UpdateClient(client Client) (Client, error)
	DeleteClient(clientId string) error
}

type Token interface {
	GetId() string
	SetId(id string)
	GetType() string
	SetType(t string)
	GetUserId() string
	SetUserId(userId string)
	GetClientId() string
	SetClientId(clientId string)
	GetExpires() time.Time
	SetExpires(expires time.Time)
	GetScope() []string
	SetScope(scope []string)
	GetAccessType() string
	SetAccessType(accessType string)
	GetRefreshToken() string
	SetRefreshToken(refreshToken string)
}

type User interface {
	GetId() string
	SetId(id string)
	GetName() string
	SetName(name string)
	GetConcents(clientId string) []string
	SetConcents(clientId string, concents []string)
}

type Client interface {
	GetId() string
	SetId(id string)
	GetSecret() string
	SetSecret(secret string)
	GetName() string
	SetName(name string)
	GetType() string
	SetType(t string)
	GetInternal() bool
	SetInternal(internal bool)
	GetRedirectURIs() []string
	SetRedirectURIs(redirectURIs []string)
}

type UserIder interface {
	UserId(id string)
}
type ClientIder interface {
	ClientId(id string)
}

func setValuesOnContext(ctx context.Context, userId string, clientId string) {
	if uier := ctx.Value("userider"); uier != nil {
		if uierv, ok := uier.(UserIder); ok {
			uierv.UserId(userId)
		}
	}
	if cier := ctx.Value("clientider"); cier != nil {
		if cierv, ok := cier.(ClientIder); ok {
			cierv.ClientId(clientId)
		}
	}
}

type ctxkey int

var tokenKey ctxkey = 0
var userKey ctxkey = 1
var clientKey ctxkey = 2

func newContext(ctx context.Context, t Token, u User, c Client) context.Context {
	if t == nil || c == nil {
		return ctx
	}
	tc := context.WithValue(ctx, tokenKey, t)
	tc = context.WithValue(ctx, userKey, u)
	tc = context.WithValue(ctx, clientKey, c)
	return tc
}

func FromContext(ctx context.Context) (t Token, u User, c Client, ok bool) {
	if t, ok = ctx.Value(tokenKey).(Token); !ok {
		return
	}
	if c, ok = ctx.Value(clientKey).(Client); !ok {
		return
	}
	u, _ = ctx.Value(userKey).(User)
	return
}
