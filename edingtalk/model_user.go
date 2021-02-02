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

type userCreateReq = User

type userCreateRes struct {
	OpenAPIResponse
	Result struct {
		UserId string `json:"userid"`
	} `json:"result"`
}

type userUpdateReq struct {
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

func NewUserUpdateReq(uid string) *userUpdateReq {
	return &userUpdateReq{UserId: uid}
}

func (u *userUpdateReq) SetName(name string) *userUpdateReq {
	u.Name = &name
	return u
}

func (u *userUpdateReq) SetMobile(mobile string) *userUpdateReq {
	u.Mobile = &mobile
	return u
}

func (u *userUpdateReq) SetHideMobile(hideMobile bool) *userUpdateReq {
	u.HideMobile = &hideMobile
	return u
}

func (u *userUpdateReq) SetTelephone(telephone string) *userUpdateReq {
	u.Telephone = &telephone
	return u
}

func (u *userUpdateReq) SetJobNumber(jobNumber string) *userUpdateReq {
	u.JobNumber = &jobNumber
	return u
}

func (u *userUpdateReq) SetTitle(title string) *userUpdateReq {
	u.Title = &title
	return u
}

func (u *userUpdateReq) SetEmail(email string) *userUpdateReq {
	u.Email = &email
	return u
}

func (u *userUpdateReq) SetOrgEmail(orgEmail string) *userUpdateReq {
	u.OrgEmail = &orgEmail
	return u
}

func (u *userUpdateReq) SetWorkPlace(workPlace string) *userUpdateReq {
	u.WorkPlace = &workPlace
	return u
}

func (u *userUpdateReq) SetRemark(remark string) *userUpdateReq {
	u.Remark = &remark
	return u
}

func (u *userUpdateReq) SetDeptIdList(deptIdList []int) *userUpdateReq {
	u.DeptIdList = (*Ints)(&deptIdList)
	return u
}

func (u *userUpdateReq) SetDeptOrderList(deptOrderList []UserDeptOrderList) *userUpdateReq {
	u.DeptOrderList = &deptOrderList
	return u
}
func (u *userUpdateReq) SetDeptTitleList(deptTitleList []UserDeptTitleList) *userUpdateReq {
	u.DeptTitleList = &deptTitleList
	return u
}
func (u *userUpdateReq) SetSeniorMode(seniorMode bool) *userUpdateReq {
	u.SeniorMode = &seniorMode
	return u
}
func (u *userUpdateReq) SetHiredDate(hiredDate int64) *userUpdateReq {
	u.HiredDate = &hiredDate
	return u
}

type userListIDRes struct {
	OpenAPIResponse
	Result struct {
		UserIDList []string `json:"userid_list"`
	} `json:"result"`
}
