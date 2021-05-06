/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package utils

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"reflect"
	"time"
)

type Debounce struct {
	Interval  time.Duration
	lastRun   time.Time
	lastValue []interface{}
}

func NewDebounce(interval time.Duration) *Debounce {
	return &Debounce{
		Interval:  interval,
		lastValue: []interface{}{},
	}
}

func (d *Debounce) Run(fn interface{}, args ...interface{}) []interface{} {
	if time.Now().Sub(d.lastRun) < d.Interval {
		log.Debug().Msg("returning debounced data")
		return d.lastValue
	}

	log.Debug().Msg("returning true data")
	refFn := reflect.ValueOf(fn)
	if refFn.Kind() != reflect.Func {
		panic(fmt.Errorf("debounce function is not a func"))
	}

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	var values []interface{}
	for _, value := range refFn.Call(in) {
		values = append(values, value.Interface())
	}

	d.lastValue = values
	d.lastRun = time.Now()

	return d.lastValue
}
