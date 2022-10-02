/* Copyright 2022 Freerware
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

package work

import (
	"testing"

	"github.com/freerware/work/v4/internal/test"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
)

type UnitCacheTestSuite struct {
	suite.Suite

	// system under test.
	sut UnitCache
}

func TestUnitCacheTestSuite(t *testing.T) {
	suite.Run(t, new(UnitCacheTestSuite))
}

func (s *UnitCacheTestSuite) SetupTest() {
	s.sut = UnitCache{scope: tally.NoopScope}
}

func (s *UnitCacheTestSuite) TestUnitCache_Delete() {
	// arrange.
	baz := test.Baz{Identifier: "1"}

	// action.
	s.sut.delete(baz)

	// assert.
	_, ok := s.sut.Load(TypeNameOf(baz), baz.ID())
	s.False(ok)
}

func (s *UnitCacheTestSuite) TestUnitCache_Load_Exists() {
	// arrange.
	baz := test.Baz{Identifier: "1"}
	s.sut.store(baz)

	// action.
	actual, ok := s.sut.Load(TypeNameOf(baz), baz.ID())

	// assert.
	s.True(ok)
	s.Equal(baz, actual)
}

func (s *UnitCacheTestSuite) TestUnitCache_Load_EntityNotExists() {
	// arrange.
	baz := test.Baz{Identifier: "1"}

	// action.
	_, ok := s.sut.Load(TypeNameOf(baz), baz.ID())

	// assert.
	s.False(ok)
}

func (s *UnitCacheTestSuite) TestUnitCache_Load_TypeNotExists() {
	// arrange.
	baz := test.Baz{Identifier: "1"}

	// action.
	_, ok := s.sut.Load("main.Oops", baz.ID())

	// assert.
	s.False(ok)
}

func (s *UnitCacheTestSuite) TestUnitCache_Store_DifferentID() {
	// arrange.
	baz := test.Baz{Identifier: "2"}
	bar := test.Bar{ID: "1"}

	// action.
	errBaz := s.sut.store(baz)
	errBar := s.sut.store(bar)

	// assert.
	s.NoError(errBaz)
	actualBaz, ok := s.sut.Load(TypeNameOf(baz), baz.ID())
	s.True(ok)
	s.Equal(baz, actualBaz)
	s.NoError(errBar)
	actualBar, ok := s.sut.Load(TypeNameOf(bar), bar.Identifier())
	s.True(ok)
	s.Equal(bar, actualBar)
}

func (s *UnitCacheTestSuite) TestUnitCache_Store_SameID() {
	// arrange.
	baz := test.Baz{Identifier: "1"}
	bar := test.Bar{ID: "1"}

	// action.
	errBaz := s.sut.store(baz)
	errBar := s.sut.store(bar)

	// assert.
	s.NoError(errBaz)
	actualBaz, ok := s.sut.Load(TypeNameOf(baz), baz.ID())
	s.True(ok)
	s.Equal(baz, actualBaz)
	s.NoError(errBar)
	actualBar, ok := s.sut.Load(TypeNameOf(bar), bar.Identifier())
	s.True(ok)
	s.Equal(bar, actualBar)
}

func (s *UnitCacheTestSuite) TestUnitCache_Store_UncachableEntityError() {
	// arrange.
	biz := test.Biz{Identifier: "1"}

	// action.
	err := s.sut.store(biz)

	// assert.
	s.Error(err)
	s.ErrorIs(err, ErrUncachableEntity)
}
