/* Copyright 2021 Freerware
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package work_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freerware/work/v4"
	"github.com/freerware/work/v4/internal/mock"
	"github.com/stretchr/testify/suite"
)

type UniterTestSuite struct {
	suite.Suite

	// system under test.
	sut work.Uniter

	// mocks.
	db      *sql.DB
	_db     sqlmock.Sqlmock
	mappers map[work.TypeName]*mock.DataMapper
}

func TestUniterTestSuite(t *testing.T) {
	suite.Run(t, new(UniterTestSuite))
}

func (s *UniterTestSuite) SetupTest() {

	// test entities.
	foo := Foo{id: 28}
	fooTypeName := work.TypeNameOf(foo)
	bar := Bar{id: "28"}
	barTypeName := work.TypeNameOf(bar)

	// initialize mocks.
	s.mappers = make(map[work.TypeName]*mock.DataMapper)
	s.mappers[fooTypeName] = &mock.DataMapper{}
	s.mappers[barTypeName] = &mock.DataMapper{}

	var err error
	s.db, s._db, err = sqlmock.New()
	s.Require().NoError(err)

	// construct SUT.
	dm := make(map[work.TypeName]work.DataMapper)
	for t, m := range s.mappers {
		dm[t] = m
	}
	s.sut = work.NewUniter(work.UnitDataMappers(dm), work.UnitDB(s.db))
}

func (s *UniterTestSuite) TestUniter() {
	// test cases.
	tests := []struct {
		name string
		err  error
	}{
		{name: "Unit", err: nil},
		{name: "UnitError", err: nil},
	}
	// execute test cases.
	for _, test := range tests {
		s.Run(test.name, func() {
			// action.
			_, err := s.sut.Unit()

			// assert.
			if test.err != nil {
				s.Require().Error(err)
				s.Require().EqualError(err, test.err.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *UniterTestSuite) TearDownTest() {
	s.sut = nil
	s.mappers = nil
	s.db = nil
}
