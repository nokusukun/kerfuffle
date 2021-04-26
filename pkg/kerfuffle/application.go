/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package kerfuffle

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
	"github.com/tv42/slug"
	_ "kerfuffle/pkg/logging"
	"kerfuffle/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("resource not found")
)

var (
	StatusBooting = "booting"
	StatusRunning = "running"
	StatusFailed  = "failed"
	StatusCrashed = "crashed"
	StatusUnknown = "unknown"
)

type Application struct {
	ID                   string                `json:"id"`
	InstallConfiguration *InstallConfiguration `json:"install_configuration"`
	Meta                 *Meta                 `json:"meta"`
	Status               string                `json:"status"`
	Created              time.Time             `json:"created"`
	MaintenanceMode      bool                  `json:"maintenance_mode"`

	appPath    string
	process    map[string]*Process
	provisions map[string]*Provision
	proxies    map[string]*Proxy
	cfs        map[string]*Cloudflare
}

func NewApplication(config *InstallConfiguration) *Application {
	s, err := slug.URLString(config.Repository)
	if err != nil {
		hash := md5.Sum([]byte(config.Repository + config.Branch + config.BootstrapPath))
		s = hex.EncodeToString(hash[:])
	}
	s = fmt.Sprintf("%v@%v", s, config.Branch)
	return &Application{ID: s,
		InstallConfiguration: config,
		process:              map[string]*Process{},
		Status:               StatusBooting,
		Created:              time.Now(),
	}
}

func (a *Application) AppPath() string {
	return a.appPath
}

func (a *Application) SetAppPath(appPath string) {
	a.appPath = appPath
}

func (a *Application) GetProcess(id string) *Process {
	return a.process[id]
}

func (a *Application) GetAllProcessIds() []string {
	var keys []string
	for s := range a.process {
		keys = append(keys, s)
	}
	return keys
}

func (a *Application) GetProvision(id string) *Provision {
	return a.provisions[id]
}

func (a *Application) GetAllProvisions() map[string]*Provision {
	return a.provisions
}

func (a *Application) GetProcessStatus(id string) (*BasicProcessState, error) {
	proc := a.process[id]
	if proc == nil {
		return nil, ErrNotFound
	}
	return proc.Status(), nil
}

func (a *Application) GetAllProcessStatus() map[string]*BasicProcessState {
	var statuses = map[string]*BasicProcessState{}
	for s, process := range a.process {
		statuses[s] = process.Status()
	}
	return statuses
}

func (a *Application) GetProxy(id string) *Proxy {
	return a.proxies[id]
}

func (a *Application) GetAllProxies() map[string]*Proxy {
	return a.proxies
}

func (a *Application) GetCf(id string) *Cloudflare {
	return a.cfs[id]
}

func (a *Application) GetAllCfs() map[string]*Cloudflare {
	return a.cfs
}

func (a *Application) BootstrapConfigs() error {
	tomlPath := filepath.Join(a.AppPath(), a.InstallConfiguration.BootstrapPath)
	config, err := toml.LoadFile(tomlPath)
	if err != nil {
		log.Err(err).Str("path", tomlPath).Msg("failed to read toml file")
		return err
	}

	a.Meta = new(Meta)
	err = config.Get("meta").(*toml.Tree).Unmarshal(a.Meta)
	if err != nil {
		return err
	}
	log.Debug().Interface("meta", a.Meta).Msg("")

	a.provisions = make(map[string]*Provision)
	for _, key := range config.GetArray("provision").(*toml.Tree).Keys() {
		p := new(Provision)
		err := config.GetArray("provision").(*toml.Tree).Get(key).(*toml.Tree).Unmarshal(p)
		if err != nil {
			return err
		}
		p.Id = key
		log.Debug().Interface("provision", p).Str("id", key).Msg("loaded provision")
		a.provisions[key] = p
	}

	a.proxies = make(map[string]*Proxy)
	for _, key := range config.GetArray("proxy").(*toml.Tree).Keys() {
		p := new(Proxy)
		err := config.GetArray("proxy").(*toml.Tree).Get(key).(*toml.Tree).Unmarshal(p)
		if err != nil {
			return err
		}
		log.Debug().Interface("proxy", p).Str("id", key).Msg("loaded proxy")
		a.proxies[key] = p
	}

	a.cfs = make(map[string]*Cloudflare)
	for _, key := range config.GetArray("cloudflare").(*toml.Tree).Keys() {
		p := new(Cloudflare)
		err := config.GetArray("cloudflare").(*toml.Tree).Get(key).(*toml.Tree).Unmarshal(p)
		if err != nil {
			return err
		}
		log.Debug().Interface("cloudflare", p).Str("id", key).Msg("loaded cloudflare")
		a.cfs[key] = p
	}

	return nil
}

