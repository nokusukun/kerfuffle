/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package kerfuffle

import (
	"bytes"
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/process"
	"kerfuffle/pkg/utils"
	"os/exec"
)

type Process struct {
	directory string
	cmd       *exec.Cmd
	env       []string
	log       *bytes.Buffer
	err       *bytes.Buffer
	Errors    []error
	done      chan interface{}
	provision *Provision

	killFunction context.CancelFunc
}

func (p *Process) Kill() error {
	log.Trace().Str("process", p.cmd.String()).Msg("killing process...")
	proc, err := process.NewProcess(int32(p.cmd.Process.Pid))
	if err != nil {
		return err
	}
	err = utils.KillAllFamilyTree(proc)
	if err != nil {
		return err
	}
	_, err = p.cmd.Process.Wait()
	log.Trace().Str("process", p.cmd.String()).Msg("killed")
	return err
}

func (p *Process) Log() *bytes.Buffer {
	return p.log
}

func (p *Process) Err() *bytes.Buffer {
	return p.err
}

func (p *Process) GetErrors() []error {
	return p.Errors
}

func (p *Process) Wait() {
	for range p.done {
	}
}

func (p *Process) Status() *BasicProcessState {
	if p.cmd.ProcessState != nil {
		return &BasicProcessState{
			false,
			p.cmd.ProcessState.String(),
		}
	}
	return &BasicProcessState{
		true,
		fmt.Sprintf("running: %v", p.cmd.String()),
	}
}

type BasicProcessState struct {
	Alive  bool   `json:"alive"`
	Status string `json:"status,omitempty"`
}
