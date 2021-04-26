/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package proxy_handler

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/txn2/txeh"
	"io/ioutil"
	_ "kerfuffle/pkg/logging"
	"net/http"
	"testing"
	"time"
)

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "error")
	}
}

func TestHttpReverseProxyManager_Launch(t *testing.T) {
	const targetAUrl = "http://localhost:5123"
	const targetBUrl = "http://localhost:5124"
	const hostA = "host-aa.local"
	const hostB = "host-b.local"
	const hostAPort = ":5123"
	const hostBPort = ":5124"

	var done = make(chan interface{})

	var ok = []byte("OK_A")
	var ok_1 = []byte("OK_A_1")
	var ok_b = []byte("OK_B")
	var ok_1_b = []byte("OK_B_1")

	t.Run("target-a-http_server", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				errorHandler(w, r, http.StatusNotFound)
				return
			}
			_, err := w.Write(ok)
			if err != nil {
				t.Error(err)
			}
		})
		mux.HandleFunc("/1", func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(ok_1)
			if err != nil {
				t.Error(err)
			}
		})
		srv := &http.Server{Addr: hostAPort, Handler: mux}
		go func(srv *http.Server) {
			err := srv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				t.Error(err)
				return
			}
		}(srv)

		for range done {
		}
		log.Info().Msg("closing server A")
		err := srv.Shutdown(context.Background())
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("target-b-http_server", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				errorHandler(w, r, http.StatusNotFound)
				return
			}
			_, err := w.Write(ok_b)
			if err != nil {
				t.Error(err)
			}
		})
		mux.HandleFunc("/1", func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(ok_1_b)
			if err != nil {
				t.Error(err)
			}
		})

		srv := &http.Server{Addr: hostBPort, Handler: mux}
		go func(srv *http.Server) {
			err := srv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				t.Error(err)
				return
			}
		}(srv)

		for range done {
		}
		log.Info().Msg("closing server B")
		err := srv.Shutdown(context.Background())
		if err != nil {
			t.Error(err)
			return
		}
	})

	// Provisioning
	{
		hosts, err := txeh.NewHostsDefault()
		if err != nil {
			log.Panic().Err(err)
		}
		hosts.AddHost("127.0.0.1", hostA)
		hosts.AddHost("127.0.0.1", hostB)
		hosts.AddHost("127.0.0.1", "host-c.local")
		err = hosts.Save()
		if err != nil {
			log.Panic().Err(err)
		}
	}

	t.Run("proxy-handler", func(t *testing.T) {
		t.Parallel()

		proxyManager := NewHttpReverseProxyManager()
		err := proxyManager.InstallRoute(hostA, targetAUrl)
		if err != nil {
			t.Error()
		}
		err = proxyManager.InstallRoute(hostB, targetBUrl)
		if err != nil {
			panic(err)
		}
		log.Info().Msg("Listening on :80")
		proxyManager.Launch(":80")
		for range done {
		}
		proxyManager.Stop()
	})

	time.Sleep(time.Second)

	t.Run("request check", func(t *testing.T) {
		t.Parallel()
		type I struct {
			host   string
			path   string
			expect []byte
		}

		var items = []I{
			{
				hostA,
				"/",
				ok,
			},
			{
				hostA,
				"/1",
				ok_1,
			},
			{
				hostA,
				"/not_exist",
				[]byte("error"),
			},
			{
				hostB,
				"/",
				ok_b,
			},
			{
				hostB,
				"/1",
				ok_1_b,
			},
		}

		for _, item := range items {
			log.Info().Msgf("Testing: %v", item)
			client := &http.Client{}
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://%v%v", item.host, item.path), nil)
			req.Header.Set("Host", item.host)
			res, err := client.Do(req)
			if err != nil {
				log.Err(err).Msg("failed request")
				t.Fail()
			}

			all, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Err(err).Msg("")
				t.Fail()
			}

			if string(all) != string(item.expect) {
				log.Error().Str("got", string(all)).Str("expect", string(item.expect)).Msg("failed to get expected value")
				t.Fail()
			}
		}
		log.Info().Msg("Finishing testing")
		close(done)
	})
}
