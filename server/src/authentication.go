package main

import (
	"encoding/json"
	"fmt"
	"github.com/Neirpyc/cfimager/server/src/auth"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type Authentication struct {
	Success bool
	Id      uint64
}

type authHandler func(w http.ResponseWriter, r *http.Request, auth Authentication)

func authenticate(w http.ResponseWriter, r *http.Request, callback authHandler) {
	var err error
	var cookie *http.Cookie
	var authentication Authentication

	defer func() {
		if !authentication.Success {
			http.Redirect(w, r, "https://cfimager.neirpyc.ovh/login", http.StatusTemporaryRedirect)
		} else {
			callback(w, r, authentication)
		}
	}()

	if cookie, err = r.Cookie("token"); err != nil {
		return
	}
	elems := strings.Split(cookie.Value, ":")
	if len(elems) != 2 {
		return
	}

	var token auth.Token
	if token, authentication.Success = auth.ValidateSelfEncoded(elems[0]); authentication.Success {
		if str, ok := token.Metadata.(string); !ok || str != "auth" {
			return
		}
		authentication.Id = token.Id
		authentication.Success = true
		return
	}

	var dbToken []byte
	if authentication.Id, authentication.Success, dbToken = auth.ValidateRefreshToken(elems[1]); authentication.Success {
		if err = DB.RetrieveAuthToken(dbToken); err != nil {
			L.Info("Could not retrieve refresh token for ", authentication.Id)
			authentication.Success = false
			authentication.Id = 0
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    auth.SelfEncodeToken(authentication.Id, "auth") + ":" + elems[1],
			Expires:  cookie.Expires,
			Secure:   Config.Secure,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})
	}
}

func revokeApi(w http.ResponseWriter, r *http.Request, authentication Authentication) {
	var err error

	send := func(code int, level logrus.Level, serverMsg string, resp SuccessError) {
		w.WriteHeader(code)
		if code != http.StatusOK || serverMsg != "" || !resp.Success {
			if err != nil {
				L.Logf(level, "Failed to login with error \"%s\"\nErr: %s\nRequest: %+v\n", serverMsg, err.Error(), r)
			} else {
				L.Logf(level, "Failed to login with error \"%s\"\nRequest: %+v\n", serverMsg, r)
			}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			L.Warn("Could not write response in login.")
		}
		w.Header().Set("content-type", "text/html; charset=utf-8")
		http.Redirect(w, r, "../login", http.StatusTemporaryRedirect)
	}

	var resp SuccessError

	if r.Method != http.MethodPost {
		resp.Error = fmt.Sprintf("received %s request instead of %s", r.Method, http.MethodPost)
		send(http.StatusBadRequest, logrus.InfoLevel, resp.Error, resp)
		return
	}
	auth.Revoke(authentication.Id)
	if err = DB.RevokeUsersTokens(authentication.Id); err != nil {
		resp.Error = "could not revoke tokens, you should change your password as soon as possible"
		send(http.StatusInternalServerError, logrus.ErrorLevel, err.Error(), resp)
		return
	}
	resp.Success = true
	send(http.StatusOK, logrus.InfoLevel, "", resp)
}

func revoke(w http.ResponseWriter, r *http.Request, authentication Authentication) {
	if err := templateRevoke.Execute(w, nil); err != nil {
		L.Warn(err)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		L.Infof("received %s request instead of %s", r.Method, http.MethodPost)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		Secure:   Config.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	if cookie, err := r.Cookie("token"); err != nil {
		return
	} else {
		elems := strings.Split(cookie.Value, ":")
		if len(elems) != 2 {
			return
		}
		_, v, db := auth.ValidateRefreshToken(elems[1])
		if v {
			if err := DB.DeleteAuthToken(db); err != nil {
				L.Error("Could not delete auth token")
			}
		}
	}
	w.WriteHeader(http.StatusOK)
}
