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
	"go.uber.org/zap"
)

// ZapLogger represents an adapter for the Zap logger.
type ZapLogger struct {
	l *zap.Logger
}

// NewZapLogger creates a Zap logger adapter for the provided logger.
func NewZapLogger(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{l: logger}
}

// message extracts the message from the provided arguments.
func (adapter *ZapLogger) message(args ...any) (msg string, ok bool) {
	if len(args) == 0 {
		return
	}

	msg, ok = args[0].(string)
	return
}

// Debug logs the provided arguments as a 'debug' level message.
func (adapter *ZapLogger) Debug(args ...any) {
	if msg, ok := adapter.message(args...); ok {
		adapter.l.Sugar().Debugw(msg, args[1:]...)
	}
}

// Info logs the provided arguments as a 'info' level message.
func (adapter *ZapLogger) Info(args ...any) {
	if msg, ok := adapter.message(args...); ok {
		adapter.l.Sugar().Infow(msg, args[1:]...)
	}
}

// Warn logs the provided arguments as a 'warn' level message.
func (adapter *ZapLogger) Warn(args ...any) {
	if msg, ok := adapter.message(args...); ok {
		adapter.l.Sugar().Warnw(msg, args[1:]...)
	}
}

// Error logs the provided arguments as an 'error' level message.
func (adapter *ZapLogger) Error(args ...any) {
	if msg, ok := adapter.message(args...); ok {
		adapter.l.Sugar().Errorw(msg, args[1:]...)
	}
}
