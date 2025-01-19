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

// UnitLogger represents a type responsible for performing logging behaviors.
type UnitLogger interface {
	// Debug logs the provided arguments as a 'debug' level message.
	Debug(...any)

	// Info logs the provided arguments as a 'info' level message.
	Info(...any)

	// Warn logs the provided arguments as a 'warn' level message.
	Warn(...any)

	// Error logs the provided arguments as an 'error' level message.
	Error(...any)
}
