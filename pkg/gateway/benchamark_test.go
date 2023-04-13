package gateway

import (
	"io"
	"net/http"
	"testing"
	"time"
)

func BenchmarkRequest1(b *testing.B) {
	s := startWebServer(0)
	go s.ListenAndServe()

	loadConf("test.yaml")
	gwServer := NewGateway(":8080", ":8443")
	defer func() {
		gwServer.Stop()
		s.Close()
	}()
	gwServer.Update()

	go gwServer.Start()

	time.Sleep(2 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := http.Get("http://127.0.0.1:8080")
		if err != nil {
			b.Fatal(err)
		}
		if res.StatusCode != 200 {
			b.Fatal("expected 200")
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			b.Error(err)
		}
		if string(body) != "done" {
			b.Fatal("expected 'done'")
		}
	}
}
