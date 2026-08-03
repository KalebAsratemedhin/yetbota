package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap/zapcore"
)

type Option func(*logger)

func OptNoop() Option {
	return func(logger *logger) {
		logger.noopLogger = true
	}
}

func MaskEnabled() Option {
	return func(logger *logger) {
		logger.maskEnabled = true
	}
}

func WithStdout() Option {
	return func(logger *logger) {
		// Wire STD output for both type
		logger.writers = append(logger.writers, os.Stdout)
	}
}

type WrapKafkaWriter struct {
	topic    string
	producer sarama.SyncProducer
}

func (w *WrapKafkaWriter) Write(p []byte) (n int, err error) {
	_, _, err = w.producer.SendMessage(&sarama.ProducerMessage{
		Topic: w.topic,
		Key:   sarama.StringEncoder(fmt.Sprint(time.Now().UTC())),
		Value: sarama.ByteEncoder(p),

		// Below this point are filled in by the producer as the message is processed
		Offset:    0,
		Partition: 0,
		Timestamp: time.Time{},
	})

	return
}

var _ io.Writer = (*WrapKafkaWriter)(nil)

// WithCustomWriter add custom writer, so you can write using any storage method
// without waiting this package to be updated.
func WithCustomWriter(writer io.WriteCloser) Option {
	if writer == nil {
		return func(logger *logger) {}
	}
	return func(logger *logger) {
		// wire custom writer to log
		logger.writers = append(logger.writers, writer)
		logger.closer = append(logger.closer, writer)
	}
}

// WithLevel set level of logger
func WithLevel(level zapcore.Level) Option {
	return func(logger *logger) {
		logger.level = level
	}
}
