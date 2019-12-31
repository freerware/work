package work

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
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
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) SetupTest() {
	c := zap.NewDevelopmentConfig()
	c.DisableStacktrace = true
	l, _ := c.Build()
	m := tally.NewTestScope("test", map[string]string{})
	options := UnitOptions{
		Logger: l,
		Scope:  m,
	}
	s.sut = newUnit(options)
}

func (s *UnitTestSuite) TestUnit_Add_Empty() {

	// arrange.
	entities := []interface{}{}
	c := func(t TypeName) bool { return true }

	// action.
	err := s.sut.add(c, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Add_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
	}
	c := func(t TypeName) bool { return false }

	// action.
	err := s.sut.add(c, entities...)

	// assert.
	s.Require().Error(err)
	s.EqualError(err, ErrMissingDataMapper.Error())
}

func (s *UnitTestSuite) TestUnit_Add() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	c := func(t TypeName) bool { return true }

	// action.
	err := s.sut.add(c, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Alter_Empty() {

	// arrange.
	entities := []interface{}{}
	c := func(t TypeName) bool { return false }

	// action.
	err := s.sut.alter(c, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Alter_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
	}
	c := func(t TypeName) bool { return false }

	// action.
	err := s.sut.alter(c, entities...)

	// assert.
	s.Require().Error(err)
	s.EqualError(err, ErrMissingDataMapper.Error())
}

func (s *UnitTestSuite) TestUnit_Alter() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	c := func(t TypeName) bool { return true }

	// action.
	err := s.sut.alter(c, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Remove_Empty() {

	// arrange.
	entities := []interface{}{}
	c := func(t TypeName) bool { return true }

	// action.
	err := s.sut.remove(c, entities...)

	// assert.
	s.NoError(err)
}

func (s *UnitTestSuite) TestUnit_Remove_MissingDataMapper() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	c := func(t TypeName) bool { return false }

	// action.
	err := s.sut.remove(c, entities...)

	// assert.
	s.Require().Error(err)
	s.EqualError(err, ErrMissingDataMapper.Error())
}

func (s *UnitTestSuite) TestUnit_Remove() {

	// arrange.
	entities := []interface{}{
		Foo{ID: 28},
		Bar{ID: "28"},
	}
	c := func(t TypeName) bool { return true }

	// action.
	err := s.sut.remove(c, entities...)

	// assert.
	s.NoError(err)
}
