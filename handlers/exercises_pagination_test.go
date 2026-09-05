package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPaginationDefaultsAndCaps(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		target    string
		wantPage  int64
		wantLimit int64
	}{
		{name: "defaults", target: "/exercises", wantPage: 1, wantLimit: 50},
		{name: "requested page", target: "/exercises?page=3&limit=25", wantPage: 3, wantLimit: 25},
		{name: "invalid values", target: "/exercises?page=0&limit=-1", wantPage: 1, wantLimit: 50},
		{name: "maximum limit", target: "/exercises?limit=1000", wantPage: 1, wantLimit: 100},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest("GET", test.target, nil)

			page, limit := pagination(context)
			if page != test.wantPage || limit != test.wantLimit {
				t.Fatalf("pagination() = (%d, %d), want (%d, %d)", page, limit, test.wantPage, test.wantLimit)
			}
		})
	}
}
