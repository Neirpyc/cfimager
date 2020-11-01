package main

import (
	"archive/tar"
	bytes "bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Neirpyc/cfimager/server/src/database"
	"github.com/Neirpyc/cfimager/server/src/forms"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	mathrand "math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	funcNameRegex *regexp.Regexp
)

func init() {
	funcNameRegex = regexp.MustCompile("^.*\\/function\\/(.*)\\/(cfimager\\.(?:worker\\.)?(?:js|wasm))$")
	mathrand.Seed(time.Now().Unix())
}

type createFunctionForm struct {
	Name string `json:"name"`
}

type SuccessError struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func createFunction(w http.ResponseWriter, r *http.Request, auth Authentication) {
	var resp SuccessError
	var req createFunctionForm
	var err error
	var id uint64
	var respBody []byte
	if r.Method != http.MethodPost {
		goto invalid
	}
	if auth.Success == false {
		return
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		goto invalid
	}

	if id, err = DB.NewFunction(auth.Id, req.Name); err != nil {
		goto invalid
	}

	resp.Success = true
	resp.Error = strconv.FormatUint(id, 10)
	if respBody, err = json.Marshal(resp); err != nil {
		goto invalid
	}
	_, _ = w.Write(respBody)
	return
invalid:
	w.WriteHeader(http.StatusBadRequest)
}

type editFunctionForm struct {
	Id              uint64         `json:"id"`
	NewName         string         `json:"newname"`
	HCaptcha        forms.HCaptcha `json:"hcaptcha"`
	NewContent      string         `json:"content"`
	ContentModified bool           `json:"modified"`
}

type EditFunctionResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
	Source  string `json:"source,omitempty"`
}

func editFunction(w http.ResponseWriter, r *http.Request, auth Authentication) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		L.Warn("Could not decode edit function request body.")
	}
	send := func(code int, level logrus.Level, serverMsg string, resp EditFunctionResponse) {
		w.WriteHeader(code)
		if code != http.StatusOK || serverMsg != "" {
			if err != nil {
				L.Logf(level, "Failed to edit function with error \"%s\"\nErr: %s\nRequest: %+v\nBody: %s", serverMsg, err.Error(), r, string(body))
			} else {
				L.Logf(level, "Failed to edit function with error \"%s\"\nRequest: %+v\nBody: %s", serverMsg, r, string(body))
			}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			L.Warn("Could not write response in edit function.")
		}
	}
	var resp EditFunctionResponse
	var req editFunctionForm
	if r.Method != http.MethodPost {
		resp.Error = fmt.Sprintf("received %s request instead of %s", r.Method, http.MethodPost)
		send(http.StatusBadRequest, logrus.InfoLevel, resp.Error, resp)
		return
	}
	if auth.Success == false {
		resp.Error = fmt.Sprintf("invalid authentication")
		send(http.StatusBadRequest, logrus.InfoLevel, resp.Error, resp)
		return
	}
	if err = json.Unmarshal(body, &req); err != nil {
		resp.Error = "could not decode request body"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}
	if !<-req.HCaptcha.IsValid(r.Header.Get("X-Forwarded-For")) {
		resp.Error = "invalid captcha"
		send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
		return
	}

	if req.ContentModified && req.NewName == "" {
		if err = DB.SetSource(auth.Id, req.Id, req.NewContent); err != nil {
			resp.Error = "could not edit function content"
			send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
			return
		}
	} else if req.NewName != "" && req.ContentModified {
		if err = DB.SetSourceAndName(auth.Id, req.Id, req.NewContent, req.NewName); err != nil {
			resp.Error = "could not edit function content and name"
			send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
			return
		}
	} else if req.NewName != "" {
		if err = DB.SetName(auth.Id, req.Id, req.NewName); err != nil {
			resp.Error = "could not edit function name"
			send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
			return
		}
	}

	var data []byte
	if data, err = compileFunction(auth.Id, req.Id); err != nil {
		if err.Error() == compilationError().Error() {
			resp.Error = "compilation failed"
			resp.Message = string(data)
			send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
			return
		} else if err.Error() == "ERR_TIMEOUT" {
			resp.Error = err.Error()
			resp.Message = "compilation timeout"
			send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
			return
		} else if err.Error() == "ERR_NOT_ENOUGH_SPACE" {
			resp.Error = err.Error()
			resp.Message = "You have reached the maximum size you can store on our server"
			send(http.StatusOK, logrus.InfoLevel, resp.Error, resp)
			return
		} else {
			resp.Error = "could not compile function"
			send(http.StatusInternalServerError, logrus.ErrorLevel, resp.Error, resp)
			return
		}
	}

	resp.Success = true
	send(http.StatusOK, logrus.InfoLevel, "", resp)
}

type FunctionEditorContent struct {
	FunctionName  string
	FunctionCode  string
	FunctionError string
}

