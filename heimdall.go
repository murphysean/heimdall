package heimdall

import (
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
