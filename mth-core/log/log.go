// Package log redirects output from the standard library's package-global
// logger to the wrapped zap logger at InfoLevel. It also has a predefined 'standard'
// Logger accessible through helper functions Print[f|ln], Fatal[f|ln], and
// Panic[f|ln], which are easier to use than creating a Logger manually.
package log

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	mu               sync.RWMutex
	wrappedZapLogger *zap.Logger
	wrappedStdLogger *log.Logger
)

func init() {
	// Using global zap logger, by default it's a no-op Logger. It never writes out logs or internal errors,
	// and it never runs user-defined hooks..
	ReplaceGlobals(zap.L())
}

// NewDevelopment creates logger for development
func NewDevelopment() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// turn off logging stack traces at warn level
	config.Development = false
	l, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build development logger: %v", err))
	}
	return l
}

// NewProduction creates logger for production
func NewProduction() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "sts"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = syslogLevelEncoder
	l, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build production logger: %v", err))
	}
	return l
}

// Logger extends zap.Logger
type Logger zap.Logger

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func With(fields ...zapcore.Field) *Logger {
	return (*Logger)(zap.L().With(fields...))
}

// WithOptions creates a child logger with options.
func (l *Logger) WithOptions(opts ...zap.Option) *Logger {
	lo := (*zap.Logger)(l)
	return (*Logger)(lo.WithOptions(opts...))
}

// Err constructs a "error" field with the error text.
func Err(err error) zapcore.Field {
	return zap.Error(err)
}

// FieldsFrom returns fields from a map[string]interface{}.
func FieldsFrom(m map[string]interface{}) (s []zapcore.Field) {
	for k, v := range m {
		s = append(s, zap.Any(k, v))
	}
	return
}

// UserID constructs a "user_id" field with the given ID.
func UserID(id int64) zapcore.Field {
	return zap.Int64("user_id", id)
}

// UserHash constructs a "user_hash" field with the given hash.
func UserHash(hash string) zapcore.Field {
	return zap.String("user_hash", hash)
}

// UserUUID constructs a "user_uuid" field with the given hash.
func UserUUID(uuid string) zapcore.Field {
	return zap.String("user_uuid", uuid)
}

// DealID constructs a "deal_id" field with the given ID.
func DealID(id int64) zapcore.Field {
	return zap.Int64("deal_id", id)
}

// ClaimID constructs a "claim_id" field with the given ID.
func ClaimID(id int64) zapcore.Field {
	return zap.Int64("claim_id", id)
}

// PhoneNumber constructs a "phone_number" field with the given phone number.
func PhoneNumber(number string) zapcore.Field {
	return zap.String("phone_number", number)
}

// PhoneCode constructs a "phone_code" field with the given phone code.
func PhoneCode(code string) zapcore.Field {
	return zap.String("phone_code", code)
}

// CorrelationID constructs a "correlation_id field with the given id.
func CorrelationID(ID string) zapcore.Field {
	return zap.String("correlation_id", ID)
}

// MerchantID constructs a "merchant_id" field with given ID.
func MerchantID(id int64) zapcore.Field {
	return zap.Int64("merchant_id", id)
}

// CallbackURL constructs a "callback_url" field with given url.
func CallbackURL(url string) zapcore.Field {
	return zap.String("callback_url", url)
}

// HTTPRequest constructs a "http_request" field with given req.
func HTTPRequest(req interface{}) zapcore.Field {
	return zap.Reflect("http_request", req)
}

// HTTPResponse constructs a "http_response" field with given resp.
func HTTPResponse(resp interface{}) zapcore.Field {
	return zap.Reflect("http_response", resp)
}

// HTTPHeader constructs a "http_header" field with the given header.
func HTTPHeader(h http.Header) zap.Field {
	// Copy header values to do not modify current request header
	headers := make(http.Header)
	for key, values := range h {
		for _, value := range values {
			headers.Add(key, value)
		}
	}

	// Remove auth token, but leave header to know that request was authorized
	if len(headers["Authorization"]) > 0 {
		headers.Set("Authorization", "***")
	}

	return zap.Any("http_header", headers)
}

// RequestURL constructs a "request_url" field with the given url
func RequestURL(url url.URL) zap.Field {
	return zap.String("request_url", url.String())
}

// RequestMethod constructs a "request_method" field with the given method.
func RequestMethod(m string) zap.Field {
	return zap.String("request_method", m)
}

// StatusCode constructs a "status_code" field with the given status code.
func StatusCode(s int) zap.Field {
	return zap.Int("status_code", s)
}

// BlockNumber constructs a "block_number" field with the given block number.
func BlockNumber(n int64) zap.Field {
	return zap.Int64("block_number", n)
}

// FieldFloatInt64 creates a new zapcore field
func FieldFloatInt64(field string, val int64) zapcore.Field {
	return zap.Int64(field, val)
}

// FieldFloat64 creates a new zapcore field
func FieldFloat64(field string, val float64) zapcore.Field {
	return zap.Float64(field, val)
}

// WithFieldInt creates a new zapcore field
func (l *Logger) WithFieldInt(field string, val int64) *Logger {
	zapField := FieldFloatInt64(field, val)
	return l.With(zapField)
}

