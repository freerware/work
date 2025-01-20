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

// Debug logs the provided message with arguments as a 'debug' level message.
func (adapter *ZapLogger) Debug(msg string, args ...any) {
	adapter.l.Sugar().Debugw(msg, args...)
}

// Info logs the provided message with arguments as a 'info' level message.
func (adapter *ZapLogger) Info(msg string, args ...any) {
	adapter.l.Sugar().Infow(msg, args...)
}

// Warn logs the provided message with arguments as a 'warn' level message.
func (adapter *ZapLogger) Warn(msg string, args ...any) {
	adapter.l.Sugar().Warnw(msg, args...)
}

// Error logs the provided message with arguments as an 'error' level message.
func (adapter *ZapLogger) Error(msg string, args ...any) {
	adapter.l.Sugar().Errorw(msg, args...)
}
