package util

import (
	"errors"

	"github.com/xzzpig/k8s-dns-manager/pkg/config"
	"github.com/xzzpig/k8s-dns-manager/util/cip"
)

var ErrNoPublicIP = errors.New("can't get public ip")

func GetPublicIP() (ip string, err error) {
	ip = cip.MyIPv4()
	if ip == "" {
		err = ErrNoPublicIP
	}
	return
}

func init() {
	cip.MinTimeout = config.GetConfig().Default.Generator.DDNS.Timeout
	cip.ApiIPv4 = append(cip.ApiIPv4, config.GetConfig().Default.Generator.DDNS.ExtraApis...)
}
