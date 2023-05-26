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

// +kubebuilder:validation:Enum=DDNS;CNAME
type DNSGeneratorType string

const (
	DNSGeneratorTypeDDNS  DNSGeneratorType = "DDNS"
	DNSGeneratorTypeCNAME DNSGeneratorType = "CNAME"
)

// DNSGeneratorSpec defines the desired state of DNSGenerator
type DNSGeneratorSpec struct {
	GeneratorType DNSGeneratorType `json:"generatorType"`
	// +optional
	DDNS DDNSGeneratorConfig `json:"ddns,omitempty"`
	// +optional
	CNAME CNAMEGeneratorConfig `json:"cname,omitempty"`
}

type DDNSGeneratorConfig struct {
	// +optional
	// +kubebuilder:default=2
	// The timeout for ddns service (seconds)
	Timeout int64 `json:"timeout"`
	// +optional
	// The extra apis to get public ip
	ExtraApis []string `json:"extraApis"`
	// +optional
	// +kubebuilder:default=60
	// The expire time for public ip cache (seconds)
	CacheExpire int64 `json:"cacheExpire"`
	// +optional
	// +kubebuilder:default=30
	// The interval to clean the public ip cache (seconds)
	CleanInterval int64 `json:"cleanInterval"`
	// +optional
	// +kubebuilder:default=600
	// The interval to refresh the public ip (seconds)
	RefreshInternal int64 `json:"refreshInternal"`
}

func (d *DDNSGeneratorConfig) WithDefault() *DDNSGeneratorConfig {
	if d.Timeout == 0 {
		d.Timeout = 2
	}
	if d.CacheExpire == 0 {
		d.CacheExpire = 60
	}
	if d.CleanInterval == 0 {
		d.CleanInterval = 30
	}
	if d.RefreshInternal == 0 {
		d.RefreshInternal = 600
	}
	return d
}

type CNAMEGeneratorConfig struct {
	Value string `json:"value"`
}

// DNSGeneratorStatus defines the observed state of DNSGenerator
type DNSGeneratorStatus struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.generatorType`
//+kubebuilder:printcolumn:name="Valid",type=boolean,JSONPath=`.status.valid`
//+kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.message`,priority=1

// DNSGenerator is the Schema for the dnsgenerators API
type DNSGenerator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSGeneratorSpec   `json:"spec,omitempty"`
	Status DNSGeneratorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DNSGeneratorList contains a list of DNSGenerator
type DNSGeneratorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSGenerator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DNSGenerator{}, &DNSGeneratorList{})
}
