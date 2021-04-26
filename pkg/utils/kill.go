/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package utils

import (
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/process"
)

func KillAllFamilyTree(proc *process.Process) error {
	children, err := proc.Children()
	if err != nil {
		log.Err(err).Str("process", proc.String()).Msg("failed to query process children")
	}

	log.Trace().Str("process", proc.String()).Msg("killing children process")
	for _, child := range children {
		_ = KillAllFamilyTree(child)
	}

	log.Trace().Str("process", proc.String()).Msg("killing parent process")
	return proc.Terminate()
}
