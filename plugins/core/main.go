package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/takevox/loki/gen/loki/v1/lokiv1connect"
)

func main() {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	version := "1"

	var (
		err  error
		port int
		ln   net.Listener
	)
	tryCount := 0
	for {
		if tryCount >= 3 {
			slog.Error("リトライ回数が %d 回に達した")
			return
		}
		tryCount += 1

		port = 52000 + rnd.Intn(1000) // 52000-52999
		ln, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		break
	}

	endpoint := fmt.Sprintf("http://127.0.0.1:%d", port)

	fmt.Println(version)
	fmt.Println(endpoint)

	path, handler := lokiv1connect.NewPluginServiceHandler(&PluginServer{})
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	srv := &http.Server{Handler: mux}
	err = srv.Serve(ln)
	if err != nil && err != http.ErrServerClosed {
		slog.Error(err.Error())
	}
}
