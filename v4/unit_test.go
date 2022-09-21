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

package work_test

import (
	"sync"
	"testing"

	"github.com/freerware/work/v4"
	"github.com/freerware/work/v4/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

type UnitTestSuite struct {
	suite.Suite

	// system under test.
	sut work.Unit

	// mocks.
	scope   tally.TestScope
	mappers map[work.TypeName]*mock.DataMapper
	mc      *gomock.Controller

	// metrics scope names and tags.
	scopePrefix string
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) SetupTest() {
	// test entities.
	foo := Foo{ID: 28}
	fooTypeName := work.TypeNameOf(foo)
	bar := Bar{ID: "28"}
	barTypeName := work.TypeNameOf(bar)

	// initialize mocks.
	s.mc = gomock.NewController(s.T())
	s.mappers = make(map[work.TypeName]*mock.DataMapper)
	s.mappers[fooTypeName] = mock.NewDataMapper(s.mc)
	s.mappers[barTypeName] = mock.NewDataMapper(s.mc)

	// construct SUT.
	dm := make(map[work.TypeName]work.DataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}

	s.scopePrefix = "test"
	c := zap.NewDevelopmentConfig()
	c.DisableStacktrace = true
	l, _ := c.Build()
	ts := tally.NewTestScope(s.scopePrefix, map[string]string{})
	s.scope = ts
	var err error
	opts := []work.UnitOption{work.UnitDataMappers(dm), work.UnitLogger(l), work.UnitScope(ts)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)
}

/*
  Combine _Empty(), _Add(), _MissingDataMapper()

  Use table driven tests with these elements:
  - unit --> either s.sut or result of constructor call.
  - entities --> slice of entities to pass in as args.
  - err --> nil when no error is expected, or expected error.
*/

func (s *UnitTestSuite) TestUnit_NewUnit_NoDataMappers() {

	// action.
	var err error
	dm := map[work.TypeName]work.DataMapper{}
	opts := []work.UnitOption{work.UnitDataMappers(dm)}
	s.sut, err = work.NewUnit(opts...)

	// assert.
	s.EqualError(err, work.ErrNoDataMapper.Error())
}

func (s *UnitTestSuite) TestUnit_Add_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Add(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Add_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
	}
	mappers := map[work.TypeName]work.DataMapper{
		work.TypeNameOf(Bar{}): &mock.DataMapper{},
	}
	var err error
	opts := []work.UnitOption{work.UnitDataMappers(mappers)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)

	// action.
	err = s.sut.Add(entities...)

	// assert.
	s.Error(err)
}

func (s *UnitTestSuite) TestUnit_Add() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}

	// action.
	err := s.sut.Add(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_ConcurrentAdd() {

	// arrange.
	foo := Foo{ID: 28}
	bar := Bar{ID: "28"}

	// action.
	var err, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err = s.sut.Add(foo)
		wg.Done()
	}()
	go func() {
		err2 = s.sut.Add(bar)
		wg.Done()
	}()
	wg.Wait()

	// assert.
	s.NoError(err)
	s.NoError(err2)
}

func (s *UnitTestSuite) TestUnit_Alter_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Alter(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Alter_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
	}
	mappers := map[work.TypeName]work.DataMapper{
		work.TypeNameOf(Bar{}): &mock.DataMapper{},
	}
	var err error
	opts := []work.UnitOption{work.UnitDataMappers(mappers)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)

	// action.
	err = s.sut.Alter(entities...)

	// assert.
	s.Error(err)
}

func (s *UnitTestSuite) TestUnit_Alter() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}

	// action.
	err := s.sut.Alter(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_ConcurrentAlter() {

	// arrange.
	foo := Foo{ID: 28}
	bar := Bar{ID: "28"}

	// action.
	var err, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err = s.sut.Alter(foo)
		wg.Done()
	}()
	go func() {
		err2 = s.sut.Alter(bar)
		wg.Done()
	}()
	wg.Wait()

	// assert.
	s.NoError(err)
	s.NoError(err2)
}

func (s *UnitTestSuite) TestUnit_Remove_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Remove(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Remove_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Bar{ID: "28"},
	}
	mappers := map[work.TypeName]work.DataMapper{
		work.TypeNameOf(Foo{}): &mock.DataMapper{},
	}
	var err error
	opts := []work.UnitOption{work.UnitDataMappers(mappers)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)

	// action.
	err = s.sut.Remove(entities...)

	// assert.
	s.Error(err)
}

func (s *UnitTestSuite) TestUnit_Remove() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}

	// action.
	err := s.sut.Remove(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_ConcurrentRemove() {

	// arrange.
	foo := Foo{ID: 28}
	bar := Bar{ID: "28"}

	// action.
	var err, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err = s.sut.Remove(foo)
		wg.Done()
	}()
	go func() {
		err2 = s.sut.Remove(bar)
		wg.Done()
	}()
	wg.Wait()

	// assert.
	s.NoError(err)
	s.NoError(err2)
}

func (s *UnitTestSuite) TestUnit_Register_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Register(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Register_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Bar{ID: "28"},
	}
	mappers := map[work.TypeName]work.DataMapper{
		work.TypeNameOf(Foo{}): &mock.DataMapper{},
	}
	var err error
	opts := []work.UnitOption{work.UnitDataMappers(mappers)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)

	// action.
	err = s.sut.Register(entities...)

	// assert.
	s.Require().Error(err)
	s.EqualError(err, work.ErrMissingDataMapper.Error())
}

func (s *UnitTestSuite) TestUnit_Register() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}

	// action.
	err := s.sut.Register(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_ConcurrentRegister() {

	// arrange.
	foo := Foo{ID: 28}
	bar := Bar{ID: "28"}

	// action.
	var err, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err = s.sut.Register(foo)
		wg.Done()
	}()
	go func() {
		err2 = s.sut.Register(bar)
		wg.Done()
	}()
	wg.Wait()

	// assert.
	s.NoError(err)
	s.NoError(err2)
}

func (s *UnitTestSuite) TearDownTest() {
	s.sut = nil
	s.mc.Finish()
}
