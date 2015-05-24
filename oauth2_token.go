package heimdall

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"mime"
	"net/http"
	"strings"
)

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

		//Coolness all is in order to give away the access token requested
		tokenId := uuid.New()
		token := h.DB.NewToken()
		token.SetId(tokenId)
		token.SetType(TokenTypeBearer)
		token.SetScope(code.GetScope())
		token.SetUserId(code.GetUserId())
		token.SetClientId(code.GetClientId())
		h.DB.CreateToken(token)

		//Finally remove the code so it can't be reused
		h.DB.DeleteToken(authorizationCode)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)

		//Maybe not always create a refresh token?
		refreshTokenId := ""

		if code.GetAccessType() == TokenAccessTypeOffline {
			refreshTokenId = uuid.New()
			refreshToken := h.DB.NewToken()
			refreshToken.SetId(refreshTokenId)
			refreshToken.SetType(TokenTypeRefresh)
			refreshToken.SetScope(code.GetScope())
			refreshToken.SetUserId(code.GetUserId())
			refreshToken.SetClientId(code.GetClientId())
			h.DB.CreateToken(refreshToken)
		}

		if refreshTokenId == "" {
			responseJSON := `{"access_token":"%v","token_type":"%v","expires_in":3600,"scope":"%v"}`
			fmt.Fprintf(w, responseJSON, token.GetId(), token.GetType(), strings.Join(token.GetScope(), " "))
		} else {
			responseJSON := `{"access_token":"%v","token_type":"%v","expires_in":3600,"scope":"%v","refresh_token":"%v"}`
			fmt.Fprintf(w, responseJSON, token.GetId(), token.GetType(), strings.Join(token.GetScope(), " "), refreshTokenId)
		}

	case "client_credentials":
		if clientId == "" || clientSecret == "" {
			writeTokenErrorResponse(w, r, "invalid_client", "Client is required to authenticate, client_id/client_secret is missing", "https://tools.ietf.org/html/rfc6749")
			return
		}

		_, err := h.DB.VerifyClient(clientId, clientSecret)
		if err != nil {
			writeTokenErrorResponse(w, r, "invalid_client", "Unknown Client", "https://tools.ietf.org/html/rfc6749")
			return
		}

		//TODO They get what they ask for?
		scopeString := r.PostFormValue("scope")
		scope := strings.Split(scopeString, " ")

		//Coolness, all is in order to give away the access token requested
		tokenId := uuid.New()
		token := h.DB.NewToken()
		token.SetId(tokenId)
		token.SetType(TokenTypeBearer)
		token.SetScope(scope)
		token.SetClientId(clientId)
		h.DB.CreateToken(token)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)

		responseJSON := `{"access_token":"%v","token_type":"%v","expires_in":3600,"scope":"%v"}`
		fmt.Fprintf(w, responseJSON, tokenId, token.GetType(), strings.Join(scope, " "))
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
		if client.GetType() == "confidential" {
			if clientId == "" || clientSecret == "" {
				writeTokenErrorResponse(w, r, "invalid_client", "Required param client_id/client_secret is missing (Or basic Auth with client credentials)", "https://tools.ietf.org/html/rfc6749")
				return
			}
		}

		//They get a subset of the original scope
		scopeRequest := r.PostFormValue("scope")
		scopeRequests := strings.Split(scopeRequest, " ")
		scope := make([]string, 0)
		if scopeRequest == "" {
			scope = refreshToken.GetScope()
		} else {
			for _, sco := range refreshToken.GetScope() {
				for _, sc := range scopeRequests {
					if sc == sco {
						scope = append(scope, sc)
					}
				}
			}
		}

		userId := refreshToken.GetUserId()

		//Coolness all is in order to give away the access token requested
		tokenId := uuid.New()
		token := h.DB.NewToken()
		token.SetId(tokenId)
		token.SetType(TokenTypeBearer)
		token.SetScope(scope)
		token.SetClientId(clientId)
		token.SetUserId(userId)
		token.SetRefreshToken(refreshTokenId)
		h.DB.CreateToken(token)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)

		responseJSON := `{"access_token":"%v","token_type":"%v","expires_in":3600,"scope":"%v"}`
		fmt.Fprintf(w, responseJSON, tokenId, token.GetType(), strings.Join(scope, " "))
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

		//TODO They get what they ask for?
		scopeString := r.PostFormValue("scope")
		scope := strings.Split(scopeString, " ")

		//Coolness all is in order to give away the access token requested
		tokenId := uuid.New()
		token := h.DB.NewToken()
		token.SetId(tokenId)
		token.SetType(TokenTypeBearer)
		token.SetScope(scope)
		token.SetClientId(clientId)
		token.SetUserId(userId)
		h.DB.CreateToken(token)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)

		//Maybe not always create a refresh token?
		refreshTokenId := ""
		if r.PostFormValue("access_type") == TokenAccessTypeOffline {
			refreshTokenId = uuid.New()
			refreshToken := h.DB.NewToken()
			refreshToken.SetId(refreshTokenId)
			refreshToken.SetType(TokenTypeRefresh)
			refreshToken.SetScope(scope)
			refreshToken.SetUserId(userId)
			refreshToken.SetClientId(clientId)
			h.DB.CreateToken(refreshToken)
		}
		if refreshTokenId == "" {
			responseJSON := `{"access_token":"%v","token_type":"%v","expires_in":3600,"scope":"%v"}`
			fmt.Fprintf(w, responseJSON, tokenId, token.GetType(), strings.Join(scope, " "))
		} else {
			responseJSON := `{"access_token":"%v","token_type":"%v","expires_in":3600,"scope":"%v","refresh_token":"%v"}`
			fmt.Fprintf(w, responseJSON, tokenId, token.GetType(), strings.Join(scope, " "), refreshTokenId)
		}
	}
}
