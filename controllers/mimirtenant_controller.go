package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	domain "mimir-operator/api/v1alpha1"
)

const (
	mimirFinalizer = "mimir.grafana.net/finalizer"
	temporaryFiles = "/tmp/"
)

// MimirTenantReconciler reconciles a MimirTenant object
type MimirTenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mimir.grafana.com,resources=mimirtenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mimir.grafana.com,resources=mimirtenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mimir.grafana.com,resources=mimirtenants/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MimirTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the MimirTenant tenant
	tenant := &domain.MimirTenant{}
	err := r.Get(ctx, req.NamespacedName, tenant)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	log.FromContext(ctx).Info("Running reconcile on MimirTenant")

	// Examine DeletionTimestamp to determine if object is under deletion
	if tenant.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer
		if !controllerutil.ContainsFinalizer(tenant, mimirFinalizer) {
			controllerutil.AddFinalizer(tenant, mimirFinalizer)
			if err := r.Update(ctx, tenant); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(tenant, mimirFinalizer) {
			if err := r.handleDeletion(ctx, tenant); err != nil {
				return ctrl.Result{}, err
			}

			// Remove our finalizer from the list and update it
			controllerutil.RemoveFinalizer(tenant, mimirFinalizer)
			return ctrl.Result{}, r.Update(ctx, tenant)
		}
	}

	return ctrl.Result{}, r.handleReconcile(ctx, tenant)
}

// handleReconcile handles reconciliation of MimirTenants for events that are not a deletion
// This means that this function will be called for any modification in a MimirTenant or for
// any creation of a new tenant in the API. It is also called periodically for scheduled
// reconciliation and at the startup of the controller.
func (r *MimirTenantReconciler) handleReconcile(ctx context.Context, tenant *domain.MimirTenant) error {
	if err := r.reconcileRules(ctx, tenant); err != nil {
		return err
	}

	return nil
}

// handleDeletion handles cleaning up after the deletion of a MimirTenant
func (r *MimirTenantReconciler) handleDeletion(ctx context.Context, tenant *domain.MimirTenant) error {
	log.FromContext(ctx).Info("Running reconciliation on deletion of a MimirTenant")

	if tenant.Spec.Rules != nil {
		log.FromContext(ctx).Info("Deleting rules from Mimir")
		return r.deleteRulesForTenant(ctx, tenant)
	}

	return nil
}

// reconcileRules ensures Mimir is synced with the PrometheusRules associated with a tenant
func (r *MimirTenantReconciler) reconcileRules(ctx context.Context, tenant *domain.MimirTenant) error {
	if tenant.Spec.Rules != nil && tenant.Spec.Rules.Selectors != nil {
		log.FromContext(ctx).Info("Running reconciliation of the rules")
		if err := r.syncRulesForTenant(ctx, tenant); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MimirTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domain.MimirTenant{}).
		Complete(r)
}
