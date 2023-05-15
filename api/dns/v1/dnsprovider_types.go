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

// +kubebuilder:validation:Enum=ALIYUN
type DNSProviderType string

const (
	DNSProviderTypeAliyun DNSProviderType = "ALIYUN"
)

type AliyunProviderConfig struct {
	AccessKeyID     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
}

// DNSProviderSpec defines the desired state of DNSProvider
type DNSProviderSpec struct {
	DomainName   string          `json:"domainName"`
	ProviderType DNSProviderType `json:"providerType"`
	// +kubebuilder:validation:Optional
	Aliyun AliyunProviderConfig `json:"aliyun,omitempty"`
}

// DNSProviderStatus defines the observed state of DNSProvider
type DNSProviderStatus struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

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
