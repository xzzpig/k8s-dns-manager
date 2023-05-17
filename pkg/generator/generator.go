package generator

import (
	"context"
	"time"

	dnsv1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
	netv1 "k8s.io/api/networking/v1"
)

const (
	AnnotationKeyGenerator    = "dns.xzzpig.com/generator"
	AnnotationKeyRecordPrefix = "dns.xzzpig.com/record-"
)

type DNSGeneratorSource string

const (
	DNSGeneratorSourceIngress DNSGeneratorSource = "ingress"
)

type ContextKey string

const (
	ContextKeyIngress        ContextKey = "ingress"
	ContextKeyShowResultFunc ContextKey = "showResultFunc"
)

type DNSGenerator interface {
	Generate(ctx context.Context, source DNSGeneratorSource) ([]dnsv1.DNSRecordSpec, error)
	Support(source DNSGeneratorSource) bool
	RequeueAfter(ctx context.Context, source DNSGeneratorSource) time.Duration
}

var generators = map[string]DNSGenerator{}

func Register(name string, generator DNSGenerator) {
	generators[name] = generator
}

func Get(name string) DNSGenerator {
	return generators[name]
}

func GetIngress(ctx context.Context) *netv1.Ingress {
	return ctx.Value(ContextKeyIngress).(*netv1.Ingress)
}

type ShowResultFunc = func(reason string, message string, err error)

func GetShowResultFunc(ctx context.Context) ShowResultFunc {
	return ctx.Value(ContextKeyShowResultFunc).(ShowResultFunc)
}

func RegistedGenerators() []string {
	var names []string
	for name := range generators {
		names = append(names, name)
	}
	return names
}
