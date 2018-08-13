package server

import (
	"context"
	"fmt"
		"github.com/rafaelgfirmino/heimdall/configuration"
	"github.com/rafaelgfirmino/heimdall/gateway"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
	"path/filepath"
	"plugin"
)

func StartHeimdall() {
	logger := log.New(os.Stdout, "server: ", log.LstdFlags)

	listenAddr := fmt.Sprintf(":%v", configuration.Env.Get("server.port"))

	server := http.Server{
		Addr:         listenAddr,
		Handler:      receiver()(nil),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Heimdall is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Heimdall is ready to handle requests at", listenAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Heimdall stopped")
}

func redirectRequestToService(r *http.Request, handler gateway.Handler) *http.Request {

	req, err := http.NewRequest(r.Method, handler.ServiceFullURL, r.Body)
	CheckErr(err)
	req.Header.Set("Content-Type", handler.ContentType)
	return req
}

func CheckErr(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func receiver() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			originalPathRequest := r.URL.Path
			var handler gateway.Handler
			for _, service := range gateway.NewGateway.Services {
				for _, gatewayhandler := range service.Handlers {
					if gatewayhandler.Listen == originalPathRequest {
						handler = gatewayhandler
					}
				}
			}

			if handler.Listen == "" {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not Found"))
				return
			}

			var req *http.Request
			all_plugins, err := filepath.Glob("plugins/*.so")
			if err != nil {
				panic(err)
			}

			for _, filename := range (all_plugins) {
				p, err := plugin.Open(filename)
				if err != nil {
					panic(err)
				}

				symbol, err := p.Lookup("Plugin")
				if err != nil {
					panic(err)
				}

				pluginFunc, ok := symbol.(func(http.Request))
				if !ok {
					panic("Plugin has no 'Sort([]int) []int' function")
				}

				pluginFunc(*r)
			}
			req = redirectRequestToService(r, handler)
			client := &http.Client{}
			resp, err := client.Do(req)
			CheckErr(err)

			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			CheckErr(err)

			//resp with original Content-Type
			headerResp := strings.Join(resp.Header["Content-Type"], "")
			w.Header().Set("Content-Type", headerResp)
			w.Write([]byte(body))
			// fmt.Fprintf(w, string(body))
		})
	}
}
