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

package test

type Foo struct {
	ID int
}

func (f Foo) Identifier() interface{} { return f.ID }

type Bar struct {
	ID string
}

func (b Bar) Identifier() interface{} { return b.ID }

type Baz struct {
	Identifier string
}

func (b Baz) ID() interface{} { return b.Identifier }

type Biz struct {
	Identifier string
}
