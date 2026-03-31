package requestid

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"localaihub/localaihub_go/internal/pkg/appctx"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}

		w.Header().Set("X-Request-Id", requestID)
		ctx := context.WithValue(r.Context(), appctx.RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
