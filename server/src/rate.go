package main

import (
	"github.com/Neirpyc/cfimager/server/src/rate"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var (
	limiterLoginIp, limiterLoginTargetEmail              rate.Limiter
	limiterRegisterIp                                    rate.Limiter
	limiterResendEmailIp, limiterResendEmailTargetEmail  rate.Limiter
	limiterValidateEmailIp, limiterValidateEmailTargetId rate.Limiter
	limiterFunctionActionIp, limiterFunctionActionId     rate.Limiter
	limiterLightDBCallIp, limiterLightDBCallId           rate.Limiter
	limiterTemplateIp                                    rate.Limiter
	limiterFileServerIp                                  rate.Limiter
	limiterGetFunctionsIp                                rate.Limiter
	limiterGetFunctionsId                                rate.Limiter
	limiterSafetyId                                      rate.Limiter
	limiterSafetyIp                                      rate.Limiter
)

func init() {
	limiterLoginIp = rate.NewLimiter(rate.Rule(10, 1*time.Minute), rate.Rule(30, time.Hour))
	limiterLoginTargetEmail = rate.NewLimiter(rate.Rule(15, 1*time.Minute), rate.Rule(45, time.Hour))

	limiterRegisterIp = rate.NewLimiter(rate.Rule(10, 1*time.Minute), rate.Rule(30, time.Hour))

	limiterResendEmailIp = rate.NewLimiter(rate.Rule(2, 1*time.Minute), rate.Rule(4, 1*time.Hour))
	limiterResendEmailTargetEmail = rate.NewLimiter(rate.Rule(3, 1*time.Minute), rate.Rule(6, time.Hour))

	limiterValidateEmailIp = rate.NewLimiter(rate.Rule(2, 1*time.Minute), rate.Rule(4, 1*time.Hour))
	limiterValidateEmailTargetId = rate.NewLimiter(rate.Rule(3, 1*time.Minute), rate.Rule(6, time.Hour))

	limiterGetFunctionsIp = rate.NewLimiter(rate.Rule(100, 1*time.Second), rate.Rule(1000, time.Hour))
	limiterGetFunctionsId = rate.NewLimiter(rate.Rule(100, 1*time.Second), rate.Rule(1000, time.Hour))

	limiterFunctionActionId = rate.NewLimiter(
		rate.Rule(3, 1*time.Minute),
		rate.Rule(15, 10*time.Minute),
		rate.Rule(30, 1*time.Hour),
	)
	limiterFunctionActionIp = rate.NewLimiter(
		rate.Rule(4, 1*time.Minute),
		rate.Rule(20, 10*time.Minute),
		rate.Rule(40, 1*time.Hour),
	)

	limiterLightDBCallIp = rate.NewLimiter(rate.Rule(15, 1*time.Minute), rate.Rule(150, 30*time.Minute))
	limiterLightDBCallId = rate.NewLimiter(rate.Rule(20, 1*time.Minute), rate.Rule(200, 30*time.Minute))

	limiterTemplateIp = rate.NewLimiter(rate.Rule(45, 1*time.Minute), rate.Rule(225, 30*time.Minute))
	limiterFileServerIp = rate.NewLimiter(rate.Rule(60, 1*time.Minute), rate.Rule(300, 10*time.Minute))

	limiterSafetyIp = rate.NewLimiter(rate.Rule(30, 1*time.Minute), rate.Rule(90, 1*time.Hour))
	limiterSafetyId = rate.NewLimiter(rate.Rule(45, 1*time.Minute), rate.Rule(120, 1*time.Hour))

	go func() {
		t := time.NewTicker(1 * time.Minute)
		for {
			<-t.C
			limiterLoginIp.Clean()
			limiterLoginTargetEmail.Clean()
			limiterRegisterIp.Clean()
			limiterResendEmailIp.Clean()
			limiterResendEmailTargetEmail.Clean()
			limiterValidateEmailIp.Clean()
			limiterValidateEmailTargetId.Clean()
			limiterFunctionActionIp.Clean()
			limiterFunctionActionId.Clean()
			limiterLightDBCallIp.Clean()
			limiterLightDBCallId.Clean()
			limiterTemplateIp.Clean()
			limiterFileServerIp.Clean()
			limiterGetFunctionsIp.Clean()
			limiterGetFunctionsId.Clean()
			limiterSafetyId.Clean()
			limiterSafetyIp.Clean()
		}
	}()
}

type rateHandler func(w http.ResponseWriter, r *http.Request)

func rateLimitIp(w http.ResponseWriter, r *http.Request, limiter rate.Limiter, callback rateHandler) {
	ip := r.Header.Get("X-Forwarded-For")
	if limiter.Consume(ip) {
		callback(w, r)
	} else {
		rateLimitFail(w, r, ip)
		return
	}
}

func rateLimitIpUser(w http.ResponseWriter, r *http.Request, ipLimiter, userLimiter rate.Limiter, callback authHandler) {
	ip := r.Header.Get("X-Forwarded-For")
	if ipLimiter.Consume(ip) {
		authenticate(w, r, func(w http.ResponseWriter, r *http.Request, auth Authentication) {
			if userLimiter.Consume(auth.Id) {
				callback(w, r, auth)
			} else {
				rateLimitFail(w, r, auth.Id)
				return
			}
		})
	} else {
		rateLimitFail(w, r, ip)
		return
	}
}

func rateLimit(w http.ResponseWriter, r *http.Request, limiter rate.Limiter, key interface{}, callback rateHandler) {
	if limiter.Consume(key) {
		callback(w, r)
	} else {
		rateLimitFail(w, r, key)
	}
}

func rateLimitFail(w http.ResponseWriter, r *http.Request, key interface{}) {
	logrus.Infof("Rate limit reached for key %+v on path %s", key, r.URL.Path)
	http.Error(w, "ERR_RATE_LIMIT", http.StatusTooManyRequests)
}
