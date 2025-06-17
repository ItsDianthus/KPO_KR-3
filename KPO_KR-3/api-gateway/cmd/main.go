package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Оборачивает handler и добавляет CORS-заголовки
func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем любому фронту обращаться к API
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Браузер может сначала слать preflight OPTIONS
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func newReverseProxy(target *url.URL) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		// Просто перенаправляем на тот же путь, что и был
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		// req.URL.Path остаётся без изменений
		req.Host = target.Host
	}
	return proxy
}

func main() {
	ordersURL, _ := url.Parse("http://orders-service:8081")
	paymentsURL, _ := url.Parse("http://payments-service:8082")

	mux := http.NewServeMux()

	mux.Handle("/orders", newReverseProxy(ordersURL))
	mux.Handle("/orders/", newReverseProxy(ordersURL))

	mux.Handle("/payments", newStripPrefixProxy(paymentsURL, "/payments"))
	mux.Handle("/payments/", newStripPrefixProxy(paymentsURL, "/payments"))

	handler := withCORS(mux)

	log.Println("API Gateway listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func newStripPrefixProxy(target *url.URL, prefix string) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(target)
	orig := proxy.Director
	proxy.Director = func(req *http.Request) {
		orig(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
	}
	return proxy
}
