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
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freerware/work"
	"github.com/freerware/work/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

type SQLUnitTestSuite struct {
	suite.Suite

	// system under test.
	sut work.Unit

	// mocks.
	db      *sql.DB
	_db     sqlmock.Sqlmock
	scope   tally.TestScope
	mc      *gomock.Controller
	mappers map[work.TypeName]*mock.SQLDataMapper

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

func TestSQLUnitTestSuite(t *testing.T) {
	suite.Run(t, new(SQLUnitTestSuite))
}

func (s *SQLUnitTestSuite) SetupTest() {

	// initialize metric names.
	sep := "+"
	s.scopePrefix = "test"
	s.tags = "unit_type=sql"
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
	s.mappers = make(map[work.TypeName]*mock.SQLDataMapper)
	s.mappers[fooTypeName] = mock.NewSQLDataMapper(s.mc)
	s.mappers[barTypeName] = mock.NewSQLDataMapper(s.mc)

	var err error
	s.db, s._db, err = sqlmock.New()
	s.Require().NoError(err)

	// construct SUT.
	dm := make(map[work.TypeName]work.SQLDataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}

	c := zap.NewDevelopmentConfig()
	c.DisableStacktrace = true
	l, _ := c.Build()
	ts := tally.NewTestScope(s.scopePrefix, map[string]string{})
	s.scope = ts
	s.sut, err = work.NewSQLUnit(dm, s.db, work.UnitLogger(l), work.UnitScope(ts))
	s.Require().NoError(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_NewSQLUnit_MissingDataMappers() {

	// action.
	var err error
	s.sut, err = work.NewSQLUnit(map[work.TypeName]work.SQLDataMapper{}, s.db)

	// assert.
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_TransactionBeginError() {

	// arrange.
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
	alterError := s.sut.Remove(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	s._db.ExpectBegin().WillReturnError(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 1)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_InsertError() {

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
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	s._db.ExpectBegin()
	s._db.ExpectRollback()
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_InsertAndRollbackError() {

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
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	s._db.ExpectBegin()
	s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_UpdateError() {

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
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	s._db.ExpectBegin()
	s._db.ExpectRollback()
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(gomock.Any(), addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(gomock.Any(), updatedEntities[0]).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_UpdateAndRollbackError() {

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
	addError := s.sut.Add(addedEntities...)
	alterError := s.sut.Alter(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	s._db.ExpectBegin()
	s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(gomock.Any(), addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(gomock.Any(), updatedEntities[0]).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_DeleteError() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback()
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(gomock.Any(), addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(gomock.Any(), updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(gomock.Any(), updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(gomock.Any(), removedEntities[0]).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_DeleteAndRollbackError() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(gomock.Any(), addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(gomock.Any(), updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(gomock.Any(), updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(gomock.Any(), removedEntities[0]).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_Panic() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback()
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(gomock.Any(), addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(gomock.Any(), updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(gomock.Any(), updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(gomock.Any(), removedEntities[0]).
		Do(func() { panic("whoa") })

	// action + assert.
	s.Require().Panics(func() { s.sut.Save() })
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_PanicAndRollbackError() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(gomock.Any(), addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(gomock.Any(), updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(gomock.Any(), updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(gomock.Any(), removedEntities[0]).
		Do(func() { panic("whoa") })

	// action + assert.
	s.Require().Panics(func() { s.sut.Save() })
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 2)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
	s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_CommitError() {

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
	s._db.ExpectBegin()
	s._db.ExpectCommit().WillReturnError(errors.New("whoa"))
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(gomock.Any(), addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(gomock.Any(), updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(gomock.Any(), updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(gomock.Any(), removedEntities[0]).Return(nil)

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Require().Error(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 1)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save() {

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
	s._db.ExpectBegin()
	s._db.ExpectCommit()
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(gomock.Any(), addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(gomock.Any(), updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(gomock.Any(), updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(gomock.Any(), removedEntities[0]).Return(nil)

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.NoError(err)
	s.Len(s.scope.Snapshot().Counters(), 1)
	s.Contains(s.scope.Snapshot().Counters(), s.saveSuccessScopeNameWithTags)
	s.Len(s.scope.Snapshot().Timers(), 1)
	s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_NoOptions() {

	// arrange.
	dm := make(map[work.TypeName]work.SQLDataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}
	var err error
	s.sut, err = work.NewSQLUnit(dm, s.db)
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
	s._db.ExpectBegin()
	s._db.ExpectCommit()
	s.mappers[fooType].EXPECT().Insert(gomock.Any(), addedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Insert(gomock.Any(), addedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Update(gomock.Any(), updatedEntities[0]).Return(nil)
	s.mappers[barType].EXPECT().Update(gomock.Any(), updatedEntities[1]).Return(nil)
	s.mappers[fooType].EXPECT().Delete(gomock.Any(), removedEntities[0]).Return(nil)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.NoError(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Add_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Add(entities...)

	// assert.
	s.NoError(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Add_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
	}
	mappers := map[work.TypeName]work.SQLDataMapper{
		work.TypeNameOf(Bar{}): &mock.SQLDataMapper{},
	}
	var err error
	s.sut, err = work.NewSQLUnit(mappers, s.db)
	s.Require().NoError(err)

	// action.
	err = s.sut.Add(entities...)

	// assert.
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Add() {

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

func (s *SQLUnitTestSuite) TestSQLUnit_Alter_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Alter(entities...)

	// assert.
	s.NoError(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Alter_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
	}
	mappers := map[work.TypeName]work.SQLDataMapper{
		work.TypeNameOf(Bar{}): &mock.SQLDataMapper{},
	}
	var err error
	s.sut, err = work.NewSQLUnit(mappers, s.db)
	s.Require().NoError(err)

	// action.
	err = s.sut.Alter(entities...)

	// assert.
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Alter() {

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

func (s *SQLUnitTestSuite) TestSQLUnit_Remove_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Remove(entities...)

	// assert.
	s.NoError(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Remove_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Bar{ID: "28"},
	}
	mappers := map[work.TypeName]work.SQLDataMapper{
		work.TypeNameOf(Foo{}): &mock.SQLDataMapper{},
	}
	var err error
	s.sut, err = work.NewSQLUnit(mappers, s.db)
	s.Require().NoError(err)

	// action.
	err = s.sut.Remove(entities...)

	// assert.
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Remove() {

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

func (s *SQLUnitTestSuite) TestSQLUnit_Register_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Register(entities...)

	// assert.
	s.NoError(err)
}

func (s *SQLUnitTestSuite) TestUnit_Register_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Bar{ID: "28"},
	}
	mappers := map[work.TypeName]work.SQLDataMapper{
		work.TypeNameOf(Foo{}): &mock.SQLDataMapper{},
	}
	var err error
	s.sut, err = work.NewSQLUnit(mappers, s.db)
	s.Require().NoError(err)

	// action.
	err = s.sut.Register(entities...)

	// assert.
	s.Require().Error(err)
	s.EqualError(err, work.ErrMissingDataMapper.Error())
}

func (s *SQLUnitTestSuite) TestSQLUnit_Register() {

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

func (s *SQLUnitTestSuite) TearDownTest() {
	s.db.Close()
	s.mc.Finish()
	s.sut = nil
	s.scope = nil
}
