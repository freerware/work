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

package work

import "database/sql"

type sqlUniter struct {
	mappers map[TypeName]SQLDataMapper
	db      *sql.DB
	options []Option
}

// NewSQLUniter constructs a new SQL work unit factory.
func NewSQLUniter(
	mappers map[TypeName]SQLDataMapper,
	db *sql.DB,
	options ...Option,
) Uniter {
	return &sqlUniter{
		mappers: mappers,
		db:      db,
		options: options,
	}
}

// Unit constructs a new SQL work unit.
func (u *sqlUniter) Unit() (Unit, error) {
	return NewSQLUnit(u.mappers, u.db, u.options...)
}
