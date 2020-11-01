package forms

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

var (
	emailRegexp          *regexp.Regexp
	publicKey, secretKey string
)

func init() {
	emailRegexp = regexp.MustCompile("(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])")
	publicKey = os.Getenv("HCAPTCHA_PUBLIC")
	secretKey = os.Getenv("HCAPTCHA_SECRET")
	_ = os.Setenv("HCAPTCHA_SECRET", "42")
}

type Email string
type HCaptcha string

func (e Email) IsValid() bool {
	return emailRegexp.MatchString(string(e))
}

type hCaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTs string   `json:"challenge_ts"`
	HostName    string   `json:"hostname"`
	Credit      bool     `json:"credit"`
	ErrorCodes  []string `json:"error-codes"`
}

func (h HCaptcha) IsValid(ip string) <-chan bool {
	valid := make(chan bool, 1)
	go func(valid chan<- bool) {
		formData := url.Values{}
		formData.Add("response", string(h))
		formData.Add("secret", secretKey)
		if ip != "" {
			formData.Add("remoteip", ip)
		}
		formData.Add("sitekey", publicKey)
		resp, err := http.PostForm("https://hcaptcha.com/siteverify", formData)
		if err != nil {
			valid <- false
			return
		}
		var captchaResp hCaptchaResponse
		_ = json.NewDecoder(resp.Body).Decode(&captchaResp)
		if captchaResp.Success && captchaResp.HostName == "neirpyc.ovh" {
			valid <- true
			return
		}
		valid <- false
	}(valid)
	return valid
}
