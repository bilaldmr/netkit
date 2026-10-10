package headers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "a=1")
		w.Header().Add("Set-Cookie", "b=2")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusTeapot)
	}))
	defer srv.Close()

	res, err := Service{HTTP: srv.Client()}.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != http.StatusTeapot || res.Headers["Set-Cookie"] != "a=1\nb=2" || res.Headers["Cache-Control"] != "no-store" {
		t.Fatalf("unexpected result: %+v", res)
	}
	if _, err := (Service{}).Get(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty url")
	}
}
