package controllers

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"mimir-operator/internal/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	domain "mimir-operator/api/v1alpha1"
)

const (
	mimirFinalizer = "mimir.randgen.xyz/finalizer"
	temporaryFiles = "/tmp/"
)

// MimirRulesReconciler reconciles a MimirRules object
type MimirRulesReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mimir.randgen.xyz,resources=mimirrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mimir.randgen.xyz,resources=mimirrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mimir.randgen.xyz,resources=mimirrules/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *MimirRulesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the MimirRules
	mr := &domain.MimirRules{}
	err := r.Get(ctx, req.NamespacedName, mr)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	log.FromContext(ctx).Info("Running reconcile on MimirRules")

	// Examine DeletionTimestamp to determine if object is under deletion
	if mr.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer
		if !controllerutil.ContainsFinalizer(mr, mimirFinalizer) {
			controllerutil.AddFinalizer(mr, mimirFinalizer)
			if err := r.Update(ctx, mr); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(mr, mimirFinalizer) {
			if err := r.handleDeletion(ctx, mr); err != nil {
				return ctrl.Result{}, err
			}

			// Remove our finalizer from the list and update it
			controllerutil.RemoveFinalizer(mr, mimirFinalizer)
			return ctrl.Result{}, r.Update(ctx, mr)
		}
	}

	return ctrl.Result{}, r.handleReconcile(ctx, mr)
}

// handleReconcile handles reconciliation of MimirRules for events that are not a deletion
// This means that this function will be called for any modification in a MimirRules or for
// any creation of a new MimirRules in the API. It is also called periodically for scheduled
// reconciliation and at the startup of the controller.
func (r *MimirRulesReconciler) handleReconcile(ctx context.Context, mr *domain.MimirRules) error {
	if err := r.reconcileRules(ctx, mr); err != nil {
		return err
	}

	return nil
}

// handleDeletion handles cleaning up after the deletion of a MimirRules
func (r *MimirRulesReconciler) handleDeletion(ctx context.Context, mr *domain.MimirRules) error {
	log.FromContext(ctx).Info("Running reconciliation on deletion of a MimirRules")

	if mr.Spec.Rules != nil {
		log.FromContext(ctx).Info("Deleting rules from Mimir")

		auth, err := utils.ExtractAuth(ctx, r.Client, mr.Spec.Auth, mr.ObjectMeta.Namespace)
		if err != nil {
			return fmt.Errorf("failed to extract authentication settings: %w", err)
		}

		return r.deleteRulesForTenant(ctx, auth, mr)
	}

	return nil
}

// reconcileRules ensures Mimir is synced with the PrometheusRules associated with a MimirRules
func (r *MimirRulesReconciler) reconcileRules(ctx context.Context, mr *domain.MimirRules) error {
	if mr.Spec.Rules != nil && mr.Spec.Rules.Selectors != nil {
		log.FromContext(ctx).Info("Running reconciliation of the rules")

		auth, err := utils.ExtractAuth(ctx, r.Client, mr.Spec.Auth, mr.ObjectMeta.Namespace)
		if err != nil {
			return fmt.Errorf("failed to extract authentication settings: %w", err)
		}

		if err := r.syncRulesToRuler(ctx, auth, mr); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MimirRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domain.MimirRules{}).
		Complete(r)
}
