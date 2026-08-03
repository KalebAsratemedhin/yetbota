package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	// used by options
	writers     []io.Writer
	maskEnabled bool
	noopLogger  bool
	closer      []io.Closer

	// initiated by this application newLogger
	zapLogger *zap.Logger
	level     zapcore.Level
}

var _ Logger = (*logger)(nil)

var (
	defaultLogger *logger
	once          sync.Once
)

func InitDefault(opt ...Option) {
	once.Do(func() {
		defaultLogger = NewLogger(opt...).(*logger)
	})
}

func Default() Logger {
	if defaultLogger == nil {
		return NewLogger()
	}
	return defaultLogger
}

func NewLogger(opts ...Option) Logger {
	logger := &logger{
		writers:     make([]io.Writer, 0),
		maskEnabled: true,
		noopLogger:  false,
		closer:      make([]io.Closer, 0),
		level:       zapcore.InfoLevel,
	}
	if defaultLogger != nil {
		logger.writers = defaultLogger.writers
		logger.maskEnabled = defaultLogger.maskEnabled
		logger.noopLogger = defaultLogger.noopLogger
		logger.closer = defaultLogger.closer
		logger.level = defaultLogger.level
	}

	for _, o := range opts {
		o(logger)
	}

	// set logger here instead in options to make easy and consistent initiation
	// set multiple writer as already set in options
	logger.zapLogger = newZapLogger(logger.level, logger.writers...)

	// use stdout only when writer is not specified
	if len(logger.writers) <= 0 {
		logger.zapLogger = newZapLogger(logger.level, zapcore.AddSync(os.Stdout))
	}

	// if noop logger enabled, then use discard all print
	if logger.noopLogger {
		logger.zapLogger = zap.NewNop()
	}

	return logger
}

func (d *logger) Close() error {
	if d.closer == nil {
		return nil
	}

	var err error
	for _, closer := range d.closer {
		if closer == nil {
			continue
		}

		if e := closer.Close(); e != nil {
			err = fmt.Errorf("%w: %q", e, err)
		}
	}

	return err
}

func (d *logger) Debug(ctx context.Context, message string, fields ...Field) {
	zapLogs := []zap.Field{
		zap.String("logType", LogTypeSYS),
		zap.String("level", "debug"),
	}

	zapLogs = append(zapLogs, formatLogs(ctx, message, d.maskEnabled, fields...)...)
	d.zapLogger.Debug(separator, zapLogs...)
}

func (d *logger) Info(ctx context.Context, message string, fields ...Field) {
	zapLogs := []zap.Field{
		zap.String("logType", LogTypeSYS),
		zap.String("level", "info"),
	}

	zapLogs = append(zapLogs, formatLogs(ctx, message, d.maskEnabled, fields...)...)
	d.zapLogger.Info(separator, zapLogs...)
}

func (d *logger) Warn(ctx context.Context, message string, fields ...Field) {
	zapLogs := []zap.Field{
		zap.String("logType", LogTypeSYS),
		zap.String("level", "warn"),
	}

	zapLogs = append(zapLogs, formatLogs(ctx, message, d.maskEnabled, fields...)...)
	d.zapLogger.Warn(separator, zapLogs...)
}

func (d *logger) Error(ctx context.Context, message string, fields ...Field) {
	zapLogs := []zap.Field{
		zap.String("logType", LogTypeSYS),
		zap.String("level", "error"),
	}

	zapLogs = append(zapLogs, formatLogs(ctx, message, d.maskEnabled, fields...)...)
	d.zapLogger.Error(separator, zapLogs...)
}

func (d *logger) Fatal(ctx context.Context, message string, fields ...Field) {
	zapLogs := []zap.Field{
		zap.String("logType", LogTypeSYS),
		zap.String("level", "fatal"),
	}

	zapLogs = append(zapLogs, formatLogs(ctx, message, d.maskEnabled, fields...)...)
	d.zapLogger.Fatal(separator, zapLogs...)
}

func (d *logger) Panic(ctx context.Context, message string, fields ...Field) {
	zapLogs := []zap.Field{
		zap.String("logType", LogTypeSYS),
		zap.String("level", "panic"),
	}

	zapLogs = append(zapLogs, formatLogs(ctx, message, d.maskEnabled, fields...)...)
	d.zapLogger.Panic(separator, zapLogs...)
}
