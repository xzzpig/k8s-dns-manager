package cip

import (
	"regexp"
	"time"
)

type Cip struct {
	MinTimeout time.Duration
	ApiIPv4    []string
	ApiIPv6    []string
}

const RegxIPv4 = `(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)`

const RegxIPv6 = `([0-9A-Fa-f]{0,4}:){2,7}([0-9A-Fa-f]{1,4}$|((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4})`

func New() (cip *Cip) {
	cip = new(Cip)
	cip.MinTimeout = 1000 * time.Millisecond
	cip.ApiIPv4 = []string{
		"http://www.net.cn/static/customercare/yourip.asp", "http://ddns.oray.com/checkip", "http://speedtest.ecnu.edu.cn/getIP.php",
		"http://members.3322.org/dyndns/getip", "http://ifconfig.cc", "http://cip.cc", "https://v6r.ipip.net",
		"http://pv.sohu.com/cityjson?ie=utf-8", "http://whois.pconline.com.cn/ipJson.jsp",
		"http://ipba.cc", "http://v4.myip.la", "https://api.ipify.org", "http://ip-api.com", "http://whatismyip.akamai.com", "https://ip.cn/api/index?ip=&type=0",
	}
	cip.ApiIPv6 = []string{
		"http://speed.neu6.edu.cn/getIP.php", "http://v6.myip.la", "https://api64.ipify.org", "http://speedtest6.ecnu.edu.cn/getIP.php",
		"http://ip6only.me/api/", "http://v6.ipv6-test.com/api/myip.php", "https://v6.ident.me",
	}
	return
}

func (cip *Cip) MyIPv4() (ip string) {
	regx := regexp.MustCompile(RegxIPv4)
	return cip.FastWGetWithVailder(cip.ApiIPv4, func(s string) string {
		return regx.FindString((s))
	})
}

func (cip *Cip) MyIPv6() (ip string) {
	regx := regexp.MustCompile(RegxIPv6)
	return cip.FastWGetWithVailder(cip.ApiIPv6, func(s string) string {
		return regx.FindString((s))
	})
}

func (cip *Cip) FastWGetWithVailder(ipAPI []string, vailder func(string) string) (ip string) {
	var (
		length   = len(ipAPI)
		ipMap    = make(map[string]int, length/5)
		cchan    = make(chan string, length/2)
		maxCount = -1
	)
	for _, url := range ipAPI {
		go func(url string) {
			cchan <- vailder(wGet(url, cip.MinTimeout))
		}(url)
	}
	for i := 0; i < length; i++ {
		v := <-cchan
		if len(v) == 0 {
			continue
		}
		if ipMap[v]++; ipMap[v] >= length/2 {
			return v
		}
	}
	for k, v := range ipMap {
		if v > maxCount {
			maxCount = v
			ip = k
		}
	}

	// Use First ipAPI as failsafe
	if len(ip) == 0 {
		ip = vailder(wGet(ipAPI[0], 5*cip.MinTimeout))
	}
	return
}
