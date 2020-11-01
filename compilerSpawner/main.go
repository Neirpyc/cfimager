package main

import (
	"fmt"
	"io/ioutil"
	mathrand "math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var (
	containerId string
)

func init() {
	mathrand.Seed(time.Now().UnixNano())
}

func main() {
	containerId = os.Getenv("BUILD_CONTAINER_ID")
	http.HandleFunc("/compile", compile)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_ = http.ListenAndServe(":8080", nil)
	return
}

func compile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("begin")
	cName := fmt.Sprintf("cfimager-compiler-%x-%x-%x", time.Now().Unix(), time.Now().UnixNano(), mathrand.Uint32()%512)
	cmd := exec.Command(
		"docker",
		"run",
		"--rm",
		"--name",
		cName,
		"--network",
		"cfimager-compilers",
		"--cap-drop",
		"all",
		"--memory=50m",
		"--memory-swap=75m",
		"--kernel-memory=25m",
		"--cpus=1.5",
		"--cpu-shares=512",
		containerId,
	)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var err error
	if err = cmd.Start(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		stopCmd := exec.Command("docker", "kill", cName)
		_ = stopCmd.Start()
	}()

	pingUrl := fmt.Sprintf("http://%s:8080/ping", cName)
	stop := time.After(5 * time.Second)
	for {
		if resp, err := http.Get(pingUrl); err == nil {
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		select {
		case <-stop:
			w.WriteHeader(http.StatusInternalServerError)
			return
		default:
		}
		time.Sleep(1 * time.Millisecond)
	}

	client := http.Client{Timeout: 20 * time.Second}
	var resp *http.Response
	if resp, err = client.Post(fmt.Sprintf("http://%s:8080/compile", cName), "", r.Body); err != nil {
		if os.IsTimeout(err) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(append([]byte{0}, []byte("ERR_TIMEOUT")...))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if resp.ContentLength > 104857600 {
		_, _ = w.Write(append([]byte{0}, []byte("ERR_FILE_TOO_BIG")...))
	}

	var data []byte
	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
