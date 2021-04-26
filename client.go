/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

// client.go handles all of the embedding of the built client
// frontend through the use of go:embed

package kerfuffle

import (
	"embed"
	"io/fs"
)

//go:embed client/kerfuffle-web/build/*
var clientFS embed.FS
var ClientFS fs.FS

func init() {
	sub, err := fs.Sub(clientFS, "client/kerfuffle-web/build")
	if err != nil {
		panic(err)
	}
	ClientFS = sub
}
