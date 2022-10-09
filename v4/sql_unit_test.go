/* Copyright 2022 Freerware
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.

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
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freerware/work/v4"
	"github.com/freerware/work/v4/internal/mock"
	"github.com/freerware/work/v4/internal/test"
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
	mappers map[work.TypeName]*mock.UnitDataMapper

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
	retryAttemptScopeName            string
	retryAttemptScopeNameWithTags    string
	insertScopeName                  string
	insertScopeNameWithTags          string
	updateScopeName                  string
	updateScopeNameWithTags          string
	deleteScopeName                  string
	deleteScopeNameWithTags          string
	tags                             string

	// suite state.
	isSetup    bool
	isTornDown bool

	retryCount int
}

func TestSQLUnitTestSuite(t *testing.T) {
	suite.Run(t, new(SQLUnitTestSuite))
}

func (s *SQLUnitTestSuite) Setup() {
	defer func() { s.isSetup, s.isTornDown = true, false }()

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
	s.retryAttemptScopeName = fmt.Sprintf("%s.%s", s.scopePrefix, "unit.retry.attempt")
	s.retryAttemptScopeNameWithTags = fmt.Sprintf("%s%s%s", s.retryAttemptScopeName, sep, s.tags)
	s.insertScopeName = fmt.Sprintf("%s.%s", s.scopePrefix, "unit.insert")
	s.insertScopeNameWithTags = fmt.Sprintf("%s%s%s", s.insertScopeName, sep, s.tags)
	s.updateScopeName = fmt.Sprintf("%s.%s", s.scopePrefix, "unit.update")
	s.updateScopeNameWithTags = fmt.Sprintf("%s%s%s", s.updateScopeName, sep, s.tags)
	s.deleteScopeName = fmt.Sprintf("%s.%s", s.scopePrefix, "unit.delete")
	s.deleteScopeNameWithTags = fmt.Sprintf("%s%s%s", s.deleteScopeName, sep, s.tags)

	// test entities.
	foo := test.Foo{ID: 28}
	fooTypeName := work.TypeNameOf(foo)
	bar := test.Bar{ID: "28"}
	barTypeName := work.TypeNameOf(bar)

	// initialize mocks.
	s.mc = gomock.NewController(s.T())
	s.mappers = make(map[work.TypeName]*mock.UnitDataMapper)
	s.mappers[fooTypeName] = mock.NewUnitDataMapper(s.mc)
	s.mappers[barTypeName] = mock.NewUnitDataMapper(s.mc)

	var err error
	s.db, s._db, err = sqlmock.New()
	s.Require().NoError(err)

	// construct SUT.
	dm := make(map[work.TypeName]work.UnitDataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}

	c := zap.NewDevelopmentConfig()
	c.DisableStacktrace = true
	l, _ := c.Build()
	ts := tally.NewTestScope(s.scopePrefix, map[string]string{})
	s.retryCount = 2
	s.scope = ts
	opts := []work.UnitOption{
		work.UnitDataMappers(dm),
		work.UnitLogger(l),
		work.UnitScope(ts),
		work.UnitDB(s.db),
		work.UnitRetryAttempts(s.retryCount),
	}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)
}

func (s *SQLUnitTestSuite) SetupTest() {
	if !s.isSetup {
		s.Setup()
	}
}

func (s *SQLUnitTestSuite) subtests() []TableDrivenTest {
	foos := []interface{}{test.Foo{ID: 28}, test.Foo{ID: 1992}, test.Foo{ID: 2}}
	bars := []interface{}{test.Bar{ID: "ID"}, test.Bar{ID: "1992"}}
	fooType, barType := work.TypeNameOf(test.Foo{}), work.TypeNameOf(test.Bar{})
	return []TableDrivenTest{
		{
			name:      "TransactionBeginError",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin().WillReturnError(errors.New("whoa"))
				}
			},
			ctx:        context.Background(),
			err:        errors.New("whoa"),
			assertions: func() {},
		},
		{
			name:      "TransactionBeginError_MetricsEmitted",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin().WillReturnError(errors.New("whoa"))
				}
			},
			ctx: context.Background(),
			err: errors.New("whoa"),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 2)
				s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.retryAttemptScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 1)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
			},
		},
		{
			name:      "InsertError",
			additions: []interface{}{foos[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback()
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(errors.New("whoa")).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("whoa"),
			assertions: func() {},
		},
		{
			name:      "InsertError_MetricsEmitted",
			additions: []interface{}{foos[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback()
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(errors.New("whoa")).Times(s.retryCount)
			},
			ctx: context.Background(),
			err: errors.New("whoa"),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 2)
				s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.retryAttemptScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 2)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
			},
		},
		{
			name:      "InsertAndRollbackError",
			additions: []interface{}{foos[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(errors.New("ouch")).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("ouch; whoa"),
			assertions: func() {},
		},
		{
			name:      "InsertAndRollbackError_MetricsEmitted",
			additions: []interface{}{foos[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(errors.New("ouch")).Times(s.retryCount)
			},
			ctx: context.Background(),
			err: errors.New("ouch; whoa"),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 2)
				s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.retryAttemptScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 2)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
			},
		},
		{
			name:      "UpdateError",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback()
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(errors.New("whoa")).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("whoa"),
			assertions: func() {},
		},
		{
			name:      "UpdateError_MetricsEmitted",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback()
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(errors.New("whoa")).Times(s.retryCount)
			},
			ctx: context.Background(),
			err: errors.New("whoa"),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 2)
				s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.retryAttemptScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 2)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
			},
		},
		{
			name:      "UpdateAndRollbackError",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(errors.New("ouch")).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("ouch; whoa"),
			assertions: func() {},
		},
		{
			name:      "UpdateAndRollbackError_MetricsEmitted",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(errors.New("ouch")).Times(s.retryCount)
			},
			ctx: context.Background(),
			err: errors.New("ouch; whoa"),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 2)
				s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.retryAttemptScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 2)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
			},
		},
		{
			name:      "DeleteError",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback()
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("whoa")).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("whoa"),
			assertions: func() {},
		},
		{
			name:      "DeleteError_MetricsEmitted",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback()
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("whoa")).Times(s.retryCount)
			},
			ctx: context.Background(),
			err: errors.New("whoa"),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 2)
				s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.retryAttemptScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 2)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
			},
		},
		{
			name:      "DeleteAndRollbackError",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("ouch")).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("ouch; whoa"),
			assertions: func() {},
		},
		{
			name:      "DeleteAndRollbackError_MetricsEmitted",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("ouch")).Times(s.retryCount)
			},
			ctx: context.Background(),
			err: errors.New("ouch; whoa"),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 2)
				s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.retryAttemptScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 2)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
			},
		},
		{
			name:      "Panic",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				s._db.ExpectBegin()
				s._db.ExpectRollback()
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Do(func() { panic("whoa") })
			},
			ctx:        context.Background(),
			assertions: func() {},
			panics:     true,
		},
		{
			name:      "Panic_MetricsEmitted",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				s._db.ExpectBegin()
				s._db.ExpectRollback()
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Do(func() { panic("whoa") })
			},
			ctx: context.Background(),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 1)
				s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 2)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
			},
			panics: true,
		},
		{
			name:      "PanicAndRollbackError",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				s._db.ExpectBegin()
				s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Do(func() { panic("ouch") })
			},
			ctx:        context.Background(),
			assertions: func() {},
			panics:     true,
		},
		{
			name:      "PanicAndRollbackError_MetricsEmitted",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				s._db.ExpectBegin()
				s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Do(func() { panic("ouch") })
			},
			ctx: context.Background(),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 1)
				s.Contains(s.scope.Snapshot().Counters(), s.rollbackFailureScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 2)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Timers(), s.rollbackScopeNameWithTags)
			},
			panics: true,
		},
		{
			name:      "CommitError",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectCommit().WillReturnError(errors.New("whoa"))
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(nil).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("whoa"),
			assertions: func() {},
		},
		{
			name:      "CommitError_MetricsEmitted",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				for i := 0; i < s.retryCount; i++ {
					s._db.ExpectBegin()
					s._db.ExpectCommit().WillReturnError(errors.New("whoa"))
				}
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(nil).Times(s.retryCount)
			},
			ctx: context.Background(),
			err: errors.New("whoa"),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 2)
				s.Contains(s.scope.Snapshot().Counters(), s.rollbackSuccessScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.retryAttemptScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 1)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
			},
		},
		{
			name:      "Success",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				s._db.ExpectBegin()
				s._db.ExpectCommit()
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(nil)
			},
			ctx:        context.Background(),
			assertions: func() {},
		},
		{
			name:      "Success_MetricsEmitted",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				s._db.ExpectBegin()
				s._db.ExpectCommit()
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)
				s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(nil)
			},
			ctx: context.Background(),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 4)
				s.Contains(s.scope.Snapshot().Counters(), s.saveSuccessScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.insertScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.updateScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.deleteScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 1)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
			},
		},
		{
			name:      "Success_RetrySucceeds",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				s._db.ExpectBegin()
				s._db.ExpectRollback()
				s._db.ExpectBegin()
				s._db.ExpectCommit()
				s.mappers[fooType].EXPECT().
					Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(2)
				s.mappers[barType].EXPECT().
					Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(2)
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), alters[0]).Return(nil).Times(2)
				s.mappers[barType].EXPECT().
					Update(ctx, gomock.Any(), alters[1]).Return(nil).Times(2)
				deletionFailure := s.mappers[fooType].EXPECT().
					Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("whoa"))
				s.mappers[fooType].EXPECT().
					Delete(ctx, gomock.Any(), removals[0]).Return(nil).After(deletionFailure)
			},
			ctx:        context.Background(),
			assertions: func() {},
		},
		{
			name:      "Success_RetrySucceeds_MetricsEmitted",
			additions: []interface{}{foos[0], bars[0]},
			alters:    []interface{}{foos[1], bars[1]},
			removals:  []interface{}{foos[2]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []interface{}) {
				s._db.ExpectBegin()
				s._db.ExpectRollback()
				s._db.ExpectBegin()
				s._db.ExpectCommit()
				s.mappers[fooType].EXPECT().
					Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(2)
				s.mappers[barType].EXPECT().
					Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(2)
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), alters[0]).Return(nil).Times(2)
				s.mappers[barType].EXPECT().
					Update(ctx, gomock.Any(), alters[1]).Return(nil).Times(2)
				deletionFailure := s.mappers[fooType].EXPECT().
					Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("whoa"))
				s.mappers[fooType].EXPECT().
					Delete(ctx, gomock.Any(), removals[0]).Return(nil).After(deletionFailure)
			},
			ctx: context.Background(),
			assertions: func() {
				s.Len(s.scope.Snapshot().Counters(), 6)
				s.Contains(s.scope.Snapshot().Counters(), s.saveSuccessScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.retryAttemptScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.insertScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.updateScopeNameWithTags)
				s.Contains(s.scope.Snapshot().Counters(), s.deleteScopeNameWithTags)
				s.Len(s.scope.Snapshot().Timers(), 2)
				s.Contains(s.scope.Snapshot().Timers(), s.saveScopeNameWithTags)
			},
		},
	}
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save() {
	// test cases.
	tests := s.subtests()
	// execute test cases.
	for _, test := range tests {
		s.Run(test.name, func() {
			// setup.
			s.Setup()

			// arrange.
			s.Require().NoError(s.sut.Add(test.additions...))
			s.Require().NoError(s.sut.Alter(test.alters...))
			s.Require().NoError(s.sut.Remove(test.removals...))
			test.expectations(test.ctx, test.registers, test.additions, test.alters, test.removals)

			// action + assert.
			if test.panics {
				s.Require().Panics(func() { s.sut.Save(test.ctx) })
			} else {
				err := s.sut.Save(test.ctx)
				if test.err != nil {
					s.Require().EqualError(err, test.err.Error())
				} else {
					s.Require().NoError(err)
				}
			}
			s.Require().NoError(s._db.ExpectationsWereMet())
			test.assertions()

			// tear down.
			s.TearDown()
		})
	}
}

func (s *SQLUnitTestSuite) TearDown() {
	defer func() { s.isSetup, s.isTornDown = false, true }()

	s.db.Close()
	s.mc.Finish()
	s.sut = nil
	s.scope = nil
}

func (s *SQLUnitTestSuite) TearDownTest() {
	if !s.isTornDown {
		s.TearDown()
	}
}
