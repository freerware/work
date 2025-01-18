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
	"log/slog"
)

type StructuredLogger struct {
	l *slog.Logger
}

func NewStructuredLogger(logger *slog.Logger) *StructuredLogger {
	return &StructuredLogger{l: logger}
}

func (adapter *StructuredLogger) message(args ...any) (msg string, ok bool) {
	if len(args) == 0 {
		return
	}

	msg, ok = args[0].(string)
	return
}

func (adapter *StructuredLogger) Debug(args ...any) {
	if msg, ok := adapter.message(args...); ok {
		adapter.l.Debug(msg, args...)
	}
}

func (adapter *StructuredLogger) Info(args ...any) {
	if msg, ok := adapter.message(args...); ok {
		adapter.l.Info(msg, args...)
	}
}

func (adapter *StructuredLogger) Warn(args ...any) {
	if msg, ok := adapter.message(args...); ok {
		adapter.l.Warn(msg, args...)
	}
}

func (adapter *StructuredLogger) Error(args ...any) {
	if msg, ok := adapter.message(args...); ok {
		adapter.l.Error(msg, args...)
	}
}
