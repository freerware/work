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

package work_test

import (
	"context"
	"testing"

	"github.com/freerware/work/v4/unit"
)

const EntityCount = 500

func setupEntities() (entities []interface{}) {
	for idx := 0; idx < EntityCount; idx++ {
		entities = append(entities, Foo{ID: idx})
	}
	return
}

// BenchmarkRegister benchmarks the Register method for work units.
func BenchmarkRegister(b *testing.B) {
	entities := setupEntities()
	mappers := map[unit.TypeName]unit.DataMapper{
		unit.TypeNameOf(Foo{}): NoOpDataMapper{},
	}
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unit, err := unit.New(unit.DataMappers(mappers))
		if err != nil {
			b.FailNow()
		}
		b.StartTimer()
		if err = unit.Register(entities...); err != nil {
			b.FailNow()
		}
		b.StopTimer()
	}
}

// BenchmarkAdd benchmarks the Add method for work units.
func BenchmarkAdd(b *testing.B) {
	entities := setupEntities()
	mappers := map[unit.TypeName]unit.DataMapper{
		unit.TypeNameOf(Foo{}): NoOpDataMapper{},
	}
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unit, err := unit.New(unit.DataMappers(mappers))
		if err != nil {
			b.FailNow()
		}
		b.StartTimer()
		if err = unit.Add(entities...); err != nil {
			b.FailNow()
		}
		b.StopTimer()
	}
}

// BenchmarkAlter benchmarks the Alter method for work units.
func BenchmarkAlter(b *testing.B) {
	entities := setupEntities()
	mappers := map[unit.TypeName]unit.DataMapper{
		unit.TypeNameOf(Foo{}): NoOpDataMapper{},
	}
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unit, err := unit.New(unit.DataMappers(mappers))
		if err != nil {
			b.FailNow()
		}
		b.StartTimer()
		if err = unit.Alter(entities...); err != nil {
			b.FailNow()
		}
		b.StopTimer()
	}
}

// BenchmarkRemove benchmarks the Remove method for work units.
func BenchmarkRemove(b *testing.B) {
	entities := setupEntities()
	mappers := map[unit.TypeName]unit.DataMapper{
		unit.TypeNameOf(Foo{}): NoOpDataMapper{},
	}
	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unit, err := unit.New(unit.DataMappers(mappers))
		if err != nil {
			b.FailNow()
		}
		b.StartTimer()
		if err = unit.Remove(entities...); err != nil {
			b.FailNow()
		}
		b.StopTimer()
	}
}

func BenchmarkSave(b *testing.B) {
	ctx := context.Background()
	entities := setupEntities()
	mappers := map[unit.TypeName]unit.DataMapper{
		unit.TypeNameOf(Foo{}): NoOpDataMapper{},
	}
	b.StopTimer()
	b.ResetTimer()
	b.Run("BestEffort", func(b *testing.B) {
		b.StopTimer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			unit, err := unit.New(unit.DataMappers(mappers))
			if err != nil {
				b.FailNow()
			}
			if err = unit.Add(entities...); err != nil {
				b.FailNow()
			}
			if err = unit.Alter(entities...); err != nil {
				b.FailNow()
			}
			if err = unit.Remove(entities...); err != nil {
				b.FailNow()
			}
			b.StartTimer()
			if err = unit.Save(ctx); err != nil {
				b.FailNow()
			}
			b.StopTimer()
		}
	})
}
