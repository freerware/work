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
	mappers map[TypeName]*mocks.DataMapper
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
	s.mappers = make(map[TypeName]*mocks.DataMapper)
	s.mappers[fooTypeName] = &mocks.DataMapper{}
	s.mappers[barTypeName] = &mocks.DataMapper{}

	// construct SUT.
	dm := make(map[TypeName]DataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}

	s.sut = NewBestEffortUniter(dm)
}

func (s *BestEffortUniterTestSuite) TestBestEffortUniter_Unit() {

	//action.
	_, err := s.sut.Unit()

	//assert.
	s.NoError(err)
}
