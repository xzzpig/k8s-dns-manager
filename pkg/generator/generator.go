package generator

import (
	"context"
	"errors"
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

var (
	ErrGeneratorNotFound = errors.New("generator not found")
)

type IDNSGenerator interface {
	Generate(ctx context.Context, source DNSGeneratorSource) ([]dnsv1.DNSRecordSpec, error)
	Support(source DNSGeneratorSource) bool
	RequeueAfter(ctx context.Context, source DNSGeneratorSource) time.Duration
}

type GeneratorFactoryArgs struct {
	Spec *dnsv1.DNSGeneratorSpec
	Ctx  context.Context
}

type GeneratorFactory func(*GeneratorFactoryArgs) (IDNSGenerator, error)

var generatorFactorys = map[string]GeneratorFactory{}
var generators = map[string]IDNSGenerator{}

func Register(name string, factory GeneratorFactory) {
	generatorFactorys[name] = factory
}

func New(generator *dnsv1.DNSGenerator, ctx context.Context) error {
	factory, ok := generatorFactorys[string(generator.Spec.GeneratorType)]
	if !ok {
		return ErrGeneratorNotFound
	}
	g, err := factory(&GeneratorFactoryArgs{
		Spec: &generator.Spec,
		Ctx:  ctx,
	})
	if err != nil {
		return err
	}
	generators[generator.Name] = g
	return nil
}

func Get(name string) IDNSGenerator {
	return generators[name]
}

func GetIngress(ctx context.Context) *netv1.Ingress {
	return ctx.Value(ContextKeyIngress).(*netv1.Ingress)
}

type ShowResultFunc = func(reason string, message string, err error)

func GetShowResultFunc(ctx context.Context) ShowResultFunc {
	return ctx.Value(ContextKeyShowResultFunc).(ShowResultFunc)
}

func RegistedFactories() []string {
	var names []string
	for name := range generatorFactorys {
		names = append(names, name)
	}
	return names
}
