package heimdall

import (
	"errors"
	"html/template"
	"net/http"
	"strings"
	"time"
)

const (
	Permit = iota
	Deny
	Indeterminate
	NotApplicable
)

type AuthZHandler func(r *http.Request, token Token, client Client, user User) (status int, message string)

func NewHeimdall(handler http.Handler, authzfunc AuthZHandler) *Heimdall {
	h := new(Heimdall)
	h.Handler = handler
	h.AuthZFunction = authzfunc

	h.SessionDuration = 4 * time.Hour
	h.AccessTokenDuration = time.Hour
	h.RefreshTokenDuration = 100 * 365 * 24 * time.Hour
	h.AuthCodeDuration = 10 * time.Minute
	h.UserConcentDuration = 5 * time.Minute

	return h
}

type Heimdall struct {
	Handler       http.Handler
	DB            HeimdallDB
	AuthZFunction AuthZHandler
	Templates     *template.Template

	SessionDuration      time.Duration
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	AuthCodeDuration     time.Duration
	UserConcentDuration  time.Duration
}

//The purpose of heimdalls handler is to protect another handler. It
//will first determine authentication through basic authentication,
//cookies, and authorization tokens. The second step will then call
//an authorization function with the incoming request as well as
//the user or token information.
func (h *Heimdall) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Protect(w, r, h.Handler, h.AuthZFunction)
}

func (h *Heimdall) Protect(w http.ResponseWriter, r *http.Request, handler http.Handler, az AuthZHandler) {
	token, client, user := h.scrapeRequest(r)
	//Send information to authz function
	s, m := az(r, token, client, user)
	//If function returns anything other than permit write failure here
	if s != Permit {
		http.Error(w, m, http.StatusForbidden)
		return
	}
	//And now let the original handler do it's job
	handler.ServeHTTP(w, r)
}

//This function will allow you to leverage Heimdall to create fine grained policies on each
//handlerfunction you might have.
func (h *Heimdall) CreateHandlerFunc(handlerFunc http.HandlerFunc, az AuthZHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, client, user := h.scrapeRequest(r)
		//Send information to authz function
		s, m := az(r, token, client, user)
		//If function returns anything other than permit write failure here
		if s != Permit {
			http.Error(w, m, http.StatusForbidden)
			return
		}
		//And now let the original handler do it's job
		handlerFunc(w, r)
	}
}

func (h *Heimdall) getLoggedInUser(w http.ResponseWriter, r *http.Request) (User, error) {
	//Is the user logged in?
	if cookie, err := r.Cookie("session-id"); err == nil && cookie.Value != "" {
		session, err := h.DB.GetToken(cookie.Value)
		if err == nil {
			userId := session.GetUserId()
			session.SetExpires(time.Now().Add(h.SessionDuration))
			h.DB.UpdateToken(session)
			return h.DB.GetUser(userId)
		}
	}
	//Is the user directly credentialing?
	if username, password, ok := r.BasicAuth(); ok {
		return h.DB.VerifyUser(username, password)
	}
	return nil, errors.New("User not logged in")
}

func (h *Heimdall) scrapeRequest(r *http.Request) (Token, Client, User) {
	//Check for token presence
	var t string
	var token Token
	var client Client
	var user User
	var err error

	if username, password, ok := r.BasicAuth(); ok {
		user, err = h.DB.VerifyUser(username, password)
		if err == nil {
			//In this case the client will just be heimdall
			client, _ = h.DB.GetClient("heimdall")
			//And the token will be a Basic One Use Token
			token = h.DB.NewToken()
			token.SetType(TokenTypeBasic)
			token.SetExpires(time.Now().UTC())
			token.SetUserId(user.GetId())
			token.SetClientId(client.GetId())
		}
		client, err = h.DB.VerifyClient(username, password)
		if err == nil {
			token = h.DB.NewToken()
			token.SetType(TokenTypeBasic)
			token.SetExpires(time.Now().UTC())
			token.SetClientId(client.GetId())
		}
	} else if ah := r.Header.Get("Authorization"); strings.HasPrefix(ah, "Bearer ") {
		t = ah[7:]
		if t != "" {
			token, err = h.DB.GetToken(t)
			//If present, gather token information
			if err == nil {
				client, _ = h.DB.GetClient(token.GetClientId())
				user, _ = h.DB.GetUser(token.GetUserId())
			}
		}
	} else if cookie, err := r.Cookie("session-id"); err == nil && cookie.Value != "" {
		token, err = h.DB.GetToken(cookie.Value)
		if err == nil {
			user, _ = h.DB.GetUser(token.GetUserId())
			client, _ = h.DB.GetClient(token.GetClientId())
			token.SetExpires(time.Now().Add(h.SessionDuration))
			h.DB.UpdateToken(token)
		}
	}

	return token, client, user
}