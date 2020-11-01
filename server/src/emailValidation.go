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

func sendEmailValidation(email forms.Email, token []byte) error {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(MailerRequest{
		Email: email,
		Token: token,
	}); err != nil {
		return err
	}
	if resp, err := http.Post("http://cfimager-mailer-api:8080/sendRegisterCode", "", buf); err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return errors.New("could not send email")
	}
	return nil
}

func resendEmailValidation(w http.ResponseWriter, r *http.Request) {
	if err := templateReSendEmail.Execute(w, nil); err != nil {
		L.Warn(err)
	}
}

func resendEmailValidationApi(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    forms.Email    `json:"email"`
		HCaptcha forms.HCaptcha `json:"hcaptcha"`
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		L.Warn("Could not decode send email request body.")
	}
	send := func(code int, level logrus.Level, serverMsg string, resp SuccessError) {
		w.WriteHeader(code)
		if code != http.StatusOK || serverMsg != "" || resp.Success == false {
			if err != nil {
				L.Logf(level, "Failed to send email with error \"%s\"\nErr: %s\nRequest: %+v\nBody: %s", serverMsg, err.Error(), r, string(body))
			} else {
				L.Logf(level, "Failed to send emailwith error \"%s\"\nRequest: %+v\nBody: %s", serverMsg, r, string(body))
			}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			L.Warn("Could not write response in send email.")
		}
	}
	var resp SuccessError
	if r.Method != http.MethodPost {
		resp.Error = fmt.Sprintf("received %s request instead of %s", r.Method, http.MethodPost)
		send(http.StatusBadRequest, logrus.InfoLevel, resp.Error, resp)
		return
	}
	if err = json.Unmarshal(body, &req); err != nil {
		resp.Error = "could not decode request body"
		send(http.StatusBadRequest, logrus.WarnLevel, resp.Error, resp)
		return
	}
	ok := false
	rateLimit(w, r, limiterResendEmailTargetEmail, req.Email, func(w http.ResponseWriter, r *http.Request) {
		ok = true
	})
	if !ok {
		return
	}
	if !req.Email.IsValid() {
		resp.Error = fmt.Sprintf("invalid email %s", req.Email)
		send(http.StatusOK, logrus.WarnLevel, resp.Error, resp)
		return
	}
	if !<-req.HCaptcha.IsValid(r.Header.Get("X-Forwarded-For")) {
		resp.Error = "invalid captcha"
		send(http.StatusOK, logrus.WarnLevel, resp.Error, resp)
		return
	}
	var id uint64
	if id, err = DB.GetTempUserId(string(req.Email)); err != nil {
		resp.Error = fmt.Sprintf("no email validation pending for email %s", req.Email)
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}

	tok, _ := base64.StdEncoding.DecodeString(auth.SelfEncodeToken(id, "email"))
	if err = sendEmailValidation(req.Email, tok); err != nil {
		resp.Error = "Internal server error"
		send(http.StatusOK, logrus.WarnLevel, resp.Error, resp)
		return
	}
	resp.Success = true
	send(http.StatusOK, logrus.InfoLevel, "", resp)
}

func validateEmail(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp SuccessError
	token := r.URL.Query().Get("token")
	send := func(code int, level logrus.Level, serverMsg string, resp SuccessError) {
		if code != http.StatusOK || serverMsg != "" || !resp.Success {
			w.WriteHeader(code)
			if err != nil {
				L.Logf(level, "Failed to validate email with error \"%s\"\nErr: %s\nBody:%+v\nToken: %s\nT", serverMsg, err.Error(), r, token)
			} else {
				L.Logf(level, "Failed to validate email with error \"%s\"\nBody:%+v\nToken: %s\n", serverMsg, r, token)
			}
			if err = templatePostReSendEmail.Execute(w, resp); err != nil {
				L.Warn("Could not write response in validate email.")
			}
		} else {
			http.Redirect(w, r, "../login", http.StatusTemporaryRedirect)
		}
	}
	var t auth.Token
	var valid bool
	if t, valid = auth.ValidateSelfEncoded(token); !valid {
		resp.Error = "invalid or expired token"
		send(http.StatusBadRequest, logrus.InfoLevel, resp.Error, resp)
		return
	}
	if str, ok := t.Metadata.(string); !ok || str != "email" {
		resp.Error = "wrong token type"
		send(http.StatusBadRequest, logrus.InfoLevel, fmt.Sprintf("Wrong token type %+v instead of email", t.Metadata), resp)
		return
	}
	id := t.Id
	ok := false
	rateLimit(w, r, limiterValidateEmailTargetId, id, func(w http.ResponseWriter, r *http.Request) {
		ok = true
	})
	if !ok {
		return
	}
	if err = DB.ValidateUser(id); err != nil {
		resp.Error = "internal server error"
		send(http.StatusBadRequest, logrus.WarnLevel, resp.Error, resp)
		return
	}
	resp.Success = true
	send(http.StatusOK, logrus.InfoLevel, "", resp)
}
