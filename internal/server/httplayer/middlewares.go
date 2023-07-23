package httplayer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (api *httpAPI) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)
		duration := time.Since(start)
		api.logger.Info("",
			zap.String("URI", r.URL.String()),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)
	})
}

func (api *httpAPI) hashing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hashing")
		headerHashValue := r.Header.Get("HashSHA256")
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			api.logger.Error("", zap.Error(err))
		}
		fmt.Println(body)
		// подписываем алгоритмом HMAC, используя SHA-256
		h := hmac.New(sha256.New, []byte(api.hashKey))
		h.Write(body)
		expectedHash := h.Sum(nil)

		//fmt.Printf("%x", expectedHash)
		expectedHashString := fmt.Sprintf("%x", expectedHash)
		fmt.Println(expectedHashString)
		fmt.Println(headerHashValue)
		if expectedHashString == headerHashValue {
			fmt.Println("hashing ServeHTTP")
			r.Body = io.NopCloser(bytes.NewReader(body))
			fmt.Println(r.Body)
			next.ServeHTTP(w, r)
		} else {
			api.logger.Info("Client send incorrect HashSHA256 header value")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Incorrect HashSHA256 header value"))
		}
	})
}
