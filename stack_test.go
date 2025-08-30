package stack_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glue-africa/stack"
)

func TestRouter(t *testing.T) {
	used := ""
	m := mw(&used, "1", "2", "3", "4", "5", "6")

	hf := func(w http.ResponseWriter, r *http.Request) {}

	r := stack.NewRouter()
	r.Use(m[0], m[1])

	r.HandleFunc("GET /{$}", hf)

	r.Group(func(r *stack.Router) {
		r.Use(m[2], m[3])
		r.HandleFunc("GET /foo", hf)

		r.Group(func(r *stack.Router) {
			r.Use(m[4])
			r.HandleFunc("GET /nested/foo", hf)
		})
	})

	r.Group(func(r *stack.Router) {
		r.Use(m[5])
		r.HandleFunc("GET /bar", hf)
	})

	r.HandleFunc("GET /baz", hf)

	cases := []struct {
		name           string
		method         string
		path           string
		expectedUsed   string
		expectedStatus int
	}{
		{
			name:           "root path",
			method:         "GET",
			path:           "/",
			expectedUsed:   "12",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "foo path",
			method:         "GET",
			path:           "/foo",
			expectedUsed:   "1234",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "nested foo path",
			method:         "GET",
			path:           "/nested/foo",
			expectedUsed:   "12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "bar path",
			method:         "GET",
			path:           "/bar",
			expectedUsed:   "126",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "baz path",
			method:         "GET",
			path:           "/baz",
			expectedUsed:   "12",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found pat",
			method:         "GET",
			path:           "/notfound",
			expectedUsed:   "12",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "method not allowed",
			method:         "POST",
			path:           "/nested/foo",
			expectedUsed:   "12",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			used = ""

			req, err := http.NewRequest(tc.method, tc.path, nil)
			if err != nil {
				t.Fatalf("NewRequest: %s", err)
			}

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			if used != tc.expectedUsed {
				t.Errorf("expected middleware %q, got %q", tc.expectedUsed, used)
			}
		})
	}
}

func mw(used *string, ids ...string) []func(http.Handler) http.Handler {
	var middlewares []func(http.Handler) http.Handler
	for _, id := range ids {
		middlewares = append(middlewares, func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				*used += id
				next.ServeHTTP(w, r)
			})
		})
	}
	return middlewares
}
