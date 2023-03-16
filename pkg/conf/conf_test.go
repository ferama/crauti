package conf

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func loadConf(file string) {
	path := filepath.Join("testdata", file)
	viper.SetConfigFile(path)
	viper.ReadInConfig()
	Update()
}

// func Test1(t *testing.T) {
// 	loadConf("test1.yaml")

// 	if !ConfInst.Middlewares.Cors.Enabled {
// 		t.Error("expected global cors enabled")
// 	}

// 	if ConfInst.MountPoints[0].Middlewares.Cors.Enabled {
// 		t.Error("cors should be disabled in mountpoints")
// 	}
// }

func Test2(t *testing.T) {
	loadConf("test2.yaml")
	if len(ConfInst.MountPoints[0].Middlewares.Cache.Methods) != 0 {
		t.Error("empty methods expected")
	}

	if len(ConfInst.MountPoints[0].Middlewares.Cache.KeyHeaders) == 0 {
		t.Fatal("keyheaders expected")
	}
	if ConfInst.MountPoints[0].Middlewares.Cache.KeyHeaders[0] != "header1" {
		t.Error("header1 expected")
	}
}

func Test3(t *testing.T) {
	loadConf("test3.yaml")
	if len(ConfInst.MountPoints[0].Middlewares.Cache.KeyHeaders) != 1 {
		t.Fatal("wrong keyheaders count")
	}
	if ConfInst.MountPoints[0].Middlewares.Cache.KeyHeaders[0] != "header1" {
		t.Fatal("header1 expected")
	}
}

func TestBooleans(t *testing.T) {
	loadConf("test3.yaml")
	if !ConfInst.MountPoints[0].Middlewares.Cache.IsEnabled() {
		t.Fatal("cache should be enabled on mount point")
	}

	if ConfInst.MountPoints[0].Middlewares.Cors.IsEnabled() {
		t.Fatal("cors should not be enabled on mount point")
	}
}
