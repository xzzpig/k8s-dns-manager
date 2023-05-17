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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=ALIYUN;CLOUDFLARE
type DNSProviderType string

const (
	DNSProviderTypeAliyun     DNSProviderType = "ALIYUN"
	DNSProviderTypeCloudflare DNSProviderType = "CLOUDFLARE"
)

type AliyunProviderConfig struct {
	AccessKeyID     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
}

type CloudflareProviderConfig struct {
	// +kubebuilder:validation:Optional
	// If empty, spec.domainName will be used as zone name
	ZoneName string `json:"zoneName,omitempty"`
	// +kubebuilder:validation:Optional
	APIToken string `json:"apiToken"`
	// +kubebuilder:validation:Optional
	Key string `json:"key"`
	// +kubebuilder:validation:Optional
	Email string `json:"email"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// If true, the DNS record will be proxied by Cloudflare, can be overrided by Annotation `dns.xzzpig.com/record-proxied`
	Proxied bool `json:"proxied"`
}

// DNSProviderSpec defines the desired state of DNSProvider
type DNSProviderSpec struct {
	DomainName   string          `json:"domainName"`
	ProviderType DNSProviderType `json:"providerType"`
	// +kubebuilder:validation:Optional
	Aliyun AliyunProviderConfig `json:"aliyun,omitempty"`
	// +kubebuilder:validation:Optional
	Cloudflare CloudflareProviderConfig `json:"cloudflare,omitempty"`
}

// DNSProviderStatus defines the observed state of DNSProvider
type DNSProviderStatus struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Domain",type=string,JSONPath=`.spec.domainName`
//+kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.providerType`
//+kubebuilder:printcolumn:name="Valid",type=boolean,JSONPath=`.status.valid`
//+kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`,priority=1

// DNSProvider is the Schema for the dnsproviders API
type DNSProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSProviderSpec   `json:"spec,omitempty"`
	Status DNSProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DNSProviderList contains a list of DNSProvider
type DNSProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DNSProvider{}, &DNSProviderList{})
}
