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
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freerware/work/v4"
	"github.com/freerware/work/v4/internal/mock"
	"github.com/freerware/work/v4/internal/test"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
)

type UnitOptionsTestSuite struct {
	suite.Suite

	// system under test.
	sut *work.UnitOptions
}

func TestUnitOptionsTestSuite(t *testing.T) {
	suite.Run(t, new(UnitOptionsTestSuite))
}

func (s *UnitOptionsTestSuite) SetupTest() {
	s.sut = &work.UnitOptions{}
}

func (s *UnitOptionsTestSuite) TestUnitDBOption() {
	// arrange.
	db, _, _ := sqlmock.New()

	// action.
	work.UnitDB(db)(s.sut)

	// assert.
	s.Equal(db, s.sut.DB)
}

func (s *UnitOptionsTestSuite) TestUnitDataMappers_Nil() {
	// arrange.
	var dm map[work.TypeName]work.UnitDataMapper

	// action.
	work.UnitDataMappers(dm)(s.sut)

	// assert.
	s.Nil(s.sut.InsertFuncs)
	s.Nil(s.sut.UpdateFuncs)
	s.Nil(s.sut.DeleteFuncs)
}

func (s *UnitOptionsTestSuite) TestUnitDataMappers_NotNil() {
	// arrange.
	dm := make(map[work.TypeName]work.UnitDataMapper)
	mc := gomock.NewController(s.T())
	fooTypeName := work.TypeNameOf(test.Foo{})
	dm[fooTypeName] = mock.NewUnitDataMapper(mc)

	// action.
	work.UnitDataMappers(dm)(s.sut)

	// assert.
	s.NotNil(s.sut.InsertFuncs)
	s.NotNil(s.sut.UpdateFuncs)
	s.NotNil(s.sut.DeleteFuncs)
}

func (s *UnitOptionsTestSuite) TestUnitInsertFunc() {
	// arrange.
	t := work.TypeNameOf(test.Foo{})
	var f work.UnitDataMapperFunc

	// action.
	work.UnitInsertFunc(t, f)(s.sut)

	// assert.
	s.NotNil(s.sut.InsertFuncs)
}

func (s *UnitOptionsTestSuite) TestUnitUpdateFunc() {
	// arrange.
	t := work.TypeNameOf(test.Foo{})
	var f work.UnitDataMapperFunc

	// action.
	work.UnitUpdateFunc(t, f)(s.sut)

	// assert.
	s.NotNil(s.sut.UpdateFuncs)
}

func (s *UnitOptionsTestSuite) TestUnitDeleteFunc() {
	// arrange.
	t := work.TypeNameOf(test.Foo{})
	var f work.UnitDataMapperFunc

	// action.
	work.UnitDeleteFunc(t, f)(s.sut)

	// assert.
	s.NotNil(s.sut.DeleteFuncs)
}

func (s *UnitOptionsTestSuite) TestUnitLogger() {
	// arrange.
	c := zap.NewDevelopmentConfig()
	c.DisableStacktrace = true
	l, _ := c.Build()

	// action.
	work.UnitLogger(l)(s.sut)

	// assert.
	s.Equal(l, s.sut.Logger)
}

func (s *UnitOptionsTestSuite) TestUnitScope() {
	// arrange.
	ts := tally.NewTestScope("test", map[string]string{})

	// action.
	work.UnitScope(ts)(s.sut)

	// assert.
	s.Equal(ts, s.sut.Scope)
}

func (s *UnitOptionsTestSuite) TestUnitAfterRegisterActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitAfterRegisterActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeAfterRegister]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterAddActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitAfterAddActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeAfterAdd]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterAlterActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitAfterAlterActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeAfterAlter]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterRemoveActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitAfterRemoveActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeAfterRemove]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterInsertsActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitAfterInsertsActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeAfterInserts]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterUpdatesActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitAfterUpdatesActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeAfterUpdates]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterDeletesActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitAfterDeletesActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeAfterDeletes]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterRollbackActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitAfterRollbackActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeAfterRollback]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterSaveActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitAfterSaveActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeAfterSave]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitBeforeInsertsActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitBeforeInsertsActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeBeforeInserts]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitBeforeUpdatesActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitBeforeUpdatesActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeBeforeUpdates]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitBeforeDeletesActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitBeforeDeletesActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeBeforeDeletes]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitBeforeRollbackActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitBeforeRollbackActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeBeforeRollback]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitBeforeSaveActions() {
	// arrange.
	same := false
	a := func(context work.UnitActionContext) { same = true }

	// action.
	work.UnitBeforeSaveActions(a)(s.sut)

	// assert.
	actions := s.sut.Actions[work.UnitActionTypeBeforeSave]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](work.UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestDisableDefaultLoggingActions() {

	// action.
	work.DisableDefaultLoggingActions()(s.sut)

	// assert.
	s.True(s.sut.DisableDefaultLoggingActions)
}

func (s *UnitOptionsTestSuite) TestUnitRetryAttempts_Negative() {

	// action.
	work.UnitRetryAttempts(-1)(s.sut)

	// assert.
	s.Zero(s.sut.RetryAttempts)
}

func (s *UnitOptionsTestSuite) TestUnitRetryAttempts_NotNegative() {
	// arrange.
	attempts := 2

	// action.
	work.UnitRetryAttempts(attempts)(s.sut)

	// assert.
	s.Equal(attempts, s.sut.RetryAttempts)
}

func (s *UnitOptionsTestSuite) TestUnitRetryDelay() {
	// arrange.
	delay := 10 * time.Second

	// action.
	work.UnitRetryDelay(delay)(s.sut)

	// assert.
	s.Equal(delay, s.sut.RetryDelay)
}

func (s *UnitOptionsTestSuite) TestUnitRetryMaximumJitter() {
	// arrange.
	delay := 10 * time.Second

	// action.
	work.UnitRetryMaximumJitter(delay)(s.sut)

	// assert.
	s.Equal(delay, s.sut.RetryMaximumJitter)
}

func (s *UnitOptionsTestSuite) TestUnitRetryType() {
	// arrange.
	var t work.UnitRetryDelayType = work.UnitRetryDelayTypeBackOff

	// action.
	work.UnitRetryType(t)(s.sut)

	// assert.
	s.Equal(t, s.sut.RetryType)
}

func (s *UnitOptionsTestSuite) TearDownTest() {
	s.sut = nil
}
