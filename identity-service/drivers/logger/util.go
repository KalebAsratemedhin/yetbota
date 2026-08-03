package logger

import (
	"context"
	"io"
	"reflect"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/segmentio/encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	maskTag       = "mask"
	sliceByteMask = "X@BQ1"
	stringMask    = "***"
)

func newZapLogger(level zapcore.Level, writers ...io.Writer) (logger *zap.Logger) {
	zapWriters := make([]zapcore.WriteSyncer, 0)
	for _, writer := range writers {
		if writer == nil {
			continue
		}

		zapWriters = append(zapWriters, zapcore.AddSync(writer))
	}

	core := zapcore.NewCore(getEncoder(), zapcore.NewMultiWriteSyncer(zapWriters...), zapcore.Level(level))
	logger = zap.New(core)
	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "xtime",
		MessageKey:     "x",
		EncodeDuration: millisDurationEncoder,
		EncodeTime:     timeEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
	}

	return zapcore.NewJSONEncoder(encoderConfig)
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.999"))
}

func millisDurationEncoder(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(d.Nanoseconds() / 1000000)
}

func formatLogs(ctx context.Context, msg string, mask bool, fields ...Field) (logRecord []zap.Field) {
	ctxVal := ExtractCtx(ctx)

	// add global value from context that must be exist on all logs!
	logRecord = append(logRecord, zap.String("message", msg))

	logRecord = append(logRecord, zap.String("_app_name", ctxVal.ServiceName))
	logRecord = append(logRecord, zap.String("_app_version", ctxVal.ServiceVersion))
	logRecord = append(logRecord, zap.Int("_app_port", ctxVal.ServicePort))
	logRecord = append(logRecord, zap.String("_x_correlation_id", ctxVal.XCorrelationID))
	logRecord = append(logRecord, zap.String("_app_tag", ctxVal.Tag))
	logRecord = append(logRecord, zap.String("_app_method", ctxVal.ReqMethod))
	logRecord = append(logRecord, zap.String("_app_uri", ctxVal.ReqURI))
	logRecord = append(logRecord, zap.String("_error", ctxVal.Error))

	// add additional data that available across all log, such as user_id
	if ctxVal.AdditionalData != nil {
		logRecord = append(logRecord, zap.Any("_app_data", ctxVal.AdditionalData))
	}

	if ctxVal.Request.Val != nil {
		logRecord = append(logRecord, formatLog("_request", ctxVal.Request.Val, mask))
	}

	if ctxVal.Response.Val != nil {
		logRecord = append(logRecord, formatLog("_response", ctxVal.Response.Val, mask))
	}

	for _, field := range fields {
		logRecord = append(logRecord, formatLog(field.Key, field.Val, mask))
	}

	return
}

func formatLog(key string, msg interface{}, mask bool) (logRecord zap.Field) {
	if msg == nil {
		logRecord = zap.Any(key, struct{}{})
		return
	}

	// handle proto message
	p, ok := msg.(proto.Message)
	if ok {
		b, _err := json.Marshal(p)
		if _err != nil {
			logRecord = zap.Any(key, p.String())
			return
		}

		var data interface{}
		if _err = json.Unmarshal(b, &data); _err != nil {
			// string cannot be masked, so only try to marshal as json object
			logRecord = zap.Any(key, p.String())
			return
		}

		// use object json
		logRecord = zap.Any(key, data)
		return
	}

	// handle string, string is cannot be masked, just write it
	// but try to parse as json object if possible
	if str, ok := msg.(string); ok {
		var data interface{}
		if _err := json.Unmarshal([]byte(str), &data); _err != nil {
			logRecord = zap.String(key, str)
			return
		}

		logRecord = zap.Any(key, data)
		return
	}

	// if masking is disabled then just set as field log
	if !mask {
		logRecord = zap.Any(key, msg)
		return
	}

	// if masking is enabled and one of type supported by masking function
	switch reflect.ValueOf(msg).Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Struct:
		msgMasking := masking(msg)

		if convert, ok := msgMasking.(reflect.Value); ok {
			value := convert.Interface()
			logRecord = zap.Any(key, value)
			return
		}
	}

	// not masked since it failed to convert to reflect.Value above
	logRecord = zap.Any(key, msg)
	return
}

