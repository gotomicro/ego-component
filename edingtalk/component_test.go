package edingtalk

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gotomicro/ego-component/eredis"
	"github.com/gotomicro/ego/core/econf"
	"github.com/stretchr/testify/assert"
)

func newCmp() *Component {
	conf := `
[dingtalk]
	rawDebug = true
	debug = false
	enableAccessInterceptor = false
	enableAccessInterceptorReply = false
	corpId = "%s"
	agentid = %d 
	appKey = "%s"
	appSecret = "%s"
	oauth2AppKey = "%s"
	oauth2AppSecret = "%s"
	oauth2RedirectUri = "%s"
[redis]
	addr = "127.0.0.1:6379"
`
	aid, _ := strconv.Atoi(os.Getenv("AGENTID"))
	conf = fmt.Sprintf(conf,
		os.Getenv("CORPID"), aid, os.Getenv("APPKEY"), os.Getenv("APPSECRET"),
		os.Getenv("OAUTH2_APPKEY"), os.Getenv("OAUTH2_APPSECRET"), os.Getenv("OAUTH2_REDIRECT_URI"),
	)
	if err := econf.LoadFromReader(strings.NewReader(conf), toml.Unmarshal); err != nil {
		panic("load conf fail," + err.Error())
	}
	cmp := Load("dingtalk").Build(WithERedis(eredis.Load("redis").Build(eredis.WithStub())))
	return cmp
}

func getUid(cmp *Component) string {
	userIds, _ := cmp.UserListID(1)
	return userIds[0]
}

func TestUserListID(t *testing.T) {
	cmp := newCmp()
	user, err := cmp.UserListID(1)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	t.Log("user", user)
}

func TestUserCreateGetUpdateDelete(t *testing.T) {
	cmp := newCmp()
	const userid = "user01"
	// test create user
	uid, err := cmp.UserCreate(UserCreateReq{
		UserId:        userid,
		Name:          "user01",
		Mobile:        os.Getenv("USER_PHONE"),
		HideMobile:    false,
		Telephone:     os.Getenv("USER_PHONE"),
		JobNumber:     "9999",
		Title:         "developer",
		Email:         "user01@gmail.com",
		OrgEmail:      "user01@gmaol.com",
		WorkPlace:     "BeiJing",
		Remark:        "a remark",
		DeptIdList:    []int{1},
		DeptOrderList: nil,
		DeptTitleList: nil,
		SeniorMode:    false,
		HiredDate:     time.Now().Unix(),
	})
	assert.NoError(t, err)
	assert.Equal(t, userid, uid)
	t.Log("user", uid)

	// test update user
	newName := "test02"
	err = cmp.UserUpdate(NewUserUpdateReq(uid).SetName(newName))
	assert.NoError(t, err)

	// test get user
	res, err := cmp.UserGet(userid)
	t.Log("updated user", res)
	assert.NoError(t, err)
	assert.Equal(t, newName, res.Name)

	// test delete user
	err = cmp.UserDelete(userid)
	assert.NoError(t, err)
}

func TestUserList(t *testing.T) {
	cmp := newCmp()
	res, err := cmp.UserList(1, 0, 100)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	t.Log("userList res", res)

	res, err = cmp.UserList(1, 0, 101)
	assert.Error(t, err)
	assert.Nil(t, res)
	t.Log("userList err", err)
}

func TestDepartmentCreateGetUpdateDelete(t *testing.T) {
	cmp := newCmp()

	// test create department
	deptId, err := cmp.DepartmentCreate(DepartmentCreateReq{
		Name:             "dep01",
		ParentId:         1,
		OuterPermitUsers: []string{"manager440"},
		OuterDept:        true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, deptId)
	t.Log("dep", deptId)

	// test update department
	newName := "test02"
	err = cmp.DepartmentUpdate(NewDepartmentUpdateReq(deptId).SetName(newName))
	assert.NoError(t, err)

	// test get department
	res, err := cmp.DepartmentGet(deptId)
	t.Log("updated dep", res)
	assert.NoError(t, err)
	assert.Equal(t, newName, res.Name)
	assert.Equal(t, true, res.OuterDept)
	assert.ElementsMatch(t, []string{"manager440"}, res.OuterPermitUsers)

	// test delete department
	err = cmp.DepartmentDelete(res.DeptId)
	assert.NoError(t, err)
}

func TestComponent_CorpconversationAsyncsendV2(t *testing.T) {
	cmp := newCmp()
	// 发送文本消息
	text := &Text{
		Content: "这是一段文本消息",
	}
	msg := &Msg{
		Msgtype: MsgText,
		Text:    text,
	}
	res, err := cmp.CorpconversationAsyncsendV2(CorpconversationAsyncsendV2Req{
		Msg:        msg,
		UseridList: "xxxxxxx",
	})
	assert.NoError(t, err)
	assert.Equal(t, res.ErrCode, 0)
	// 发送 链接消息
	link := &Link{
		PicURL:     "xxxxx",
		MessageURL: "xxx",
		Text:       "这是一段链接消息text",
		Title:      "这是一段链接消息title",
	}
	msg = &Msg{
		Msgtype: MsgLink,
		Link:    link,
	}
	res, err = cmp.CorpconversationAsyncsendV2(CorpconversationAsyncsendV2Req{
		Msg:        msg,
		UseridList: "xxxxxx",
	})
	assert.NoError(t, err)
	assert.Equal(t, res.ErrCode, 0)
	// 其他方式的消息 ........
}
