package work

import (
	"testing"

	"github.com/freerware/work/mocks"
	"github.com/stretchr/testify/suite"
)

type BestEffortUniterTestSuite struct {
	suite.Suite

	// system under test.
	sut Uniter

	// mocks.
	inserters map[TypeName]*mocks.Inserter
	updaters  map[TypeName]*mocks.Updater
	deleters  map[TypeName]*mocks.Deleter
}

func TestBestEffortUniterTestSuite(t *testing.T) {
	suite.Run(t, new(BestEffortUniterTestSuite))
}

func (s *BestEffortUniterTestSuite) SetupTest() {

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

	params := UnitParameters{
		Inserters: i,
		Updaters:  u,
		Deleters:  d,
	}
	s.sut = NewBestEffortUniter(params)
}

func (s *BestEffortUniterTestSuite) TestBestEffortUniter_Unit() {

	//action.
	_, err := s.sut.Unit()

	//assert.
	s.NoError(err)
}