func ToField(key string, val interface{}) (field Field) {
	field = Field{
		Key: key,
		Val: val,
	}
	return
}

func masking(data interface{}) interface{} {
	original := reflect.ValueOf(data)
	altered := reflect.New(original.Type()).Elem()

	switch original.Kind() {
	case reflect.Ptr:
		// check if value is nil
		if !isNil(original) {
			elem := original.Elem()
			switch elem.Kind() {
			case reflect.Struct, reflect.Interface, reflect.Ptr:
				altered.Set(masking(elem.Interface()).(reflect.Value).Addr())
			case reflect.Slice:
				altered = maskSlice(elem)
			case reflect.Map:
				altered = maskMap(elem)
			default:
				altered.Set(elem.Addr())
			}
		}
	case reflect.Slice:
		altered = maskSlice(original)
	case reflect.Map:
		altered = maskMap(original)
	case reflect.Struct:
		for i := 0; i < original.NumField(); i++ {
			field := original.Field(i)
			switch field.Kind() {
			case reflect.Struct, reflect.Map, reflect.Interface, reflect.Slice, reflect.Ptr:
				if altered.Field(i).CanSet() && !isNil(field) {
					// []byte mostly used for byte file
					if field.Type() == TypeSliceOfBytes {
						if _, ok := original.Type().Field(i).Tag.Lookup(maskTag); ok {
							if !original.Field(i).IsNil() {
								altered.Field(i).SetBytes([]byte(sliceByteMask))
							}
						} else {
							altered.Field(i).Set(original.Field(i))
						}
					} else {
						altered.Field(i).Set(masking(field.Interface()).(reflect.Value))
					}
				}
			default:
				if _, ok := original.Type().Field(i).Tag.Lookup(maskTag); ok {
					if original.Field(i).Kind() == reflect.String && original.Field(i).Len() > 0 {
						altered.Field(i).SetString(stringMask)
					} else {
						altered.Field(i).Set(original.Field(i))
					}
				} else {
					if altered.Field(i).CanSet() {
						altered.Field(i).Set(original.Field(i))
					} else {
						switch original.Type() {
						case TypeTime:
							altered.Set(original)
							i += 2
						}
					}
				}
			}
		}
	default:
		altered.Set(original)
	}

	return altered
}

func maskSlice(elem reflect.Value) (altered reflect.Value) {
	altered = reflect.MakeSlice(elem.Type(), elem.Len(), elem.Len())
	for i := 0; i < elem.Len(); i++ {
		value := elem.Index(i)
		switch value.Kind() {
		case reflect.Struct, reflect.Map, reflect.Interface, reflect.Slice, reflect.Ptr:
			// check if value is nil
			if !isNil(value) {
				altered.Index(i).Set(masking(value.Interface()).(reflect.Value))
			}
		default:
			altered.Index(i).Set(value)
		}
	}

	return
}

func maskMap(elem reflect.Value) (altered reflect.Value) {
	altered = reflect.MakeMapWithSize(elem.Type(), len(elem.MapKeys()))
	mapRange := elem.MapRange()
	for mapRange.Next() {
		switch mapRange.Value().Kind() {
		case reflect.Struct, reflect.Map, reflect.Interface, reflect.Slice, reflect.Ptr:
			// check if value is nil
			if !isNil(mapRange.Value()) {
				altered.SetMapIndex(
					mapRange.Key(),
					masking(mapRange.Value().Interface()).(reflect.Value),
				)
			}
		default:
			altered.SetMapIndex(mapRange.Key(), mapRange.Value())
		}
	}

	return
}

func isNil(elem reflect.Value) bool {
	return elem.Interface() == nil ||
		(reflect.ValueOf(elem.Interface()).Kind() == reflect.Ptr && reflect.ValueOf(elem.Interface()).IsNil())
}
