package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Neirpyc/cfimager/server/src/auth"
	"github.com/Neirpyc/cfimager/server/src/forms"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

type LoginRequest struct {
	Email    forms.Email    `json:"email"`
	Password []byte         `json:"password"`
	HCaptcha forms.HCaptcha `json:"hcaptcha"`
}

func loginApi(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		L.Warn("Could not decode login request body.")
	}
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
	}

	var req LoginRequest
	var resp SuccessError

	if r.Method != http.MethodPost {
		resp.Error = fmt.Sprintf("received %s request instead of %s", r.Method, http.MethodPost)
		send(http.StatusBadRequest, logrus.InfoLevel, resp.Error, resp)
		return
	}
	if err = json.Unmarshal(body, &req); err != nil {
		resp.Error = "could not decode request body"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}
	ok := false
	rateLimit(w, r, limiterLoginTargetEmail, req.Email, func(w http.ResponseWriter, r *http.Request) {
		ok = true
	})
	if !ok {
		return
	}
	ch := req.HCaptcha.IsValid(r.Header.Get("X-Forwarded-For"))
	if len(req.Password) != 64 || !req.Email.IsValid() {
		resp.Error = "invalid request"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}
	if !<-ch {
		resp.Error = "invalid captcha"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}

	var salt, expectedHash []byte
	var id uint64
	if id, err = DB.GetUserAuthData(string(req.Email), &expectedHash, &salt); err != nil {
		resp.Error = "invalid account -> register/validate email"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}
	hash := hashPassword(req.Password, salt)
	if !bytes.Equal(hash, expectedHash) {
		resp.Error = "invalid password"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}

	token, dbToken := auth.RefreshToken(id)
	if err = DB.RegisterAuthToken(id, dbToken); err != nil {
		resp.Error = "internal server error"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}
	resp.Success = true
	fmt.Printf("Issuing self encoded token for id %d\n", id)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    auth.SelfEncodeToken(id, "auth") + ":" + token,
		Expires:  time.Now().Add(Config.TokenForceExpire),
		Secure:   Config.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	send(http.StatusOK, logrus.InfoLevel, "", resp)
}

func login(w http.ResponseWriter, _ *http.Request) {
	if err := templateLogin.Execute(w, nil); err != nil {
		L.Warn(err)
	}
}
