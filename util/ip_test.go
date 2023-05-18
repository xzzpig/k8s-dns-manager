package util_test

import (
	"testing"

	"github.com/xzzpig/k8s-dns-manager/util"
)

func TestPublicIP(t *testing.T) {
	ip, err := util.GetPublicIP()
	if err != nil {
		t.Error(err)
	}
	t.Log(ip)
}
