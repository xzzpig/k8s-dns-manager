package util_test

import (
	"testing"

	"github.com/xzzpig/k8s-dns-manager/util"
)

func TestPublicIP1(t *testing.T) {
	ip, err := util.GetPublicIP1()
	if err != nil {
		t.Error(err)
	}
	t.Log(ip)
}

func TestPublicIP2(t *testing.T) {
	ip, err := util.GetPublicIP2()
	if err != nil {
		t.Error(err)
	}
	t.Log(ip)
}

func TestPublicIP3(t *testing.T) {
	ip, err := util.GetPublicIP3()
	if err != nil {
		t.Error(err)
	}
	t.Log(ip)
}

func TestPublicIP(t *testing.T) {
	ip, err := util.GetPublicIP()
	if err != nil {
		t.Error(err)
	}
	t.Log(ip)
}
