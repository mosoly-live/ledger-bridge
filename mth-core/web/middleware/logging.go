package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/log"
	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/web"
	webcontext "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/web/context"
	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/web/header"
	"go.uber.org/zap"
)

// LoggingParameters contains additional data
type LoggingParameters struct {
	HeaderKeys    map[string]struct{}
	PayloadFields map[string]struct{}
}

// LoggingHandler is a middleware that will write the log to 'out' writer.
func LoggingHandler(h http.Handler, parameters LoggingParameters) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		path := r.URL.Path
		raw := r.URL.RawQuery

		cacheReader := newLimitCacheReader(r.Body, 8192)
		r.Body = cacheReader

		// call inner handler
		lrw := web.NewLogStatusReponseWriter(w)
		recoveryErr := handleRequestWithRecovery(h, lrw, r)

		end := time.Now()
		latency := end.Sub(start)

		method := r.Method
		statusCode := lrw.Status()

		correlationID := webcontext.CorrelationID(r.Context())
		clientIP := header.ClientIP(r)
		authClaims := header.AuthClaims(r)

		if raw != "" {
			path = path + "?" + raw
		}

		l := log.With(
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status_code", statusCode),
			zap.String("correlation_id", correlationID),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
		)
		l = l.WithOptions(zap.AddStacktrace(zap.DPanicLevel)) // Do not include stacktrace for Error level and lower
		l = l.With(log.FieldsFrom(adjustKeysWithPrefix(authClaims, "c"))...)

		reqHeaderFields := getHeaderFields(r.Header, parameters.HeaderKeys, "ih")
		reqHeaderLoggingFields := log.FieldsFrom(reqHeaderFields)
		l = l.With(reqHeaderLoggingFields...)

		respHeaderFields := getHeaderFields(lrw.Header(), parameters.HeaderKeys, "oh")
		respHeaderLoggingFields := log.FieldsFrom(respHeaderFields)
		l = l.With(respHeaderLoggingFields...)

		logMsg := "[HTTP]"
		if recoveryErr != nil {
			logMsg += " " + recoveryErr.Error()
		}

		// Include request payload when we generate 400+ responses.
		if statusCode >= 400 {
			var rb bytes.Buffer

			replacedBytes := obfuscateSensitiveBodyFields(cacheReader.Bytes(), parameters.PayloadFields)
			rb.Write(replacedBytes)

			l = l.With(zap.String("payload", rb.String()))
		}

		if statusCode >= 500 {
			l.Error(logMsg)
		} else {
			l.Info(logMsg)
		}
	})
}

func obfuscateSensitiveBodyFields(bodyBytes []byte, sensitiveFields map[string]struct{}) []byte {
	var replacedBytes []byte
	m := make(map[string]interface{})
	err := json.Unmarshal(bodyBytes, &m)

	if err == nil {
		m = obfuscateSensitiveFieldsInternal(m, sensitiveFields)
		replacedBytes, _ = json.Marshal(m)
	}

	return replacedBytes
}

func obfuscateSensitiveFieldsInternal(m map[string]interface{}, sensitiveFields map[string]struct{}) map[string]interface{} {
	for key := range m {
		if _, ok := sensitiveFields[key]; ok {
			m[key] = "*****"
		} else {
			if complexValue, ok := m[key].(map[string]interface{}); ok {
				m[key] = obfuscateSensitiveFieldsInternal(complexValue, sensitiveFields)
			}
		}
	}

	return m
}

func adjustKeysWithPrefix(m map[string]interface{}, prefix string) map[string]interface{} {
	ret := make(map[string]interface{})
	for k, v := range m {
		ret[prefix+"_"+k] = v
	}
	return ret
}

func getHeaderFields(h http.Header, exclude map[string]struct{}, prefix string) (m map[string]interface{}) {
	m = make(map[string]interface{})
	for key := range h {
		if _, ok := exclude[http.CanonicalHeaderKey(key)]; ok {
			m[prefix+"_"+strings.ToLower(key)] = "*****"
		} else {
			m[prefix+"_"+strings.ToLower(key)] = h.Get(key)
		}
	}
	return
}

func handleRequestWithRecovery(h http.Handler, w http.ResponseWriter, r *http.Request) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("PANIC: %v\n%v", e, string(debug.Stack()))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	h.ServeHTTP(w, r)
	return
}

func newLimitCacheReader(r io.ReadCloser, limit int64) *limitCacheReader {
	return &limitCacheReader{r: r, n: limit}
}

// limitCacheReader caches first `limit` bytes from underlying reader
type limitCacheReader struct {
	r   io.ReadCloser
	n   int64 // max bytes remaining
	buf bytes.Buffer
}

func (c *limitCacheReader) Read(p []byte) (n int, err error) {
	n, err = c.r.Read(p)
	if c.n <= 0 || n <= 0 {
		return
	}
	var r int64
	if int64(n) > c.n {
		r = c.n
	} else {
		r = int64(n)
	}
	c.buf.Write(p[:r])
	c.n -= r
	return
}

func (c *limitCacheReader) Close() error { return c.r.Close() }

func (c *limitCacheReader) Bytes() []byte { return c.buf.Bytes() }
