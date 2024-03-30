package protectedbranches

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/xanzy/go-gitlab"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	secretstoreapi "github.com/crossplane-contrib/provider-gitlab/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients/projects"
	"github.com/crossplane-contrib/provider-gitlab/pkg/features"
)

const (
	errNotProtectedBranch = "managed resource is not a protected branch"
	errGetFailed          = "cannot get protected branch"
	errCreateFailed       = "cannot create protected branch"
	errUpdateFailed       = "cannot update protected branch"
	errDeleteFailed       = "cannot delete protected branch"
	errProjectIDMissing   = "ProjectID is missing"
)

// SetupProtectedBranch adds a controller that reconciles protected branches.
func SetupProtectedBranch(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ProtectedBranchKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), secretstoreapi.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newGitlabClientFn: projects.NewProtectedBranchClient}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ProtectedBranchGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.ProtectedBranch{}).
		Complete(r)
}

type connector struct {
	kube              client.Client
	newGitlabClientFn func(cfg clients.Config) projects.ProtectedBranchClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ProtectedBranch)
	if !ok {
		return nil, errors.New(errNotProtectedBranch)
	}
	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}
	return &external{kube: c.kube, client: c.newGitlabClientFn(*cfg)}, nil
}

type external struct {
	kube   client.Client
	client projects.ProtectedBranchClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.ProtectedBranch)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotProtectedBranch)
	}
	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalObservation{}, errors.New(errProjectIDMissing)
	}

	pb, res, err := e.client.GetProtectedBranch(*cr.Spec.ForProvider.ProjectID, cr.Spec.ForProvider.Name)
	if err != nil {
		if clients.IsResponseNotFound(res) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	projects.LateInitializeProtectedBranch(&cr.Spec.ForProvider, pb)

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        projects.IsProtectedBranchUpToDate(&cr.Spec.ForProvider, pb),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.ProtectedBranch)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProtectedBranch)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalCreation{}, errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(xpv1.Creating())
	rsp, _, err := e.client.ProtectRepositoryBranches(
		*cr.Spec.ForProvider.ProjectID,
		projects.GenerateProtectRepositoryBranchesOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	cr.Spec.ForProvider.ID = &rsp.ID

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.ProtectedBranch)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotProtectedBranch)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return managed.ExternalUpdate{}, errors.New(errProjectIDMissing)
	}

	_, _, err := e.client.UpdateProtectedBranch(
		*cr.Spec.ForProvider.ProjectID,
		cr.Spec.ForProvider.Name,
		projects.GenerateUpdateProtectedBranchOptions(&cr.Spec.ForProvider),
		gitlab.WithContext(ctx),
	)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.ProtectedBranch)
	if !ok {
		return errors.New(errNotProtectedBranch)
	}

	if cr.Spec.ForProvider.ProjectID == nil {
		return errors.New(errProjectIDMissing)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := e.client.UnprotectRepositoryBranches(
		*cr.Spec.ForProvider.ProjectID,
		cr.Spec.ForProvider.Name,
		gitlab.WithContext(ctx),
	)
	return errors.Wrap(err, errDeleteFailed)
}
