package ddns

import (
	"context"
	"errors"
	"time"

	gocache "github.com/patrickmn/go-cache"
	v1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
	"github.com/xzzpig/k8s-dns-manager/pkg/generator"
	"github.com/xzzpig/k8s-dns-manager/util/cip"
)

var ErrNoPublicIP = errors.New("can't get public ip")

type DDNSGenerator struct {
	cache           *gocache.Cache
	refreshInternal time.Duration
	ip              *cip.Cip
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
	return g.refreshInternal
}

const cacheKeyPublicIP = "publicIP"

func (g *DDNSGenerator) GetPublicIP() (string, error) {
	ip, ok := g.cache.Get(cacheKeyPublicIP)
	if ok {
		return ip.(string), nil
	}
	ip = g.ip.MyIPv4()
	if ip == "" {
		return "", ErrNoPublicIP
	}
	g.cache.SetDefault(cacheKeyPublicIP, ip)
	return ip.(string), nil
}

func init() {
	generator.Register("DDNS", func(gfa *generator.GeneratorFactoryArgs) (generator.IDNSGenerator, error) {
		ip := cip.New()
		config := gfa.Spec.DDNS.DeepCopy().WithDefault()
		ip.ApiIPv4 = append(ip.ApiIPv4, config.ExtraApis...)
		ip.MinTimeout = time.Duration(config.Timeout) * time.Second
		return &DDNSGenerator{
			cache:           gocache.New(time.Second*time.Duration(config.CacheExpire), time.Second*time.Duration(config.CleanInterval)),
			refreshInternal: time.Second * time.Duration(config.RefreshInternal),
			ip:              ip,
		}, nil
	})
}
