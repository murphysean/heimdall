package heimdall

import (
	"net/http"
	"time"
)

func (h *Heimdall) Login(w http.ResponseWriter, r *http.Request) {
	user, err := h.getLoggedInUser(w, r)
	if err != nil {
		if r.Method == "POST" {
			username := r.PostFormValue("login")
			password := r.PostFormValue("password")
			if username != "" && password != "" {
				user, err = h.DB.VerifyUser(username, password)
				if err == nil {
					session := h.DB.NewToken()
					session.SetType(TokenTypeSession)
					session.SetClientId("heimdall")
					session.SetUserId(user.GetId())
					session.SetExpires(time.Now().UTC().Add(h.SessionDuration))
					h.DB.CreateToken(session)
					cookie := http.Cookie{}
					cookie.Name = "session-id"
					cookie.Value = session.GetId()
					cookie.Secure = true
					cookie.HttpOnly = true
					w.Header().Add("Set-Cookie", cookie.String())
					//Set Headers so the rest of the application can get at the user and client ids
					r.Header.Set("X-User-Id", user.GetId())
					r.Header.Set("X-Client-Id", "heimdall")
				}
			}
		}
	}
	if user == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if w.Header().Get("Authorization") != "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Heimdall"`)
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		h.Templates.ExecuteTemplate(w, "login.html", r.URL.Query().Get("return_to"))
		return
	}

	if returnTo := r.FormValue("return_to"); returnTo != "" {
		w.Header().Set("Location", returnTo)
		w.WriteHeader(http.StatusFound)
		return
	} else {
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusFound)
		return
	}
}
