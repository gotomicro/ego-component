package edingtalk

import (
	"strconv"
	"strings"
)

type Ints []int

// MarshalJSON 覆盖[]int 的MarshalJSON()方法
func (i Ints) MarshalJSON() ([]byte, error) {
	return []byte(`"` + IntsJoin(i, ",") + `"`), nil
}

type Strings []string

// MarshalJSON 覆盖[]string 的MarshalJSON()方法
func (s Strings) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strings.Join(s, ",") + `"`), nil
}

// IntsJoin []int{"a","b"} => `"a,b"`
func IntsJoin(a []int, sep string) string {
	b := make([]string, len(a), len(a))
	for i, v := range a {
		b[i] = strconv.Itoa(v)
	}
	return strings.Join(b, sep)
}

type User struct {
	UserId        string              `json:"userid,omitempty"`
	Name          string              `json:"name,omitempty"`
	Mobile        string              `json:"mobile,omitempty"`
	HideMobile    bool                `json:"hide_mobile,omitempty"`
	Telephone     string              `json:"telephone,omitempty"`
	JobNumber     string              `json:"job_number,omitempty"`
	Title         string              `json:"title,omitempty"`
	Email         string              `json:"email,omitempty"`     // 员工私人邮箱
	OrgEmail      string              `json:"org_email,omitempty"` // 员工企业邮箱
	WorkPlace     string              `json:"work_place,omitempty"`
	Remark        string              `json:"remark,omitempty"`
	DeptIdList    Ints                `json:"dept_id_list,omitempty"`
	DeptOrderList []UserDeptOrderList `json:"dept_order_list,omitempty"`
	DeptTitleList []UserDeptTitleList `json:"dept_title_list,omitempty"`
	SeniorMode    bool                `json:"senior_mode,omitempty"`
	HiredDate     int64               `json:"hired_date,omitempty"`
}

type UserDeptOrderList struct {
	DeptId int `json:"dept_id"`
	Order  int `json:"order"`
}

type UserDeptTitleList struct {
	DeptId int `json:"dept_id"`
	Title  int `json:"title"`
}

type userGetRes struct {
	OpenAPIResponse
	Result User `json:"result"`
}

type UserCreateReq = User

type userCreateRes struct {
	OpenAPIResponse
	Result struct {
		UserId string `json:"userid"`
	} `json:"result"`
}

type UserUpdateReq struct {
	UserId        string               `json:"userid,omitempty"`
	Name          *string              `json:"name,omitempty"`
	Mobile        *string              `json:"mobile,omitempty"`
	HideMobile    *bool                `json:"hide_mobile,omitempty"`
	Telephone     *string              `json:"telephone,omitempty"`
	JobNumber     *string              `json:"job_number,omitempty"`
	Title         *string              `json:"title,omitempty"`
	Email         *string              `json:"email,omitempty"`     // 员工私人邮箱
	OrgEmail      *string              `json:"org_email,omitempty"` // 员工企业邮箱
	WorkPlace     *string              `json:"work_place,omitempty"`
	Remark        *string              `json:"remark,omitempty"`
	DeptIdList    *Ints                `json:"dept_id_list,omitempty"`
	DeptOrderList *[]UserDeptOrderList `json:"dept_order_list,omitempty"`
	DeptTitleList *[]UserDeptTitleList `json:"dept_title_list,omitempty"`
	SeniorMode    *bool                `json:"senior_mode,omitempty"`
	HiredDate     *int64               `json:"hired_date,omitempty"`
}

func NewUserUpdateReq(uid string) *UserUpdateReq {
	return &UserUpdateReq{UserId: uid}
}

func (u *UserUpdateReq) SetName(name string) *UserUpdateReq {
	u.Name = &name
	return u
}

func (u *UserUpdateReq) SetMobile(mobile string) *UserUpdateReq {
	u.Mobile = &mobile
	return u
}

func (u *UserUpdateReq) SetHideMobile(hideMobile bool) *UserUpdateReq {
	u.HideMobile = &hideMobile
	return u
}

func (u *UserUpdateReq) SetTelephone(telephone string) *UserUpdateReq {
	u.Telephone = &telephone
	return u
}

func (u *UserUpdateReq) SetJobNumber(jobNumber string) *UserUpdateReq {
	u.JobNumber = &jobNumber
	return u
}

func (u *UserUpdateReq) SetTitle(title string) *UserUpdateReq {
	u.Title = &title
	return u
}

func (u *UserUpdateReq) SetEmail(email string) *UserUpdateReq {
	u.Email = &email
	return u
}

func (u *UserUpdateReq) SetOrgEmail(orgEmail string) *UserUpdateReq {
	u.OrgEmail = &orgEmail
	return u
}

func (u *UserUpdateReq) SetWorkPlace(workPlace string) *UserUpdateReq {
	u.WorkPlace = &workPlace
	return u
}

func (u *UserUpdateReq) SetRemark(remark string) *UserUpdateReq {
	u.Remark = &remark
	return u
}

func (u *UserUpdateReq) SetDeptIdList(deptIdList []int) *UserUpdateReq {
	u.DeptIdList = (*Ints)(&deptIdList)
	return u
}

func (u *UserUpdateReq) SetDeptOrderList(deptOrderList []UserDeptOrderList) *UserUpdateReq {
	u.DeptOrderList = &deptOrderList
	return u
}

func (u *UserUpdateReq) SetDeptTitleList(deptTitleList []UserDeptTitleList) *UserUpdateReq {
	u.DeptTitleList = &deptTitleList
	return u
}

func (u *UserUpdateReq) SetSeniorMode(seniorMode bool) *UserUpdateReq {
	u.SeniorMode = &seniorMode
	return u
}

func (u *UserUpdateReq) SetHiredDate(hiredDate int64) *UserUpdateReq {
	u.HiredDate = &hiredDate
	return u
}

type userListIDRes struct {
	OpenAPIResponse
	Result struct {
		UserIDList []string `json:"userid_list"`
	} `json:"result"`
}

type userListRes struct {
	OpenAPIResponse
	Result *UserListRes `json:"result"`
}

type UserListRes struct {
	HasMore    bool   `json:"has_more"`
	NextCursor int    `json:"next_cursor"`
	List       []User `json:"list"`
}
