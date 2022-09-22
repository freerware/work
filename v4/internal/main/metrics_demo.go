/* Copyright 2022 Freerware
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/freerware/work/v4"
	"github.com/uber-go/tally"
	tstatsd "github.com/uber-go/tally/statsd"
)

/* Data Mapper Definition */

type demoDataMapper struct{}

func (dm *demoDataMapper) simulateLatency() {
	latency := time.Duration(rand.Intn(maximumLatencyMilliseconds)) * time.Millisecond
	time.Sleep(latency)
}

func (dm *demoDataMapper) simulateOperation() (err error) {
	isError := rand.Intn(5) == 4
	if isError {
		err = errors.New("oops")
	}
	return
}

func (dm *demoDataMapper) simulate() error {
	dm.simulateLatency()
	return dm.simulateOperation()
}

func (dm *demoDataMapper) Insert(ctx context.Context, mCtx work.MapperContext, e ...interface{}) error {
	return dm.simulate()
}

func (dm *demoDataMapper) Update(ctx context.Context, mCtx work.MapperContext, e ...interface{}) error {
	return dm.simulate()
}

func (dm *demoDataMapper) Delete(ctx context.Context, mCtx work.MapperContext, e ...interface{}) error {
	return dm.simulate()
}

/* Setup Options */

func setupScope() tally.Scope {
	statter, err := statsd.NewBufferedClient("127.0.0.1:8125", "demo", 150, 512)
	if err != nil {
		panic(err)
	}
	reporter := tstatsd.NewReporter(statter, tstatsd.Options{
		SampleRate: 1.0,
	})
	scope, _ := tally.NewRootScope(tally.ScopeOptions{
		Tags:     map[string]string{},
		Reporter: reporter,
	}, time.Second)
	return scope
}

func setupDataMapper() map[work.TypeName]work.DataMapper {
	dm := &demoDataMapper{}
	return map[work.TypeName]work.DataMapper{
		work.TypeNameOf(foo{}): dm,
	}
}

func o() []work.UnitOption {
	return []work.UnitOption{
		work.UnitScope(setupScope()),
		work.UnitDataMappers(setupDataMapper()),
	}
}

/* Entity Definition */

type foo struct{}

/* Demo */

const (
	maximumLatencyMilliseconds  = 150
	saveAttempts                = 500
	maximumEntitiesPerOperation = 50
)

func main() {
	for i := 0; i < saveAttempts; i++ {
		unit, err := work.NewUnit(o()...)
		if err != nil {
			panic(err)
		}

		registrations := []interface{}{}
		for j := 0; j < rand.Intn(maximumEntitiesPerOperation); j++ {
			registrations = append(registrations, foo{})
		}
		if err = unit.Register(registrations...); err != nil {
			panic(err)
		}

		additions := []interface{}{}
		for j := 0; j < rand.Intn(maximumEntitiesPerOperation); j++ {
			additions = append(additions, foo{})
		}
		if err = unit.Add(additions...); err != nil {
			panic(err)
		}

		alters := []interface{}{}
		for j := 0; j < rand.Intn(maximumEntitiesPerOperation); j++ {
			alters = append(alters, foo{})
		}
		if err = unit.Alter(alters...); err != nil {
			panic(err)
		}

		removals := []interface{}{}
		for j := 0; j < rand.Intn(maximumEntitiesPerOperation); j++ {
			removals = append(removals, foo{})
		}
		if err = unit.Remove(removals...); err != nil {
			panic(err)
		}

		unit.Save(context.Background())
	}
}
