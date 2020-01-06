package work_test

import (
	"testing"

	"github.com/freerware/work/v3"
	"github.com/freerware/work/v3/internal/mock"
	"github.com/stretchr/testify/suite"
)

type BestEffortUniterTestSuite struct {
	suite.Suite

	// system under test.
	sut work.Uniter

	// mocks.
	mappers map[work.TypeName]*mock.DataMapper
}

func TestBestEffortUniterTestSuite(t *testing.T) {
	suite.Run(t, new(BestEffortUniterTestSuite))
}

func (s *BestEffortUniterTestSuite) SetupTest() {

	// test entities.
	foo := Foo{ID: 28}
	fooTypeName := work.TypeNameOf(foo)
	bar := Bar{ID: "28"}
	barTypeName := work.TypeNameOf(bar)

	// initialize mocks.
	s.mappers = make(map[work.TypeName]*mock.DataMapper)
	s.mappers[fooTypeName] = &mock.DataMapper{}
	s.mappers[barTypeName] = &mock.DataMapper{}

	// construct SUT.
	dm := make(map[work.TypeName]work.DataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}

	s.sut = work.NewBestEffortUniter(dm)
}

func (s *BestEffortUniterTestSuite) TestBestEffortUniter_Unit() {

	//action.
	_, err := s.sut.Unit()

	//assert.
	s.NoError(err)
}
