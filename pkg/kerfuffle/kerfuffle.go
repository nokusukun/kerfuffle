/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package kerfuffle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/phayes/freeport"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"kerfuffle/pkg/cloudflare"
	_ "kerfuffle/pkg/logging"
	"kerfuffle/pkg/proxy_handler"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type SystemConfiguration struct {
	CloudflareCredentials string
}

type Manager struct {
	AppDataPath             string
	HttpReverseProxyManager *proxy_handler.HttpReverseProxyManager
	CloudflareZoneDir       string
	applications            map[string]*Application
	system                  *SystemConfiguration
	shutdown                chan interface{}
	installedCf             []*Cloudflare
}

func (m *Manager) GetApplication(id string) *Application {
	return m.applications[id]
}

func (m *Manager) GetAllApplications() []*Application {
	var apps []*Application
	for _, application := range m.applications {
		apps = append(apps, application)
	}
	return apps
}

func (m *Manager) SetAppMaintenanceMode(id string, state bool) error {
	app, exists := m.applications[id]
	if !exists {
		return ErrNotFound
	}
	app.MaintenanceMode = state
	for _, proxy := range app.proxies {
		for _, s := range proxy.Host {
			err := m.HttpReverseProxyManager.SetHold(s, state)
			if err != nil {
				log.Err(err).Str("route", s).Msg("failed to set hold mode on route")
			}
		}
	}
	return nil
}

// Shutdown attempts to shutdown all of the running applications peacefully and closes the m.shutdown channel
func (m *Manager) Shutdown() {
	for _, application := range m.applications {
		application.Shutdown()
	}
	close(m.shutdown)
}

// SetShutdown lets you set a channel which gets closed by the manager as soon as it's safe to shutdown.
func (m *Manager) SetShutdown(shutdown chan interface{}) {
	m.shutdown = shutdown
}

func (m *Manager) SetSystemConfiguration(system *SystemConfiguration) {
	m.system = system
}

func (m *Manager) SetHttpReverseProxyManager(HttpReverseProxyManager *proxy_handler.HttpReverseProxyManager) {
	m.HttpReverseProxyManager = HttpReverseProxyManager
}

func NewManager() *Manager {
	return &Manager{
		AppDataPath:       "app_data",
		applications:      map[string]*Application{},
		CloudflareZoneDir: ".cf-zones",
		installedCf:       []*Cloudflare{},
	}
}

func (m *Manager) Load() {
	glob, err := filepath.Glob(filepath.Join(m.AppDataPath, "*.install-info"))
	if err != nil {
		log.Err(err).Msg("failed to read directory")
		return
	}

	var configs []*InstallConfiguration

	for _, f := range glob {
		install := &InstallConfiguration{}
		open, err := os.Open(f)
		if err != nil {
			log.Err(err).Str("file", f).Msg("failed to read install file")
			continue
		}
		err = json.NewDecoder(open).Decode(install)
		if err != nil {
			err = open.Close()
			if err != nil {
				log.Err(err).Str("file", f).Msg("failed to close install file")
			}
			continue
		}

		configs = append(configs, install)

		err = open.Close()
		if err != nil {
			log.Err(err).Str("file", f).Msg("failed to close install file")
			continue
		}
	}

	for _, config := range configs {
		git, err := m.InstallFromGit(config)
		if err != nil {
			log.Err(err).Str("application", git.ID).Msg("failed to install")
			continue
		}
	}
}

// InstallConfiguration are a set of parameters needed to successfully launch
// an application in Kerfuffle. Default values are supplied through the
// DefaultInstallConfiguration(string) *InstallConfiguration method.
type InstallConfiguration struct {
	Repository    string `json:"repository,omitempty"`
	Branch        string `json:"branch,omitempty"`
	BootstrapPath string `json:"bootstrap,omitempty"`
}

func (i *InstallConfiguration) LoadDefaults() {
	if i.Branch == "" {
		i.Branch = "master"
	}

	if i.BootstrapPath == "" {
		i.BootstrapPath = ".kerfuffle"
	}
}

