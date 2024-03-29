package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProtectedBranchParameters defines the desired state of a protected branch
// https://docs.gitlab.com/ee/api/protected_branches.html
type ProtectedBranchParameters struct {
	// ProjectID is the ID of the project to protect a branch on.
	// +optional
	// +immutable
	ProjectID *int `json:"projectId,omitempty"`

	// +optional
	// +immutable
	ProjectIDRef *xpv1.Reference `json:"projectIdRef,omitempty"`

	// +optional
	ProjectIDSelector *xpv1.Selector `json:"projectIdSelector,omitempty"`

	// ID is the internal id of the protected branch.
	// +optional
	ID *int `json:"id,omitempty"`

	// Name is the name or a wildcard of the protected branch.
	Name string `json:"name"`

	// +optional
	AllowedToPush []*BranchPermissionOptions `json:"allowedToPush,omitempty"`

	// +optional
	AllowedToMerge []*BranchPermissionOptions `json:"allowedToMerge,omitempty"`

	// +optional
	AllowedToUnprotect []*BranchPermissionOptions `json:"AllowedToUnprotect,omitempty"`

	// +optional
	AllowForcePush *bool `json:"allowForcePush,omitempty"`

	// +optional
	CodeOwnerApprovalRequired *bool `json:"codeOwnerApprovalRequired,omitempty"`
}

// BranchPermissionOptions defines a Permission.
type BranchPermissionOptions struct {
	UserID *int `url:"user_id,omitempty" json:"user_id,omitempty"`

	// +optional
	GroupID *int `url:"group_id,omitempty" json:"group_id,omitempty"`

	// +optional
	// +immutable
	GroupIDRef *xpv1.Reference `json:"groupIdRef,omitempty"`

	// +optional
	GroupIDSelector *xpv1.Selector `json:"groupIdSelector,omitempty"`

	// +optional
	DeployKeyID *int `url:"deploy_key_id,omitempty" json:"deploy_key_id,omitempty"`

	// +optional
	// +immutable
	DeployKeyIDRef *xpv1.Reference `json:"deployKeyIdRef,omitempty"`

	// +optional
	DeployKeyIDSelector *xpv1.Selector `json:"deployKeyIdSelector,omitempty"`

	// +optional
	AccessLevel *AccessLevelValue `url:"access_level,omitempty" json:"access_level,omitempty"`
}

// A ProtectedBranchSpec defines the desired state of a protected branch.
type ProtectedBranchSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ProtectedBranchParameters `json:"forProvider"`
}

// A ProtectedBranchStatus represents the observed state of a protected branch.
type ProtectedBranchStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A ProtectedBranch is a managed resource that represents a protected branch.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type ProtectedBranch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProtectedBranchSpec   `json:"spec"`
	Status ProtectedBranchStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProtectedBranchList contains a list of ProtectedBranch items.
type ProtectedBranchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProtectedBranch `json:"items"`
}
