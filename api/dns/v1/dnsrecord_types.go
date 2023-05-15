/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Enum=A;CNAME;TXT;MX;SRV;AAAA;NS;CAA
type DNSRecordType string

const (
	DNSRecordTypeA     DNSRecordType = "A"
	DNSRecordTypeCNAME DNSRecordType = "CNAME"
	DNSRecordTypeTXT   DNSRecordType = "TXT"
	DNSRecordTypeMX    DNSRecordType = "MX"
	DNSRecordTypeSRV   DNSRecordType = "SRV"
	DNSRecordTypeAAAA  DNSRecordType = "AAAA"
	DNSRecordTypeNS    DNSRecordType = "NS"
	DNSRecordTypeCAA   DNSRecordType = "CAA"
)

type NamespacedName struct {
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
}

// DNSRecordSpec defines the desired state of DNSRecord
type DNSRecordSpec struct {
	RecordType DNSRecordType `json:"recordType"`
	Name       string        `json:"name"`
	Value      string        `json:"value"`
	// +kubebuilder:validation:Optional
	TTL *int `json:"ttl"`
}

type DNSRecordStatusPhase string

const (
	DNSRecordStatusPhasePending  DNSRecordStatusPhase = "Pending"
	DNSRecordStatusPhaseMatching DNSRecordStatusPhase = "Matching"
	DNSRecordStatusPhaseSyncing  DNSRecordStatusPhase = "Syncing"
	DNSRecordStatusPhaseSuccess  DNSRecordStatusPhase = "Success"
	DNSRecordStatusPhaseFailed   DNSRecordStatusPhase = "Failed"
)

// DNSRecordStatus defines the observed state of DNSRecord
type DNSRecordStatus struct {
	ProviderRef NamespacedName       `json:"providerRef"`
	Status      DNSRecordStatusPhase `json:"status"`
	Message     string               `json:"message"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.providerRef.name",priority=1
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.message",priority=1

// DNSRecord is the Schema for the dnsrecords API
type DNSRecord struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSRecordSpec   `json:"spec,omitempty"`
	Status DNSRecordStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DNSRecordList contains a list of DNSRecord
type DNSRecordList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSRecord `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DNSRecord{}, &DNSRecordList{})
}

func (record *DNSRecordSpec) Match(provider *DNSProviderSpec) bool {
	domainName := "." + provider.DomainName
	return strings.HasSuffix(record.Name, domainName)
}

func (record *DNSRecordSpec) RR(provider *DNSProviderSpec) string {
	domainName := "." + provider.DomainName
	rr := strings.TrimSuffix(record.Name, domainName)
	return rr
}

func (record *DNSRecordSpec) SpinalName() string {
	return strings.TrimSpace(strings.ToLower(strings.ReplaceAll(record.Name, ".", "-")))
}
