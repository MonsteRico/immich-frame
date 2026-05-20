package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeEmbeddedFrameIndexWhenNoDistIsConfigured(t *testing.T) {
	server := Server{}
	request := httptest.NewRequest(http.MethodGet, "/frame", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET /frame status = %d, want 200", response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", contentType)
	}
	if !strings.Contains(response.Body.String(), "<html") {
		t.Fatal("embedded frame index response should contain HTML")
	}
}

func TestCleanAssetPathRejectsTraversal(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "app.js", want: "app.js"},
		{name: "leading slash", in: "/app.js", want: "app.js"},
		{name: "nested", in: "nested/app.css", want: "nested/app.css"},
		{name: "parent", in: "../app.js", want: ""},
		{name: "windows parent", in: `..\app.js`, want: ""},
		{name: "current", in: ".", want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := cleanAssetPath(tc.in); got != tc.want {
				t.Fatalf("cleanAssetPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
