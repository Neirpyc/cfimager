package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Neirpyc/cfimager/server/src/auth"
	"github.com/Neirpyc/cfimager/server/src/forms"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type RegisterRequest struct {
	Email    forms.Email    `json:"email"`
	Password []byte         `json:"password"`
	HCaptcha forms.HCaptcha `json:"hcaptcha"`
}

func registerApi(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		L.Warn("Could not decode register request body.")
	}
	send := func(code int, level logrus.Level, serverMsg string, resp SuccessError) {
		w.WriteHeader(code)
		if code != http.StatusOK || serverMsg != "" {
			if err != nil {
				L.Logf(level, "Failed to register with error \"%s\"\nErr: %s\nRequest: %+v\n", serverMsg, err.Error(), r)
			} else {
				L.Logf(level, "Failed to register with error \"%s\"\nRequest: %+v\n", serverMsg, r)
			}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			L.Warn("Could not write response in register.")
		}
	}

	var req RegisterRequest
	var resp SuccessError

	if r.Method != http.MethodPost {
		resp.Error = fmt.Sprintf("received %s request instead of %s", r.Method, http.MethodPost)
		send(http.StatusBadRequest, logrus.InfoLevel, resp.Error, resp)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		resp.Error = "could not decode request body"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}
	captcha := req.HCaptcha.IsValid(r.Header.Get("X-Forwarded-For"))
	if len(req.Password) != 64 {
		resp.Error = "bad request"
		send(http.StatusBadRequest, logrus.WarnLevel, resp.Error, resp)
		return
	}
	if !req.Email.IsValid() {
		resp.Error = "invalid email"
		send(http.StatusOK, logrus.WarnLevel, resp.Error, resp)
		return
	}
	if !<-captcha {
		resp.Error = "invalid captcha"
		send(http.StatusOK, logrus.WarnLevel, resp.Error, resp)
		return
	}

	if id, err := DB.GetId(string(req.Email)); err == nil || id != 0 {
		resp.Error = "email already taken"
		send(http.StatusOK, logrus.WarnLevel, resp.Error, resp)
		return
	}

	salt := auth.Salt()
	saltCpy := make([]byte, len(salt))
	copy(saltCpy, salt)
	hash := hashPassword(req.Password, salt)
	var id uint64
	if id, err = DB.AddTempUser(string(req.Email), hash, saltCpy); err != nil {
		resp.Error = "Internal Server Error"
		send(http.StatusInternalServerError, logrus.WarnLevel, "could not create user to database", resp)
		return
	}
	tok, _ := base64.StdEncoding.DecodeString(auth.SelfEncodeToken(id, "email"))
	if err = sendEmailValidation(req.Email, tok); err != nil {
		resp.Error = "internal server error"
		send(http.StatusInternalServerError, logrus.ErrorLevel, "could not send email validation", resp)
		return
	}
	resp.Success = true
	send(http.StatusOK, logrus.InfoLevel, "", resp)
}

func register(w http.ResponseWriter, r *http.Request) {
	if err := templateRegister.Execute(w, nil); err != nil {
		L.Warn(err)
	}
}

func postRegister(w http.ResponseWriter, r *http.Request) {
	if err := templatePostRegister.Execute(w, struct {
		Email string
	}{Email: r.URL.Query().Get("email")}); err != nil {
		L.Warn(err)
	}
}
