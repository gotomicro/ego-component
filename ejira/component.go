package ejira

import (
	"github.com/gotomicro/ego/client/ehttp"
	"github.com/gotomicro/ego/core/elog"
)

// PackageName 包名
const PackageName = "component.ejira"

// Component Component
type Component struct {
	config *config
	ehttp  *ehttp.Component
	logger *elog.Component
}

// newComponent newComponent
func newComponent(compName string, config *config, logger *elog.Component) *Component {
	ehttpClient := ehttp.DefaultContainer().Build(
		ehttp.WithDebug(config.Debug),
		ehttp.WithRawDebug(config.RawDebug),
		ehttp.WithAddr(config.Addr),
		ehttp.WithReadTimeout(config.ReadTimeout),
		ehttp.WithSlowLogThreshold(config.SlowLogThreshold),
		ehttp.WithEnableAccessInterceptor(config.EnableAccessInterceptor),
		ehttp.WithEnableAccessInterceptorRes(config.EnableAccessInterceptorRes),
	)

	return &Component{
		config: config,
		ehttp:  ehttpClient,
		logger: logger,
	}
}

// Bool return bool pointer
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}
