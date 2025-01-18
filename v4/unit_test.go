/* Copyright 2025 Freerware
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
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/freerware/work/v4"
	"github.com/freerware/work/v4/internal/mock"
	"github.com/freerware/work/v4/internal/test"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
)

type UnitTestSuite struct {
	suite.Suite

	// system under test.
	sut work.Unit

	// mocks.
	scope   tally.TestScope
	mappers map[work.TypeName]*mock.UnitDataMapper
	mc      *gomock.Controller

	// metrics scope names and tags.
	scopePrefix string
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) SetupTest() {
	// test entities.
	foo := test.Foo{ID: 28}
	fooTypeName := work.TypeNameOf(foo)
	bar := test.Bar{ID: "28"}
	barTypeName := work.TypeNameOf(bar)
	baz := test.Baz{Identifier: "28"}
	bazTypeName := work.TypeNameOf(baz)
	biz := test.Biz{Identifier: "28"}
	bizTypeName := work.TypeNameOf(biz)

	// initialize mocks.
	s.mc = gomock.NewController(s.T())
	s.mappers = make(map[work.TypeName]*mock.UnitDataMapper)
	s.mappers[fooTypeName] = mock.NewUnitDataMapper(s.mc)
	s.mappers[barTypeName] = mock.NewUnitDataMapper(s.mc)
	s.mappers[bizTypeName] = mock.NewUnitDataMapper(s.mc)
	s.mappers[bazTypeName] = mock.NewUnitDataMapper(s.mc)

	// construct SUT.
	dm := make(map[work.TypeName]work.UnitDataMapper)
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
	opts := []work.UnitOption{work.UnitDataMappers(dm), work.UnitZapLogger(l), work.UnitTallyMetricScope(ts)}
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
	dm := map[work.TypeName]work.UnitDataMapper{}
	opts := []work.UnitOption{work.UnitDataMappers(dm)}
	s.sut, err = work.NewUnit(opts...)

	// assert.
	s.EqualError(err, work.ErrNoDataMapper.Error())
}

func (s *UnitTestSuite) TestUnit_NewUnit_WithDataMappers() {

	// action.
	var err error
	mappers := map[work.TypeName]work.UnitDataMapper{
		work.TypeNameOf(test.Bar{}): &mock.UnitDataMapper{},
	}
	opts := []work.UnitOption{work.UnitDataMappers(mappers)}
	s.sut, err = work.NewUnit(opts...)

	// assert.
	s.NoError(err)
	s.NotNil(s.sut)
}

func (s *UnitTestSuite) TestUnit_NewUnit_NoDataMapperFunctions() {

	// action.
	var err error
	s.sut, err = work.NewUnit()

	// assert.
	s.EqualError(err, work.ErrNoDataMapper.Error())
}

func (s *UnitTestSuite) TestUnit_NewUnit_WithSomeDataMapperFuncs() {

	// action.
	var err error
	mapper := &mock.UnitDataMapper{}
	t := work.TypeNameOf(test.Bar{})
	opts := []work.UnitOption{
		work.UnitInsertFunc(t, mapper.Insert),
		work.UnitUpdateFunc(t, mapper.Update),
	}
	s.sut, err = work.NewUnit(opts...)

	// assert.
	s.NoError(err)
	s.NotNil(s.sut)
}

func (s *UnitTestSuite) TestUnit_NewUnit_WithAllDataMapperFuncs() {

	// action.
	var err error
	mapper := &mock.UnitDataMapper{}
	t := work.TypeNameOf(test.Bar{})
	opts := []work.UnitOption{
		work.UnitInsertFunc(t, mapper.Insert),
		work.UnitUpdateFunc(t, mapper.Update),
		work.UnitDeleteFunc(t, mapper.Delete),
	}
	s.sut, err = work.NewUnit(opts...)

	// assert.
	s.NoError(err)
	s.NotNil(s.sut)
}

func (s *UnitTestSuite) TestUnit_Add_Empty() {

	// arrange.
	entities := []interface{}{}
	ctx := context.Background()

	// action.
	err := s.sut.Add(ctx, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Add_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		test.Foo{ID: 28},
	}
	mappers := map[work.TypeName]work.UnitDataMapper{
		work.TypeNameOf(test.Bar{}): &mock.UnitDataMapper{},
	}
	ctx := context.Background()
	var err error
	opts := []work.UnitOption{work.UnitDataMappers(mappers)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)

	// action.
	err = s.sut.Add(ctx, entities...)

	// assert.
	s.Error(err)
}

func (s *UnitTestSuite) TestUnit_Add() {

	// arrange.
	entities := []interface{}{
		test.Foo{ID: 28},
		test.Bar{ID: "28"},
	}
	ctx := context.Background()

	// action.
	err := s.sut.Add(ctx, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_ConcurrentAdd() {

	// arrange.
	foo := test.Foo{ID: 28}
	bar := test.Bar{ID: "28"}
	ctx := context.Background()

	// action.
	var err, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err = s.sut.Add(ctx, foo)
		wg.Done()
	}()
	go func() {
		err2 = s.sut.Add(ctx, bar)
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
	ctx := context.Background()

	// action.
	err := s.sut.Alter(ctx, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Alter_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		test.Foo{ID: 28},
	}
	mappers := map[work.TypeName]work.UnitDataMapper{
		work.TypeNameOf(test.Bar{}): &mock.UnitDataMapper{},
	}
	ctx := context.Background()
	var err error
	opts := []work.UnitOption{work.UnitDataMappers(mappers)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)

	// action.
	err = s.sut.Alter(ctx, entities...)

	// assert.
	s.Error(err)
}

func (s *UnitTestSuite) TestUnit_Alter() {

	// arrange.
	entities := []interface{}{
		test.Foo{ID: 28},
		test.Bar{ID: "28"},
	}
	ctx := context.Background()

	// action.
	err := s.sut.Alter(ctx, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_ConcurrentAlter() {

	// arrange.
	foo := test.Foo{ID: 28}
	bar := test.Bar{ID: "28"}
	ctx := context.Background()

	// action.
	var err, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err = s.sut.Alter(ctx, foo)
		wg.Done()
	}()
	go func() {
		err2 = s.sut.Alter(ctx, bar)
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
	ctx := context.Background()

	// action.
	err := s.sut.Remove(ctx, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Remove_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		test.Bar{ID: "28"},
	}
	mappers := map[work.TypeName]work.UnitDataMapper{
		work.TypeNameOf(test.Foo{}): &mock.UnitDataMapper{},
	}
	ctx := context.Background()
	var err error
	opts := []work.UnitOption{work.UnitDataMappers(mappers)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)

	// action.
	err = s.sut.Remove(ctx, entities...)

	// assert.
	s.Error(err)
}

func (s *UnitTestSuite) TestUnit_Remove() {

	// arrange.
	ctx := context.Background()
	entities := []interface{}{
		test.Foo{ID: 28},
		test.Bar{ID: "28"},
	}

	// action.
	err := s.sut.Remove(ctx, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_ConcurrentRemove() {

	// arrange.
	foo := test.Foo{ID: 28}
	bar := test.Bar{ID: "28"}
	ctx := context.Background()

	// action.
	var err, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err = s.sut.Remove(ctx, foo)
		wg.Done()
	}()
	go func() {
		err2 = s.sut.Remove(ctx, bar)
		wg.Done()
	}()
	wg.Wait()

	// assert.
	s.NoError(err)
	s.NoError(err2)
}

func (s *UnitTestSuite) TestUnit_Register_Empty() {

	// arrange.
	ctx := context.Background()
	entities := []interface{}{}

	// action.
	err := s.sut.Register(ctx, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Register_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		test.Bar{ID: "28"},
	}
	mappers := map[work.TypeName]work.UnitDataMapper{
		work.TypeNameOf(test.Foo{}): &mock.UnitDataMapper{},
	}
	ctx := context.Background()
	var err error
	opts := []work.UnitOption{work.UnitDataMappers(mappers)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)

	// action.
	err = s.sut.Register(ctx, entities...)

	// assert.
	s.Require().Error(err)
	s.EqualError(err, work.ErrMissingDataMapper.Error())
}

func (s *UnitTestSuite) TestUnit_Register() {

	// arrange.
	entities := []interface{}{
		test.Foo{ID: 28},
		test.Biz{Identifier: "28"},
	}
	ctx := context.Background()

	// action.
	err := s.sut.Register(ctx, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_ConcurrentRegister() {

	// arrange.
	foo := test.Foo{ID: 28}
	bar := test.Bar{ID: "28"}
	ctx := context.Background()

	// action.
	var err, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err = s.sut.Register(ctx, foo)
		wg.Done()
	}()
	go func() {
		err2 = s.sut.Register(ctx, bar)
		wg.Done()
	}()
	wg.Wait()

	// assert.
	s.NoError(err)
	s.NoError(err2)
}

func (s *UnitTestSuite) TestUnit_Cache() {
	// arrange.
	foo := test.Foo{ID: 28}
	baz := test.Baz{Identifier: "28"}
	ctx := context.Background()
	tFoo := work.TypeNameOf(foo)
	tBaz := work.TypeNameOf(baz)
	s.sut.Register(ctx, foo, baz)

	// action.
	cached := s.sut.Cached()

	// assert.
	cachedFoo, err := cached.Load(ctx, tFoo, foo.ID)
	s.Require().NoError(err)
	s.Equal(foo, cachedFoo)
	cachedBaz, err := cached.Load(ctx, tBaz, baz.Identifier)
	s.Require().NoError(err)
	s.Equal(baz, cachedBaz)
}

func (s *UnitTestSuite) TestUnit_Remove_InvalidatesCache() {
	// arrange.
	foo := test.Foo{ID: 28}
	baz := test.Baz{Identifier: "28"}
	ctx := context.Background()
	tFoo := work.TypeNameOf(foo)
	tBaz := work.TypeNameOf(baz)
	s.sut.Register(ctx, foo, baz)

	// action.
	err := s.sut.Remove(ctx, foo)

	// assert.
	s.NoError(err)
	cached := s.sut.Cached()
	cachedFoo, err := cached.Load(ctx, tFoo, foo.ID)
	s.Require().NoError(err)
	s.Nil(cachedFoo)
	cachedBaz, err := cached.Load(ctx, tBaz, baz.Identifier)
	s.Require().NoError(err)
	s.Equal(baz, cachedBaz)
}

func (s *UnitTestSuite) TestUnit_Alter_InvalidatesCache() {
	// arrange.
	foo := test.Foo{ID: 28}
	baz := test.Baz{Identifier: "28"}
	ctx := context.Background()
	tFoo := work.TypeNameOf(foo)
	tBaz := work.TypeNameOf(baz)
	s.sut.Register(ctx, foo, baz)

	// action.
	err := s.sut.Alter(ctx, foo)

	// assert.
	s.NoError(err)
	cached := s.sut.Cached()
	cachedFoo, err := cached.Load(ctx, tFoo, foo.ID)
	s.Require().NoError(err)
	s.Nil(cachedFoo)
	cachedBaz, err := cached.Load(ctx, tBaz, baz.Identifier)
	s.Require().NoError(err)
	s.Equal(baz, cachedBaz)
}

func (s *UnitTestSuite) TestUnit_Alter_CacheInvalidationError() {
	// arrange.
	ctx := context.Background()
	foo := test.Foo{ID: 28}
	baz := test.Baz{Identifier: "28"}
	tFoo := work.TypeNameOf(foo)
	tBaz := work.TypeNameOf(baz)

	// initialize mocks.
	s.mc = gomock.NewController(s.T())
	cacheClient := mock.NewUnitCacheClient(s.mc)
	cacheInvalidationError := errors.New("cache invalidation failed!")
	cacheClient.
		EXPECT().
		Set(ctx, fmt.Sprintf("%s-%v", string(tFoo), foo.ID), foo).
		Return(nil)
	cacheClient.
		EXPECT().
		Set(ctx, fmt.Sprintf("%s-%v", string(tBaz), baz.Identifier), baz).
		Return(nil)
	cacheClient.
		EXPECT().
		Delete(ctx, fmt.Sprintf("%s-%v", string(tFoo), foo.ID)).
		Return(cacheInvalidationError)

	s.mappers = make(map[work.TypeName]*mock.UnitDataMapper)
	s.mappers[tFoo] = mock.NewUnitDataMapper(s.mc)
	s.mappers[tBaz] = mock.NewUnitDataMapper(s.mc)

	// construct SUT.
	dm := make(map[work.TypeName]work.UnitDataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}

	var err error
	opts := []work.UnitOption{work.UnitDataMappers(dm), work.UnitWithCacheClient(cacheClient)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)

	s.sut.Register(ctx, foo, baz)

	// action.
	err = s.sut.Alter(ctx, foo)

	// assert.
	s.EqualError(err, cacheInvalidationError.Error())
}

func (s *UnitTestSuite) TestUnit_Remove_CacheInvalidationError() {
	// arrange.
	ctx := context.Background()
	foo := test.Foo{ID: 28}
	baz := test.Baz{Identifier: "28"}
	tFoo := work.TypeNameOf(foo)
	tBaz := work.TypeNameOf(baz)

	// initialize mocks.
	s.mc = gomock.NewController(s.T())
	cacheClient := mock.NewUnitCacheClient(s.mc)
	cacheInvalidationError := errors.New("cache invalidation failed!")
	cacheClient.
		EXPECT().
		Set(ctx, fmt.Sprintf("%s-%v", string(tFoo), foo.ID), foo).
		Return(nil)
	cacheClient.
		EXPECT().
		Set(ctx, fmt.Sprintf("%s-%v", string(tBaz), baz.Identifier), baz).
		Return(nil)
	cacheClient.
		EXPECT().
		Delete(ctx, fmt.Sprintf("%s-%v", string(tFoo), foo.ID)).
		Return(cacheInvalidationError)

	s.mappers = make(map[work.TypeName]*mock.UnitDataMapper)
	s.mappers[tFoo] = mock.NewUnitDataMapper(s.mc)
	s.mappers[tBaz] = mock.NewUnitDataMapper(s.mc)

	// construct SUT.
	dm := make(map[work.TypeName]work.UnitDataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}

	var err error
	opts := []work.UnitOption{work.UnitDataMappers(dm), work.UnitWithCacheClient(cacheClient)}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)

	s.sut.Register(ctx, foo, baz)

	// action.
	err = s.sut.Remove(ctx, foo)

	// assert.
	s.EqualError(err, cacheInvalidationError.Error())
}

func (s *UnitTestSuite) TearDownTest() {
	s.sut = nil
}
