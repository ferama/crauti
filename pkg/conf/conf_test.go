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

func Test1(t *testing.T) {
	loadConf("test1.yaml")

	if !Crauti.Middlewares.Cors.Enabled {
		t.Error("expected global cors enabled")
	}

	if Crauti.Middlewares.Cors.Val != "test1" {
		t.Error("test1 expected")
	}

	if Crauti.MountPoints[0].Middlewares.Cors.Enabled {
		t.Error("cors should be disabled in mountpoints")
	}

	if Crauti.MountPoints[0].Middlewares.Cors.Val != "test1" {
		t.Error("test1 expected")
	}

	if Crauti.MountPoints[1].Middlewares.Cors.Val != "test2" {
		t.Error("test1 expected")
	}
}
