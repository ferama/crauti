package cors

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferama/crauti/pkg/chaincontext"
	"github.com/ferama/crauti/pkg/conf"
)

func TestMustHaveCors(t *testing.T) {
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"I won?": true}`))
	})
	m := &corsMiddleware{
		next: root,
	}

	var enabled *bool
	b := true
	enabled = &b
	chain := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), chaincontext.ChainContextKey, chaincontext.ChainContext{
			Conf: &conf.MountPoint{
				Path:     "/",
				Upstream: "https://httpbin.org/get",
				Middlewares: conf.Middlewares{
					Cors: conf.Cors{
						Enabled: enabled,
					},
				},
			},
		})
		r = r.WithContext(ctx)
		m.ServeHTTP(w, r)
	})
	s := httptest.NewServer(chain)
	defer s.Close()

	req, _ := http.NewRequest("GET", s.URL, nil)

	// if I have an Origin header, I expect cors headers
	// to be set server side
	req.Header.Set("Origin", "http://localhost")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fail()
	}

	passed := 0
	for k, v := range resp.Header {
		t.Logf("%s: %s", k, v)

		want := "http://localhost"
		got := v[0]
		if k == "Access-Control-Allow-Origin" {
			if want != got {
				t.Fatalf("expected a %s, instead got: %s", want, got)
			} else {
				passed++
			}
		}

		want = "GET"
		got = v[0]
		if k == "Access-Control-Allow-Methods" {
			if want != got {
				t.Fatalf("expected a %s, instead got: %s", want, got)
			} else {
				passed++
			}
		}
	}
	if passed != 2 {
		t.Fatalf("not all headers are present: %d", passed)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	t.Log(string(bodyBytes))
}

func TestMustNotHaveCors(t *testing.T) {
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"I won?": true}`))
	})
	m := &corsMiddleware{
		next: root,
	}
	var enabled *bool
	b := true
	enabled = &b
	chain := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), chaincontext.ChainContextKey, chaincontext.ChainContext{
			Conf: &conf.MountPoint{
				Path:     "/",
				Upstream: "https://httpbin.org/get",
				Middlewares: conf.Middlewares{
					Cors: conf.Cors{
						Enabled: enabled,
					},
				},
			},
		})
		r = r.WithContext(ctx)
		m.ServeHTTP(w, r)
	})
	s := httptest.NewServer(chain)
	defer s.Close()

	req, _ := http.NewRequest("GET", s.URL, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fail()
	}

	for k := range resp.Header {
		switch k {
		case "Access-Control-Allow-Origin",
			"Access-Control-Allow-Methods",
			"Access-Control-Allow-Headers",
			"Access-Control-Allow-Credentials":
			t.Fatalf("unexpected header %s", k)
		}
	}
}
