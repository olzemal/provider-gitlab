package projects

import (
	"testing"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"
)

var (
	protectedBranchType                      = v1alpha1.ProtectedBranchKind
	protectedBranchName                      = "main"
	protectedBranchID                        = 1
	protectedBranchAllowForcePush            = true
	protectedBranchCodeOwnerApprovalRequired = true
	protectedBranchPermission                = []*v1alpha1.BranchPermissionOptions{
		{
			AccessLevel: (*v1alpha1.AccessLevelValue)(toPtr(40)),
			UserID:      toPtr(0),
			GroupID:     toPtr(0),
		},
	}
)

func TestLateInitializeProtectedBranch(t *testing.T) {
	cases := map[string]struct {
		parameters *v1alpha1.ProtectedBranchParameters
		pb         *gitlab.ProtectedBranch
		want       *v1alpha1.ProtectedBranchParameters
	}{
		"AllOptionalFields": {
			parameters: &v1alpha1.ProtectedBranchParameters{
				Name: protectedBranchName,
			},
			pb: &gitlab.ProtectedBranch{
				ID:   protectedBranchID,
				Name: protectedBranchName,
				PushAccessLevels: []*gitlab.BranchAccessDescription{
					{
						ID:                     1,
						AccessLevel:            40,
						AccessLevelDescription: "Maintainers",
						UserID:                 0,
						GroupID:                0,
					},
				},
				MergeAccessLevels: []*gitlab.BranchAccessDescription{
					{
						ID:                     1,
						AccessLevel:            40,
						AccessLevelDescription: "Maintainers",
						UserID:                 0,
						GroupID:                0,
					},
				},
				UnprotectAccessLevels:     nil,
				AllowForcePush:            protectedBranchAllowForcePush,
				CodeOwnerApprovalRequired: protectedBranchCodeOwnerApprovalRequired,
			},
			want: &v1alpha1.ProtectedBranchParameters{
				ID:                        &protectedBranchID,
				Name:                      protectedBranchName,
				AllowedToPush:             protectedBranchPermission,
				AllowedToMerge:            protectedBranchPermission,
				AllowedToUnprotect:        []*v1alpha1.BranchPermissionOptions{},
				AllowForcePush:            &protectedBranchAllowForcePush,
				CodeOwnerApprovalRequired: &protectedBranchCodeOwnerApprovalRequired,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeProtectedBranch(tc.parameters, tc.pb)
			if diff := cmp.Diff(tc.want, tc.parameters); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}
func TestGenerateProtectRepositoryBranchOptions(t *testing.T) {
	cases := map[string]struct {
		parameters *v1alpha1.ProtectedBranchParameters
		want       *gitlab.ProtectRepositoryBranchesOptions
	}{
		"MinimalProtectedBranch": {
			parameters: &v1alpha1.ProtectedBranchParameters{
				Name: protectedBranchName,
			},
			want: &gitlab.ProtectRepositoryBranchesOptions{
				Name:               &protectedBranchName,
				AllowedToPush:      &[]*gitlab.BranchPermissionOptions{},
				AllowedToMerge:     &[]*gitlab.BranchPermissionOptions{},
				AllowedToUnprotect: &[]*gitlab.BranchPermissionOptions{},
			},
		},
		"WithAllowedToPush": {
			parameters: &v1alpha1.ProtectedBranchParameters{
				Name:          protectedBranchName,
				AllowedToPush: protectedBranchPermission,
			},
			want: &gitlab.ProtectRepositoryBranchesOptions{
				Name: &protectedBranchName,
				AllowedToPush: &[]*gitlab.BranchPermissionOptions{
					{
						ID:          toPtr(1),
						AccessLevel: (*gitlab.AccessLevelValue)(toPtr(40)),
						UserID:      toPtr(0),
						GroupID:     toPtr(0),
					},
				},
				AllowedToMerge:     &[]*gitlab.BranchPermissionOptions{},
				AllowedToUnprotect: &[]*gitlab.BranchPermissionOptions{},
			},
		},
		"WithAllowForcePush": {
			parameters: &v1alpha1.ProtectedBranchParameters{
				Name:           protectedBranchName,
				AllowForcePush: &protectedBranchAllowForcePush,
			},
			want: &gitlab.ProtectRepositoryBranchesOptions{
				Name:               &protectedBranchName,
				AllowedToPush:      &[]*gitlab.BranchPermissionOptions{},
				AllowedToMerge:     &[]*gitlab.BranchPermissionOptions{},
				AllowedToUnprotect: &[]*gitlab.BranchPermissionOptions{},
				AllowForcePush:     &protectedBranchAllowForcePush,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateProtectRepositoryBranchesOptions(tc.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
func TestUpdateProtectedBranchOptions(t *testing.T) {
	cases := map[string]struct {
		parameters *v1alpha1.ProtectedBranchParameters
		want       *gitlab.UpdateProtectedBranchOptions
	}{
		"UpdateName": {
			parameters: &v1alpha1.ProtectedBranchParameters{
				Name: protectedBranchName,
			},
			want: &gitlab.UpdateProtectedBranchOptions{
				Name:               &protectedBranchName,
				AllowedToPush:      &[]*gitlab.BranchPermissionOptions{},
				AllowedToMerge:     &[]*gitlab.BranchPermissionOptions{},
				AllowedToUnprotect: &[]*gitlab.BranchPermissionOptions{},
			},
		},
		"UpdateAllowedToMerge": {
			parameters: &v1alpha1.ProtectedBranchParameters{
				Name:           protectedBranchName,
				AllowedToMerge: protectedBranchPermission,
			},
			want: &gitlab.UpdateProtectedBranchOptions{
				Name:          &protectedBranchName,
				AllowedToPush: &[]*gitlab.BranchPermissionOptions{},
				AllowedToMerge: &[]*gitlab.BranchPermissionOptions{
					{
						ID:          toPtr(1),
						AccessLevel: (*gitlab.AccessLevelValue)(toPtr(40)),
						UserID:      toPtr(0),
						GroupID:     toPtr(0),
					},
				},
				AllowedToUnprotect: &[]*gitlab.BranchPermissionOptions{},
			},
		},
		"UpdateCodeOwnerApprovalRequired": {
			parameters: &v1alpha1.ProtectedBranchParameters{
				Name:                      protectedBranchName,
				CodeOwnerApprovalRequired: &protectedBranchCodeOwnerApprovalRequired,
			},
			want: &gitlab.UpdateProtectedBranchOptions{
				Name:                      &protectedBranchName,
				AllowedToPush:             &[]*gitlab.BranchPermissionOptions{},
				AllowedToMerge:            &[]*gitlab.BranchPermissionOptions{},
				AllowedToUnprotect:        &[]*gitlab.BranchPermissionOptions{},
				CodeOwnerApprovalRequired: &protectedBranchCodeOwnerApprovalRequired,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateProtectedBranchOptions(tc.parameters)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsProtectedBranchUpToDate(t *testing.T) {
	cases := map[string]struct {
		parameters *v1alpha1.ProtectedBranchParameters
		pb         *gitlab.ProtectedBranch
		want       bool
	}{
		"MinimalProtectedBranch": {
			parameters: &v1alpha1.ProtectedBranchParameters{
				Name: protectedBranchName,
			},
			pb: &gitlab.ProtectedBranch{
				Name: protectedBranchName,
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeProtectedBranch(tc.parameters, tc.pb)
			got := IsProtectedBranchUpToDate(tc.parameters, tc.pb)
			if got != tc.want {
				t.Errorf("got %v but want %v, when checking if...\n%+v\n...is up to date with...\n%+v", got, tc.want, tc.parameters, tc.pb)
			}
		})
	}
}

func toPtr[T any](i T) *T {
	return &i
}
