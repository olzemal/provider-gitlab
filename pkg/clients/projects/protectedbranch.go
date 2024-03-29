package projects

import (
	"slices"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/xanzy/go-gitlab"
)

type ProtectedBranchClient interface {
	GetProtectedBranch(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error)
	ProtectRepositoryBranches(pid interface{}, opt *gitlab.ProtectRepositoryBranchesOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error)
	UpdateProtectedBranch(pid interface{}, branch string, opt *gitlab.UpdateProtectedBranchOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error)
	UnprotectRepositoryBranches(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

func NewProtectedBranchClient(cfg clients.Config) ProtectedBranchClient {
	git := clients.NewClient(cfg)
	return git.ProtectedBranches
}

func LateInitializeProtectedBranch(p *v1alpha1.ProtectedBranchParameters, g *gitlab.ProtectedBranch) {
	if p.ID == nil {
		p.ID = &g.ID
	}

	if p.AllowForcePush == nil {
		p.AllowForcePush = &g.AllowForcePush
	}

	if p.AllowedToPush == nil {
		p.AllowedToPush = accessLevelsToBranchPermissionOptionsList(g.PushAccessLevels)
	}

	if p.AllowedToMerge == nil {
		p.AllowedToMerge = accessLevelsToBranchPermissionOptionsList(g.MergeAccessLevels)
	}

	if p.AllowedToUnprotect == nil {
		p.AllowedToUnprotect = accessLevelsToBranchPermissionOptionsList(g.UnprotectAccessLevels)
	}

	if p.CodeOwnerApprovalRequired == nil {
		p.CodeOwnerApprovalRequired = &g.CodeOwnerApprovalRequired
	}
}

func GenerateProtectRepositoryBranchesOptions(p *v1alpha1.ProtectedBranchParameters) *gitlab.ProtectRepositoryBranchesOptions {
	return &gitlab.ProtectRepositoryBranchesOptions{
		Name:                      &p.Name,
		AllowForcePush:            p.AllowForcePush,
		AllowedToPush:             GenerateBranchPermissionOptions(p.AllowedToPush),
		AllowedToMerge:            GenerateBranchPermissionOptions(p.AllowedToMerge),
		AllowedToUnprotect:        GenerateBranchPermissionOptions(p.AllowedToUnprotect),
		CodeOwnerApprovalRequired: p.CodeOwnerApprovalRequired,
	}
}

func GenerateUpdateProtectedBranchOptions(p *v1alpha1.ProtectedBranchParameters) *gitlab.UpdateProtectedBranchOptions {
	return &gitlab.UpdateProtectedBranchOptions{
		Name:                      &p.Name,
		AllowForcePush:            p.AllowForcePush,
		AllowedToPush:             GenerateBranchPermissionOptions(p.AllowedToPush),
		AllowedToMerge:            GenerateBranchPermissionOptions(p.AllowedToMerge),
		AllowedToUnprotect:        GenerateBranchPermissionOptions(p.AllowedToUnprotect),
		CodeOwnerApprovalRequired: p.CodeOwnerApprovalRequired,
	}
}

func GenerateBranchPermissionOptions(in []*v1alpha1.BranchPermissionOptions) *[]*gitlab.BranchPermissionOptions {
	o := []*gitlab.BranchPermissionOptions{}
	for _, item := range in {
		o = append(o, &gitlab.BranchPermissionOptions{
			UserID:      item.UserID,
			GroupID:     item.GroupID,
			DeployKeyID: item.DeployKeyID,
			AccessLevel: (*gitlab.AccessLevelValue)(item.AccessLevel),
		})
	}
	return &o
}

func GenerateProviderBranchPermissionOptions(in []*gitlab.BranchPermissionOptions) []*v1alpha1.BranchPermissionOptions {
	o := []*v1alpha1.BranchPermissionOptions{}
	for _, item := range in {
		o = append(o, &v1alpha1.BranchPermissionOptions{
			UserID:      item.UserID,
			GroupID:     item.GroupID,
			DeployKeyID: item.DeployKeyID,
			AccessLevel: (*v1alpha1.AccessLevelValue)(item.AccessLevel),
		})
	}
	return o
}

func IsProtectedBranchUpToDate(p *v1alpha1.ProtectedBranchParameters, g *gitlab.ProtectedBranch) bool {
	if p == nil {
		return true
	}

	return cmp.Equal(*p,
		ProtectedBranchToParameters(*g),
		cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}, &xpv1.SecretKeySelector{}),
		cmpopts.IgnoreFields(v1alpha1.ProtectedBranchParameters{}, "ProjectID"),
	)
}

func ProtectedBranchToParameters(in gitlab.ProtectedBranch) v1alpha1.ProtectedBranchParameters {
	return v1alpha1.ProtectedBranchParameters{
		ID:                        &in.ID,
		Name:                      in.Name,
		AllowedToPush:             accessLevelsToBranchPermissionOptionsList(in.PushAccessLevels),
		AllowedToMerge:            accessLevelsToBranchPermissionOptionsList(in.MergeAccessLevels),
		AllowedToUnprotect:        accessLevelsToBranchPermissionOptionsList(in.UnprotectAccessLevels),
		AllowForcePush:            &in.AllowForcePush,
		CodeOwnerApprovalRequired: &in.CodeOwnerApprovalRequired,
	}
}

func accessLevelsToBranchPermissionOptionsList(in []*gitlab.BranchAccessDescription) []*v1alpha1.BranchPermissionOptions {
	slices.SortFunc(in, func(a, b *gitlab.BranchAccessDescription) int {
		if a.ID > b.ID {
			return 1
		}
		if a.ID < b.ID {
			return -1
		}
		return 0
	})

	o := []*v1alpha1.BranchPermissionOptions{}
	for _, i := range in {
		o = append(o, accessLevelToBranchPermissionOptions(i))
	}
	return o
}

func accessLevelToBranchPermissionOptions(in *gitlab.BranchAccessDescription) *v1alpha1.BranchPermissionOptions {
	o := v1alpha1.BranchPermissionOptions{
		GroupID:     &in.GroupID,
		AccessLevel: (*v1alpha1.AccessLevelValue)(&in.AccessLevel),
	}
	if in.AccessLevelDescription == "Deploy key" {
		o.DeployKeyID = &in.UserID
	} else {
		o.UserID = &in.UserID
	}
	return &o
}
