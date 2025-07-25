package httpserver

import (
	"net/http"
	"time"

	"github.com/kagent-dev/kagent/go/controller/internal/httpserver/handlers"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log := ctrllog.Log.WithName("http").WithValues(
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		if userID := r.URL.Query().Get("user_id"); userID != "" {
			log = log.WithValues("user_id", userID)
		}

		ww := newStatusResponseWriter(w)
		ctx := ctrllog.IntoContext(r.Context(), log)
		log.V(1).Info("Request started")
		next.ServeHTTP(ww, r.WithContext(ctx))
		log.Info("Request completed",
			"status", ww.status,
			"duration", time.Since(start),
		)
	})
}

// For streaming responses in A2A lib
var _ http.Flusher = &statusResponseWriter{}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{w, http.StatusOK}
}

func (w *statusResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Forward RespondWithError to underlying writer if it implements ErrorResponseWriter
func (w *statusResponseWriter) RespondWithError(err error) {
	if errWriter, ok := w.ResponseWriter.(handlers.ErrorResponseWriter); ok {
		errWriter.RespondWithError(err)
		w.status = 500
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			w.Header().Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}
