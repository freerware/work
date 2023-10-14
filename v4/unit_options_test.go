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

package work

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freerware/work/v4/internal/test"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
)

type UnitOptionsTestSuite struct {
	suite.Suite

	// system under test.
	sut *UnitOptions
}

func TestUnitOptionsTestSuite(t *testing.T) {
	suite.Run(t, new(UnitOptionsTestSuite))
}

func (s *UnitOptionsTestSuite) SetupTest() {
	s.sut = &UnitOptions{}
}

func (s *UnitOptionsTestSuite) TestUnitDBOption() {
	// arrange.
	db, _, _ := sqlmock.New()

	// action.
	UnitDB(db)(s.sut)

	// assert.
	s.Equal(db, s.sut.db)
}

func (s *UnitOptionsTestSuite) TestUnitDataMappers_Nil() {
	// arrange.
	var dm map[TypeName]UnitDataMapper

	// action.
	UnitDataMappers(dm)(s.sut)

	// assert.
	s.Nil(s.sut.insertFuncs)
	s.Nil(s.sut.updateFuncs)
	s.Nil(s.sut.deleteFuncs)
}

func (s *UnitOptionsTestSuite) TestUnitDataMappers_NotNil() {
	// arrange.
	dm := make(map[TypeName]UnitDataMapper)
	fooTypeName := TypeNameOf(test.Foo{})
	dm[fooTypeName] = &noOpDataMapper{}

	// action.
	UnitDataMappers(dm)(s.sut)

	// assert.
	s.NotNil(s.sut.insertFuncs)
	s.NotNil(s.sut.updateFuncs)
	s.NotNil(s.sut.deleteFuncs)
}

func (s *UnitOptionsTestSuite) TestUnitInsertFunc() {
	// arrange.
	t := TypeNameOf(test.Foo{})
	var f UnitDataMapperFunc

	// action.
	UnitInsertFunc(t, f)(s.sut)

	// assert.
	s.NotNil(s.sut.insertFuncs)
}

func (s *UnitOptionsTestSuite) TestUnitUpdateFunc() {
	// arrange.
	t := TypeNameOf(test.Foo{})
	var f UnitDataMapperFunc

	// action.
	UnitUpdateFunc(t, f)(s.sut)

	// assert.
	s.NotNil(s.sut.updateFuncs)
}

func (s *UnitOptionsTestSuite) TestUnitDeleteFunc() {
	// arrange.
	t := TypeNameOf(test.Foo{})
	var f UnitDataMapperFunc

	// action.
	UnitDeleteFunc(t, f)(s.sut)

	// assert.
	s.NotNil(s.sut.deleteFuncs)
}

func (s *UnitOptionsTestSuite) TestUnitLogger() {
	// arrange.
	c := zap.NewDevelopmentConfig()
	c.DisableStacktrace = true
	l, _ := c.Build()

	// action.
	UnitZapLogger(l)(s.sut)

	// assert.
	s.Equal(l, s.sut.logger)
}

func (s *UnitOptionsTestSuite) TestUnitScope() {
	// arrange.
	ts := tally.NewTestScope("test", map[string]string{})

	// action.
	UnitTallyMetricScope(ts)(s.sut)

	// assert.
	s.Equal(ts, s.sut.scope)
}

func (s *UnitOptionsTestSuite) TestUnitAfterRegisterActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitAfterRegisterActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeAfterRegister]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterAddActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitAfterAddActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeAfterAdd]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterAlterActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitAfterAlterActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeAfterAlter]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterRemoveActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitAfterRemoveActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeAfterRemove]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterInsertsActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitAfterInsertsActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeAfterInserts]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterUpdatesActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitAfterUpdatesActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeAfterUpdates]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterDeletesActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitAfterDeletesActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeAfterDeletes]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterRollbackActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitAfterRollbackActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeAfterRollback]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitAfterSaveActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitAfterSaveActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeAfterSave]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitBeforeInsertsActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitBeforeInsertsActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeBeforeInserts]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitBeforeUpdatesActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitBeforeUpdatesActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeBeforeUpdates]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitBeforeDeletesActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitBeforeDeletesActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeBeforeDeletes]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitBeforeRollbackActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitBeforeRollbackActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeBeforeRollback]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestUnitBeforeSaveActions() {
	// arrange.
	same := false
	a := func(context UnitActionContext) { same = true }

	// action.
	UnitBeforeSaveActions(a)(s.sut)

	// assert.
	actions := s.sut.actions[UnitActionTypeBeforeSave]
	s.Len(actions, 1)
	s.Condition(func() bool {
		actions[0](UnitActionContext{})
		return same
	})
}

func (s *UnitOptionsTestSuite) TestDisableDefaultLoggingActions() {

	// action.
	DisableDefaultLoggingActions()(s.sut)

	// assert.
	s.True(s.sut.disableDefaultLoggingActions)
}

func (s *UnitOptionsTestSuite) TestUnitRetryAttempts_Negative() {

	// action.
	UnitRetryAttempts(-1)(s.sut)

	// assert.
	s.Zero(s.sut.retryAttempts)
}

func (s *UnitOptionsTestSuite) TestUnitRetryAttempts_NotNegative() {
	// arrange.
	attempts := 2

	// action.
	UnitRetryAttempts(attempts)(s.sut)

	// assert.
	s.Equal(attempts, s.sut.retryAttempts)
}

func (s *UnitOptionsTestSuite) TestUnitRetryDelay() {
	// arrange.
	delay := 10 * time.Second

	// action.
	UnitRetryDelay(delay)(s.sut)

	// assert.
	s.Equal(delay, s.sut.retryDelay)
}

func (s *UnitOptionsTestSuite) TestUnitRetryMaximumJitter() {
	// arrange.
	delay := 10 * time.Second

	// action.
	UnitRetryMaximumJitter(delay)(s.sut)

	// assert.
	s.Equal(delay, s.sut.retryMaximumJitter)
}

func (s *UnitOptionsTestSuite) TestUnitRetryType() {
	// arrange.
	var t UnitRetryDelayType = UnitRetryDelayTypeBackOff

	// action.
	UnitRetryType(t)(s.sut)

	// assert.
	s.Equal(t, s.sut.retryType)
}

func (s *UnitOptionsTestSuite) TestUnitWithCacheClient() {
	// arrange.
	cacheClient := &memoryCacheClient{}

	// action.
	UnitWithCacheClient(cacheClient)(s.sut)

	// assert.
	s.Equal(cacheClient, s.sut.cacheClient)
}

func (s *UnitOptionsTestSuite) TearDownTest() {
	s.sut = nil
}

type noOpDataMapper struct{}

func (dm noOpDataMapper) Insert(ctx context.Context, mCtx UnitMapperContext, e ...interface{}) error {
	return nil
}

func (dm noOpDataMapper) Update(ctx context.Context, mCtx UnitMapperContext, e ...interface{}) error {
	return nil
}

func (dm noOpDataMapper) Delete(ctx context.Context, mCtx UnitMapperContext, e ...interface{}) error {
	return nil
}
