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
	"errors"
	"sync"

	"github.com/uber-go/tally"
)

// UnitCache represents the cache that the work unit manipulates as a result
// of entity registration.
type UnitCache struct {
	m sync.Map

	scope tally.Scope
}

var (
	// ErrUncachableEntity represents the error that is returned when an attempt
	// to cache an entity with in unresolvable ID occurs.
	ErrUncachableEntity = errors.New("unable to cache entity - does not implement supported interfaces")
)

// Delete removes an entity from the work unit cache.
func (uc *UnitCache) delete(entity interface{}) {
	t := TypeNameOf(entity)
	if id, ok := id(entity); ok {
		if entitiesByID, ok := uc.m.Load(t); ok {
			if entityMap, ok := entitiesByID.(*sync.Map); ok {
				entityMap.Delete(id)
				uc.scope.Counter(cacheInvalidate).Inc(1)
			}
		}
	}
}

// Store places the provided entity in the work unit cache.
func (uc *UnitCache) store(entity interface{}) (err error) {
	id, ok := id(entity)
	if !ok {
		err = ErrUncachableEntity
		return
	}
	t := TypeNameOf(entity)
	if cached, ok := uc.m.Load(t); !ok {
		entitiesByID := &sync.Map{}
		entitiesByID.Store(id, entity)
		uc.m.Store(t, entitiesByID)
		uc.scope.Counter(cacheInsert).Inc(1)
		return
	} else {
		if entityMap, ok := cached.(*sync.Map); ok {
			entityMap.Store(id, entity)
			uc.scope.Counter(cacheInsert).Inc(1)
		}
		return
	}
}

// Load retrieves the entity with the provided type name and ID from the work
// unit cache.
func (uc *UnitCache) Load(t TypeName, id interface{}) (entity interface{}, loaded bool) {
	if entitiesByID, ok := uc.m.Load(t); ok {
		if entityMap, ok := entitiesByID.(*sync.Map); ok {
			entity, loaded = entityMap.Load(id)
		}
	}
	return
}
