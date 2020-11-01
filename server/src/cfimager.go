package main

import (
	"net/http"
	"strconv"
)

type cfimagerContent struct {
	Id uint64
}

func cfimager(w http.ResponseWriter, r *http.Request, auth Authentication) {
	if auth.Success == false {
		return
	}

	var id uint64
	var err error
	if id, err = strconv.ParseUint(r.URL.Query().Get("id"), 10, 64); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := templateCFImager.Execute(w, cfimagerContent{Id: id}); err != nil {
		L.Warn(err)
	}
}
