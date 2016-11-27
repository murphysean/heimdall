package heimdall

import (
	"errors"
	"github.com/murphysean/advhttp"
	"html/template"
	"net/http"
	"time"
)

const (
	Permit = iota
	Deny
	Indeterminate
	NotApplicable
)

type PreAuthZHandler func(r *http.Request, scope string, client Client, user User) (status int, message string)
type AuthZHandler func(r *http.Request, token Token, client Client, user User) (status int, message string)
type NoPermitHandler func(w http.ResponseWriter, r *http.Request, status int, message string, token Token, client Client, user User)

func NewHeimdall(handler http.Handler, preauthzfunc PreAuthZHandler, authzfunc AuthZHandler, nopermitfunc NoPermitHandler) *Heimdall {
	h := new(Heimdall)
	h.Handler = handler
	h.PreAuthZFunction = preauthzfunc
	h.AuthZFunction = authzfunc
	h.NoPermitFunction = nopermitfunc

	h.RewriteMe = false

	h.SessionDuration = 4 * time.Hour
	h.AccessTokenDuration = time.Hour
	h.RefreshTokenDuration = 100 * 365 * 24 * time.Hour
	h.AuthCodeDuration = 10 * time.Minute
	h.UserConcentDuration = 5 * time.Minute

	return h
}

type Heimdall struct {
	Handler          http.Handler
	DB               HeimdallDB
	PreAuthZFunction PreAuthZHandler
	AuthZFunction    AuthZHandler
	NoPermitFunction NoPermitHandler
	Templates        *template.Template

	RewriteMe bool

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
	token, client, user := h.ExpandRequest(r)
	//Send information to authz function
	s, m := az(r, token, client, user)
	//If function returns anything other than permit hand off response to the no permit handler
	if s != Permit {
		h.NoPermitFunction(w, r, s, m, token, client, user)
		return
	}
	if h.RewriteMe {
		//TODO Do a find and replace for /me/ or /me$ on path and replace with userid (or clientid if client token) from token
	}
	ctx := newContext(r.Context(), token, user, client)
	//And now let the original handler do it's job
	handler.ServeHTTP(w, r.WithContext(ctx))
}

//This function will allow you to leverage Heimdall to create fine grained policies on each
//handlerfunction you might have.
func (h *Heimdall) CreateHandlerFunc(handlerFunc http.HandlerFunc, az AuthZHandler, np NoPermitHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, client, user := h.ExpandRequest(r)
		//Send information to authz function
		s, m := az(r, token, client, user)
		//If function returns anything other than permit write failure here
		if s != Permit {
			np(w, r, s, m, token, client, user)
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
			setValuesOnContext(r.Context(), userId, session.GetClientId())
			//r.Header.Set("X-User-Id", userId)
			//r.Header.Set("X-Client-Id", session.GetClientId())
			return h.DB.GetUser(userId)
		}
	}
	//Is the user directly credentialing?
	if username, password, ok := r.BasicAuth(); ok {
		user, err := h.DB.VerifyUser(username, password)
		if err == nil {
			setValuesOnContext(r.Context(), user.GetId(), "heimdall")
			//r.Header.Set("X-User-Id", user.GetId())
			//r.Header.Set("X-Client-Id", "heimdall")
		}
		return user, err
	}
	return nil, errors.New("User not logged in")
}

func (h *Heimdall) ExpandRequest(r *http.Request) (Token, Client, User) {
	//Check for token presence
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
			token.SetClientId("heimdall")
		} else {
			client, err = h.DB.VerifyClient(username, password)
			if err == nil {
				token = h.DB.NewToken()
				token.SetType(TokenTypeBasic)
				token.SetExpires(time.Now().UTC())
				token.SetClientId(client.GetId())
			}
		}
	} else if at, ok := advhttp.BearerAuth(r); ok {
		if at != "" {
			token, err = h.DB.GetToken(at)
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

	if token != nil {
		if token.GetUserId() != "" {
			setValuesOnContext(r.Context(), token.GetUserId(), token.GetClientId())
			//r.Header.Set("X-User-Id", token.GetUserId())
		} else {
			setValuesOnContext(r.Context(), token.GetClientId(), token.GetClientId())
			//r.Header.Set("X-User-Id", token.GetClientId())
		}
		//r.Header.Set("X-Client-Id", token.GetClientId())
	}
	return token, client, user
}
