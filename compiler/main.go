package main

import (
	"archive/tar"
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	http.HandleFunc("/compile", compile)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_ = http.ListenAndServe(":8080", nil)
}

func compile(w http.ResponseWriter, r *http.Request) {
	var err error
	var src []byte

	var errBytes []byte
	errBuffer := bytes.NewBuffer(errBytes)

	defer func() {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte{0})
			_, _ = w.Write(errBuffer.Bytes())
		}
	}()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if src, err = ioutil.ReadAll(r.Body); err != nil {
		return
	}

	if err := ioutil.WriteFile("src/flib.c", src, os.ModePerm); err != nil {
		return
	}

	cmd := exec.Command(
		"/bin/sh",
		"-c",
		`emcc  -s EXTRA_EXPORTED_RUNTIME_METHODS='["cwrap", "ccall"]' -s EXPORTED_FUNCTIONS='["_main", "_draw", "_stop_render", "_set_render_settings"]' -s USE_PTHREADS=1 -O3 -s EXPORT_ALL=0 -s ALLOW_MEMORY_GROWTH=1 -s ASSERTIONS=1 -c src/flib.c -o objs/flib.o`,
	)
	cmd.Stderr = errBuffer
	if err = cmd.Run(); err != nil {
		/*if _, ok := err.(*exec.ExitError); ok {
			return
		}*/
		return
	}
	cmd = exec.Command(
		"/bin/sh",
		"-c",
		`emcc -s EXTRA_EXPORTED_RUNTIME_METHODS='["cwrap", "ccall"]' -s EXPORTED_FUNCTIONS='["_main", "_draw", "_stop_render", "_set_render_settings"]' -s USE_PTHREADS=1 -O3 -s EXPORT_ALL=0 -s ALLOW_MEMORY_GROWTH=1 -s ASSERTIONS=1 objs/*.o -o cfimager.js`,
	)
	cmd.Stderr = errBuffer
	if err = cmd.Run(); err != nil {
		/*if _, ok := err.(*exec.ExitError); ok {
			return
		}*/
		return
	}
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	files := []string{"cfimager.js", "cfimager.worker.js", "cfimager.wasm"}
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return
		}
		if err = tw.WriteHeader(&tar.Header{
			Name: file,
			Size: int64(len(data)),
			Mode: 0600,
		}); err != nil {
			return
		}
		if _, err = tw.Write(data); err != nil {
			return
		}
	}
	if err = tw.Close(); err != nil {
		return
	}

	if buf.Len() > 104857600 {
		_, _ = w.Write(append([]byte{0}, []byte("ERR_FILE_TOO_BIG")...))
	}

	if _, err = w.Write([]byte{1}); err != nil {
		return
	}
	if _, err = w.Write(buf.Bytes()); err != nil {
		return
	}
}
