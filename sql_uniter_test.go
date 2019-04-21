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
	db        *sql.DB
	_db       sqlmock.Sqlmock
	inserters map[TypeName]*mocks.Inserter
	updaters  map[TypeName]*mocks.Updater
	deleters  map[TypeName]*mocks.Deleter
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
	s.inserters = make(map[TypeName]*mocks.Inserter)
	s.inserters[fooTypeName] = &mocks.Inserter{}
	s.inserters[barTypeName] = &mocks.Inserter{}
	s.updaters = make(map[TypeName]*mocks.Updater)
	s.updaters[fooTypeName] = &mocks.Updater{}
	s.updaters[barTypeName] = &mocks.Updater{}
	s.deleters = make(map[TypeName]*mocks.Deleter)
	s.deleters[fooTypeName] = &mocks.Deleter{}
	s.deleters[barTypeName] = &mocks.Deleter{}

	var err error
	s.db, s._db, err = sqlmock.New()
	s.Require().NoError(err)

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

	params := SQLUnitParameters{
		UnitParameters: UnitParameters{
			Inserters: i,
			Updaters:  u,
			Deleters:  d,
		},
		ConnectionPool: s.db,
	}
	s.sut = NewSQLUniter(params)
}

func (s *SQLUniterTestSuite) TestSQLUniter_Unit() {

	//action.
	_, err := s.sut.Unit()

	//assert.
	s.NoError(err)
}

func (s *SQLUniterTestSuite) TestSQLUniter_UnitError() {

	//arrange.
	params := SQLUnitParameters{
		ConnectionPool: nil,
	}
	s.sut = NewSQLUniter(params)

	//action.
	_, err := s.sut.Unit()

	//assert.
	s.Error(err)
}
