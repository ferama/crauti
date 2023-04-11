package gateway

import (
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/utils"
	"github.com/spf13/viper"
)

func loadConf(file string) {
	path := filepath.Join("testdata", file)
	viper.SetConfigFile(path)
	viper.ReadInConfig()
	conf.Update()
}

func startWebServer(sleepTime int) http.Server {
	s := http.Server{
		Addr: ":19999",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Duration(sleepTime) * time.Second)
			w.Write([]byte("done"))
		}),
	}
	return s
}

func TestTimeout(t *testing.T) {
	s := startWebServer(2)
	go s.ListenAndServe()

	loadConf("test.yaml")
	gwServer := NewGateway(":8080", ":8443")
	defer func() {
		gwServer.Stop()
		s.Close()
	}()
	gwServer.UpdateHandlers()

	go gwServer.Start()

	// give time to gateway to raise
	time.Sleep(1 * time.Second)

	res, err := http.Get("http://127.0.0.1:8080")
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
	s := startWebServer(0)
	go s.ListenAndServe()

	loadConf("test2.yaml")
	gwServer := NewGateway(":8080", ":8443")
	defer func() {
		gwServer.Stop()
		s.Close()
	}()
	gwServer.UpdateHandlers()

	go gwServer.Start()
	// give time to gateway to raise
	time.Sleep(1 * time.Second)

	res, err := http.Get("http://127.0.0.1:8080/notexists")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 404 {
		t.Fatal("expected 404")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	if string(body) != utils.BodyResponse404 {
		t.Fatalf("expected '%s'", utils.BodyResponse404)
	}
}

func Test404MatchHost(t *testing.T) {
	s := startWebServer(0)
	go s.ListenAndServe()

	loadConf("test2.yaml")
	gwServer := NewGateway(":8080", ":8443")
	defer func() {
		gwServer.Stop()
		s.Close()
	}()
	gwServer.UpdateHandlers()

	go gwServer.Start()
	// give time to gateway to raise
	time.Sleep(1 * time.Second)

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
	req.Host = "test3.loc"

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 404 {
		t.Fatal("expected 404")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	if string(body) != utils.BodyResponse404 {
		t.Fatalf("expected '%s'", utils.BodyResponse404)
	}
}

func TestPort80(t *testing.T) {
	s := startWebServer(0)
	go s.ListenAndServe()
	loadConf("test4.yaml")
	gwServer := NewGateway(":8080", ":8443")
	defer func() {
		gwServer.Stop()
		s.Close()
	}()
	gwServer.UpdateHandlers()

	go gwServer.Start()

	time.Sleep(1 * time.Second)

	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Host = "test4.loc"

	client := &http.Client{}
	res, err := client.Do(req)
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
