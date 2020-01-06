package work_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freerware/work/v3"
	"github.com/freerware/work/v3/internal/mock"
	"github.com/stretchr/testify/suite"
)

type SQLUniterTestSuite struct {
	suite.Suite

	// system under test.
	sut work.Uniter

	// mocks.
	db      *sql.DB
	_db     sqlmock.Sqlmock
	mappers map[work.TypeName]*mock.SQLDataMapper
}

func TestSQLUniterTestSuite(t *testing.T) {
	suite.Run(t, new(SQLUniterTestSuite))
}

func (s *SQLUniterTestSuite) SetupTest() {

	// test entities.
	foo := Foo{ID: 28}
	fooTypeName := work.TypeNameOf(foo)
	bar := Bar{ID: "28"}
	barTypeName := work.TypeNameOf(bar)

	// initialize mocks.
	s.mappers = make(map[work.TypeName]*mock.SQLDataMapper)
	s.mappers[fooTypeName] = &mock.SQLDataMapper{}
	s.mappers[barTypeName] = &mock.SQLDataMapper{}

	var err error
	s.db, s._db, err = sqlmock.New()
	s.Require().NoError(err)

	// construct SUT.
	dm := make(map[work.TypeName]work.SQLDataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}
	s.sut = work.NewSQLUniter(dm, s.db)
}

func (s *SQLUniterTestSuite) TestSQLUniter_Unit() {

	//action.
	_, err := s.sut.Unit()

	//assert.
	s.NoError(err)
}

func (s *SQLUniterTestSuite) TestSQLUniter_UnitError() {

	//arrange.
	s.sut = work.NewSQLUniter(map[work.TypeName]work.SQLDataMapper{}, nil)

	//action.
	_, err := s.sut.Unit()

	//assert.
	s.Error(err)
}
