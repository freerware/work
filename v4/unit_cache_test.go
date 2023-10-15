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
	"context"
	"testing"

	"github.com/freerware/work/v4/internal/test"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally/v4"
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
	s.sut = UnitCache{cc: &memoryCacheClient{}, scope: tally.NoopScope}
}

func (s *UnitCacheTestSuite) TestUnitCache_Delete() {
	// arrange.
	ctx := context.Background()
	baz := test.Baz{Identifier: "1"}
	t := TypeNameOf(baz)

	// action.
	err := s.sut.delete(ctx, baz)

	// assert.
	s.NoError(err)
	cached, err := s.sut.Load(ctx, t, baz.ID())
	s.NoError(err)
	s.Nil(cached)
}

func (s *UnitCacheTestSuite) TestUnitCache_Load_Exists() {
	// arrange.
	ctx := context.Background()
	baz := test.Baz{Identifier: "1"}
	t := TypeNameOf(baz)
	s.sut.store(ctx, baz)

	// action.
	actual, err := s.sut.Load(ctx, t, baz.ID())

	// assert.
	s.Require().NoError(err)
	s.Equal(baz, actual)
}

func (s *UnitCacheTestSuite) TestUnitCache_Load_EntityNotExists() {
	// arrange.
	ctx := context.Background()
	baz := test.Baz{Identifier: "1"}
	t := TypeNameOf(baz)

	// action.
	actual, err := s.sut.Load(ctx, t, baz.ID())

	// assert.
	s.NoError(err)
	s.Nil(actual)
}

func (s *UnitCacheTestSuite) TestUnitCache_Load_TypeNotExists() {
	// arrange.
	ctx := context.Background()
	baz := test.Baz{Identifier: "1"}

	// action.
	actual, err := s.sut.Load(ctx, "main.Oops", baz.ID())

	// assert.
	s.NoError(err)
	s.Nil(actual)
}

func (s *UnitCacheTestSuite) TestUnitCache_Store_DifferentID() {
	// arrange.
	ctx := context.Background()
	baz := test.Baz{Identifier: "2"}
	bar := test.Bar{ID: "1"}
	tBaz := TypeNameOf(baz)
	tBar := TypeNameOf(bar)

	// action.
	errBaz := s.sut.store(ctx, baz)
	errBar := s.sut.store(ctx, bar)

	// assert.
	s.NoError(errBaz)
	actualBaz, err := s.sut.Load(ctx, tBaz, baz.ID())
	s.Require().NoError(err)
	s.Equal(baz, actualBaz)
	s.NoError(errBar)
	actualBar, err := s.sut.Load(ctx, tBar, bar.Identifier())
	s.Require().NoError(err)
	s.Equal(bar, actualBar)
}

func (s *UnitCacheTestSuite) TestUnitCache_Store_SameID() {
	// arrange.
	ctx := context.Background()
	baz := test.Baz{Identifier: "1"}
	bar := test.Bar{ID: "1"}
	tBaz := TypeNameOf(baz)
	tBar := TypeNameOf(bar)

	// action.
	errBaz := s.sut.store(ctx, baz)
	errBar := s.sut.store(ctx, bar)

	// assert.
	s.NoError(errBaz)
	actualBaz, err := s.sut.Load(ctx, tBaz, baz.ID())
	s.Require().NoError(err)
	s.Equal(baz, actualBaz)
	s.NoError(errBar)
	actualBar, err := s.sut.Load(ctx, tBar, bar.Identifier())
	s.Require().NoError(err)
	s.Equal(bar, actualBar)
}

func (s *UnitCacheTestSuite) TestUnitCache_Store_UncachableEntityError() {
	// arrange.
	ctx := context.Background()
	biz := test.Biz{Identifier: "1"}

	// action.
	err := s.sut.store(ctx, biz)

	// assert.
	s.Error(err)
	s.ErrorIs(err, ErrUncachableEntity)
}
