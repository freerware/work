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
	"testing"

	"github.com/freerware/work"
	"github.com/freerware/work/internal/mock"
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
