package work

import (
	"testing"

	"github.com/freerware/work/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type Foo struct {
	ID int
}

type Bar struct {
	ID string
}

type UnitTestSuite struct {
	suite.Suite

	// system under test.
	sut unit

	// mocks.
	inserters map[TypeName]*mocks.Inserter
	updaters  map[TypeName]*mocks.Updater
	deleters  map[TypeName]*mocks.Deleter
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) SetupTest() {

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
	s.sut = newUnit(params)
}

func (s *UnitTestSuite) TestUnit_Add_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Add(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Add_MissingInserter() {

	// arrange.
	s.sut = newUnit(UnitParameters{})
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}

	// action.
	err := s.sut.Add(entities...)

	// assert.
	s.Error(err)
}

func (s *UnitTestSuite) TestUnit_Add() {

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

func (s *UnitTestSuite) TestUnit_Alter_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Alter(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Alter_MissingUpdater() {

	// arrange.
	s.sut = newUnit(UnitParameters{})
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}

	// action.
	err := s.sut.Alter(entities...)

	// assert.
	s.Error(err)
}

func (s *UnitTestSuite) TestUnit_Alter() {

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

func (s *UnitTestSuite) TestUnit_Remove_Empty() {

	// arrange.
	entities := []interface{}{}

	// action.
	err := s.sut.Remove(entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Remove_MissingDeleter() {

	// arrange.
	s.sut = newUnit(UnitParameters{})
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}

	// action.
	err := s.sut.Remove(entities...)

	// assert.
	s.Error(err)
}

func (s *UnitTestSuite) TestUnit_Remove() {

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
