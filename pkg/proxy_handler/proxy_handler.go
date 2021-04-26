/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package proxy_handler

import (
	"context"
	_ "embed"
	"errors"
	"github.com/rs/zerolog/log"
	"kerfuffle"
	_ "kerfuffle/pkg/logging"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Host = string

var (
	//go:embed assets/html/compiled_index.html
	SiteIndex []byte
	//go:embed assets/html/compiled_maintenance.html
	SiteMaintenance []byte
)

type Route struct {
	Origin *url.URL
	Target *url.URL
	Proxy  *httputil.ReverseProxy

	hold bool
}

type HttpReverseProxyManager struct {
	routes map[Host]*Route

	stop chan interface{}
}

func NewHttpReverseProxyManager() *HttpReverseProxyManager {
	return &HttpReverseProxyManager{routes: make(map[Host]*Route), stop: make(chan interface{})}
}

func (m *HttpReverseProxyManager) UninstallRoute(originAddr string) error {
	origin, err := url.Parse(originAddr)
	if err != nil {
		return err
	}

	if origin.Host == "" && originAddr != "" {
		origin.Host = originAddr
	} else {
		return errors.New("origin host cannot be empty")
	}

	if _, exists := m.routes[origin.Host]; !exists {
		return errors.New("origin host isn't installed")
	}

	delete(m.routes, origin.Host)
	return nil
}

func (m *HttpReverseProxyManager) InstallRoute(originAddr string, targetAddr string) error {
	origin, err := url.Parse(originAddr)
	if err != nil {
		return err
	}

	if origin.Host == "" && originAddr != "" {
		origin.Host = originAddr
	} else {
		return errors.New("origin host cannot be empty")
	}

	if _, exists := m.routes[origin.Host]; exists {
		return errors.New("origin host already exists")
	}

	target, err := url.Parse(targetAddr)
	if err != nil {
		return err
	}

	if target.Host == "" {
		return errors.New("target host cannot be empty")
	}

	log.Debug().Str("origin", origin.Host).Str("target", target.Host).Msg("registering route")
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = func(response *http.Response) error {
		response.Header.Set("X-Kerfuffle-Version", kerfuffle.Version)
		return nil
	}

	m.routes[origin.Host] = &Route{
		Origin: origin,
		Target: target,
		Proxy:  proxy,
	}

	return nil
}

func (m *HttpReverseProxyManager) SetHold(originAddr string, value bool) error {
	origin, err := url.Parse(originAddr)
	if err != nil {
		return err
	}

	if origin.Host == "" && originAddr != "" {
		origin.Host = originAddr
	} else {
		return errors.New("origin host cannot be empty")
	}

	if route, exists := m.routes[origin.Host]; !exists {
		return errors.New("origin host isn't installed")
	} else {
		route.hold = value
	}
	return nil
}

func (m *HttpReverseProxyManager) Stop() {
	m.stop <- struct{}{}
}

func (m *HttpReverseProxyManager) Launch(addr string) chan error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		route, exists := m.routes[req.Host]
		if !exists {
			log.Error().Str("path", req.Host).Msgf("host not found")
			_, err := res.Write(SiteIndex)
			if err != nil {
				log.Err(err).Stack().Msg("failed to write")
			}
			return
		}

		if route.hold {
			_, err := res.Write(SiteMaintenance)
			if err != nil {
				log.Err(err).Stack().Msg("failed to write")
			}
			return
		}

		log.Info().
			Str("method", req.Method).
			Str("origin", req.Host).
			Str("target", route.Target.String()).
			Str("path", req.URL.String()).
			Msg("proxy")

		req.URL.Host = route.Target.Host
		req.URL.Scheme = route.Target.Scheme
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Host = route.Target.Host
		route.Proxy.ServeHTTP(res, req)
	})

	errChan := make(chan error)

	srv := &http.Server{Addr: addr, Handler: mux}
	go func(srv *http.Server) {
		err := srv.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}(srv)

	go func(srv *http.Server) {
		<-m.stop
		log.Debug().Msg("stopping server")
		err := srv.Shutdown(context.Background())
		if err != nil {
			errChan <- err
		}
	}(srv)

	return errChan
}
