package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelloType is a top-level type
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
	Message string `json:"message,omitempty"`
}

type TaskSpec struct {
	Name string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// no client needed for list as it's been created in above
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `son:"metadata,omitempty"`

	Items []Task `json:"items"`
}
