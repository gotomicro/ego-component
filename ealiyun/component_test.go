package ealiyun

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/gotomicro/ego/core/econf"
	"github.com/stretchr/testify/assert"
)

func newCmp() *Component {
	conf := `
[aliyun]
accessKeyId = "%s"
accessKeySecret = "%s"
endpoint = "%s"
`
	conf = fmt.Sprintf(conf,
		os.Getenv("ACCESSKEYID"), os.Getenv("ACCESSKEYSECRET"), os.Getenv("ENDPOINT"),
	)
	if err := econf.LoadFromReader(strings.NewReader(conf), toml.Unmarshal); err != nil {
		panic("load conf fail," + err.Error())
	}
	cmp := Load("aliyun").Build()
	return cmp
}

func TestComponent_ListGroups(t *testing.T) {
	cmp := newCmp()
	groups, err := cmp.ListGroups()
	assert.NoError(t, err)
	for _, group := range groups {
		log.Printf("%#v", group)
	}
}

func TestComponent_ListGroupsForUser(t *testing.T) {
	cmp := newCmp()
	groups, err := cmp.ListGroupsForUser("xxxx")
	assert.NoError(t, err)
	for _, group := range groups {
		log.Printf("%#v", group)
	}
}

func TestComponent_AddUserToGroup(t *testing.T) {
	cmp := newCmp()
	err := cmp.AddUserToGroup(AddOrRemoveUserToGroupRequest{
		GroupName:         "xxxx",
		UserPrincipalName: "xxxx",
	})
	assert.NoError(t, err)
}

func TestComponent_RemoveUserFromGroup(t *testing.T) {
	cmp := newCmp()
	err := cmp.RemoveUserFromGroup(AddOrRemoveUserToGroupRequest{
		GroupName:         "xxxx",
		UserPrincipalName: "xxxx",
	})
	assert.NoError(t, err)
}

func TestComponent_GetGroup(t *testing.T) {
	cmp := newCmp()
	group, err := cmp.GetGroup("test")
	assert.NoError(t, err)
	log.Printf("%#v", group)
}
