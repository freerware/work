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
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/freerware/work/v4"
	"github.com/freerware/work/v4/internal/mock"
	"github.com/freerware/work/v4/unit"
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

func TestBestEffortUnitTestSuite(t *testing.T) {
	suite.Run(t, new(BestEffortUnitTestSuite))
}

func (s *BestEffortUnitTestSuite) Setup() {
	defer func() { s.isSetup, s.isTornDown = true, false }()

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
	s.retryAttemptScopeName = fmt.Sprintf("%s.%s", s.scopePrefix, "unit.retry.attempt")
	s.retryAttemptScopeNameWithTags = fmt.Sprintf("%s%s%s", s.retryAttemptScopeName, sep, s.tags)
	s.insertScopeName = fmt.Sprintf("%s.%s", s.scopePrefix, "unit.insert")
	s.insertScopeNameWithTags = fmt.Sprintf("%s%s%s", s.insertScopeName, sep, s.tags)
	s.updateScopeName = fmt.Sprintf("%s.%s", s.scopePrefix, "unit.update")
	s.updateScopeNameWithTags = fmt.Sprintf("%s%s%s", s.updateScopeName, sep, s.tags)
	s.deleteScopeName = fmt.Sprintf("%s.%s", s.scopePrefix, "unit.delete")
	s.deleteScopeNameWithTags = fmt.Sprintf("%s%s%s", s.deleteScopeName, sep, s.tags)

	// test entities.
	foo := Foo{id: 28}
	fooTypeName := work.TypeNameOf(foo)
	bar := Bar{id: "28"}
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
	s.retryCount = 2
	s.scope = ts
	var err error
	opts := []work.UnitOption{
		work.UnitDataMappers(dm),
		work.UnitLogger(l),
		work.UnitScope(ts),
		work.UnitRetryAttempts(s.retryCount),
	}
	s.sut, err = work.NewUnit(opts...)
	s.Require().NoError(err)
}

func (s *BestEffortUnitTestSuite) SetupTest() {
	if !s.isSetup {
		s.Setup()
	}
}

