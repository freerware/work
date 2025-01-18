/* Copyright 2025 Freerware
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
	"errors"
	"fmt"
	"sync"

	"github.com/uber-go/tally/v4"
)

type memoryCacheClient struct {
	m sync.Map
}

func (mcc *memoryCacheClient) Delete(ctx context.Context, key string) (err error) {
	mcc.m.Delete(key)
	return
}

func (mcc *memoryCacheClient) Get(ctx context.Context, key string) (entry interface{}, err error) {
	entry, _ = mcc.m.Load(key)
	return
}

func (mcc *memoryCacheClient) Set(ctx context.Context, key string, entry interface{}) (err error) {
	mcc.m.Store(key, entry)
	return
}

// UnitCacheClient represents a client for a cache provider.
type UnitCacheClient interface {
	Get(context.Context, string) (interface{}, error)
	Set(context.Context, string, interface{}) error
	Delete(context.Context, string) error
}

// UnitCache represents the cache that the work unit manipulates as a result
// of entity registration.
type UnitCache struct {
	cc UnitCacheClient

	scope tally.Scope
}

var (
	// ErrUncachableEntity represents the error that is returned when an attempt
	// to cache an entity with an unresolvable ID occurs.
	ErrUncachableEntity = errors.New("unable to cache entity - does not implement supported interfaces")
)

func cacheKey(t TypeName, id interface{}) string {
	return fmt.Sprintf("%s-%v", string(t), id)
}

// Delete removes an entity from the work unit cache.
func (uc *UnitCache) delete(ctx context.Context, entity interface{}) (err error) {
	t := TypeNameOf(entity)
	if id, ok := id(entity); ok {
		if err = uc.cc.Delete(ctx, cacheKey(t, id)); err == nil {
			uc.scope.Counter(cacheDelete).Inc(1)
		}
	}
	return
}

// Store places the provided entity in the work unit cache.
func (uc *UnitCache) store(ctx context.Context, entity interface{}) (err error) {
	id, ok := id(entity)
	if !ok {
		return ErrUncachableEntity
	}
	t := TypeNameOf(entity)
	if err = uc.cc.Set(ctx, cacheKey(t, id), entity); err == nil {
		uc.scope.Counter(cacheInsert).Inc(1)
	}
	return
}

// Load retrieves the entity with the provided type name and ID from the work
// unit cache.
func (uc *UnitCache) Load(ctx context.Context, t TypeName, id interface{}) (entity interface{}, err error) {
	return uc.cc.Get(ctx, cacheKey(t, id))
}
