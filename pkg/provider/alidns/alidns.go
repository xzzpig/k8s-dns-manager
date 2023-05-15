package alidns

import (
	dnsv1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
	"github.com/xzzpig/k8s-dns-manager/pkg/provider"
	"github.com/xzzpig/k8s-dns-manager/util"
)

type AliDNSProvider struct {
	util *util.AliDNSUtils
	spec *dnsv1.DNSProviderSpec
}

func (p *AliDNSProvider) SearchRecord(rec *dnsv1.DNSRecord) (id string, ok bool, err error) {
	rr := rec.Spec.RR(p.spec)
	record, err := p.util.FindRecordByRR(rr)
	if err != nil {
		return "", false, err
	}
	if record == nil {
		return "", false, nil
	}
	return *record.RecordId, true, nil
}

func (p *AliDNSProvider) CreateRecord(rec *dnsv1.DNSRecord) (id string, err error) {
	rr := rec.Spec.RR(p.spec)
	return p.util.CreateRecord(rr, rec.Spec.Value, string(rec.Spec.RecordType))
}

func (p *AliDNSProvider) UpdateRecord(rec *dnsv1.DNSRecord, id *string) (err error) {
	record, err := p.util.FindRecordById(*id)
	if err != nil {
		return err
	}
	rr := rec.Spec.RR(p.spec)
	if *record.RR == rr && *record.Type == string(rec.Spec.RecordType) && *record.Value == rec.Spec.Value {
		return nil
	}
	return p.util.UpdateRecord(*id, rr, rec.Spec.Value, string(rec.Spec.RecordType))
}

func (p *AliDNSProvider) DeleteRecord(rec *dnsv1.DNSRecord, id *string) (err error) {
	return p.util.DeleteRecord(*id)
}

func init() {
	// fmt.Println("init alidns provider")
	provider.Register(string(dnsv1.DNSProviderTypeAliyun), func(spec *dnsv1.DNSProviderSpec) (provider.IDNSProvider, error) {
		dnsutil, err := util.NewAliDnsUtils(util.AliDnsAccount{
			AccessKeyID:     spec.Aliyun.AccessKeyID,
			AccessKeySecret: spec.Aliyun.AccessKeySecret,
			DomainName:      spec.DomainName,
		})
		if err != nil {
			return nil, err
		}
		return &AliDNSProvider{
			util: dnsutil,
			spec: spec,
		}, nil
	})
}
