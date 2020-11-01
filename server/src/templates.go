package main

import (
	"html/template"
	"os"
)

var (
	HCaptchaSiteKey string
)

func init() {
	HCaptchaSiteKey = os.Getenv("HCAPTCHA_PUBLIC")
}

var (
	templateLogin           *template.Template
	templateRegister        *template.Template
	templatePostRegister    *template.Template
	templateReSendEmail     *template.Template
	templatePostReSendEmail *template.Template
	templateFunctionList    *template.Template
	templateFunctionEditor  *template.Template
	templateCFImager        *template.Template
	templateRevoke          *template.Template
	templateReset           *template.Template
	templateResetReq        *template.Template
)

func getSiteKey() string {
	return HCaptchaSiteKey
}

func init() {
	var err error

	fMap := template.FuncMap{"sitekey": getSiteKey}

	if templateLogin, err = template.New("login.html").Funcs(fMap).ParseFiles("web/templates/login.html"); err != nil {
		panic(err)
	}
	if templateRegister, err = template.New("register.html").Funcs(fMap).ParseFiles("web/templates/register.html"); err != nil {
		panic(err)
	}
	if templatePostRegister, err = template.New("postregister.html").Funcs(fMap).ParseFiles("web/templates/postregister.html"); err != nil {
		panic(err)
	}
	if templateReSendEmail, err = template.New("resend.html").Funcs(fMap).ParseFiles("web/templates/resend.html"); err != nil {
		panic(err)
	}
	if templateFunctionList, err = template.New("list.html").Funcs(fMap).ParseFiles("web/templates/list.html"); err != nil {
		panic(err)
	}
	if templateFunctionEditor, err = template.New("editor.html").Funcs(fMap).ParseFiles("web/templates/editor.html"); err != nil {
		panic(err)
	}
	if templatePostReSendEmail, err = template.New("postresend.html").Funcs(fMap).ParseFiles("web/templates/postresend.html"); err != nil {
		panic(err)
	}
	if templateCFImager, err = template.New("cfimager.html").Funcs(fMap).ParseFiles("web/templates/cfimager.html"); err != nil {
		panic(err)
	}
	if templateRevoke, err = template.New("revoke.html").Funcs(fMap).ParseFiles("web/templates/revoke.html"); err != nil {
		panic(err)
	}
	if templateReset, err = template.New("passwordReset.html").Funcs(fMap).ParseFiles("web/templates/passwordReset.html"); err != nil {
		panic(err)
	}
	if templateResetReq, err = template.New("passwordResetRequest.html").Funcs(fMap).ParseFiles("web/templates/passwordResetRequest.html"); err != nil {
		panic(err)
	}
}
