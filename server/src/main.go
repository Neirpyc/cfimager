package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/Neirpyc/cfimager/server/src/database"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	DB     database.CFImager
	Config Conf
	L      *logrus.Logger
)

func init() {

}

func main() {
	stop := make(chan bool, 1)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)

	handleCommandLineArguments()

	L.SetLevel(logrus.InfoLevel)
	L.SetFormatter(&logrus.TextFormatter{})
	L.Info("Started server")

	http.HandleFunc("/login/", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterTemplateIp, login)
	})
	http.HandleFunc("/register/", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterTemplateIp, register)
	})
	http.HandleFunc("/postregister/", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterTemplateIp, postRegister)
	})
	http.HandleFunc("/resend/", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterTemplateIp, resendEmailValidation)
	})
	http.HandleFunc("/list/", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIpUser(w, r, limiterLightDBCallIp, limiterLightDBCallId, functionsList)
	})
	http.HandleFunc("/editor/", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIpUser(w, r, limiterLightDBCallIp, limiterLightDBCallId, functionEditor)
	})
	http.HandleFunc("/revoke", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIpUser(w, r, limiterSafetyIp, limiterSafetyId, revoke)
	})
	http.HandleFunc("/cfimager/", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIpUser(w, r, limiterLightDBCallIp, limiterLightDBCallId, cfimager)
	})
	http.HandleFunc("/function/", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIpUser(w, r, limiterGetFunctionsIp, limiterGetFunctionsId, serveFunction)
	})
	http.HandleFunc("/v1/register", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterRegisterIp, registerApi)
	})
	http.HandleFunc("/v1/resetPassword", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterLightDBCallIp, passwordResetApi)
	})
	http.HandleFunc("/v1/logout", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterLightDBCallIp, logout)
	})
	http.HandleFunc("/v1/resetPasswordRequest", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterResendEmailIp, passwordResetRequestApi)
	})
	http.HandleFunc("/resetPassword", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterFileServerIp, passwordReset)
	})
	http.HandleFunc("/resetPasswordRequest", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterFileServerIp, passwordResetRequest)
	})
	http.HandleFunc("/v1/revoke", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIpUser(w, r, limiterSafetyIp, limiterSafetyId, revokeApi)
	})
	http.HandleFunc("/v1/delete", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIpUser(w, r, limiterLightDBCallIp, limiterLightDBCallId, deleteFunc)
	})
	http.HandleFunc("/v1/login", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterLoginIp, loginApi)
	})
	http.HandleFunc("/v1/resendEmail", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterResendEmailIp, resendEmailValidationApi)
	})
	http.HandleFunc("/v1/validateEmail", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterValidateEmailIp, validateEmail)
	})
	http.HandleFunc("/v1/createFunction", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIpUser(w, r, limiterFunctionActionIp, limiterFunctionActionId, createFunction)
	})
	http.HandleFunc("/v1/editFunction", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIpUser(w, r, limiterFunctionActionIp, limiterFunctionActionId, editFunction)
	})
	fs := http.FileServer(http.Dir("web/static"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rateLimitIp(w, r, limiterFileServerIp, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				auth := false
				authenticate(w, r, func(w http.ResponseWriter, r *http.Request, a Authentication) {
					auth = a.Success
				})
				if !auth {
					http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
					return
				}
			}
			fs.ServeHTTP(w, r)
		})
	})

	go func() {
		L.Warn(http.ListenAndServe(Config.Port, nil))
		stop <- true
	}()

	select {
	case <-sc:
		break
	case <-stop:
		break
	}
}

type Conf struct {
	Port             string
	TokenExpire      time.Duration
	TokenForceExpire time.Duration
	Secure           bool
	CompileContainer string
}

type confToml struct {
	Port             int    `toml:"port"`
	TokenExpire      string `toml:"tokenExpire"`
	TokenForceExpire string `toml:"tokenForceExpire"`
	Secure           bool   `toml:"secure"`
	CompileContainer string `toml:"compileContainer"`
}

func (c *confToml) Config() Conf {
	var C Conf
	var err error
	C.Port = ":" + strconv.FormatInt(int64(c.Port), 10)
	C.Secure = c.Secure
	C.CompileContainer = c.CompileContainer
	if C.TokenExpire, err = time.ParseDuration(c.TokenExpire); err != nil {
		panic(err)
	}
	if C.TokenForceExpire, err = time.ParseDuration(c.TokenForceExpire); err != nil {
		panic(err)
	}
	return C
}

func handleCommandLineArguments() {

	var err error
	var fileContent []byte
	if fileContent, err = ioutil.ReadFile("config.toml"); err != nil {
		panic(err)
	}
	var confToml confToml
	if err = toml.Unmarshal(fileContent, &confToml); err != nil {
		panic(err)
	}
	Config = confToml.Config()

	L = new(logrus.Logger)
	L.SetReportCaller(true)
	L.SetOutput(os.Stderr)

	passwd := os.Getenv("MYSQL_ROOT_PASSWORD")
	_ = os.Setenv("MYSQL_ROOT_PASSWORD", "42")
	if DB, err = database.NewCFImager(fmt.Sprintf("root:%s@tcp(cfimager-mariadb:3306)/", passwd), Config.TokenForceExpire); err != nil {
		panic(err)
	}
}
