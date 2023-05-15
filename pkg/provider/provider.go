package provider

import (
	"errors"

	dnsv1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
)

type IDNSProvider interface {
	SearchRecord(record *dnsv1.DNSRecord) (id string, ok bool, err error)
	CreateRecord(record *dnsv1.DNSRecord) (id string, err error)
	UpdateRecord(record *dnsv1.DNSRecord, id *string) (err error)
	DeleteRecord(record *dnsv1.DNSRecord, id *string) (err error)
}

type DNSProviderFactory func(*dnsv1.DNSProviderSpec) (IDNSProvider, error)

var providers = map[string]DNSProviderFactory{}
var ErrProviderNotFound = errors.New("provider not found")

func Register(name string, factory DNSProviderFactory) {
	providers[name] = factory
}

func New(provider *dnsv1.DNSProviderSpec) (IDNSProvider, error) {
	if factory, ok := providers[string(provider.ProviderType)]; ok {
		return factory(provider)
	}
	return nil, ErrProviderNotFound
}
