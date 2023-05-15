package ddns

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
	v1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
	"github.com/xzzpig/k8s-dns-manager/pkg/generator"
	"github.com/xzzpig/k8s-dns-manager/util"
)

type DDNSGenerator struct {
	cache *gocache.Cache
}

func (g *DDNSGenerator) Generate(ctx context.Context, source generator.DNSGeneratorSource) ([]v1.DNSRecordSpec, error) {
	records := []v1.DNSRecordSpec{}
	if source == generator.DNSGeneratorSourceIngress {
		ip, err := g.GetPublicIP()
		if err != nil {
			return nil, err
		}
		ingress := generator.GetIngress(ctx)
		for _, rule := range ingress.Spec.Rules {
			records = append(records, v1.DNSRecordSpec{
				RecordType: v1.DNSRecordTypeA,
				Name:       rule.Host,
				Value:      ip,
			})
		}
	}
	return records, nil
}

func (g *DDNSGenerator) Support(source generator.DNSGeneratorSource) bool {
	return source == generator.DNSGeneratorSourceIngress
}

func (g *DDNSGenerator) RequeueAfter(ctx context.Context, source generator.DNSGeneratorSource) time.Duration {
	return time.Minute
}

const cacheKeyPublicIP = "publicIP"

func (g *DDNSGenerator) GetPublicIP() (string, error) {
	ip, ok := g.cache.Get(cacheKeyPublicIP)
	if ok {
		return ip.(string), nil
	}
	ip, err := util.GetPublicIP()
	if err != nil {
		return "", err
	}
	g.cache.SetDefault(cacheKeyPublicIP, ip)
	return ip.(string), nil
}

func init() {
	generator.Register("ddns", &DDNSGenerator{
		cache: gocache.New(time.Minute, time.Minute/2),
	})
}
