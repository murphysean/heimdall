package heimdall

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (h *Heimdall) OAuth2Authorize(w http.ResponseWriter, r *http.Request) {
	responseType := r.FormValue("response_type")
	if responseType != "code" && responseType != "token" {
		http.Error(w, "Invalid Response Type, Should be one of token or code", http.StatusBadRequest)
		return
	}
	user, err := h.getLoggedInUser(w, r)
	if err != nil {
		//Redirect to the login page
		values := url.Values{}
		values.Add("return_to", r.URL.Path+"?"+r.URL.Query().Encode())
		w.Header().Add("Location", "/login?"+values.Encode())
		w.WriteHeader(http.StatusFound)
		return
	}
	setValuesOnContext(r.Context(), user.GetId(), "heimdall")
	//r.Header.Set("X-User-Id", user.GetId())
	//r.Header.Set("X-Client-Id", "heimdall")

	//Grab the client
	clientId := r.FormValue("client_id")
	client, err := h.DB.GetClient(clientId)
	if err != nil {
		http.Error(w, "Invalid Client Id", http.StatusBadRequest)
		return
	}
	//r.Header.Set("X-Client-Id", clientId)
	setValuesOnContext(r.Context(), user.GetId(), clientId)
	//Is the redirectURI valid?
	valid := false
	redirectURIs := client.GetRedirectURIs()
	for _, redirect_uri := range redirectURIs {
		if redirect_uri != "" && redirect_uri == r.FormValue("redirect_uri") {
			valid = true
		}
	}
	if !valid {
		http.Error(w, "Invalid redirect uri", http.StatusBadRequest)
		return
	}

	redirect_uri, err := url.Parse(r.FormValue("redirect_uri"))
	if err != nil {
		http.Error(w, "Invalid redirect uri", http.StatusBadRequest)
		return
	}

	scope := r.FormValue("scope")
	scopes := strings.Split(scope, " ")

	grantedScopes := user.GetConcents(clientId)

	approvedScopes := make([]string, 0)
	finalScopes := make([]string, 0)

	allConcent := true
	for _, s := range scopes {
		//TODO Step 1: Find out if the user has access to the scope by executing a policy set
		//result := EnforcePolicySet(policySet, request)
		result := "Permit"
		if result == "Deny" {
			continue
		}
		//Step 2: If the client is internal, no need to check user grants
		if !client.GetInternal() {
			//Step 3: If the client is not internal, does it have the users concent for this scope
			prevGrant := contains(grantedScopes, s)
			if !prevGrant {
				//Check and see if the user has granted from the web app
				if !(r.FormValue("Authorize") != "" && r.FormValue(s) == "on") {
					allConcent = false
				}
			}
		}

		approvedScopes = append(approvedScopes, s)
	}

	if r.Method == "POST" && r.FormValue("concent_token") != "" {
		concentToken, err := h.DB.GetToken(r.FormValue("concent_token"))
		concentUId := concentToken.GetUserId()
		concentCId := concentToken.GetClientId()
		askedScopes := make([]string, 0)
		if err == nil {
			as := concentToken.GetScope()
			for _, s := range as {
				askedScopes = append(askedScopes, s)
			}
		}

		if r.PostFormValue("deny") == "Deny" || concentUId == "" {
			rq := redirect_uri.Query()
			rq.Set("error", "access_denied")
			rq.Set("error_description", "The resource owner has denied the request")
			rq.Set("error_uri", "http://tools.ietf.org/html/rfc6749")
			if r.FormValue("state") != "" {
				rq.Set("state", r.FormValue("state"))
			}
			if r.FormValue("response_type") == "code" {
				redirect_uri.RawQuery = rq.Encode()
			} else {
				redirect_uri.Fragment = rq.Encode()
				redirect_uri.RawQuery = ""
			}
			//Return the deny back to the client
			w.Header().Set("Location", redirect_uri.String())
			w.WriteHeader(http.StatusFound)
			return
		} else if r.PostFormValue("authorize") == "Authorize" && concentUId == user.GetId() && concentCId == clientId {
			allConcent = true
		}

		for _, s := range askedScopes {
			//For all the scopes in askedScopes check to see if the user approved
			if r.PostFormValue(s) == "on" {
				finalScopes = append(finalScopes, s)
			}
		}
		user.SetConcents(clientId, finalScopes)
		h.DB.UpdateUser(user)
	}

	allConcent = client.GetInternal() || allConcent

	if allConcent {
		if !(r.Method == "POST" && r.FormValue("concent_token") != "") {
			finalScopes = approvedScopes
		}
	} else {
		//Prompt the user for concent
		token := h.DB.NewToken()
		token.SetType(TokenTypeConcent)
		token.SetUserId(user.GetId())
		token.SetClientId(clientId)
		token.SetScope(approvedScopes)
		h.DB.CreateToken(token)
		rq := r.URL.Query()
		rq.Set("scope", strings.Join(approvedScopes, " "))
		rq.Set("concent_token", token.GetId())
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		dataMap := make(map[string]interface{})
		dataMap["Query"] = template.URL(rq.Encode())
		dataScopes := make([]map[string]interface{}, 0)
		for _, s := range approvedScopes {
			if s == "" {
				continue
			}
			scopeMap := make(map[string]interface{})
			scopeMap["PrevApproved"] = contains(grantedScopes, s)
			scopeMap["Scope"] = s
			dataScopes = append(dataScopes, scopeMap)
		}
		dataMap["Scopes"] = dataScopes
		err := h.Templates.ExecuteTemplate(w, "concent.html", dataMap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	if r.FormValue("response_type") == AuthorizationResponseTypeToken {
		//Create and Save the token
		token := h.DB.NewToken()
		token.SetType(TokenTypeBearer)
		token.SetScope(finalScopes)
		token.SetUserId(user.GetId())
		token.SetClientId(clientId)
		token.SetExpires(time.Now().Add(h.AccessTokenDuration))
		h.DB.CreateToken(token)

		rq := url.Values{}
		rq.Set("access_token", token.GetId())
		rq.Set("token_type", token.GetType())
		rq.Set("expires_in", fmt.Sprintf("%.f", token.GetExpires().Sub(time.Now()).Seconds()))
		rq.Set("scope", strings.Join(finalScopes, " "))
		if r.FormValue("access_type") == TokenAccessTypeOffline {
			refreshToken := h.DB.NewToken()
			refreshToken.SetType(TokenTypeRefresh)
			refreshToken.SetScope(finalScopes)
			refreshToken.SetUserId(user.GetId())
			refreshToken.SetClientId(clientId)
			refreshToken.SetExpires(time.Now().Add(h.RefreshTokenDuration))
			h.DB.CreateToken(refreshToken)
			rq.Set("refresh_token", refreshToken.GetId())
		}
		if r.FormValue("state") != "" {
			rq.Set("state", r.FormValue("state"))
		}
		redirect_uri.Fragment = rq.Encode()

		w.Header().Set("Location", redirect_uri.String())
		w.WriteHeader(http.StatusFound)
	} else if r.FormValue("response_type") == AuthorizationResponseTypeCode {
		//Create and Save the code
		code := h.DB.NewToken()
		code.SetType(TokenTypeCode)
		code.SetScope(finalScopes)
		code.SetUserId(user.GetId())
		code.SetClientId(clientId)
		code.SetExpires(time.Now().Add(h.AuthCodeDuration))
		if r.FormValue("access_type") == TokenAccessTypeOffline {
			code.SetAccessType(TokenAccessTypeOffline)
		}
		h.DB.CreateToken(code)
		rq := redirect_uri.Query()
		rq.Set("code", code.GetId())
		if r.FormValue("state") != "" {
			rq.Set("state", r.FormValue("state"))
		}
		redirect_uri.RawQuery = rq.Encode()
		w.Header().Set("Location", redirect_uri.String())
		w.WriteHeader(http.StatusFound)
	}
}
