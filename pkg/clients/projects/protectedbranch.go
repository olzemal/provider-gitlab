package projects

import (
	"slices"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

// ProtectedBranchClient defines GitLab protected branch service operations
type ProtectedBranchClient interface {
	GetProtectedBranch(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error)
	ProtectRepositoryBranches(pid interface{}, opt *gitlab.ProtectRepositoryBranchesOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error)
	UpdateProtectedBranch(pid interface{}, branch string, opt *gitlab.UpdateProtectedBranchOptions, options ...gitlab.RequestOptionFunc) (*gitlab.ProtectedBranch, *gitlab.Response, error)
	UnprotectRepositoryBranches(pid interface{}, branch string, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewProtectedBranchClient returns a new GitLab protected branch service
func NewProtectedBranchClient(cfg clients.Config) ProtectedBranchClient {
	git := clients.NewClient(cfg)
	return git.ProtectedBranches
}

// LateInitializeProtectedBranch fills the empty variables in the parameters with the values seen in gitlab.ProtectedBranch
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

// GenerateProtectRepositoryBranchesOptions generates a GitLab API struct from our internal version.
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

// GenerateUpdateProtectedBranchOptions generates a GitLab API struct from our internal version.
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

// GenerateBranchPermissionOptions generates a GitLab API struct from our internal version.
func GenerateBranchPermissionOptions(in []*v1alpha1.BranchPermissionOptions) *[]*gitlab.BranchPermissionOptions {
	o := []*gitlab.BranchPermissionOptions{}
	id := 1 // Indexing starts with 1 here (https://docs.gitlab.com/ee/api/protected_branches.html#example-with-user--group-level-access)
	for _, item := range in {
		id++
		o = append(o, &gitlab.BranchPermissionOptions{
			ID:          &id,
			UserID:      item.UserID,
			GroupID:     item.GroupID,
			DeployKeyID: item.DeployKeyID,
			AccessLevel: (*gitlab.AccessLevelValue)(item.AccessLevel),
		})
	}

	// Delete the next Permissions
	t := true
	for ; id < 100; id++ {
		o = append(o, &gitlab.BranchPermissionOptions{
			Destroy: &t,
		})
	}
	return &o
}

// IsProtectedBranchUpToDate checks if the given parameters are equal, ignoring references and the ProjectID.
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

// ProtectedBranchToParameters generates a Parameter struct from a GitLab API ProtectedBranch
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

	out := []*v1alpha1.BranchPermissionOptions{}
	for _, d := range in {
		o := v1alpha1.BranchPermissionOptions{
			GroupID:     &d.GroupID,
			AccessLevel: (*v1alpha1.AccessLevelValue)(&d.AccessLevel),
		}
		if d.AccessLevelDescription == "Deploy key" {
			o.DeployKeyID = &d.UserID
		} else {
			o.UserID = &d.UserID
		}
		out = append(out, &o)
	}
	return out
}
