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

import "context"

type Foo struct {
	ID int
}

type Bar struct {
	ID string
}

type TableDrivenTest struct {
	name         string
	registers    []interface{}
	additions    []interface{}
	alters       []interface{}
	removals     []interface{}
	expectations func(ctx context.Context, registers, additions, alters, removals []interface{})
	ctx          context.Context
	err          error
	assertions   func()
	panics       bool
}
