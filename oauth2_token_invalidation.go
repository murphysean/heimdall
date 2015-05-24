package heimdall

import (
	"net/http"
	"strings"
)

func (h *Heimdall) OAuth2TokenInvalidation(w http.ResponseWriter, r *http.Request) {
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

	if token.GetType() == TokenTypeRefresh {
		//TODO Delete any tokens that were created under this refresh token
	}

	if r.Method == "DELETE" {
		h.DB.DeleteToken(tokenId)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
}
