package work

import (
	"errors"
	"fmt"
	"testing"

	"github.com/freerware/work/internal/mocks"
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
	mappers map[TypeName]*mocks.DataMapper
	scope   tally.TestScope

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
	s.mappers = make(map[TypeName]*mocks.DataMapper)
	s.mappers[fooTypeName] = &mocks.DataMapper{}
	s.mappers[barTypeName] = &mocks.DataMapper{}

	// construct SUT.
	dm := make(map[TypeName]DataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}

	c := zap.NewDevelopmentConfig()
	c.DisableStacktrace = true
	l, _ := c.Build()
	ts := tally.NewTestScope(s.scopePrefix, map[string]string{})
	s.scope = ts
	var err error
	s.sut, err = NewBestEffortUnit(dm, UnitLogger(l), UnitScope(ts))
	s.Require().NoError(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_NewBestEffortUnit_MissingDataMappers() {

	// action.
	var err error
	s.sut, err = NewBestEffortUnit(map[TypeName]DataMapper{})

	// assert.
	s.Error(err)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_InsertError() {

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
	s.mappers[fooType].On("Insert", addedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.mappers[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].On(
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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_InsertAndRollbackError() {

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

	s.mappers[fooType].On("Insert", addedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.mappers[fooType].On(
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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_UpdateError() {

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
	s.mappers[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.mappers[barType].On("Insert", addedEntities[1]).Return(nil)
	s.mappers[fooType].On("Update", updatedEntities[0]).Return(err).Once()

	// arrange - rollback invocations.
	s.mappers[fooType].On("Delete", addedEntities[0]).Return(nil)
	s.mappers[barType].On("Delete", addedEntities[1]).Return(nil)
	s.mappers[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].On(
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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_UpdateAndRollbackError() {

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
	s.mappers[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.mappers[fooType].On("Update", updatedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.mappers[fooType].On("Delete", addedEntities[0]).Return(err)
	s.mappers[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].On(
		"Update", registeredEntities[1]).Return(nil)

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
	s.mappers[barType].On("Insert", addedEntities[1]).Return(nil)
	s.mappers[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.mappers[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.mappers[barType].On("Update", updatedEntities[1]).Return(nil)
	s.mappers[fooType].On("Delete", removedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.mappers[fooType].On("Delete", addedEntities[0]).Return(nil)
	s.mappers[barType].On("Delete", addedEntities[1]).Return(nil)
	s.mappers[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].On(
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

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_DeleteAndRollbackError() {

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
	s.mappers[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.mappers[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.mappers[barType].On("Update", updatedEntities[1]).Return(nil)
	s.mappers[fooType].On("Delete", removedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.mappers[fooType].On("Delete", addedEntities[0]).Return(err)
	s.mappers[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].On(
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
	s.mappers[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.mappers[barType].On("Insert", addedEntities[1]).Return(nil)
	s.mappers[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.mappers[barType].On("Update", updatedEntities[1]).Return(nil)
	s.mappers[fooType].
		On("Delete", removedEntities[0]).
		Return().Run(func(args mock.Arguments) { panic("whoa") })

	// arrange - rollback invocations.
	s.mappers[fooType].On("Delete", addedEntities[0]).Return(nil)
	s.mappers[barType].On("Delete", addedEntities[1]).Return(nil)
	s.mappers[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.mappers[barType].On(
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
	s.mappers[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.mappers[barType].On("Insert", addedEntities[1]).Return(nil)
	s.mappers[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.mappers[barType].On("Update", updatedEntities[1]).Return(nil)
	s.mappers[fooType].
		On("Delete", removedEntities[0]).
		Return().Run(func(args mock.Arguments) { panic("whoa") })

	// arrange - rollback invocations.
	s.mappers[fooType].On(
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
	s.mappers[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.mappers[barType].On("Insert", addedEntities[1]).Return(nil)
	s.mappers[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.mappers[barType].On("Update", updatedEntities[1]).Return(nil)
	s.mappers[fooType].
		On("Delete", removedEntities[0]).
		Return().Run(func(args mock.Arguments) { panic("whoa") })

	// arrange - rollback invocations.
	s.mappers[fooType].
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
	s.mappers[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.mappers[barType].On("Insert", addedEntities[1]).Return(nil)
	s.mappers[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.mappers[barType].On("Update", updatedEntities[1]).Return(nil)
	s.mappers[fooType].On("Delete", removedEntities[0]).Return(nil)

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
	dm := make(map[TypeName]DataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}
	var err error
	s.sut, err = NewBestEffortUnit(dm)
	s.Require().NoError(err)

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
	s.mappers[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.mappers[barType].On("Insert", addedEntities[1]).Return(nil)
	s.mappers[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.mappers[barType].On("Update", updatedEntities[1]).Return(nil)
	s.mappers[fooType].On("Delete", removedEntities[0]).Return(nil)

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
	mappers := map[TypeName]DataMapper{
		TypeNameOf(Bar{}): &mocks.DataMapper{},
	}
	var err error
	s.sut, err = NewBestEffortUnit(mappers)
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
	mappers := map[TypeName]DataMapper{
		TypeNameOf(Bar{}): &mocks.DataMapper{},
	}
	var err error
	s.sut, err = NewBestEffortUnit(mappers)
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
	mappers := map[TypeName]DataMapper{
		TypeNameOf(Foo{}): &mocks.DataMapper{},
	}
	var err error
	s.sut, err = NewBestEffortUnit(mappers)
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
	mappers := map[TypeName]DataMapper{
		TypeNameOf(Foo{}): &mocks.DataMapper{},
	}
	var err error
	s.sut, err = NewBestEffortUnit(mappers)
	s.Require().NoError(err)

	// action.
	err = s.sut.Register(entities...)

	// assert.
	s.Require().Error(err)
	s.EqualError(err, ErrMissingDataMapper.Error())
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

func (s *BestEffortUnitTestSuite) AfterTest(suiteName, testName string) {
	for _, i := range s.mappers {
		i.AssertExpectations(s.T())
	}
}

func (s *BestEffortUnitTestSuite) TearDownTest() {
	s.sut = nil
}
