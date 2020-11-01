package database

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

func zip(buf *bytes.Buffer, data []byte) (err error) {
	var w *gzip.Writer
	if w, err = gzip.NewWriterLevel(buf, gzip.BestCompression); err != nil {
		return
	}
	if _, err = w.Write(data); err != nil {
		return
	}
	err = w.Close()
	return
}

func unzip(buf *bytes.Buffer) (data []byte, err error) {
	var r *gzip.Reader
	if r, err = gzip.NewReader(buf); err != nil {
		return
	}
	if data, err = ioutil.ReadAll(r); err != nil {
		return
	}
	err = r.Close()
	return
}
