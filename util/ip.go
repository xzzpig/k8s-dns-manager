package util

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func GetPublicIP() (ip string, err error) {
	ip, err = GetPublicIP1()
	if err == nil {
		return
	}
	ip, err = GetPublicIP2()
	if err == nil {
		return
	}
	ip, err = GetPublicIP3()
	if err == nil {
		return
	}
	return
}

func GetPublicIP1() (string, error) {
	resp, err := http.Get("https://v6r.ipip.net/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("Error Status" + resp.Status)
	}
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

func GetPublicIP2() (string, error) {
	type PublicIP struct {
		Address  string `json:"address"`
		Code     int64  `json:"code"`
		IP       string `json:"ip"`
		IsDomain int64  `json:"isDomain"`
		Rs       int64  `json:"rs"`
	}
	resp, err := http.Get("https://ip.cn/api/index?ip=&type=0")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("Error Status" + resp.Status)
	}
	body, _ := io.ReadAll(resp.Body)
	// return string(body), nil
	var publicIp PublicIP
	json.Unmarshal(body, &publicIp)
	return publicIp.IP, nil
}

func GetPublicIP3() (string, error) {
	type PublicIP struct {
		Query string `json:"query"`
	}
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	var ip PublicIP
	json.Unmarshal(body, &ip)

	return ip.Query, nil
}
