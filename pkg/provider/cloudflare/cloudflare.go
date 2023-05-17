package cloudflare

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	dnsv1 "github.com/xzzpig/k8s-dns-manager/api/dns/v1"
	"github.com/xzzpig/k8s-dns-manager/pkg/provider"
)

const (
	AnnotationKeyProxied = "dns.xzzpig.com/record-proxied"
)

type CloudflareProvider struct {
	spec *dnsv1.DNSProviderSpec
	api  *cloudflare.API
	zone *cloudflare.Zone
}

func (p *CloudflareProvider) SearchRecord(ctx context.Context, rec *dnsv1.DNSRecord) (id string, ok bool, err error) {
	if rec.Status.RecordID != "" {
		record, err := p.api.GetDNSRecord(ctx, cloudflare.ZoneIdentifier(p.zone.ID), rec.Status.RecordID)
		if err == nil {
			return record.ID, true, nil
		}
	}

	records, _, err := p.api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(p.zone.ID), cloudflare.ListDNSRecordsParams{
		Name: rec.Spec.Name,
	})
	if err != nil {
		return "", false, err
	}
	if len(records) == 0 {
		return "", false, nil
	}
	rec.Status.RecordID = records[0].ID
	return rec.Status.RecordID, true, nil
}

func (p *CloudflareProvider) proxied(rec *dnsv1.DNSRecord) bool {
	if rec == nil {
		return false
	}
	proxied := rec.Annotations[AnnotationKeyProxied]
	if proxied == "" {
		return p.spec.Cloudflare.Proxied
	} else {
		return proxied == "true"
	}
}

func (p *CloudflareProvider) CreateRecord(ctx context.Context, rec *dnsv1.DNSRecord) (id string, err error) {
	record, err := p.api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(p.zone.ID), cloudflare.CreateDNSRecordParams{
		Type:    string(rec.Spec.RecordType),
		Name:    rec.Spec.Name,
		Content: rec.Spec.Value,
		Proxied: cloudflare.BoolPtr(p.proxied(rec)),
	})
	if err != nil {
		return "", err
	}
	return record.ID, nil
}

func (p *CloudflareProvider) UpdateRecord(ctx context.Context, rec *dnsv1.DNSRecord, id *string) (err error) {
	record, err := p.api.GetDNSRecord(ctx, cloudflare.ZoneIdentifier(p.zone.ID), *id)
	if err != nil {
		return err
	}
	proxied := p.proxied(rec)
	if record.Name == rec.Spec.Name && record.Type == string(rec.Spec.RecordType) && record.Content == rec.Spec.Value && *record.Proxied == proxied {
		return nil
	}
	_, err = p.api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(p.zone.ID), cloudflare.UpdateDNSRecordParams{
		ID:      *id,
		Type:    string(rec.Spec.RecordType),
		Name:    rec.Spec.Name,
		Content: rec.Spec.Value,
		Proxied: cloudflare.BoolPtr(proxied),
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *CloudflareProvider) DeleteRecord(ctx context.Context, rec *dnsv1.DNSRecord, id *string) (err error) {
	return p.api.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(p.zone.ID), *id)
}

func init() {
	provider.Register(string(dnsv1.DNSProviderTypeCloudflare), func(args *provider.DNSProviderFactoryArgs) (provider.IDNSProvider, error) {
		spec := args.Spec
		p := &CloudflareProvider{spec: spec}
		if spec.Cloudflare.APIToken != "" {
			api, err := cloudflare.NewWithAPIToken(spec.Cloudflare.APIToken)
			if err != nil {
				return nil, err
			}
			p.api = api
		} else if spec.Cloudflare.Key != "" && spec.Cloudflare.Email != "" {
			api, err := cloudflare.New(spec.Cloudflare.Key, spec.Cloudflare.Email)
			if err != nil {
				return nil, err
			}
			p.api = api
		} else {
			return nil, fmt.Errorf("cloudflare api token or key and email is required")
		}

		zoneName := spec.Cloudflare.ZoneName
		if zoneName == "" {
			zoneName = spec.DomainName
		}
		zoneID, err := p.api.ZoneIDByName(zoneName)
		if err != nil {
			return nil, err
		}

		zone, err := p.api.ZoneDetails(args.Ctx, zoneID)
		if err != nil {
			return nil, err
		}
		p.zone = &zone

		return p, nil
	})
}
