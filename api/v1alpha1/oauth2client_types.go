/*

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OAuth2ClientSpec defines the desired state of OAuth2Client
type OAuth2ClientSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// OAuth2ClientStatus defines the observed state of OAuth2Client
type OAuth2ClientStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// OAuth2Client is the Schema for the oauth2clients API
type OAuth2Client struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OAuth2ClientSpec   `json:"spec,omitempty"`
	Status OAuth2ClientStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OAuth2ClientList contains a list of OAuth2Client
type OAuth2ClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OAuth2Client `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OAuth2Client{}, &OAuth2ClientList{})
}
