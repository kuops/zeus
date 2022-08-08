package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.kubernetesVersion"
// +kubebuilder:printcolumn:name="Nodes",type="integer",JSONPath=".status.nodeCount"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.provider"
// +kubebuilder:resource:scope=Cluster

type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec"`
	Status ClusterStatus `json:"status,omitempty"`
}

type ClusterSpec struct {
	KubeConfig string `json:"kubeconfig,omitempty"`
}

type ClusterStatus struct {
	Conditions        []ClusterCondition `json:"conditions,omitempty"`
	KubernetesVersion string             `json:"kubernetesVersion,omitempty"`
	NodeCount         int                `json:"nodeCount,omitempty"`
	Provider          string             `json:"provider,omitempty"`
}

type ClusterCondition struct {
	// Type of cluster condition.
	Type ClusterConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

type ClusterConditionType string

const (
	ClusterReady    ClusterConditionType = "Ready"
	ClusterNotReady ClusterConditionType = "NotReady"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Cluster `json:"items"`
}
