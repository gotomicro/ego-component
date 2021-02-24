package miniprogram

import (
	"encoding/json"
	"fmt"

	"github.com/gotomicro/ego-component/ewechat/util"
)

const (
	getImgSecCheckURL = "https://api.weixin.qq.com/wxa/img_sec_check?access_token=%s"
	getMediaCheckURL  = "https://api.weixin.qq.com/wxa/media_check_async?access_token=%s"
	getMsgSecCheckURL = "https://api.weixin.qq.com/wxa/msg_sec_check?access_token=%s"
)

// ResAnalysisRetain 小程序留存数据返回
type ResSecurity struct {
	util.CommonError
}

// CheckImg 检测图片
// 文档地址： https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/sec-check/security.imgSecCheck.html
func (wxa *MiniProgram) CheckImg(fileName string) (response ResSecurity, err error) {
	var accessToken string
	accessToken, err = wxa.GetAccessToken()
	if err != nil {
		return
	}

	var info []byte
	info, err = wxa.Context.PostFile("media", fileName, fmt.Sprintf(getImgSecCheckURL, accessToken))
	if err != nil {
		return
	}
	err = json.Unmarshal(info, &response)
	return
}

//// CheckMedia 检测Media
//// 文档地址： https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/sec-check/security.imgSecCheck.html
//func (wxa *MiniProgram) CheckMedia(fileName string) (response ResSecurity, err error) {
//	var info  []byte
//	info,err = wxa.ctx.PostFile("media",fileName,getImgSecCheckURL)
//	if err != nil {
//		return
//	}
//	err = json.Unmarshal(info,&response)
//	return
//}

type ReqCheckMsg struct {
	Content string `json:"content"`
}

// CheckMsg 检测敏感词
// 文档地址： https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/sec-check/security.msgSecCheck.html
func (wxa *MiniProgram) CheckMsg(content string) (response ResSecurity, err error) {
	var accessToken string
	accessToken, err = wxa.GetAccessToken()
	if err != nil {
		return
	}

	var info []byte
	req := ReqCheckMsg{
		Content: content,
	}

	info, err = wxa.Context.PostJSON(fmt.Sprintf(getMsgSecCheckURL, accessToken), req)
	if err != nil {
		return
	}
	err = json.Unmarshal(info, &response)
	return
}
