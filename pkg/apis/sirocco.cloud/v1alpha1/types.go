package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"orcaoperator/pkg/flow"
)

// -------------------------------------------------------------------
// TASK
// -------------------------------------------------------------------

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=sirocco.cloud

// Task is a top-level type
type Task struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status TaskStatus `json:"status,omitempty"`
	// This is where you can define
	// your own custom spec
	Spec TaskSpec `json:"spec,omitempty"`
}

type TaskStatus struct {
	State       flow.Status `json:"state"`
	LastSuccess string      `json:"lastSuccess"`
	LastFailure string      `json:"lastFailure"`
}

type TaskSpec struct {
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	StartOnIgnition []string `json:"startOnIgnition,omitempty"`
	// +optional
	StartOnSuccess []string `json:"startOnSuccess,omitempty"`
	// +optional
	StartOnFailure []string `json:"startOnFailure,omitempty"`
	// +optional
	SuccessActions []string `json:"successActions,omitempty"`
	// +optional
	FailureActions []string `json:"failureActions,omitempty"`

	Template corev1.PodTemplateSpec `json:"template"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// no client needed for list as it's been created in above
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []Task `json:"items"`
}

// -------------------------------------------------------------------
// IGNITOR
// -------------------------------------------------------------------

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=sirocco.cloud

// Ignitor is a top-level type
type Ignitor struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status IgnitorStatus `json:"status,omitempty"`
	// This is where you can define
	// your own custom spec
	Spec IgnitorSpec `json:"spec,omitempty"`
}

type IgnitorStatus struct {
	State flow.Status `json:"state"`
}

type IgnitorSpec struct {
	// +optional
	Scheduled string `json:"scheduled,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// no client needed for list as it's been created in above
type IgnitorList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []Ignitor `json:"items"`
}
