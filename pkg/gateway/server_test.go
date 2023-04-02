package gateway

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/spf13/viper"
)

func loadConf(file string) {
	path := filepath.Join("testdata", file)
	viper.SetConfigFile(path)
	viper.ReadInConfig()
	conf.Update()
}

func startWebServer(sleepTime int) {
	http.ListenAndServe(":19999", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(sleepTime) * time.Second)
		w.Write([]byte("done"))
	}))
}

func TestTimeout(t *testing.T) {
	go startWebServer(2)

	loadConf("test.yaml")
	gwListenAddress := "localhost:39142"
	gwServer := NewServer(gwListenAddress)
	gwServer.UpdateHandlers()

	go gwServer.Start()

	// give time to gateway to raise
	time.Sleep(1 * time.Second)

	res, err := http.Get("http://" + gwListenAddress)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatal("expected 200")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	if string(body) != "done" {
		t.Fatal("expected 'done'")
	}
}

func Test404(t *testing.T) {
	go startWebServer(0)

	loadConf("test2.yaml")
	gwListenAddress := "localhost:39143"
	gwServer := NewServer(gwListenAddress)
	gwServer.UpdateHandlers()

	go gwServer.Start()

	res, err := http.Get(fmt.Sprintf("http://%s/notexists", gwListenAddress))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 404 {
		t.Fatal("expected 404")
	}
}

func BenchmarkRequest1(b *testing.B) {
	go startWebServer(0)

	loadConf("test.yaml")
	gwListenAddress := "localhost:39142"
	gwServer := NewServer(gwListenAddress)
	gwServer.UpdateHandlers()

	go gwServer.Start()

	time.Sleep(2 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := http.Get("http://" + gwListenAddress)
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