// WithFieldFloat64 creates a new zapcore field
func (l *Logger) WithFieldFloat64(field string, val float64) *Logger {
	zapField := FieldFloat64(field, val)
	return l.With(zapField)
}

// With returns new logger with new fields
func (l *Logger) With(fields ...zapcore.Field) *Logger {
	zl := (*zap.Logger)(l)
	zl2 := zl.With(fields...)
	return (*Logger)(zl2)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Warn(msg string, fields ...zapcore.Field) {
	zl := (*zap.Logger)(l).WithOptions(zap.AddCallerSkip(1))
	zl.Warn(msg, fields...)
}

// Info  logs a message at infoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Info(msg string, fields ...zapcore.Field) {
	zl := (*zap.Logger)(l).WithOptions(zap.AddCallerSkip(1))
	zl.Info(msg, fields...)
}

// Error logs a message at errorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Error(msg string, fields ...zapcore.Field) {
	zl := (*zap.Logger)(l).WithOptions(zap.AddCallerSkip(1))
	zl.Error(msg, fields...)
}

func syslogLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt(syslogCode(l))
}

const (
	syslogDebug         = 7
	syslogInformational = 6
	syslogWarning       = 4
	syslogError         = 3
	syslogCritical      = 2
	syslogAlert         = 1
	syslogEmergency     = 0
)

func syslogCode(l zapcore.Level) int {
	switch l {
	case zapcore.DebugLevel:
		return syslogDebug
	case zapcore.InfoLevel:
		return syslogInformational
	case zapcore.WarnLevel:
		return syslogWarning
	case zapcore.ErrorLevel:
		return syslogError
	case zapcore.DPanicLevel:
		return syslogCritical
	case zapcore.PanicLevel:
		return syslogCritical
	case zapcore.FatalLevel:
		return syslogAlert
	default:
		return syslogEmergency
	}
}

// ReplaceGlobals replaces the global loggers
func ReplaceGlobals(l *zap.Logger) {
	zap.ReplaceGlobals(l)
	zap.RedirectStdLog(l)
	redirectFlagLog(l)

	// add caller skip for wrapped functions
	loggerWithCallerSkip := l.WithOptions(zap.AddCallerSkip(1))

	mu.Lock()
	wrappedZapLogger = loggerWithCallerSkip
	wrappedStdLogger = zap.NewStdLog(loggerWithCallerSkip)
	mu.Unlock()
}

const (
	flagDefaultDepth  = 2
	loggerWriterDepth = 2
)

func redirectFlagLog(l *zap.Logger) {
	flag.CommandLine.SetOutput(
		&loggerWriter{l.WithOptions(
			zap.AddCallerSkip(flagDefaultDepth + loggerWriterDepth),
		).Warn},
	)
}

type loggerWriter struct {
	logFunc func(msg string, fields ...zapcore.Field)
}

func (l *loggerWriter) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	l.logFunc(string(p))
	return len(p), nil
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zapcore.Field) {
	mu.RLock()
	l := wrappedZapLogger
	mu.RUnlock()
	l.Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zapcore.Field) {
	mu.RLock()
	l := wrappedZapLogger
	mu.RUnlock()
	l.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zapcore.Field) {
	mu.RLock()
	l := wrappedZapLogger
	mu.RUnlock()
	l.Error(msg, fields...)
}

// Sync flushes any buffered log entries. Applications should take care to call Sync before exiting.
func Sync() error {
	mu.RLock()
	l := wrappedZapLogger
	mu.RUnlock()
	return l.Sync()
}

// StdLogger returns a *log.Logger which writes to the underlying logger at InfoLevel.
func StdLogger() (l *log.Logger) {
	mu.RLock()
	l = wrappedStdLogger
	mu.RUnlock()
	return
}

// Print calls log.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func Print(args ...interface{}) {
	mu.RLock()
	l := wrappedStdLogger
	mu.RUnlock()
	l.Print(args...)
}

// Printf calls log.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(template string, args ...interface{}) {
	mu.RLock()
	l := wrappedStdLogger
	mu.RUnlock()
	l.Printf(template, args...)
}

// Println calls log.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func Println(args ...interface{}) {
	mu.RLock()
	l := wrappedStdLogger
	mu.RUnlock()
	l.Println(args...)
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(args ...interface{}) {
	mu.RLock()
	l := wrappedStdLogger
	mu.RUnlock()
	l.Fatal(args...)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(template string, args ...interface{}) {
	mu.RLock()
	l := wrappedStdLogger
	mu.RUnlock()
	l.Fatalf(template, args...)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(args ...interface{}) {
	mu.RLock()
	l := wrappedStdLogger
	mu.RUnlock()
	l.Fatalln(args...)
}

// Panic is equivalent to Print() followed by a call to panic().
func Panic(args ...interface{}) {
	mu.RLock()
	l := wrappedStdLogger
	mu.RUnlock()
	l.Panic(args...)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func Panicf(template string, args ...interface{}) {
	mu.RLock()
	l := wrappedStdLogger
	mu.RUnlock()
	l.Panicf(template, args...)
}

// Panicln is equivalent to Println() followed by a call to panic().
func Panicln(args ...interface{}) {
	mu.RLock()
	l := wrappedStdLogger
	mu.RUnlock()
	l.Panicln(args...)
}
