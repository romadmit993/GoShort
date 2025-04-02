package middleware

import (
	"compress/gzip"
	//"io"
	"net/http"
	"romadmit993/GoShort/internal/storage"
	"strings"
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
	// gzipWriter struct {
	// 	http.ResponseWriter
	// 	Writer io.Writer
	// }
	gzipResponseWriter struct {
		http.ResponseWriter
		gz *gzip.Writer
	}
)

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		storage.Sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.responseData.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// func (w gzipWriter) Write(b []byte) (int, error) {
// 	return w.Writer.Write(b)
// }

// func GzipHandle(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		contentType := w.Header().Get("Content-Type")
// 		if strings.Contains(contentType, "application/json") ||
// 			strings.Contains(contentType, "text/html") {
// 			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
// 				next.ServeHTTP(w, r)
// 				return
// 			}
// 			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
// 			if err != nil {
// 				io.WriteString(w, err.Error())
// 				return
// 			}
// 			defer gz.Close()

//				w.Header().Set("Content-Encoding", "gzip")
//				next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
//				return
//			}
//			next.ServeHTTP(w, r)
//		})
//	}
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.gz == nil {
		contentType := w.Header().Get("Content-Type")
		if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
			if strings.Contains(w.Header().Get("Accept-Encoding"), "gzip") {
				w.gz = gzip.NewWriter(w.ResponseWriter)
				w.Header().Set("Content-Encoding", "gzip")
				defer w.gz.Close()
			}
		}
	}
	if w.gz != nil {
		return w.gz.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.gz == nil {
		contentType := w.Header().Get("Content-Type")
		if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
			if strings.Contains(w.Header().Get("Accept-Encoding"), "gzip") {
				w.gz = gzip.NewWriter(w.ResponseWriter)
				w.Header().Set("Content-Encoding", "gzip")
				defer w.gz.Close()
			}
		}
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func GzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gzw := &gzipResponseWriter{ResponseWriter: w}
		defer func() {
			if gzw.gz != nil {
				gzw.gz.Close()
			}
		}()
		next.ServeHTTP(gzw, r)
	})
}

func UngzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			if r.Body == nil || r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}

			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}
		next.ServeHTTP(w, r)
	})
}
