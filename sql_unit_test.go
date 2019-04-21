package work

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freerware/work-it/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type SQLUnitTestSuite struct {
	suite.Suite

	// system under test.
	sut Unit

	// mocks.
	db        *sql.DB
	_db       sqlmock.Sqlmock
	inserters map[TypeName]*mocks.Inserter
	updaters  map[TypeName]*mocks.Updater
	deleters  map[TypeName]*mocks.Deleter
}

func TestSQLUnitTestSuite(t *testing.T) {
	suite.Run(t, new(SQLUnitTestSuite))
}

func (s *SQLUnitTestSuite) SetupTest() {

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

	l, _ := zap.NewDevelopment()
	params := SQLUnitParameters{
		UnitParameters: UnitParameters{
			Inserters: i,
			Updaters:  u,
			Deleters:  d,
			Logger:    l,
		},
		ConnectionPool: s.db,
	}
	s.sut, err = NewSQLUnit(params)
	s.Require().NoError(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_TransactionBeginError() {

	// arrange.
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
	alterError := s.sut.Remove(updatedEntities...)
	removeError := s.sut.Remove(removedEntities...)
	s._db.ExpectBegin().WillReturnError(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_InserterError() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback()
	s.inserters[fooType].On(
		"Insert",
		addedEntities[0],
	).Return(errors.New("whoa"))
	s.inserters[barType].On(
		"Insert",
		addedEntities[1],
	).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_InserterAndRollbackError() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
	s.inserters[fooType].On(
		"Insert",
		addedEntities[0],
	).Return(errors.New("whoa"))
	s.inserters[barType].On(
		"Insert",
		addedEntities[1],
	).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
}
func (s *SQLUnitTestSuite) TestSQLUnit_Save_UpdaterError() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback()
	s.inserters[fooType].On(
		"Insert",
		addedEntities[0],
	).Return(nil)
	s.inserters[barType].On(
		"Insert",
		addedEntities[1],
	).Return(nil)
	s.updaters[fooType].On(
		"Update",
		updatedEntities[0],
	).Return(errors.New("whoa"))
	s.updaters[barType].On(
		"Update",
		updatedEntities[1],
	).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_UpdaterAndRollbackError() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
	s.inserters[fooType].On(
		"Insert",
		addedEntities[0],
	).Return(nil)
	s.inserters[barType].On(
		"Insert",
		addedEntities[1],
	).Return(nil)
	s.updaters[fooType].On(
		"Update",
		updatedEntities[0],
	).Return(errors.New("whoa"))
	s.updaters[barType].On(
		"Update",
		updatedEntities[1],
	).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_DeleterError() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback()
	s.inserters[fooType].On(
		"Insert",
		addedEntities[0],
	).Return(nil)
	s.inserters[barType].On(
		"Insert",
		addedEntities[1],
	).Return(nil)
	s.updaters[fooType].On(
		"Update",
		updatedEntities[0],
	).Return(nil)
	s.updaters[barType].On(
		"Update",
		updatedEntities[1],
	).Return(nil)
	s.deleters[fooType].On(
		"Delete",
		removedEntities[0],
	).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_DeleterAndRollbackError() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
	s.inserters[fooType].On(
		"Insert",
		addedEntities[0],
	).Return(nil)
	s.inserters[barType].On(
		"Insert",
		addedEntities[1],
	).Return(nil)
	s.updaters[fooType].On(
		"Update",
		updatedEntities[0],
	).Return(nil)
	s.updaters[barType].On(
		"Update",
		updatedEntities[1],
	).Return(nil)
	s.deleters[fooType].On(
		"Delete",
		removedEntities[0],
	).Return(errors.New("whoa"))

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_Panic() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback().WillReturnError(errors.New("whoa"))
	s.inserters[fooType].On(
		"Insert",
		addedEntities[0],
	).Return(nil)
	s.inserters[barType].On(
		"Insert",
		addedEntities[1],
	).Return(nil)
	s.updaters[fooType].On(
		"Update",
		updatedEntities[0],
	).Return(nil)
	s.updaters[barType].On(
		"Update",
		updatedEntities[1],
	).Return(nil)
	s.deleters[fooType].On(
		"Delete",
		removedEntities[0],
	).Return().Run(func(args mock.Arguments) { panic("whoa") })

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_PanicAndRollbackError() {

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
	s._db.ExpectBegin()
	s._db.ExpectRollback()
	s.inserters[fooType].On(
		"Insert",
		addedEntities[0],
	).Return(nil)
	s.inserters[barType].On(
		"Insert",
		addedEntities[1],
	).Return(nil)
	s.updaters[fooType].On(
		"Update",
		updatedEntities[0],
	).Return(nil)
	s.updaters[barType].On(
		"Update",
		updatedEntities[1],
	).Return(nil)
	s.deleters[fooType].On(
		"Delete",
		removedEntities[0],
	).Return().Run(func(args mock.Arguments) { panic("whoa") })

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save_CommitError() {

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
	s._db.ExpectBegin()
	s._db.ExpectCommit().WillReturnError(errors.New("whoa"))
	s.inserters[fooType].On(
		"Insert",
		addedEntities[0],
	).Return(nil)
	s.inserters[barType].On(
		"Insert",
		addedEntities[1],
	).Return(nil)
	s.updaters[fooType].On(
		"Update",
		updatedEntities[0],
	).Return(nil)
	s.updaters[barType].On(
		"Update",
		updatedEntities[1],
	).Return(nil)
	s.deleters[fooType].On(
		"Delete",
		removedEntities[0],
	).Return(nil)

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.Error(err)
}

func (s *SQLUnitTestSuite) TestSQLUnit_Save() {

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
	s._db.ExpectBegin()
	s._db.ExpectCommit()
	s.inserters[fooType].On(
		"Insert",
		addedEntities[0],
	).Return(nil)
	s.inserters[barType].On(
		"Insert",
		addedEntities[1],
	).Return(nil)
	s.updaters[fooType].On(
		"Update",
		updatedEntities[0],
	).Return(nil)
	s.updaters[barType].On(
		"Update",
		updatedEntities[1],
	).Return(nil)
	s.deleters[fooType].On(
		"Delete",
		removedEntities[0],
	).Return(nil)

	// action.
	err := s.sut.Save()

	// assert.
	s.Require().NoError(addError)
	s.Require().NoError(alterError)
	s.Require().NoError(removeError)
	s.Require().NoError(s._db.ExpectationsWereMet())
	s.NoError(err)
}

func (s *SQLUnitTestSuite) TearDownTest() {
	s.db.Close()
	s.sut = nil
}
