package work

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freerware/work/mocks"
	"github.com/stretchr/testify/suite"
)

type SQLUniterTestSuite struct {
	suite.Suite

	// system under test.
	sut Uniter

	// mocks.
	db      *sql.DB
	_db     sqlmock.Sqlmock
	mappers map[TypeName]*mocks.SQLDataMapper
}

func TestSQLUniterTestSuite(t *testing.T) {
	suite.Run(t, new(SQLUniterTestSuite))
}

func (s *SQLUniterTestSuite) SetupTest() {

	// test entities.
	foo := Foo{ID: 28}
	fooTypeName := TypeNameOf(foo)
	bar := Bar{ID: "28"}
	barTypeName := TypeNameOf(bar)

	// initialize mocks.
	s.mappers = make(map[TypeName]*mocks.SQLDataMapper)
	s.mappers[fooTypeName] = &mocks.SQLDataMapper{}
	s.mappers[barTypeName] = &mocks.SQLDataMapper{}

	var err error
	s.db, s._db, err = sqlmock.New()
	s.Require().NoError(err)

	// construct SUT.
	dm := make(map[TypeName]SQLDataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}
	s.sut = NewSQLUniter(dm, s.db)
}

func (s *SQLUniterTestSuite) TestSQLUniter_Unit() {

	//action.
	_, err := s.sut.Unit()

	//assert.
	s.NoError(err)
}

func (s *SQLUniterTestSuite) TestSQLUniter_UnitError() {

	//arrange.
	s.sut = NewSQLUniter(map[TypeName]SQLDataMapper{}, nil)

	//action.
	_, err := s.sut.Unit()

	//assert.
	s.Error(err)
}
