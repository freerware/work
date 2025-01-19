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

package adapters

// NopLogger represents an adapter for a no-op logger.
type NopLogger struct {
}

// NewNopLogger creates a no-op logger adapter that does nothing.
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

// Debug does nothing.
func (adapter *NopLogger) Debug(args ...any) {}

// Info does nothing.
func (adapter *NopLogger) Info(args ...any) {}

// Warn does nothing.
func (adapter *NopLogger) Warn(args ...any) {}

// Error does nothing.
func (adapter *NopLogger) Error(args ...any) {}
