package eredis

import (
	"context"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/gotomicro/ego/core/econf"
	"github.com/stretchr/testify/assert"
)

func newCmp() *Component {
	conf := `
[redis]
	mode = "sentinel"
	masterName = "redis-master"
	addrs = ["localhost:26379","localhost:26380","localhost:26380"]
`
	if err := econf.LoadFromReader(strings.NewReader(conf), toml.Unmarshal); err != nil {
		panic("load conf fail," + err.Error())
	}
	cmp := Load("redis").Build()
	return cmp
}

func TestSentinel(t *testing.T) {
	cmp := newCmp()
	res, err := cmp.Ping(context.TODO())
	assert.NoError(t, err)
	t.Log("ping result", res)
}
