package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Neirpyc/cfimager/server/src/auth"
	"github.com/Neirpyc/cfimager/server/src/forms"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type PasswordResetReq struct {
	HCaptcha    forms.HCaptcha `json:"hcaptcha"`
	NewPassword []byte         `json:"new_password"`
	Token       string         `json:"token"`
}

func passwordResetApi(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		L.Warn("Could not decode password reset request body.")
	}
	send := func(code int, level logrus.Level, serverMsg string, resp SuccessError) {
		w.WriteHeader(code)
		if code != http.StatusOK || serverMsg != "" {
			if err != nil {
				L.Logf(level, "Failed to password reset with error \"%s\"\nErr: %s\nRequest: %+v\nBody: %s", serverMsg, err.Error(), r, string(body))
			} else {
				L.Logf(level, "Failed to password reset with error \"%s\"\nRequest: %+v\nBody: %s", serverMsg, r, string(body))
			}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			L.Warn("Could not write response in password reset.")
		}
	}

	var req PasswordResetReq
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

	token, ok := auth.ValidateSelfEncoded(req.Token)
	if !ok {
		resp.Error = "invalid token"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}
	if str, ok := token.Metadata.(string); !ok || str != "passwordReset" {
		resp.Error = "Wrong token type"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}

	ok = false
	rateLimit(w, r, limiterLightDBCallId, token, func(w http.ResponseWriter, r *http.Request) {
		ok = true
	})
	if !ok {
		return
	}

	captcha := req.HCaptcha.IsValid(r.Header.Get("X-Forwarded-For"))
	if len(req.NewPassword) != 64 {
		resp.Error = "bad request"
		send(http.StatusBadRequest, logrus.WarnLevel, resp.Error, resp)
		return
	}
	if !<-captcha {
		resp.Error = "invalid captcha"
		send(http.StatusOK, logrus.WarnLevel, resp.Error, resp)
		return
	}

	salt := auth.Salt()
	saltCpy := make([]byte, len(salt))
	copy(saltCpy, salt)
	fmt.Printf("Password: %v\nSalt   : %v\n", req.NewPassword, salt)
	hash := hashPassword(req.NewPassword, saltCpy)
	fmt.Printf("Password: %v\nSalt   : %v\nHash    : %v\n", req.NewPassword, salt, hash)
	if err = DB.ChangePassword(token.Id, hash, saltCpy); err != nil {
		resp.Error = "internal server error"
		send(http.StatusOK, logrus.ErrorLevel, resp.Error, resp)
		return
	}
	auth.Revoke(token.Id)
	if err = DB.RevokeUsersTokens(token.Id); err != nil {
		resp.Error = "internal server error, you password has been modified"
		send(http.StatusInternalServerError, logrus.ErrorLevel, err.Error(), resp)
		return
	}

	resp.Success = true
	send(http.StatusOK, logrus.InfoLevel, "", resp)
}

type PasswordResetReqReq struct {
	HCaptcha forms.HCaptcha `json:"hcaptcha"`
	Email    forms.Email    `json:"email"`
}

func passwordResetRequestApi(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		L.Warn("Could not decode password reset request request body.")
	}
	send := func(code int, level logrus.Level, serverMsg string, resp SuccessError) {
		w.WriteHeader(code)
		if code != http.StatusOK || serverMsg != "" {
			if err != nil {
				L.Logf(level, "Failed to password reset request with error \"%s\"\nErr: %s\nRequest: %+v\nBody: %s", serverMsg, err.Error(), r, string(body))
			} else {
				L.Logf(level, "Failed to password reset request with error \"%s\"\nRequest: %+v\nBody: %s", serverMsg, r, string(body))
			}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			L.Warn("Could not write response in password request reset.")
		}
	}

	var req PasswordResetReqReq
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

	if !<-req.HCaptcha.IsValid(r.Header.Get("X-Forwarded-For")) {
		resp.Error = "invalid captcha"
		send(http.StatusOK, logrus.WarnLevel, resp.Error, resp)
		return
	}

	var id uint64
	if id, err = DB.GetId(string(req.Email)); err != nil {
		resp.Error = "Unknown email"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}
	fmt.Printf("Id: %v\n", id)

	tok, _ := base64.StdEncoding.DecodeString(auth.SelfEncodeToken(id, "passwordReset"))
	if err = sendPasswordResetEmail(req.Email, tok); err != nil {
		resp.Error = "internal server error"
		send(http.StatusOK, logrus.ErrorLevel, resp.Error, resp)
		return
	}

	resp.Success = true
	send(http.StatusOK, logrus.InfoLevel, "", resp)
}

func passwordReset(w http.ResponseWriter, r *http.Request) {
	if err := templateReset.Execute(w, r); err != nil {
		L.Warn("Could not write response in validate email.", err)
	}
}

func passwordResetRequest(w http.ResponseWriter, r *http.Request) {
	if err := templateResetReq.Execute(w, r); err != nil {
		L.Warn("Could not write response in validate email.", err)
	}
}

func sendPasswordResetEmail(email forms.Email, token []byte) error {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(MailerRequest{
		Email: email,
		Token: token,
	}); err != nil {
		return err
	}
	if resp, err := http.Post("http://cfimager-mailer-api:8080/sendPasswordResetCode", "", buf); err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return errors.New("could not send email")
	}
	return nil
}
