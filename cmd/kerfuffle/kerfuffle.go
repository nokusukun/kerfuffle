/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package main

import (
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io/ioutil"
	kerfuffleRoot "kerfuffle"
	"kerfuffle/pkg/kerfuffle"
	_ "kerfuffle/pkg/logging"
	"kerfuffle/pkg/proxy_handler"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

const (
	CfgApiBind          = "api_bind"
	CfgReverseProxyBind = "reverse_proxy_bind"
	CfgZoneDir          = "cf_zones_path"
	CFZonePath          = ".cf-zones"
)

func init() {
	viper.SetDefault(CfgApiBind, "0.0.0.0:8080")
	viper.SetDefault(CfgReverseProxyBind, "0.0.0.0:80")
	viper.SetDefault(CfgZoneDir, CFZonePath)

	viper.SetConfigName("kerfuffle")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(errors.New("failed to read configuration file"))
		}
	}

	_, err := os.Stat(CFZonePath)
	if err != nil && os.IsExist(err) {
		log.Err(err).Msg("failed to query directory")
		return
	} else {
		log.Info().Msg("Creating cloudflare configuration directory")
		err = os.MkdirAll(filepath.Join(".", viper.GetString(CfgZoneDir)), os.ModePerm)
		if err != nil {
			log.Err(err).Msgf("failed to create directory '%v'", CFZonePath)
			return
		}
	}

	// checks if the path is writeable
	err = ioutil.WriteFile(filepath.Join(viper.GetString(CfgZoneDir), ".empty"), []byte("beep"), os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	kill := make(chan interface{})

	kMan := kerfuffle.NewManager()
	kMan.CloudflareZoneDir = viper.GetString(CfgZoneDir)
	kMan.SetShutdown(kill)

	// reverse proxy bootstrapping, launches reverse proxy server, usually on port 80
	{
		revProxyMan := proxy_handler.NewHttpReverseProxyManager()
		go func(r *proxy_handler.HttpReverseProxyManager) {
			log.Info().Str("api", viper.GetString(CfgReverseProxyBind)).Msg("exposing reverse proxy")
			err := <-r.Launch(viper.GetString(CfgReverseProxyBind))
			log.Err(err).Msg("proxy manager failed")
		}(revProxyMan)
		kMan.SetHttpReverseProxyManager(revProxyMan)
	}

	// loading all of the existing stuff
	kMan.Load()

	// api services bootstrapping, starts api endpoint server on port 8080
	{
		go func(k *kerfuffle.Manager) {
			log.Info().Str("api", viper.GetString(CfgApiBind)).Msg("exposing api")

			api := NewRestApi(k).GenerateEndpoints()
			api.StaticFS("/console", http.FS(kerfuffleRoot.ClientFS))
			err := api.Run(viper.GetString(CfgApiBind))
			if err != nil {
				log.Err(err).Msg("api endpoint failed")
			}
		}(kMan)
	}

	go func() {
		<-signals
		// Shutdown signals all of the running application to at least pack up before being rudely shutdown
		kMan.Shutdown()
	}()

	// waits for the kill channel to close
	for range kill {
	}
	log.Info().Msg("kerfuffle has been terminated")
}
