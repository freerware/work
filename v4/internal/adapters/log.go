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

import (
	"log"
)

// StandardLogger represents an adapter for the standard logger.
type StandardLogger struct {
	l *log.Logger
}

// NewStandardLogger creates a standard logger adapter for the provided logger.
func NewStandardLogger(logger *log.Logger) *StandardLogger {
	return &StandardLogger{l: logger}
}

// Debug logs the provided arguments as a 'debug' level message.
func (adapter *StandardLogger) Debug(args ...any) {
	adapter.l.Println(args...)
}

// Info logs the provided arguments as a 'info' level message.
func (adapter *StandardLogger) Info(args ...any) {
	adapter.l.Println(args...)
}

// Warn logs the provided arguments as a 'warn' level message.
func (adapter *StandardLogger) Warn(args ...any) {
	adapter.l.Println(args...)
}

// Error logs the provided arguments as an 'error' level message.
func (adapter *StandardLogger) Error(args ...any) {
	adapter.l.Println(args...)
}