func (s *BestEffortUnitTestSuite) subtests() []TableDrivenTest {
	foos := []unit.Entity{Foo{id: 28}, Foo{id: 1992}, Foo{id: 2}, Foo{id: 1111}}
	bars := []unit.Entity{Bar{id: "id"}, Bar{id: "1992"}}
	fooType, barType := work.TypeNameOf(Foo{}), work.TypeNameOf(Bar{})
	return []TableDrivenTest{
		{
			name:      "InsertError",
			additions: []unit.Entity{foos[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				s.mappers[fooType].EXPECT().
					Insert(ctx, gomock.Any(), additions[0]).Return(errors.New("whoa")).Times(s.retryCount)

				// arrange - rollback invocations.
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), registers[0], registers[2]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().
					Update(ctx, gomock.Any(), registers[1]).Return(nil).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("whoa"),
			assertions: func() {},
		},
		{
			name:      "InsertError_MetricsEmitted",
			additions: []unit.Entity{foos[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				s.mappers[fooType].EXPECT().
					Insert(ctx, gomock.Any(), additions[0]).Return(errors.New("whoa")).Times(s.retryCount)

				// arrange - rollback invocations.
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), registers[0], registers[2]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().
					Update(ctx, gomock.Any(), registers[1]).Return(nil).Times(s.retryCount)
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
			additions: []unit.Entity{foos[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				s.mappers[fooType].EXPECT().
					Insert(ctx, gomock.Any(), additions[0]).
					Return(errors.New("ouch")).Times(s.retryCount)

				// arrange - rollback invocations.
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), registers[0], registers[1]).
					Return(errors.New("whoa")).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("ouch; whoa"),
			assertions: func() {},
		},
		{
			name:      "InsertAndRollbackError_MetricsEmitted",
			additions: []unit.Entity{foos[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				s.mappers[fooType].EXPECT().
					Insert(ctx, gomock.Any(), additions[0]).
					Return(errors.New("ouch")).Times(s.retryCount)

				// arrange - rollback invocations.
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), registers[0], registers[1]).
					Return(errors.New("whoa")).Times(s.retryCount)
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
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				for i := 0; i < s.retryCount; i++ {
					// arrange - encounter update error.
					applyUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(errors.New("whoa"))

					// arrange - successfully rollback updates.
					s.mappers[fooType].EXPECT().
						Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyUpdate)
					s.mappers[barType].EXPECT().
						Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyUpdate)
				}

				// arrange - successfully rollback inserts.
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Delete(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("whoa"),
			assertions: func() {},
		},
		{
			name:      "UpdateError_MetricsEmitted",
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				for i := 0; i < s.retryCount; i++ {
					// arrange - encounter update error.
					applyUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(errors.New("whoa"))

					// arrange - successfully rollback updates.
					s.mappers[fooType].EXPECT().
						Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyUpdate)
					s.mappers[barType].EXPECT().
						Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyUpdate)
				}

				// arrange - successfully rollback inserts.
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				s.mappers[barType].EXPECT().Delete(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
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
			additions: []unit.Entity{foos[0]},
			alters:    []unit.Entity{foos[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				for i := 0; i < s.retryCount; i++ {
					// arrange - encounter update error.
					applyUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(errors.New("ouch"))

					// arrange - successfully roll back updates.
					s.mappers[fooType].EXPECT().
						Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyUpdate)
					s.mappers[barType].EXPECT().
						Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyUpdate)
				}

				// arrange - encounter error when rolling back inserts.
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), additions[0]).Return(errors.New("whoa")).Times(s.retryCount)
			},
			ctx:        context.Background(),
			err:        errors.New("ouch; whoa"),
			assertions: func() {},
		},
		{
			name:      "UpdateAndRollbackError_MetricsEmitted",
			additions: []unit.Entity{foos[0]},
			alters:    []unit.Entity{foos[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				for i := 0; i < s.retryCount; i++ {
					// arrange - encounter update error.
					applyUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(errors.New("ouch"))

					// arrange - successfully roll back updates.
					s.mappers[fooType].EXPECT().
						Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyUpdate)
					s.mappers[barType].EXPECT().
						Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyUpdate)
				}

				// arrange - encounter error when rolling back inserts.
				s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), additions[0]).Return(errors.New("whoa")).Times(s.retryCount)
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
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				for i := 0; i < s.retryCount; i++ {
					// arrange - successfully apply updates.
					applyFooUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
					applyBarUpdate := s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)

					// arrange - successfully roll back updates.
					s.mappers[fooType].EXPECT().
						Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyFooUpdate)
					s.mappers[barType].EXPECT().
						Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyBarUpdate)

					// arrange - encounter delete error.
					applyDelete := s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("whoa"))

					// arrange - successfully roll back inserts.
					s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), additions[0]).Return(nil).After(applyDelete)
					s.mappers[barType].EXPECT().Delete(ctx, gomock.Any(), additions[1]).Return(nil).After(applyDelete)
				}
			},
			ctx:        context.Background(),
			err:        errors.New("whoa"),
			assertions: func() {},
		},
		{
			name:      "DeleteError_MetricsEmitted",
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil).Times(s.retryCount)
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				for i := 0; i < s.retryCount; i++ {
					// arrange - successfully apply updates.
					applyFooUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
					applyBarUpdate := s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)

					// arrange - successfully roll back updates.
					s.mappers[fooType].EXPECT().
						Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyFooUpdate)
					s.mappers[barType].EXPECT().
						Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyBarUpdate)

					// arrange - encounter delete error.
					applyDelete := s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("whoa"))

					// arrange - successfully roll back inserts.
					s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), additions[0]).Return(nil).After(applyDelete)
					s.mappers[barType].EXPECT().Delete(ctx, gomock.Any(), additions[1]).Return(nil).After(applyDelete)
				}
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
			additions: []unit.Entity{foos[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				for i := 0; i < s.retryCount; i++ {
					// arrange - successfully apply updates.
					applyFooUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
					applyBarUpdate := s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)

					// arrange - successfully roll back updates.
					s.mappers[fooType].EXPECT().
						Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyFooUpdate)
					s.mappers[barType].EXPECT().
						Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyBarUpdate)

					// arrange - encounter delete error.
					applyDelete := s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("whoa"))

					// arrange - encounter error when rolling back inserts.
					s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), additions[0]).Return(errors.New("ouch")).After(applyDelete)
				}
			},
			ctx:        context.Background(),
			err:        errors.New("whoa; ouch"),
			assertions: func() {},
		},
		{
			name:      "DeleteAndRollbackError_MetricsEmitted",
			additions: []unit.Entity{foos[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil).Times(s.retryCount)
				for i := 0; i < s.retryCount; i++ {
					// arrange - successfully apply updates.
					applyFooUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
					applyBarUpdate := s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)

					// arrange - successfully roll back updates.
					s.mappers[fooType].EXPECT().
						Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyFooUpdate)
					s.mappers[barType].EXPECT().
						Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyBarUpdate)

					// arrange - encounter delete error.
					applyDelete := s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("whoa"))

					// arrange - encounter error when rolling back inserts.
					s.mappers[fooType].EXPECT().Delete(ctx, gomock.Any(), additions[0]).Return(errors.New("ouch")).After(applyDelete)
				}
			},
			ctx: context.Background(),
			err: errors.New("whoa; ouch"),
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
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)

				// arrange - successfully apply updates.
				applyFooUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				applyBarUpdate := s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)

				// arrange - successfully roll back updates.
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyFooUpdate)
				s.mappers[barType].EXPECT().
					Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyBarUpdate)

				// arrange - encounter delete panic.
				applyDelete := s.mappers[fooType].
					EXPECT().Delete(ctx, gomock.Any(), removals[0]).Do(
					func(_ctx context.Context, _mCtx work.MapperContext, e ...unit.Entity) {
						panic("whoa")
					},
				)

				// arrange - successfully roll back inserts.
				s.mappers[fooType].EXPECT().
					Delete(ctx, gomock.Any(), additions[0]).Return(nil).After(applyDelete)
				s.mappers[barType].EXPECT().
					Delete(ctx, gomock.Any(), additions[1]).Return(nil).After(applyDelete)
			},
			ctx:        context.Background(),
			assertions: func() {},
			panics:     true,
		},
		{
			name:      "Panic_MetricsEmitted",
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)

				// arrange - successfully apply updates.
				applyFooUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				applyBarUpdate := s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)

				// arrange - successfully roll back updates.
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyFooUpdate)
				s.mappers[barType].EXPECT().
					Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyBarUpdate)

				// arrange - encounter delete panic.
				applyDelete := s.mappers[fooType].
					EXPECT().Delete(ctx, gomock.Any(), removals[0]).Do(
					func(_ctx context.Context, _mCtx work.MapperContext, e ...unit.Entity) {
						panic("whoa")
					},
				)

				// arrange - successfully roll back inserts.
				s.mappers[fooType].EXPECT().
					Delete(ctx, gomock.Any(), additions[0]).Return(nil).After(applyDelete)
				s.mappers[barType].EXPECT().
					Delete(ctx, gomock.Any(), additions[1]).Return(nil).After(applyDelete)
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
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)

				// arrange - successfully apply updates.
				applyFooUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)

				// arrange - encounter update error during roll back.
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[1]}).
					Return(errors.New("whoa")).After(applyFooUpdate)

				// arrange - encounter delete panic.
				s.mappers[fooType].
					EXPECT().Delete(ctx, gomock.Any(), removals[0]).Do(
					func(_ctx context.Context, _mCtx work.MapperContext, e ...unit.Entity) {
						panic("whoa")
					},
				)
			},
			ctx:        context.Background(),
			assertions: func() {},
			panics:     true,
			err:        errors.New("whoa"),
		},
		{
			name:      "PanicAndRollbackError_MetricsEmitted",
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)

				// arrange - successfully apply updates.
				applyFooUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)

				// arrange - encounter update error during roll back.
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[1]}).
					Return(errors.New("whoa")).After(applyFooUpdate)

				// arrange - encounter delete panic.
				s.mappers[fooType].
					EXPECT().Delete(ctx, gomock.Any(), removals[0]).Do(
					func(_ctx context.Context, _mCtx work.MapperContext, e ...unit.Entity) {
						panic("whoa")
					},
				)
			},
			ctx: context.Background(),
			err: errors.New("whoa"),
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
			name:      "PanicAndRollbackPanic",
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)

				// arrange - successfully apply updates.
				applyFooUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)

				// arrange - encounter update panic during roll back.
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[1]}).Do(
					func(_ctx context.Context, _mCtx work.MapperContext, e ...unit.Entity) {
						panic("whoa")
					},
				).After(applyFooUpdate)

				// arrange - encounter delete panic.
				s.mappers[fooType].
					EXPECT().Delete(ctx, gomock.Any(), removals[0]).Do(
					func(_ctx context.Context, _mCtx work.MapperContext, e ...unit.Entity) {
						panic("whoa")
					},
				)
			},
			ctx:        context.Background(),
			assertions: func() {},
			panics:     true,
			err:        errors.New("whoa"),
		},
		{
			name:      "PanicAndRollbackPanic_MetricsEmitted",
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				// arrange - successfully apply inserts.
				s.mappers[fooType].EXPECT().Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().Insert(ctx, gomock.Any(), additions[1]).Return(nil)

				// arrange - successfully apply updates.
				applyFooUpdate := s.mappers[fooType].EXPECT().Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().Update(ctx, gomock.Any(), alters[1]).Return(nil)

				// arrange - encounter update panic during roll back.
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[1]}).Do(
					func(_ctx context.Context, _mCtx work.MapperContext, e ...unit.Entity) {
						panic("whoa")
					},
				).After(applyFooUpdate)

				// arrange - encounter delete panic.
				s.mappers[fooType].
					EXPECT().Delete(ctx, gomock.Any(), removals[0]).Do(
					func(_ctx context.Context, _mCtx work.MapperContext, e ...unit.Entity) {
						panic("whoa")
					},
				)
			},
			ctx: context.Background(),
			err: errors.New("whoa"),
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
			name:      "Success",
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				s.mappers[fooType].EXPECT().
					Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().
					Insert(ctx, gomock.Any(), additions[1]).Return(nil)
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().
					Update(ctx, gomock.Any(), alters[1]).Return(nil)
				s.mappers[fooType].EXPECT().
					Delete(ctx, gomock.Any(), removals[0]).Return(nil)
			},
			ctx:        context.Background(),
			assertions: func() {},
		},
		{
			name:      "Success_MetricsEmitted",
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				s.mappers[fooType].EXPECT().
					Insert(ctx, gomock.Any(), additions[0]).Return(nil)
				s.mappers[barType].EXPECT().
					Insert(ctx, gomock.Any(), additions[1]).Return(nil)
				s.mappers[fooType].EXPECT().
					Update(ctx, gomock.Any(), alters[0]).Return(nil)
				s.mappers[barType].EXPECT().
					Update(ctx, gomock.Any(), alters[1]).Return(nil)
				s.mappers[fooType].EXPECT().
					Delete(ctx, gomock.Any(), removals[0]).Return(nil)
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
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				for i := 0; i < 2; i++ {
					// arrange - successfully apply inserts.
					s.mappers[fooType].EXPECT().
						Insert(ctx, gomock.Any(), additions[0]).Return(nil)
					s.mappers[barType].EXPECT().
						Insert(ctx, gomock.Any(), additions[1]).Return(nil)

					// arrange - successfully apply updates.
					applyFooUpdate := s.mappers[fooType].EXPECT().
						Update(ctx, gomock.Any(), alters[0]).Return(nil)
					applyBarUpdate := s.mappers[barType].EXPECT().
						Update(ctx, gomock.Any(), alters[1]).Return(nil)

					if i != 1 {
						// arrange - successfully roll back inserts.
						s.mappers[fooType].EXPECT().
							Delete(ctx, gomock.Any(), additions[0]).Return(nil)
						s.mappers[barType].EXPECT().
							Delete(ctx, gomock.Any(), additions[1]).Return(nil)

						// arrange - successfully roll back updates.
						s.mappers[fooType].EXPECT().
							Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyFooUpdate)
						s.mappers[barType].EXPECT().
							Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyBarUpdate)

						// arrange - encounter transient delete error.
						deletionFailure := s.mappers[fooType].EXPECT().
							Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("whoa"))

						// arrange - successfully apply deletes on retry.
						s.mappers[fooType].EXPECT().
							Delete(ctx, gomock.Any(), removals[0]).Return(nil).After(deletionFailure)
					}
				}

			},
			ctx:        context.Background(),
			assertions: func() {},
		},
		{
			name:      "Success_RetrySucceeds_MetricsEmitted",
			additions: []unit.Entity{foos[0], bars[0]},
			alters:    []unit.Entity{foos[1], bars[1]},
			removals:  []unit.Entity{foos[2]},
			registers: []unit.Entity{foos[1], bars[1], foos[3]},
			expectations: func(ctx context.Context, registers, additions, alters, removals []unit.Entity) {
				for i := 0; i < 2; i++ {
					// arrange - successfully apply inserts.
					s.mappers[fooType].EXPECT().
						Insert(ctx, gomock.Any(), additions[0]).Return(nil)
					s.mappers[barType].EXPECT().
						Insert(ctx, gomock.Any(), additions[1]).Return(nil)

					// arrange - successfully apply updates.
					applyFooUpdate := s.mappers[fooType].EXPECT().
						Update(ctx, gomock.Any(), alters[0]).Return(nil)
					applyBarUpdate := s.mappers[barType].EXPECT().
						Update(ctx, gomock.Any(), alters[1]).Return(nil)

					if i != 1 {
						// arrange - successfully roll back inserts.
						s.mappers[fooType].EXPECT().
							Delete(ctx, gomock.Any(), additions[0]).Return(nil)
						s.mappers[barType].EXPECT().
							Delete(ctx, gomock.Any(), additions[1]).Return(nil)

						// arrange - successfully roll back updates.
						s.mappers[fooType].EXPECT().
							Update(ctx, gomock.Any(), []unit.Entity{registers[0], registers[2]}).Return(nil).After(applyFooUpdate)
						s.mappers[barType].EXPECT().
							Update(ctx, gomock.Any(), registers[1]).Return(nil).After(applyBarUpdate)

						// arrange - encounter transient delete error.
						deletionFailure := s.mappers[fooType].EXPECT().
							Delete(ctx, gomock.Any(), removals[0]).Return(errors.New("whoa"))

						// arrange - successfully apply deletes on retry.
						s.mappers[fooType].EXPECT().
							Delete(ctx, gomock.Any(), removals[0]).Return(nil).After(deletionFailure)
					}
				}

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save() {
	// test cases.
	tests := s.subtests()
	// execute test cases.
	for _, test := range tests {
		s.Run(test.name, func() {
			// setup.
			s.Setup()

			// arrange.
			s.Require().NoError(s.sut.Register(test.registers...))
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
			test.assertions()

			// tear down.
			s.TearDown()
		})
	}
}

func (s *BestEffortUnitTestSuite) TearDown() {
	defer func() { s.isSetup, s.isTornDown = false, true }()

	s.mc.Finish()
	s.sut = nil
	s.scope = nil
}

func (s *BestEffortUnitTestSuite) TearDownTest() {
	if !s.isTornDown {
		s.TearDown()
	}
}
