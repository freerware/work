package work

import (
	"errors"
	"testing"

	"github.com/freerware/work/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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
}

func TestBestEffortUnitTestSuite(t *testing.T) {
	suite.Run(t, new(BestEffortUnitTestSuite))
}

func (s *BestEffortUnitTestSuite) SetupTest() {

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

	l, _ := zap.NewDevelopment()
	params := UnitParameters{
		Inserters: i,
		Updaters:  u,
		Deleters:  d,
		Logger:    l,
	}
	s.sut = NewBestEffortUnit(params)
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_InserterError() {

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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(err)
	s.inserters[barType].On("Insert", addedEntities[1]).Return(err)

	// arrange - rollback invocations.
	s.inserters[fooType].On("Insert", removedEntities[0]).Return(nil)
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
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_InserterAndRollbackError() {

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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.inserters[fooType].On("Insert", removedEntities[0]).Return(nil)
	s.deleters[fooType].On("Delete", addedEntities[0]).Return(nil)
	s.deleters[barType].On("Delete", addedEntities[1]).Return(nil)
	s.updaters[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.updaters[barType].On(
		"Update", registeredEntities[1]).Return(err)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Error(err)
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
	s.inserters[barType].On("Insert", addedEntities[1]).Return(nil)
	s.updaters[barType].On("Update", updatedEntities[1]).Return(err)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.inserters[fooType].On("Insert", removedEntities[0]).Return(nil)
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
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_UpdaterAndRollbackError() {

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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.inserters[fooType].On("Insert", removedEntities[0]).Return(nil)
	s.deleters[fooType].On("Delete", addedEntities[0]).Return(nil)
	s.deleters[barType].On("Delete", addedEntities[1]).Return(err)
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
	s.inserters[fooType].On("Insert", removedEntities[0]).Return(nil)
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
}

func (s *BestEffortUnitTestSuite) TestBestEffortUnit_Save_DeleterAndRollbackError() {

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
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.updaters[barType].On("Update", updatedEntities[1]).Return(nil)
	s.deleters[fooType].On("Delete", removedEntities[0]).Return(err)

	// arrange - rollback invocations.
	s.inserters[fooType].On("Insert", removedEntities[0]).Return(nil)
	s.deleters[fooType].On("Delete", addedEntities[0]).Return(err)
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
	err := errors.New("whoa")
	s.inserters[fooType].On("Insert", addedEntities[0]).Return(nil)
	s.updaters[fooType].On("Update", updatedEntities[0]).Return(nil)
	s.updaters[barType].On("Update", updatedEntities[1]).Return(nil)
	s.deleters[fooType].
		On("Delete", removedEntities[0]).
		Return().Run(func(args mock.Arguments) { panic("whoa") })

	// arrange - rollback invocations.
	s.inserters[fooType].On("Insert", removedEntities[0]).Return(nil)
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
	s.deleters[fooType].
		On("Delete", removedEntities[0]).
		Return().Run(func(args mock.Arguments) { panic("whoa") })

	// arrange - rollback invocations.
	s.inserters[fooType].On("Insert", removedEntities[0]).Return(nil)
	s.deleters[fooType].On("Delete", addedEntities[0]).Return(nil)
	s.deleters[barType].On("Delete", addedEntities[1]).Return(nil)
	s.updaters[fooType].On(
		"Update", registeredEntities[0], registeredEntities[2]).Return(nil)
	s.updaters[barType].On(
		"Update", registeredEntities[1]).Return(err)

	// action.
	err = s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Error(err)
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
}

func (s *BestEffortUnitTestSuite) TearDownTest() {
	s.sut = nil
}
