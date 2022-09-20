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
	"testing"

	"github.com/freerware/work/v3"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
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

func (s *UnitOptionsTestSuite) TearDownTest() {
	s.sut = nil
}
