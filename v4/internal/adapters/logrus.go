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
	"github.com/sirupsen/logrus"
)

// LogrusLogger represents an adapter for the Logrus logger.
type LogrusLogger struct {
	l *logrus.Logger
}

// NewLogrusLogger creates a Logrus logger adapter for the provided logger.
func NewLogrusLogger(logger *logrus.Logger) *LogrusLogger {
	return &LogrusLogger{l: logger}
}

// Debug logs the provided message with arguments as a 'debug' level message.
func (adapter *LogrusLogger) Debug(msg string, args ...any) {
	adapter.l.Debug(append([]any{msg}, args...))
}

// Info logs the provided message with arguments as a 'info' level message.
func (adapter *LogrusLogger) Info(msg string, args ...any) {
	adapter.l.Info(append([]any{msg}, args...))
}

// Warn logs the provided message with arguments as a 'warn' level message.
func (adapter *LogrusLogger) Warn(msg string, args ...any) {
	adapter.l.Warn(append([]any{msg}, args...))
}

// Error logs the provided message with arguments as an 'error' level message.
func (adapter *LogrusLogger) Error(msg string, args ...any) {
	adapter.l.Error(append([]any{msg}, args...))
}
