package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDocumentationRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerDocumentationRoutes(router)

	for _, test := range []struct {
		path        string
		contentType string
		body        string
	}{
		{path: "/openapi.yaml", contentType: "application/yaml", body: "openapi: 3.0.3"},
		{path: "/docs", contentType: "text/html", body: `Redoc.init("/openapi.yaml"`},
	} {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", test.path, recorder.Code, http.StatusOK)
			}
			if !strings.Contains(recorder.Header().Get("Content-Type"), test.contentType) {
				t.Fatalf("GET %s content type = %q, want %q", test.path, recorder.Header().Get("Content-Type"), test.contentType)
			}
			if !strings.Contains(recorder.Body.String(), test.body) {
				t.Fatalf("GET %s body does not contain %q", test.path, test.body)
			}
		})
	}
}

func TestOpenAPISpecCoversRegisteredRoutes(t *testing.T) {
	spec := string(openAPISpec)
	for _, route := range []string{
		"/openapi.yaml:", "/docs:",
		"/register:", "/login:", "/healthcheck:", "/health/storage:", "/me:",
		"/exercises:", "/exercises/{id}:", "/routines:", "/routines/{id}:",
		"/recommendations/routine:", "/recommendations/alternative/{id}:",
		"/permissions:", "/permissions/role/{subject}:", "/permissions/groups:",
		"/permissions/groups/{user}:",
	} {
		if !strings.Contains(spec, route) {
			t.Errorf("OpenAPI spec does not document %s", route)
		}
	}
}
