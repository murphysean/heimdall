package heimdall

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"time"
)

func genUUIDv4() string {
	u := make([]byte, 16)
	rand.Read(u)
	//Set the version to 4
	u[6] = (u[6] | 0x40) & 0x4F
	u[8] = (u[8] | 0x80) & 0xBF
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

type tokenResponse struct {
	AccessToken  string   `json:"access_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	Scope        []string `json:"scope"`
	RefreshToken string   `json:"refresh_token,omitempty"`
}

type tokenError struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
	URI         string `json:"error_uri"`
}

func writeTokenErrorResponse(w http.ResponseWriter, r *http.Request, errorString, errorDescription, errorURI string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	if errorString == "invalid_client" {
		if r.Header.Get("Authorization") != "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Token"`)
		}
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	te := tokenError{Code: errorString, Description: errorDescription, URI: errorURI}
	e := json.NewEncoder(w)
	err := e.Encode(&te)
	if err != nil {
		fmt.Println(err)
	}
}

func (h *Heimdall) OAuth2Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "The Token endpoint only supports POST requests", http.StatusNotFound)
		return
	}
	if mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mediaType != "application/x-www-form-urlencoded" {
		//writeTokenErrorResponse("invalid_request", "Token endpoint only supports a content-type of application/x-www-form-urlencoded", "https://tools.ietf.org/html/rfc6749", w, r)
		http.Error(w, "Token endpoint only supports a content-type of application/x-www-form-urlencoded", http.StatusNotFound)
		return
	}

	clientId, clientSecret, basicAuth := r.BasicAuth()
	if !basicAuth {
		clientId = r.PostFormValue("client_id")
		clientSecret = r.PostFormValue("client_secret")
	}
	grantType := r.PostFormValue("grant_type")
	if grantType != TokenGrantTypeAuthCode &&
		grantType != TokenGrantTypeClientCredentials &&
		grantType != TokenGrantTypeRefreshToken &&
		grantType != TokenGrantTypeRefreshToken {
		writeTokenErrorResponse(w, r, "invalid_grant", "Grant Type must be one of authorization_code, client_credentials, refresh_token, or password", "https://tools.ietf.org/html/rfc6749")
		return
	}
	switch grantType {
	case "authorization_code":
		authorizationCode := r.PostFormValue("code")
		redirectURI := r.PostFormValue("redirect_uri")

		if clientId == "" {
			writeTokenErrorResponse(w, r, "invalid_client", "Required param client_id is missing", "https://tools.ietf.org/html/rfc6749")
			return
		}

		if authorizationCode == "" {
			writeTokenErrorResponse(w, r, "invalid_request", "Required param code is missing", "https://tools.ietf.org/html/rfc6749")
			return
		}

		if redirectURI == "" {
			writeTokenErrorResponse(w, r, "invalid_request", "Required param redirect_uri is missing", "https://tools.ietf.org/html/rfc6749")
			return
		}

		//Grab the client
		client, err := h.DB.GetClient(clientId)
		if err != nil {
			writeTokenErrorResponse(w, r, "invalid_client", "Unknown Client", "https://tools.ietf.org/html/rfc6749")
			return
		}
		r.Header.Set("X-User-Id", clientId)
		r.Header.Set("X-Client-Id", clientId)

		//The big question now is whether I should _force_ the client to auth, the spec recommends that any client that is confidential should
		if client.GetType() == "confidential" {
			if clientSecret != "" {
				if client.GetSecret() != clientSecret {
					writeTokenErrorResponse(w, r, "invalid_client", "A confidential client is required to authenticate (client_id and client_secret)", "https://tools.ietf.org/html/rfc6749")
					return
				}
			}
		}

		//Is the code valid?
		code, err := h.DB.GetToken(authorizationCode)
		if err != nil {
			writeTokenErrorResponse(w, r, "invalid_grant", "Invalid or Expired Authorization Code", "https://tools.ietf.org/html/rfc6749")
			return
		}

		//Does the client_id match the code?
		if code.GetClientId() != clientId {
			writeTokenErrorResponse(w, r, "invalid_request", "client_id does not match authorization code grant", "https://tools.ietf.org/html/rfc6749")
			return
		}

		//Is the redirect_uri valid
		valid := false
		redirectURIs := client.GetRedirectURIs()
		for _, ruri := range redirectURIs {
			if ruri == r.PostFormValue("redirect_uri") {
				valid = true
			}
		}
		if !valid {
			writeTokenErrorResponse(w, r, "invalid_grant", "The provided redirect_uri does not match the redirection URI used in the authorization request, or was issued to another client", "https://tools.ietf.org/html/rfc6749")
			return
		}
		r.Header.Set("X-User-Id", code.GetUserId())

		//Coolness all is in order to give away the access token requested
		tokenId := genUUIDv4()
		token := h.DB.NewToken()
		token.SetId(tokenId)
		token.SetType(TokenTypeBearer)
		token.SetScope(code.GetScope())
		token.SetUserId(code.GetUserId())
		token.SetClientId(code.GetClientId())
		token.SetExpires(time.Now().UTC().Add(h.AccessTokenDuration))
		h.DB.CreateToken(token)

		//Finally remove the code so it can't be reused
		h.DB.DeleteToken(authorizationCode)

		//Maybe not always create a refresh token?
		refreshTokenId := ""

		if code.GetAccessType() == TokenAccessTypeOffline {
			refreshTokenId = genUUIDv4()
			refreshToken := h.DB.NewToken()
			refreshToken.SetId(refreshTokenId)
			refreshToken.SetType(TokenTypeRefresh)
			refreshToken.SetScope(code.GetScope())
			refreshToken.SetUserId(code.GetUserId())
			refreshToken.SetClientId(code.GetClientId())
			refreshToken.SetExpires(time.Now().UTC().Add(h.RefreshTokenDuration))
			h.DB.CreateToken(refreshToken)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)

		tr := tokenResponse{AccessToken: token.GetId(), TokenType: token.GetType(), ExpiresIn: int64(token.GetExpires().Sub(time.Now()).Seconds()), Scope: token.GetScope(), RefreshToken: refreshTokenId}
		e := json.NewEncoder(w)
		err = e.Encode(tr)
		if err != nil {
			fmt.Println(err)
		}
	case "client_credentials":
		if clientId == "" || clientSecret == "" {
			writeTokenErrorResponse(w, r, "invalid_client", "Client is required to authenticate, client_id/client_secret is missing", "https://tools.ietf.org/html/rfc6749")
			return
		}

		client, err := h.DB.VerifyClient(clientId, clientSecret)
		if err != nil {
			writeTokenErrorResponse(w, r, "invalid_client", "Unknown Client", "https://tools.ietf.org/html/rfc6749")
			return
		}

		asked_scope := strings.Split(r.PostFormValue("scope"), " ")
		scope := make([]string, 0)
		for _, s := range asked_scope {
			if z, _ := h.PreAuthZFunction(r, s, client, nil); z == Permit {
				scope = append(scope, s)
			}
		}

		r.Header.Set("X-User-Id", clientId)
		r.Header.Set("X-Client-Id", clientId)
		//Coolness, all is in order to give away the access token requested
		tokenId := genUUIDv4()
		token := h.DB.NewToken()
		token.SetId(tokenId)
		token.SetType(TokenTypeBearer)
		token.SetScope(scope)
		token.SetClientId(clientId)
		token.SetExpires(time.Now().UTC().Add(h.AccessTokenDuration))
		h.DB.CreateToken(token)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)

		tr := tokenResponse{AccessToken: token.GetId(), TokenType: token.GetType(), ExpiresIn: int64(token.GetExpires().Sub(time.Now()).Seconds()), Scope: token.GetScope()}
		e := json.NewEncoder(w)
		err = e.Encode(tr)
		if err != nil {
			fmt.Println(err)
		}
	case "refresh_token":
		refreshTokenId := r.PostFormValue("refresh_token")

		if refreshTokenId == "" {
			writeTokenErrorResponse(w, r, "invalid_request", "Required param refresh_token is missing", "https://tools.ietf.org/html/rfc6749")
			return
		}

		refreshToken, err := h.DB.GetToken(refreshTokenId)
		if err != nil {
			writeTokenErrorResponse(w, r, "invalid_grant", "Refresh Token is invalid, expired, or revoked", "https://tools.ietf.org/html/rfc6749")
			return
		}
		client, err := h.DB.GetClient(refreshToken.GetClientId())
		if err != nil {
			writeTokenErrorResponse(w, r, "invalid_client", "The refresh_token provided does not tie to a valid client", "https://tools.ietf.org/html/rfc6749")
			return
		}
		r.Header.Set("X-User-Id", clientId)
		r.Header.Set("X-Client-Id", clientId)

		if client.GetType() == "confidential" {
			if clientId == "" || clientSecret == "" {
				writeTokenErrorResponse(w, r, "invalid_client", "Required param client_id/client_secret is missing (Or basic Auth with client credentials)", "https://tools.ietf.org/html/rfc6749")
				return
			}
		}

		//They get a subset of the original scope
		scopeRequest := strings.Split(r.PostFormValue("scope"), " ")
		scope := make([]string, 0)
		if len(scopeRequest) == 0 {
			scope = refreshToken.GetScope()
		} else {
			for _, sco := range refreshToken.GetScope() {
				for _, sc := range scopeRequest {
					if sc == sco {
						scope = append(scope, sc)
					}
				}
			}
		}

		userId := refreshToken.GetUserId()
		r.Header.Set("X-User-Id", userId)

		//Coolness all is in order to give away the access token requested
		tokenId := genUUIDv4()
		token := h.DB.NewToken()
		token.SetId(tokenId)
		token.SetType(TokenTypeBearer)
		token.SetScope(scope)
		token.SetClientId(clientId)
		token.SetUserId(userId)
		token.SetRefreshToken(refreshTokenId)
		token.SetExpires(time.Now().UTC().Add(h.AccessTokenDuration))
		h.DB.CreateToken(token)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)

		tr := tokenResponse{AccessToken: token.GetId(), TokenType: token.GetType(), ExpiresIn: int64(token.GetExpires().Sub(time.Now()).Seconds()), Scope: token.GetScope()}
		e := json.NewEncoder(w)
		err = e.Encode(tr)
		if err != nil {
			fmt.Println(err)
		}
	case "password":
		if clientId == "" || clientSecret == "" {
			writeTokenErrorResponse(w, r, "invalid_request", "Required param client_id/client_secret is missing", "https://tools.ietf.org/html/rfc6749")
			return
		}

		client, err := h.DB.VerifyClient(clientId, clientSecret)
		if err != nil {
			writeTokenErrorResponse(w, r, "invalid_client", "Unknown Client", "https://tools.ietf.org/html/rfc6749")
			return
		}
		r.Header.Set("X-User-Id", clientId)
		r.Header.Set("X-Client-Id", clientId)
		clientInternal := client.GetInternal()
		clientType := client.GetType()
		if !clientInternal || clientType != "confidential" {
			writeTokenErrorResponse(w, r, "unauthorized_client", "Unauthorized Client", "https://tools.ietf.org/html/rfc6749")
			return
		}

		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		user, err := h.DB.VerifyUser(username, password)
		if err != nil {
			//TODO Maybe rate limit for the requested user, or something
			writeTokenErrorResponse(w, r, "invalid_grant", "Invalid User Credentials", "https://tools.ietf.org/html/rfc6749")
			return
		}
		userId := user.GetId()
		r.Header.Set("X-User-Id", userId)

		asked_scope := strings.Split(r.PostFormValue("scope"), " ")
		scope := make([]string, 0)
		for _, s := range asked_scope {
			if z, _ := h.PreAuthZFunction(r, s, client, user); z == Permit {
				scope = append(scope, s)
			}
		}

		//Coolness all is in order to give away the access token requested
		tokenId := genUUIDv4()
		token := h.DB.NewToken()
		token.SetId(tokenId)
		token.SetType(TokenTypeBearer)
		token.SetScope(scope)
		token.SetClientId(clientId)
		token.SetUserId(userId)
		token.SetExpires(time.Now().UTC().Add(h.AccessTokenDuration))
		h.DB.CreateToken(token)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)

		//Maybe not always create a refresh token?
		refreshTokenId := ""
		if r.PostFormValue("access_type") == TokenAccessTypeOffline {
			refreshTokenId = genUUIDv4()
			refreshToken := h.DB.NewToken()
			refreshToken.SetId(refreshTokenId)
			refreshToken.SetType(TokenTypeRefresh)
			refreshToken.SetScope(scope)
			refreshToken.SetUserId(userId)
			refreshToken.SetClientId(clientId)
			refreshToken.SetExpires(time.Now().UTC().Add(h.RefreshTokenDuration))
			h.DB.CreateToken(refreshToken)
		}
		tr := tokenResponse{AccessToken: token.GetId(), TokenType: token.GetType(), ExpiresIn: int64(token.GetExpires().Sub(time.Now()).Seconds()), Scope: token.GetScope(), RefreshToken: refreshTokenId}
		e := json.NewEncoder(w)
		err = e.Encode(tr)
		if err != nil {
			fmt.Println(err)
		}
	}
}
