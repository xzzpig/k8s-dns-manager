package cname

import (
	"context"
	"errors"
	"time"

	v1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
	"github.com/xzzpig/k8s-dns-manager/pkg/generator"
)

const (
	AnnotationKeyCname = "dns.xzzpig.com/cname"
)

type CNameGenerator struct {
	Value string
}

func (g *CNameGenerator) Generate(ctx context.Context, source generator.DNSGeneratorSource) ([]v1.DNSRecordSpec, error) {
	records := []v1.DNSRecordSpec{}
	if source == generator.DNSGeneratorSourceIngress {
		ingress := generator.GetIngress(ctx)
		cnameValue, ok := ingress.Annotations[AnnotationKeyCname]
		if !ok {
			cnameValue = g.Value
		}
		if cnameValue == "" {
			showResult := generator.GetShowResultFunc(ctx)
			showResult("Error", "generate cname record error, annotation "+AnnotationKeyCname+" not found", errors.New("cname annotation not found"))
			return nil, nil
		}
		for _, rule := range ingress.Spec.Rules {
			records = append(records, v1.DNSRecordSpec{
				RecordType: v1.DNSRecordTypeCNAME,
				Name:       rule.Host,
				Value:      cnameValue,
			})
		}
	}
	return records, nil
}

func (g *CNameGenerator) Support(source generator.DNSGeneratorSource) bool {
	return source == generator.DNSGeneratorSourceIngress
}

func (g *CNameGenerator) RequeueAfter(ctx context.Context, source generator.DNSGeneratorSource) time.Duration {
	return 0
}

func init() {
	// generator.Register("CNAME", &CNameGenerator{})
	generator.Register("CNAME", func(gfa *generator.GeneratorFactoryArgs) (generator.IDNSGenerator, error) {
		return &CNameGenerator{
			Value: gfa.Spec.CNAME.Value,
		}, nil
	})
}