func DefaultInstallConfiguration(repository string) *InstallConfiguration {
	cfg := &InstallConfiguration{
		Repository: repository,
	}
	cfg.LoadDefaults()
	return cfg
}

func (m *Manager) InstallFromGit(config *InstallConfiguration) (*Application, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	config.LoadDefaults()
	app := NewApplication(config)
	app.SetAppPath(filepath.Join(m.AppDataPath, app.ID))
	log.Debug().Str("id", app.ID).Str("repository", config.Repository).Interface("config", config).Msg("installing application")

	a := m.GetApplication(app.ID)
	if a != nil {
		return nil, errors.New("application already exists")
	}

	log.Debug().Str("app", app.ID).Str("destination", app.AppPath()).Msg("cloning application")
	err := clone(app)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("app", app.ID).Msg("bootstrapping application")
	err = app.BootstrapConfigs()
	if err != nil {
		return nil, err
	}

	log.Debug().Str("app", app.ID).Msg("bootstrapping provisions")
	err = app.BootstrapProvisions()
	if err != nil {
		return nil, err
	}

	log.Debug().Str("app", app.ID).Msg("bootstrapping reverse proxies")
	err = m.bootstrapProxies(app)
	if err != nil {
		return nil, err
	}

	// todo: bootstrap cloudflare
	for _, cf := range app.cfs {
		err := m.InstallCloudflareConfiguration(cf)
		if err != nil {
			return nil, err
		}
	}

	err = m.saveConfiguration(config, app)
	if err != nil {
		return nil, err
	}

	m.applications[app.ID] = app
	return app, nil
}

func (m *Manager) InstallCloudflareConfiguration(cf *Cloudflare) error {
	// do nothing on example domains
	if cf.Zone == "example.com" {
		return nil
	}

	for _, c := range m.installedCf {
		if fmt.Sprintf("%v", c.Host) == fmt.Sprintf("%v", cf.Host) {
			log.Info().Msg("A duplicate has already been installed, will reinstall anyways")
		}
	}

	tokenBytes, err := ioutil.ReadFile(path.Join(m.CloudflareZoneDir, cf.Zone))
	if os.IsNotExist(err) {
		return fmt.Errorf("no cloudflare token found for '%v', add it to '%v'", cf.Zone, m.CloudflareZoneDir)
	}
	if err != nil {
		return err
	}
	token := strings.TrimSpace(string(tokenBytes))
	for _, host := range cf.Host {
		_, err := cloudflare.
			AutoCloudflare(token).
			SetZone(cf.Zone).
			SetDomain(host).
			Proxied(cf.Proxied).
			SendConfiguration()
		if err != nil {
			return err
		}
	}

	m.installedCf = append(m.installedCf, cf)
	return nil
}

func (m *Manager) saveConfiguration(config *InstallConfiguration, app *Application) error {
	{
		cfgBytes, err := json.Marshal(config)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(m.AppDataPath, app.ID+".install-info"), cfgBytes, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) bootstrapProxies(app *Application) error {
	// Abort when there's no HTTPManager installedCf
	if m.HttpReverseProxyManager == nil {
		return errors.New("no HttpReverseProxyManager installedCf")
	}
	for _, proxy := range app.proxies {
		if proxy.BindPort == "" {
			port, err := freeport.GetFreePort()
			if err != nil {
				return err
			}
			log.Debug().Int("port", port).Msg("using generated port")
			proxy.BindPort = fmt.Sprintf("%v", port)
		}

		target := fmt.Sprintf("http://localhost:%v", proxy.BindPort)
		for _, origin := range proxy.Host {
			err := m.HttpReverseProxyManager.InstallRoute(origin, target)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func clone(app *Application) error {
	err := os.RemoveAll(app.AppPath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	config := app.InstallConfiguration
	args := []string{"clone"}
	if config.Branch != "" {
		args = append(args, "--branch", config.Branch)
	}
	args = append(args, config.Repository, app.AppPath())

	b, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %v", string(b), err)
	}
	return nil
}
