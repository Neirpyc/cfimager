package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/gomail.v2"
	"net/http"
	url2 "net/url"
)

type Request struct {
	Email string
	Token []byte
}

func main() {
	http.HandleFunc("/sendRegisterCode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println("bad method")
			return
		}

		var req Request
		var err error
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(err)
			return
		}

		url := url2.Values{}
		url.Add("token", base64.StdEncoding.EncodeToString(req.Token))

		m := gomail.NewMessage()
		m.SetHeader("From", "no-reply@cfimager.neirpyc.ovh")
		m.SetHeader("To", req.Email)
		m.SetHeader("Subject", "CF Imager - Account Validation")
		m.SetBody("text/html", fmt.Sprintf(
			` <a href="%s" target="_blank">Validate your email.</a> `,
			"https://cfimager.neirpyc.ovh/v1/validateEmail?"+url.Encode()))

		d := gomail.Dialer{
			Host:      "cfimager-mailer",
			Port:      587,
			TLSConfig: &tls.Config{InsecureSkipVerify: true},
		}
		if err = d.DialAndSend(m); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Println(err)
		return
	})
	http.HandleFunc("/sendPasswordResetCode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println("bad method")
			return
		}

		var req Request
		var err error
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(err)
			return
		}

		url := url2.Values{}
		url.Add("token", base64.StdEncoding.EncodeToString(req.Token))

		m := gomail.NewMessage()
		m.SetHeader("From", "no-reply@cfimager.neirpyc.ovh")
		m.SetHeader("To", req.Email)
		m.SetHeader("Subject", "CF Imager - Password Reset")
		m.SetBody("text/html", fmt.Sprintf(
			`Click <a href="%s" target="_blank">here</a> to change your CFImager password. If you did not request this, you can safely ignore this email.`,
			"https://cfimager.neirpyc.ovh/resetPassword?"+url.Encode()))

		d := gomail.Dialer{
			Host:      "cfimager-mailer",
			Port:      587,
			TLSConfig: &tls.Config{InsecureSkipVerify: true},
		}
		if err = d.DialAndSend(m); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	})
	panic(http.ListenAndServe(":8080", nil))
}
