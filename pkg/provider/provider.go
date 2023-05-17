package provider

import (
	"context"
	"errors"

	dnsv1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
)

type IDNSProvider interface {
	SearchRecord(ctx context.Context, record *dnsv1.DNSRecord) (id string, ok bool, err error)
	CreateRecord(ctx context.Context, record *dnsv1.DNSRecord) (id string, err error)
	UpdateRecord(ctx context.Context, record *dnsv1.DNSRecord, id *string) (err error)
	DeleteRecord(ctx context.Context, record *dnsv1.DNSRecord, id *string) (err error)
}

type DNSProviderFactoryArgs struct {
	Spec *dnsv1.DNSProviderSpec
	Ctx  context.Context
}

type DNSProviderFactory func(*DNSProviderFactoryArgs) (IDNSProvider, error)

var providers = map[string]DNSProviderFactory{}
var ErrProviderNotFound = errors.New("provider not found")

func Register(name string, factory DNSProviderFactory) {
	providers[name] = factory
}

func New(ctx context.Context, provider *dnsv1.DNSProviderSpec) (IDNSProvider, error) {
	if factory, ok := providers[string(provider.ProviderType)]; ok {
		return factory(&DNSProviderFactoryArgs{
			Spec: provider,
			Ctx:  ctx,
		})
	}
	return nil, ErrProviderNotFound
}

func RegistedProviders() []string {
	var names []string
	for name := range providers {
		names = append(names, name)
	}
	return names
}
