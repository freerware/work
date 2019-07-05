package work

import (
	"errors"
	"fmt"
	"testing"

	"github.com/freerware/work/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

type BestEffortUnitTestSuite struct {
	suite.Suite

	// system under test.
	sut Unit

	// mocks.
	inserters map[TypeName]*mocks.Inserter
	updaters  map[TypeName]*mocks.Updater
	deleters  map[TypeName]*mocks.Deleter
	scope     tally.TestScope

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
	fooTypeName := TypeNameOf(foo)
	bar := Bar{ID: "28"}
	barTypeName := TypeNameOf(bar)

	// initialize mocks.
	s.inserters = make(map[TypeName]*mocks.Inserter)
	s.inserters[fooTypeName] = &mocks.Inserter{}
	s.inserters[barTypeName] = &mocks.Inserter{}
	s.updaters = make(map[TypeName]*mocks.Updater)
	s.updaters[fooTypeName] = &mocks.Updater{}
	s.updaters[barTypeName] = &mocks.Updater{}
	s.deleters = make(map[TypeName]*mocks.Deleter)
	s.deleters[fooTypeName] = &mocks.Deleter{}
	s.deleters[barTypeName] = &mocks.Deleter{}

	// construct SUT.
	i := make(map[TypeName]Inserter)
	for t, m := range s.inserters {
		i[t] = m
	}
	u := make(map[TypeName]Updater)
	for t, m := range s.updaters {
		u[t] = m
	}
	d := make(map[TypeName]Deleter)
	for t, m := range s.deleters {
		d[t] = m
	}

	c := zap.NewDevelopmentConfig()
	c.DisableStacktrace = true
	l, _ := c.Build()
	ts := tally.NewTestScope(s.scopePrefix, map[string]string{})
	params := UnitParameters{
		Inserters: i,
		Updaters:  u,
		Deleters:  d,
		Logger:    l,
		Scope:     ts,
	}
	s.scope = ts
	s.sut = NewBestEffortUnit(params)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_InserterError() {

	// arrange.
	fooType := TypeNameOf(Foo{})
	barType := TypeNameOf(Bar{})
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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.updaters[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.updaters[barType].On(
		"Update", registeredEntities[1]).Return(nil)

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_InserterAndRollbackError() {

	// arrange.
	fooType := TypeNameOf(Foo{})
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

	s.inserters[fooType].On("Insert", addedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.updaters[fooType].On(
		"Update", registeredEntities[0], registeredEntities[1]).Return(err)

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_UpdaterError() {

	// arrange.
	fooType := TypeNameOf(Foo{})
	barType := TypeNameOf(Bar{})
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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.inserters[barType].On("Insert", addedEntities[1]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(err).Once()

	// arrange - rollback invocations.
	s.deleters[fooType].On("Delete", addedEntities[0]).Return(nil)
	s.deleters[barType].On("Delete", addedEntities[1]).Return(nil)
	s.updaters[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.updaters[barType].On(
		"Update", registeredEntities[1]).Return(nil)

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_UpdaterAndRollbackError() {

	// arrange.
	fooType := TypeNameOf(Foo{})
	barType := TypeNameOf(Bar{})
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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.deleters[fooType].On("Delete", addedEntities[0]).Return(err)
	s.updaters[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.updaters[barType].On(
		"Update", registeredEntities[1]).Return(nil)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Error(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_DeleterError() {

	// arrange.
	fooType := TypeNameOf(Foo{})
	barType := TypeNameOf(Bar{})
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
	s.inserters[barType].On("Insert", addedEntities[1]).Return(nil)
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.updaters[barType].On("Update", updatedEntities[1]).Return(nil)
	s.deleters[fooType].On("Delete", removedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.deleters[fooType].On("Delete", addedEntities[0]).Return(nil)
	s.deleters[barType].On("Delete", addedEntities[1]).Return(nil)
	s.updaters[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.updaters[barType].On(
		"Update", registeredEntities[1]).Return(nil)

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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_DeleterAndRollbackError() {

	// arrange.
	fooType := TypeNameOf(Foo{})
	barType := TypeNameOf(Bar{})
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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.updaters[barType].On("Update", updatedEntities[1]).Return(nil)
	s.deleters[fooType].On("Delete", removedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.deleters[fooType].On("Delete", addedEntities[0]).Return(err)
	s.updaters[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.updaters[barType].On(
		"Update", registeredEntities[1]).Return(nil)

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
	fooType := TypeNameOf(Foo{})
	barType := TypeNameOf(Bar{})
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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.inserters[barType].On("Insert", addedEntities[1]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.updaters[barType].On("Update", updatedEntities[1]).Return(nil)
	s.deleters[fooType].
		On("Delete", removedEntities[0]).
		Return().Run(func(args mock.Arguments) { panic("whoa") })

	// arrange - rollback invocations.
	s.deleters[fooType].On("Delete", addedEntities[0]).Return(nil)
	s.deleters[barType].On("Delete", addedEntities[1]).Return(nil)
	s.updaters[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.updaters[barType].On(
		"Update", registeredEntities[1]).Return(nil)

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
	fooType := TypeNameOf(Foo{})
	barType := TypeNameOf(Bar{})
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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.inserters[barType].On("Insert", addedEntities[1]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.updaters[barType].On("Update", updatedEntities[1]).Return(nil)
	s.deleters[fooType].
		On("Delete", removedEntities[0]).
		Return().Run(func(args mock.Arguments) { panic("whoa") })

	// arrange - rollback invocations.
	s.updaters[fooType].On(
		"Update", registeredEntities[0], registeredEntities[1]).Return(err)

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
	fooType := TypeNameOf(Foo{})
	barType := TypeNameOf(Bar{})
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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.inserters[barType].On("Insert", addedEntities[1]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.updaters[barType].On("Update", updatedEntities[1]).Return(nil)
	s.deleters[fooType].
		On("Delete", removedEntities[0]).
		Return().Run(func(args mock.Arguments) { panic("whoa") })

	// arrange - rollback invocations.
	s.updaters[fooType].
		On("Update", registeredEntities[0], registeredEntities[1]).
		Return().Run(func(args mock.Arguments) { panic("whoa") })

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
	fooType := TypeNameOf(Foo{})
	barType := TypeNameOf(Bar{})
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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.inserters[barType].On("Insert", addedEntities[1]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.updaters[barType].On("Update", updatedEntities[1]).Return(nil)
	s.deleters[fooType].On("Delete", removedEntities[0]).Return(nil)

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

func (s *BestEffortUnitTestSuite) AfterTest(suiteName, testName string) {
	for _, i := range s.inserters {
		i.AssertExpectations(s.T())
	}
	for _, u := range s.updaters {
		u.AssertExpectations(s.T())
	}
	for _, d := range s.deleters {
		d.AssertExpectations(s.T())
	}
}

func (s *BestEffortUnitTestSuite) TearDownTest() {
	s.sut = nil
}
