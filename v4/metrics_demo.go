/* Copyright 2021 Freerware
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

package work

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/uber-go/tally"
	tstatsd "github.com/uber-go/tally/statsd"
)

/* Data Mapper Definition */

type demoDataMapper struct{}

func (dm *demoDataMapper) simulateLatency() {
	latency := time.Duration(rand.Intn(500)) * time.Millisecond
	time.Sleep(latency)
}

func (dm *demoDataMapper) simulateOperation() (err error) {
	isError := rand.Intn(5) == 5
	if isError {
		err = errors.New("oops")
	}
	return
}

func (dm *demoDataMapper) simulate() error {
	dm.simulateLatency()
	return dm.simulateOperation()
}

func (dm *demoDataMapper) Insert(ctx context.Context, mCtx MapperContext, e ...interface{}) error {
	return dm.simulate()
}

func (dm *demoDataMapper) Update(ctx context.Context, mCtx MapperContext, e ...interface{}) error {
	return dm.simulate()
}

func (dm *demoDataMapper) Delete(ctx context.Context, mCtx MapperContext, e ...interface{}) error {
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

func setupDataMapper() map[TypeName]DataMapper {
	dm := &demoDataMapper{}
	return map[TypeName]DataMapper{
		TypeNameOf(foo{}): dm,
	}
}

func o() []UnitOption {
	return []UnitOption{
		UnitScope(setupScope()),
		UnitDataMappers(setupDataMapper()),
	}
}

/* Entity Definition */

type foo struct{}

/* Demo */

func main() {
	for i := 0; i < 10000; i++ {
		unit, err := NewUnit(o()...)
		if err != nil {
			panic(err)
		}

		additions := []interface{}{}
		for j := 0; j < rand.Intn(50); j++ {
			additions = append(additions, foo{})
		}
		if err = unit.Add(additions...); err != nil {
			panic(err)
		}

		alters := []interface{}{}
		for j := 0; j < rand.Intn(50); j++ {
			alters = append(alters, foo{})
		}
		if err = unit.Alter(alters...); err != nil {
			panic(err)
		}

		removals := []interface{}{}
		for j := 0; j < rand.Intn(50); j++ {
			removals = append(removals, foo{})
		}
		if err = unit.Remove(removals...); err != nil {
			panic(err)
		}

		unit.Save(context.Background())
	}
}
