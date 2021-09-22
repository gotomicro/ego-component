package eminio

import (
	"strings"
)

const usEast1 = "us-east-1"

var regionMap = map[string]struct{}{
	usEast1:          {},
	"us-east-2":      {},
	"us-west-1":      {},
	"us-west-2":      {},
	"ca-central-1":   {},
	"eu-west-1":      {},
	"eu-west-2":      {},
	"eu-west-3":      {},
	"eu-central-1":   {},
	"eu-north-1":     {},
	"ap-south-1":     {},
	"ap-southeast-1": {},
	"ap-southeast-2": {},
	"ap-northeast-1": {},
	"ap-northeast-2": {},
	"ap-northeast-3": {},
	"me-south-1":     {},
	"sa-east-1":      {},
	"us-gov-west-1":  {},
	"us-gov-east-1":  {},
	"cn-north-1":     {},
	"cn-northwest-1": {},
}

func checkRegion(region string) bool {
	region = strings.TrimSpace(region)
	_, ok := regionMap[region]
	return ok
}