func (a *Application) GetUnhealthyProcesses() []*Process {
	var p []*Process
	for _, process := range a.process {
		if len(process.Errors) != 0 {
			p = append(p, process)
		}
	}
	return p
}

func (a *Application) BootstrapProvisions() error {
	init, exists := a.provisions["init"]
	if exists {
		err := a.executeProvision(init, "init")
		if err != nil {
			return fmt.Errorf("init failed to finish: %v", err)
		}
	}

	for target, provision := range a.provisions {
		if target == "init" {
			continue
		}
		log.Debug().Str("target", target).Interface("provision", provision).Msg("spawning provision")
		provision := provision
		target := target
		go func() {
			err := a.executeProvision(provision, target)
			if err != nil {
				log.Err(err).Str("id", provision.Id).Msg("provision returned an error")
			}
		}()
	}
	return nil
}

func (a *Application) ReloadProvision(target string) error {
	// For tracking the provisions being executed to set the booting flag to running
	var wg sync.WaitGroup
	wg.Add(len(a.provisions) - 1)
	go func() {
		wg.Wait()
		a.Status = StatusRunning
	}()

	if provision, exists := a.provisions[target]; exists {
		log.Debug().Str("target", target).Interface("provision", provision).Msg("reloading provision")
		_ = a.process[target].Kill()
		delete(a.process, target)
		go func() {
			defer wg.Done()
			err := a.executeProvision(provision, target)
			if err != nil {
				log.Err(err).Str("id", provision.Id).Msg("provision returned an error")
			}
		}()
		return nil
	}
	return errors.New("target provision does not exist")
}

func (a *Application) executeProvision(provision *Provision, target string) error {
	process := new(Process)
	a.process[target] = process
	process.provision = provision
	process.done = make(chan interface{}, 1)
	process.Errors = []error{}
	defer close(process.done)

	process.err = bytes.NewBufferString("")
	process.log = bytes.NewBufferString("")

	process.env = os.Environ()
	process.env = append(process.env, provision.EnvironmentVariables...)
	process.directory = filepath.Join(a.AppPath(), provision.BaseDirectory)

	if proxy, exists := a.proxies[target]; exists {
		log.Debug().Str("id", target).Str("port", proxy.BindPort).Msg("assigning port")
		process.env = append(
			process.env,
			fmt.Sprintf("APP_HOST=localhost:%v", proxy.BindPort),
			fmt.Sprintf("APP_PORT=%v", proxy.BindPort),
		)
	}

	for i, commands := range provision.Run {
		log.Info().Str("base_dir", process.directory).Str("id", provision.Id).Msgf("Launching CMD (%v/%v) '%v'", i+1, len(provision.Run), commands)
		cmd := exec.Command(commands[0], commands[1:]...)
		utils.AttachSysProcAttr(cmd)
		cmd.Dir = process.directory
		cmd.Env = process.env
		process.cmd = cmd

		cmd.Stdout = process.log
		cmd.Stderr = process.err

		err := cmd.Run()
		if err != nil {
			process.Errors = append(process.Errors, err)
			return err
		}
		log.Info().Str("id", provision.Id).Msgf("Finished CMD (%v/%v) '%v'", i+1, len(provision.Run), commands)
	}
	return nil
}

func (a *Application) Shutdown() {
	log.Debug().Str("app", a.ID).Msg("shutting down application")
	for s, process := range a.process {
		err := process.Kill()
		if err != nil {
			log.Err(err).Str("process", s).Msg("failed to kill")
		}
	}
}

func (a *Application) GetLastGitCommit() (string, error) {
	output := bytes.NewBuffer([]byte{})
	cmd := exec.Command("git", "log", "-n", "1")
	cmd.Stdout = output
	cmd.Stderr = output
	cmd.Env = os.Environ()
	cmd.Dir = a.appPath
	err := cmd.Run()
	return output.String(), err
}

func (a *Application) GetStatus() string {
	return StatusUnknown
}
