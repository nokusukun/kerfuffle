/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package main

import (
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"kerfuffle/pkg/kerfuffle"
	"net/http"
	"time"
)

var (
	ErrApplicationNotExist = errors.New("application does not exist")
)

type RestApi struct {
	manager *kerfuffle.Manager
}

func NewRestApi(manager *kerfuffle.Manager) *RestApi {
	return &RestApi{manager: manager}
}

func (r *RestApi) GenerateEndpoints() *gin.Engine {
	mux := gin.Default()
	mux.Use(cors.Default())
	mux.Use(ErrMiddleware())
	api := mux.Group("/api")
	r.v1ApiGenerate(api.Group("/v1"))
	return mux
}

func (r *RestApi) v1ApiGenerate(v1 *gin.RouterGroup) {
	application := v1.Group("/application")
	{
		application.POST("", func(context *gin.Context) {
			config := &kerfuffle.InstallConfiguration{}
			err := context.ShouldBind(config)
			if err != nil {
				handleErr(context, http.StatusBadRequest, "", err)
				return
			}
			app, err := r.manager.InstallFromGit(config)
			if err != nil {
				handleErr(context, http.StatusBadRequest, "", err)
				return
			}
			context.JSON(200, app)
			return
		})

		application.GET("", func(context *gin.Context) {
			context.JSON(200, r.manager.GetAllApplications())
		})

		application.DELETE("/:id", func(context *gin.Context) {
			id := context.Param("id")
			app := r.manager.GetApplication(id)
			if app == nil {
				handleErr(context, http.StatusNotFound, id, ErrApplicationNotExist)
				return
			}
		})

		application.GET("/:id", func(context *gin.Context) {
			id := context.Param("id")
			app := r.manager.GetApplication(id)
			if app == nil {
				handleErr(context, http.StatusNotFound, id, ErrApplicationNotExist)
				return
			}
			lastCommit, _ := app.GetLastGitCommit()
			context.JSON(200, gin.H{
				"application": app,
				"provisions":  app.GetAllProvisions(),
				"proxies":     app.GetAllProxies(),
				"cfs":         app.GetAllCfs(),
				"processes":   app.GetAllProcessStatus(),
				"last_commit": lastCommit,
			})
		})

		application.PATCH("/:id/hold", func(context *gin.Context) {
			id := context.Param("id")
			app := r.manager.GetApplication(id)
			if app == nil {
				handleErr(context, http.StatusNotFound, id, ErrApplicationNotExist)
				return
			}
			err := r.manager.SetAppMaintenanceMode(app.ID, !app.MaintenanceMode)
			context.JSON(200, gin.H{"error": err})
		})

		application.GET("/:id/processes", func(context *gin.Context) {
			id := context.Param("id")
			app := r.manager.GetApplication(id)
			if app == nil {
				handleErr(context, http.StatusNotFound, id, ErrApplicationNotExist)
				return
			}
			context.JSON(200, app.GetAllProcessIds())
		})

		application.GET("/:id/provisions", func(context *gin.Context) {
			id := context.Param("id")
			app := r.manager.GetApplication(id)
			if app == nil {
				handleErr(context, http.StatusNotFound, id, ErrApplicationNotExist)
				return
			}
			context.JSON(200, app.GetAllProvisions())
		})

		application.GET("/:id/provision/:provisionId/output/:t", func(context *gin.Context) {
			id := context.Param("id")
			provision := context.Param("provisionId")
			t := context.Param("t")
			app := r.manager.GetApplication(id)
			if app == nil {
				handleErr(context, http.StatusNotFound, id, ErrApplicationNotExist)
				return
			}
			process := app.GetProcess(provision)
			log.Debug().Interface("process", process).Msg("")
			if process == nil {
				handleErr(context, http.StatusNotFound, provision, errors.New("process does not exist"))
				return
			}
			switch t {
			case "log":
				context.String(200, process.Log().String())
			case "err":
				context.String(200, process.Err().String())
			default:
				log.Error().Str("buffer", t).Msg("buffer does not exist")
				handleErr(context, http.StatusNotFound, provision, errors.New("buffer does not exist"))
			}
		})

		application.GET("/:id/provision/:provisionId/reload", func(context *gin.Context) {
			id := context.Param("id")
			target := context.Param("provisionId")
			app := r.manager.GetApplication(id)
			if app == nil {
				handleErr(context, http.StatusNotFound, id, ErrApplicationNotExist)
				return
			}
			err := app.ReloadProvision(target)
			if err != nil {
				handleErr(context, http.StatusInternalServerError, id, err)
				return
			}
			context.String(200, "ok")
		})

	}

	debug := v1.Group("/debug")

	debug.GET("/shutdown", yellowTape, func(context *gin.Context) {
		r.manager.Shutdown()
		context.String(200, "kerfuffle is shutting down in 1 second")
	})

	debug.GET("/force_error", yellowTape, func(context *gin.Context) {
		handleErr(context, http.StatusBadRequest, "hello world", errors.New("big boy error"))
	})
}

// Todo: for testing purposes only, delete in the future
func yellowTape(c *gin.Context) {
	// Get the Basic Authentication credentials
	if x, _ := c.Cookie("debug_auth"); x == "yes" {
		return
	}
	user, password, hasAuth := c.Request.BasicAuth()
	if hasAuth && user == "testuser" && password == "testpass" {
		log.Debug().Str("user", user).Msg("User authenticated")
		c.SetCookie("debug_auth", "yes", int(time.Hour.Seconds()), "", "", false, false)
	} else {
		c.Status(401)
		c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		c.Abort()
		return
	}
}
