/* Copyright 2020 Freerware
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
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/freerware/work/v3"
	"github.com/freerware/work/v3/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

type BestEffortUnitTestSuite struct {
	suite.Suite

	// system under test.
	sut work.Unit

	// mocks.
	mappers map[work.TypeName]*mock.DataMapper
	scope   tally.TestScope
	mc      *gomock.Controller

	// metrics scope names and tags.
	scopePrefix                      string
	saveScopeName                    string
	saveSuccessScopeName             string
	saveScopeNameWithTags            string
	saveSuccessScopeNameWithTags     string
	rollbackScopeNameWithTags        string
	rollbackSuccessScopeNameWithTags string
	rollbackFailureScopeNameWithTags string
	rollbackScopeName                string
	rollbackFailureScopeName         string
	rollbackSuccessScopeName         string
	tags                             string
}

func TestBestEffortUnitTestSuite(t *testing.T) {
	suite.Run(t, new(BestEffortUnitTestSuite))
}

func (s *BestEffortUnitTestSuite) SetupTest() {

	// initialize metric names.
	sep := "+"
	s.scopePrefix = "test"
	s.tags = "unit_type=best_effort"
	s.saveScopeName = fmt.Sprintf("%s.%s", s.scopePrefix, "unit.save")
	s.saveScopeNameWithTags = fmt.Sprintf("%s%s%s", s.saveScopeName, sep, s.tags)
	s.rollbackScopeName = fmt.Sprintf("%s.%s", s.scopePrefix, "unit.rollback")
	s.rollbackScopeNameWithTags = fmt.Sprintf("%s%s%s", s.rollbackScopeName, sep, s.tags)
	s.saveSuccessScopeName = fmt.Sprintf("%s.success", s.saveScopeName)
	s.rollbackSuccessScopeName = fmt.Sprintf("%s.success", s.rollbackScopeName)
	s.rollbackFailureScopeName = fmt.Sprintf("%s.failure", s.rollbackScopeName)
	s.saveSuccessScopeNameWithTags = fmt.Sprintf("%s%s%s", s.saveSuccessScopeName, sep, s.tags)
	s.rollbackSuccessScopeNameWithTags = fmt.Sprintf("%s%s%s", s.rollbackSuccessScopeName, sep, s.tags)
	s.rollbackFailureScopeNameWithTags = fmt.Sprintf("%s%s%s", s.rollbackFailureScopeName, sep, s.tags)

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

	c := zap.NewDevelopmentConfig()
	c.DisableStacktrace = true
	l, _ := c.Build()
	ts := tally.NewTestScope(s.scopePrefix, map[string]string{})
	s.scope = ts
	var err error
	s.sut, err = work.NewBestEffortUnit(dm, work.UnitLogger(l), work.UnitScope(ts))
	s.Require().NoError(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_NewBestEffortUnit_MissingDataMappers() {

	// action.
	var err error
	s.sut, err = work.NewBestEffortUnit(map[work.TypeName]work.DataMapper{})

	// assert.
	s.Error(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_InsertError() {

	// arrange.
	fooType := work.TypeNameOf(Foo{})
	barType := work.TypeNameOf(Bar{})
	addedEntities := []interface{}{
		Foo{ID: 28},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	registeredEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
		Foo{ID: 1111},
	}
	s.sut.Register(registeredEntities...)
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	err := errors.New("whoa")
	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.mappers[fooType].EXPECT().
		Update(registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].EXPECT().
		Update(registeredEntities[1]).Return(nil)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_InsertAndRollbackError() {

	// arrange.
	fooType := work.TypeNameOf(Foo{})
	addedEntities := []interface{}{
		Foo{ID: 28},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	registeredEntities := []interface{}{
		Foo{ID: 1992},
		Foo{ID: 1111},
	}
	s.sut.Register(registeredEntities...)
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	err := errors.New("whoa")

	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.mappers[fooType].EXPECT().
		Update(registeredEntities[0], registeredEntities[1]).Return(err)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_UpdateError() {

	// arrange.
	fooType := work.TypeNameOf(Foo{})
	barType := work.TypeNameOf(Bar{})
	addedEntities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	registeredEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
		Foo{ID: 1111},
	}
	s.sut.Register(registeredEntities...)
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	err := errors.New("whoa")
	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(updatedEntities[0]).Return(err).Times(1)

	// arrange - rollback invocations.
	s.mappers[fooType].EXPECT().Delete(addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Delete(addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().
		Update(registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].EXPECT().
		Update(registeredEntities[1]).Return(nil)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_UpdateAndRollbackError() {

	// arrange.
	fooType := work.TypeNameOf(Foo{})
	barType := work.TypeNameOf(Bar{})
	addedEntities := []interface{}{
		Foo{ID: 28},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	registeredEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
		Foo{ID: 1111},
	}
	s.sut.Register(registeredEntities...)
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	err := errors.New("whoa")
	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(nil)
	s.mappers[fooType].EXPECT().Update(updatedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.mappers[fooType].EXPECT().Delete(addedEntities[0]).Return(err)
	s.mappers[fooType].EXPECT().
		Update(registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].EXPECT().
		Update(registeredEntities[1]).Return(nil)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Error(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_DeleteError() {

	// arrange.
	fooType := work.TypeNameOf(Foo{})
	barType := work.TypeNameOf(Bar{})
	addedEntities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	registeredEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
		Foo{ID: 1111},
	}
	s.sut.Register(registeredEntities...)
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	err := errors.New("whoa")
	s.mappers[barType].EXPECT().Insert(addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(nil)
	s.mappers[fooType].EXPECT().Update(updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(removedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.mappers[fooType].EXPECT().Delete(addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Delete(addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().
		Update(registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].EXPECT().
		Update(registeredEntities[1]).Return(nil)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_DeleteAndRollbackError() {

	// arrange.
	fooType := work.TypeNameOf(Foo{})
	barType := work.TypeNameOf(Bar{})
	addedEntities := []interface{}{
		Foo{ID: 28},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	registeredEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
		Foo{ID: 1111},
	}
	s.sut.Register(registeredEntities...)
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	err := errors.New("whoa")
	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(nil)
	s.mappers[fooType].EXPECT().Update(updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(removedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.mappers[fooType].EXPECT().Delete(addedEntities[0]).Return(err)
	s.mappers[fooType].EXPECT().
		Update(registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].EXPECT().
		Update(registeredEntities[1]).Return(nil)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_Panic() {

	// arrange.
	fooType := work.TypeNameOf(Foo{})
	barType := work.TypeNameOf(Bar{})
	addedEntities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	registeredEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
		Foo{ID: 1111},
	}
	s.sut.Register(registeredEntities...)
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(updatedEntities[1]).Return(nil)
	s.mappers[fooType].
		EXPECT().Delete(removedEntities[0]).Do(func() { panic("whoa") })

	// arrange - rollback invocations.
	s.mappers[fooType].EXPECT().Delete(addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Delete(addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().
		Update(registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].EXPECT().
		Update(registeredEntities[1]).Return(nil)

	// action + assert.
	s.Require().Panics(func() { s.sut.Save() })
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_PanicAndRollbackError() {

	// arrange.
	fooType := work.TypeNameOf(Foo{})
	barType := work.TypeNameOf(Bar{})
	addedEntities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	registeredEntities := []interface{}{
		Foo{ID: 1992},
		Foo{ID: 1111},
	}
	s.sut.Register(registeredEntities...)
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	err := errors.New("whoa")
	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(updatedEntities[1]).Return(nil)
	s.mappers[fooType].
		EXPECT().Delete(removedEntities[0]).Do(func() { panic("whoa") })

	// arrange - rollback invocations.
	s.mappers[fooType].EXPECT().
		Update(registeredEntities[0], registeredEntities[1]).Return(err)

	// action + assert.
	s.Require().Panics(func() { s.sut.Save() })
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_PanicAndRollbackPanic() {

	// arrange.
	fooType := work.TypeNameOf(Foo{})
	barType := work.TypeNameOf(Bar{})
	addedEntities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	registeredEntities := []interface{}{
		Foo{ID: 1992},
		Foo{ID: 1111},
	}
	s.sut.Register(registeredEntities...)
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(updatedEntities[1]).Return(nil)
	s.mappers[fooType].
		EXPECT().Delete(removedEntities[0]).Do(func() { panic("whoa") })

	// arrange - rollback invocations.
	s.mappers[fooType].
		EXPECT().Update(registeredEntities[0], registeredEntities[1]).
		Do(func() { panic("whoa") })

	// action + assert.
	s.Require().Panics(func() { s.sut.Save() })
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save() {

	// arrange.
	fooType := work.TypeNameOf(Foo{})
	barType := work.TypeNameOf(Bar{})
	addedEntities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(removedEntities[0]).Return(nil)

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.NoError(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.saveSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 1)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_NoOptions() {

	// arrange.
	dm := make(map[work.TypeName]work.DataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}
	var err error
	s.sut, err = work.NewBestEffortUnit(dm)
	s.Require().NoError(err)

	fooType := work.TypeNameOf(Foo{})
	barType := work.TypeNameOf(Bar{})
	addedEntities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	updatedEntities := []interface{}{
		Foo{ID: 1992},
		Bar{ID: "1992"},
	}
	removedEntities := []interface{}{
		Foo{ID: 2},
	}
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	s.mappers[fooType].EXPECT().Insert(addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(removedEntities[0]).Return(nil)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.NoError(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Add_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Add(entities...)

	// assert.
	s.NoError(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Add_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
	}
	mappers := map[work.TypeName]work.DataMapper{
		work.TypeNameOf(Bar{}): &mock.DataMapper{},
	}
	var err error
	s.sut, err = work.NewBestEffortUnit(mappers)
	s.Require().NoError(err)

	// action.
	err = s.sut.Add(entities...)

	// assert.
	s.Error(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Add() {

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_ConcurrentAdd() {

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Alter_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Alter(entities...)

	// assert.
	s.NoError(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Alter_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
	}
	mappers := map[work.TypeName]work.DataMapper{
		work.TypeNameOf(Bar{}): &mock.DataMapper{},
	}
	var err error
	s.sut, err = work.NewBestEffortUnit(mappers)
	s.Require().NoError(err)

	// action.
	err = s.sut.Alter(entities...)

	// assert.
	s.Error(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Alter() {

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_ConcurrentAlter() {

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Remove_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Remove(entities...)

	// assert.
	s.NoError(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Remove_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Bar{ID: "28"},
	}
	mappers := map[work.TypeName]work.DataMapper{
		work.TypeNameOf(Foo{}): &mock.DataMapper{},
	}
	var err error
	s.sut, err = work.NewBestEffortUnit(mappers)
	s.Require().NoError(err)

	// action.
	err = s.sut.Remove(entities...)

	// assert.
	s.Error(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Remove() {

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_ConcurrentRemove() {

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Register_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Register(entities...)

	// assert.
	s.NoError(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Register_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Bar{ID: "28"},
	}
	mappers := map[work.TypeName]work.DataMapper{
		work.TypeNameOf(Foo{}): &mock.DataMapper{},
	}
	var err error
	s.sut, err = work.NewBestEffortUnit(mappers)
	s.Require().NoError(err)

	// action.
	err = s.sut.Register(entities...)

	// assert.
	s.Require().Error(err)
	s.EqualError(err, work.ErrMissingDataMapper.Error())
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Register() {

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_ConcurrentRegister() {

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

func (s *BestEffortUnitTestSuite) TearDownTest() {
	s.sut = nil
	s.mc.Finish()
}
