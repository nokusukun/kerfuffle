/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package main

import (
	"github.com/gin-gonic/gin"
)

type restError struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
	Path string `json:"path,omitempty"`
}

func handleErr(context *gin.Context, code int, msg string, err error) {
	e := &restError{
		Code: code,
		Msg:  msg,
		Path: context.Request.URL.String(),
	}

	ginErr := &gin.Error{
		Err:  err,
		Type: gin.ErrorTypePublic,
		Meta: e,
	}
	context.Error(ginErr)
}

func ErrMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) != 0 {
			err := c.Errors[0]
			code := 500
			if rErr, ok := err.Meta.(*restError); ok {
				code = rErr.Code
			}
			c.JSON(code, err.JSON())
		}
	}
}
