package heimdall

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (h *Heimdall) OAuth2TokenInfo(w http.ResponseWriter, r *http.Request) {
	tokenId := r.FormValue("access_token")
	//If not found via query, look at the authorization header
	if authorization := r.Header.Get("Authorization"); tokenId == "" && strings.HasPrefix(authorization, "Bearer ") {
		tokenId = r.Header.Get("Authorization")[7:]
	}
	if tokenId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token, err := h.DB.GetToken(tokenId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if time.Now().After(token.GetExpires()) {
		w.WriteHeader(http.StatusNotFound)
		//TODO Delete the token?
		return
	}

	if r.Method == "GET" || r.Method == "POST" {
		tokenInfo := make(map[string]interface{})
		tokenInfo["audience"] = token.GetClientId()
		tokenInfo["scope"] = token.GetScope()
		if token.GetUserId() != "" {
			tokenInfo["userid"] = token.GetUserId()
		}
		tokenInfo["expires_in"] = fmt.Sprintf("%.f", token.GetExpires().Sub(time.Now()))
		tokenInfo["type"] = token.GetType()

		s, err := json.Marshal(&tokenInfo)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(s)
	}
}
