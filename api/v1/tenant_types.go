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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
	// Namespaces are the namespaces that belong to this tenant
	Namespaces []string `json:"namespaces,omitempty"`

	// AdminEmail is the email of the administrator
	AdminEmail string `json:"adminEmail,omitempty"`

	// AdminGroups are the admin groups for this tenant
	AdminGroups []string `json:"adminGroups,omitempty"`

	// UserGroups are the user groups for this tenant
	UserGroups []string `json:"userGroups,omitempty"`
}

// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	// NamespaceCount holds the number of namespaces that belong to this tenant
	NamespaceCount int `json:"namespaceCount"`

	// AdminEmail holds the admin email
	AdminEmail string `json:"adminEmail"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Email",type="string",JSONPath=".status.adminEmail",description="AdminEmail"
// +kubebuilder:printcolumn:name="NamespaceCount",type="integer",JSONPath=".status.namespaceCount",description="NamespaceCount"
// Tenant is the Schema for the tenants API
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TenantList contains a list of Tenant
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}
