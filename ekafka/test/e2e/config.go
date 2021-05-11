package e2e

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/gotomicro/ego/core/econf"
)

func init() {
	configFilePath := os.Getenv("TEST_CONFIG")

	f, err := os.Open(configFilePath)
	if err != nil {
		panic(err)
	}

	err = econf.LoadFromReader(f, toml.Unmarshal)
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}
