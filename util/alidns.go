package util

import (
	alidns "github.com/alibabacloud-go/alidns-20150109/client"
	openapi "github.com/alibabacloud-go/tea-rpc/client"
	"github.com/alibabacloud-go/tea/tea"
)

/**
 * 使用AK&SK初始化账号Client
 * @param accessKeyId
 * @param accessKeySecret
 * @return Client
 * @throws Exception
 */
func CreateClient(accessKeyId *string, accessKeySecret *string) (_result *alidns.Client, _err error) {
	config := &openapi.Config{}
	// 您的AccessKey ID
	config.AccessKeyId = accessKeyId
	// 您的AccessKey Secret
	config.AccessKeySecret = accessKeySecret
	// 访问的域名
	config.Endpoint = tea.String("dns.aliyuncs.com")
	_result, _err = alidns.NewClient(config)
	return _result, _err
}

type AliDnsAccount struct {
	AccessKeyID     string `json:"accessKey-id"`
	AccessKeySecret string `json:"accessKey-secret"`
	DomainName      string `json:"domain-name"`
}

type AliDNSUtils struct {
	account AliDnsAccount
	client  *alidns.Client
}

func NewAliDnsUtils(account AliDnsAccount) (*AliDNSUtils, error) {
	dnsUtils := AliDNSUtils{
		account: account,
	}
	client, err := CreateClient(tea.String(account.AccessKeyID), tea.String(account.AccessKeySecret))
	if err != nil {
		return nil, err
	}
	dnsUtils.client = client
	return &dnsUtils, nil
}

func (dns *AliDNSUtils) GetRecordCount() (int64, error) {
	resp, err := dns.client.DescribeDomainRecords(&alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(dns.account.DomainName),
		PageSize:   tea.Int64(1),
	})
	if err != nil {
		return 0, err
	}
	return *resp.TotalCount, nil
}

func (dns *AliDNSUtils) ListRecords() ([]*alidns.DescribeDomainRecordsResponseDomainRecordsRecord, error) {
	totalCount, err := dns.GetRecordCount()
	if err != nil {
		return nil, err
	}
	resp, err := dns.client.DescribeDomainRecords(&alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(dns.account.DomainName),
		PageSize:   tea.Int64(totalCount),
	})
	if err != nil {
		return nil, err
	}
	return resp.DomainRecords.Record, nil
}

func (dns *AliDNSUtils) FindRecordById(id string) (*alidns.DescribeDomainRecordInfoResponse, error) {
	resp, err := dns.client.DescribeDomainRecordInfo(&alidns.DescribeDomainRecordInfoRequest{
		RecordId: tea.String(id),
	})
	if err != nil {
		return nil, err
	}
	return resp, nil

}

func (dns *AliDNSUtils) FindRecordByRR(rr string) (*alidns.DescribeDomainRecordsResponseDomainRecordsRecord, error) {
	records, err := dns.ListRecords()
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if *record.RR == rr {
			return record, nil
		}
	}
	return nil, nil
}

func (dns *AliDNSUtils) CreateRecord(RR string, Value string, Type string) (string, error) {
	resp, err := dns.client.AddDomainRecord(&alidns.AddDomainRecordRequest{
		DomainName: tea.String(dns.account.DomainName),
		RR:         tea.String(RR),
		Type:       tea.String(Type),
		Value:      tea.String(Value),
	})
	if err != nil {
		return "", err
	}
	return *resp.RecordId, nil
}

func (dns *AliDNSUtils) DeleteRecord(recordId string) error {
	_, err := dns.client.DeleteDomainRecord(&alidns.DeleteDomainRecordRequest{
		RecordId: tea.String(recordId),
	})
	return err
}

func (dns *AliDNSUtils) DeleteRecordByRR(rr string) error {
	record, err := dns.FindRecordByRR(rr)
	if err != nil {
		return err
	}
	if record == nil {
		return nil
	}
	return dns.DeleteRecord(*record.RecordId)
}

func (dns *AliDNSUtils) UpdateRecord(RecordId string, RR string, Value string, Type string) error {
	_, err := dns.client.UpdateDomainRecord((&alidns.UpdateDomainRecordRequest{
		RecordId: tea.String(RecordId),
		RR:       tea.String(RR),
		Type:     tea.String(Type),
		Value:    tea.String(Value),
	}))
	return err
}
