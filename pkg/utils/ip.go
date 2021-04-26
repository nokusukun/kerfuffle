/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package utils

import (
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: time.Second * 15}

func GetIP() (string, error) {
	get, err := httpClient.Get("https://v4.ident.me/")
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Debug().Err(err).Msg("failed GET IP request")
		}
	}(get.Body)

	all, err := ioutil.ReadAll(get.Body)
	if err != nil {
		return "", err
	}
	return string(all), nil
}
