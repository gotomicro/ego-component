package miniprogram

import (
	"github.com/gotomicro/ego-component/ewechat/context"
)

// MiniProgram struct extends context
type MiniProgram struct {
	*context.Context
}

// NewMiniProgram 实例化小程序接口
func NewMiniProgram(context *context.Context) *MiniProgram {
	miniProgram := new(MiniProgram)
	miniProgram.Context = context
	return miniProgram
}