func functionEditor(w http.ResponseWriter, r *http.Request, auth Authentication) {
	if auth.Success == false {
		return
	}

	var err error
	var id uint64
	if id, err = strconv.ParseUint(r.URL.Query().Get("id"), 10, 64); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		L.Warn(err)
		return
	}

	var source, compileError string
	if source, compileError, err = DB.GetSourceAndError(auth.Id, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		L.Warn(err)
		return
	}

	if err = templateFunctionEditor.Execute(w, FunctionEditorContent{
		FunctionName:  r.URL.Query().Get("name"),
		FunctionCode:  source,
		FunctionError: compileError,
	}); err != nil {
		L.Warn(err)
	}
}

func functionsList(w http.ResponseWriter, r *http.Request, auth Authentication) {
	if auth.Success == false {
		return
	}

	var funcs database.Functions
	var err error

	if funcs, err = DB.ListFunctions(auth.Id); err != nil {
		L.Errorf("Could not get function list for user %+v with err %+v\n", auth, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := templateFunctionList.Execute(w, funcs); err != nil {
		L.Warn(err)
	}
}

func compileFunction(ownerId, funcId uint64) (data []byte, err error) {
	var source string
	if source, err = DB.GetSource(ownerId, funcId); err != nil {
		return
	}

	var resp *http.Response
	L.Info("post begin")
	if resp, err = http.Post(
		"http://cfimager-compilers-spawner:8080/compile",
		"",
		bytes.NewBuffer([]byte(source)),
	); err != nil {
		L.Info("err")
		return nil, err
	}
	L.Info("post end")

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid response from compiler1")
	}

	L.Info("read begin")
	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, errors.New("invalid response from compiler")
	}
	L.Info("read end")

	if len(data) < 1 {
		return nil, errors.New("invalid response from compiler2")
	}

	L.Info("DB begin")
	if data[0] != 1 {
		if string(data[1:]) == "ERR_TIMEOUT" {
			_ = DB.SetCompiledAndError(ownerId, funcId, nil, nil, nil, string(data[1:]))
			return data[1:], errors.New(string(data[1:]))
		}
		_ = DB.SetCompiledAndError(ownerId, funcId, nil, nil, nil, string(data[1:]))
		return data[1:], compilationError()
	}
	L.Info("DB end")

	data = data[1:]

	tr := tar.NewReader(bytes.NewReader(data))
	var ret, wasm, js, wjs []byte
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			L.Warn(err)
		}
		content, err := ioutil.ReadAll(tr)
		if err != nil {
			return nil, err
		}

		switch hdr.Name {
		case "cfimager.wasm":
			wasm = content
		case "cfimager.js":
			js = content
		case "cfimager.worker.js":
			wjs = content
		default:
			L.Warn("Unexpected tar content", hdr.Name)
		}
	}
	L.Info("DB begin")
	if err = DB.SetCompiledAndError(
		ownerId,
		funcId,
		wasm,
		js,
		wjs,
		"",
	); err != nil {
		return
	}
	L.Info("DB end")
	L.Info("Successfully updated function")
	return ret, nil
}

func serveFunction(w http.ResponseWriter, r *http.Request, auth Authentication) {
	if !auth.Success {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var data []byte
	var err error
	matches := funcNameRegex.FindStringSubmatch(r.RequestURI)
	if matches == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var id uint64
	if id, err = strconv.ParseUint(matches[1], 10, 64); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	L.Info(auth.Id, id, matches[2])
	switch matches[2] {
	case "cfimager.wasm":
		data, err = DB.GetWasm(auth.Id, id)
		w.Header().Set("content-type", "application/wasm")
	case "cfimager.js":
		data, err = DB.GetJS(auth.Id, id)
		w.Header().Set("content-type", "application/javascript")
	case "cfimager.worker.js":
		data, err = DB.GetWJS(auth.Id, id)
		w.Header().Set("content-type", "application/javascript")
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		L.Info(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("cross-origin-opener-policy", "same-origin")
	w.Header().Set("cross-origin-embedder-policy", "require-corp")
	_, _ = w.Write(data)
}

func deleteFunc(w http.ResponseWriter, r *http.Request, auth Authentication) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		L.Warn("Could not decode deleteFunc request body.")
	}
	send := func(code int, level logrus.Level, serverMsg string, resp SuccessError) {
		w.WriteHeader(code)
		if code != http.StatusOK || serverMsg != "" || resp.Success == false {
			if err != nil {
				L.Logf(level, "Failed to deleteFunc with error \"%s\"\nErr: %s\nRequest: %+v\nBody: %s", serverMsg, err.Error(), r, string(body))
			} else {
				L.Logf(level, "Failed to deleteFunc with error \"%s\"\nRequest: %+v\nBody: %s", serverMsg, r, string(body))
			}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			L.Warn("Could not write response in deleteFunc.")
		}
	}

	var req uint64
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

	if err = DB.DeleteFunction(auth.Id, req); err != nil {
		resp.Error = "could not delete function"
		send(http.StatusOK, logrus.WarnLevel, err.Error(), resp)
		return
	}

	resp.Success = true
	send(http.StatusOK, logrus.InfoLevel, "", resp)
}
