package egitlab

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/gotomicro/ego/core/econf"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

func newCmp() *Component {
	conf := `
[gitlab]
	token = "%s"
	baseUrl = "%s" 
`
	conf = fmt.Sprintf(conf,
		os.Getenv("TOKEN"), os.Getenv("BASEURL"),
	)
	if err := econf.LoadFromReader(strings.NewReader(conf), toml.Unmarshal); err != nil {
		panic("load conf fail," + err.Error())
	}
	return Load("gitlab").Build()
}

func TestComponent_Client(t *testing.T) {
	cmp := newCmp()
	client := cmp.Client()
	user, _, err := client.Users.GetUser(11, gitlab.GetUsersOptions{})
	assert.Equal(t, nil, err)
	log.Printf("user:%v \n", user)
}
