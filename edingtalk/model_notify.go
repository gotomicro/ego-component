package edingtalk

type SendWorkNotifyMsgRes struct {
	OpenAPIResponse
	TaskID int64 `json:"task_id"` // 创建的异步发送任务ID
}

// SendWorkNotifyMsgReq 具体注释: https://developers.dingtalk.com/document/app/asynchronous-sending-of-enterprise-session-messages
type SendWorkNotifyMsgReq struct {
	Msg        *Msg   `json:"msg,omitempty"`
	ToAllUser  string `json:"to_all_user,omitempty"`  // 是否发送给企业全部用户
	AgentID    int64  `json:"agent_id,omitempty"`     // 无需传递
	DeptIDList string `json:"dept_id_list,omitempty"` // 接收者的部门id列表，最大列表长度20
	UseridList string `json:"userid_list,omitempty"`  // 接收者的userid列表，最大用户列表长度100(多个用户用逗号间隔)
}

type Msg struct {
	Voice      *Voice      `json:"voice,omitempty"`
	Image      *File       `json:"image,omitempty"`
	Oa         *Oa         `json:"oa,omitempty"`
	File       *File       `json:"file,omitempty"`
	ActionCard *ActionCard `json:"action_card,omitempty"`
	Link       *Link       `json:"link,omitempty"`
	Markdown   *Markdown   `json:"markdown,omitempty"`
	Text       *Text       `json:"text,omitempty"`
	Msgtype    string      `json:"msgtype,omitempty"` // 消息类型 必传
}

type ActionCard struct {
	BtnJSONList    *BtnJSONList `json:"btn_json_list,omitempty"`
	SingleURL      string       `json:"single_url,omitempty"`
	BtnOrientation string       `json:"btn_orientation,omitempty"`
	SingleTitle    string       `json:"single_title,omitempty"`
	Markdown       string       `json:"markdown,omitempty"`
	Title          string       `json:"title,omitempty"`
}

type BtnJSONList struct {
	ActionURL string `json:"action_url,omitempty"`
	Title     string `json:"title,omitempty"`
}

type File struct {
	MediaID string `json:"media_id,omitempty"`
}

type Link struct {
	PicURL     string `json:"picUrl,omitempty"`
	MessageURL string `json:"messageUrl,omitempty"`
	Text       string `json:"text,omitempty"`
	Title      string `json:"title,omitempty"`
}

type Markdown struct {
	Text  string `json:"text,omitempty"`
	Title string `json:"title,omitempty"`
}

type Oa struct {
	Head         *Head      `json:"head,omitempty"`
	PCMessageURL string     `json:"pc_message_url,omitempty"`
	StatusBar    *StatusBar `json:"status_bar,omitempty"`
	Body         *Body      `json:"body,omitempty"`
	MessageURL   string     `json:"message_url,omitempty"`
}

type Body struct {
	FileCount string `json:"file_count,omitempty"`
	Image     string `json:"image,omitempty"`
	Form      *Form  `json:"form,omitempty"`
	Author    string `json:"author,omitempty"`
	Rich      *Rich  `json:"rich,omitempty"`
	Title     string `json:"title,omitempty"`
	Content   string `json:"content,omitempty"`
}

type Form struct {
	Value string `json:"value,omitempty"`
	Key   string `json:"key,omitempty"`
}

type Rich struct {
	Unit string `json:"unit,omitempty"`
	Num  string `json:"num,omitempty"`
}

type Head struct {
	Bgcolor string `json:"bgcolor,omitempty"`
	Text    string `json:"text,omitempty"`
}

type StatusBar struct {
	StatusValue string `json:"status_value,omitempty"`
	StatusBg    string `json:"status_bg,omitempty"`
}

type Text struct {
	Content string `json:"content,omitempty"`
}

type Voice struct {
	Duration string `json:"duration,omitempty"`
	MediaID  string `json:"media_id,omitempty"`
}
