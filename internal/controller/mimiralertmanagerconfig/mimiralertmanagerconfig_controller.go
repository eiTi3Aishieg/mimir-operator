package mimiralertmanagerconfig

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	domain "github.com/AmiditeX/mimir-operator/api/v1alpha1"
	"github.com/AmiditeX/mimir-operator/internal/utils"
)

const (
	alertManagerFinalizer = "mimir.randgen.xyz/finalizer"
	temporaryFiles        = "/tmp/"
)

// MimirAlertManagerConfigReconciler reconciles a MimirAlertManagerConfig object
type MimirAlertManagerConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mimir.randgen.xyz,resources=mimiralertmanagerconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mimir.randgen.xyz,resources=mimiralertmanagerconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mimir.randgen.xyz,resources=mimiralertmanagerconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *MimirAlertManagerConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	amc := &domain.MimirAlertManagerConfig{}
	err := r.Get(ctx, req.NamespacedName, amc)

	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	log.FromContext(ctx).Info("Running reconcile on MimirAlertManagerConfig")

	// Examine DeletionTimestamp to determine if object is under deletion
	if amc.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer
		if !controllerutil.ContainsFinalizer(amc, alertManagerFinalizer) {
			controllerutil.AddFinalizer(amc, alertManagerFinalizer)
			if err := r.Update(ctx, amc); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(amc, alertManagerFinalizer) {
			if err := r.handleDeletion(ctx, amc); err != nil {
				// Status is set only on failure to delete (the status is going to be deleted anyway if it succeeds)
				return ctrl.Result{}, r.setStatus(ctx, amc, err)
			}

			// Remove our finalizer from the list and update it
			controllerutil.RemoveFinalizer(amc, alertManagerFinalizer)
			return ctrl.Result{}, r.Update(ctx, amc)
		}
	}

	return ctrl.Result{}, r.handleCreationAndChanges(ctx, amc)
}

// handleCreationAndChanges handles reconciliation of Alert Manager Config for events that are not a deletion
// This means that this function will be called for any modification in an Alert Manager Config or for
// any creation of a new Alert Manager Config in the API. It is also called periodically for scheduled
// reconciliation and at the startup of the controller.
func (r *MimirAlertManagerConfigReconciler) handleCreationAndChanges(ctx context.Context, amc *domain.MimirAlertManagerConfig) error {
	reconciliationError := r.reconcileAMConfig(ctx, amc)
	if err := r.setStatus(ctx, amc, reconciliationError); err != nil {
		return err
	}

	return nil
}

// handleDeletion handles cleaning up after the deletion of a MimirAlertManagerConfig
func (r *MimirAlertManagerConfigReconciler) handleDeletion(ctx context.Context, amc *domain.MimirAlertManagerConfig) error {
	log.FromContext(ctx).Info("Running reconciliation on deletion of a MimirAlertManagerConfig")

	auth, err := utils.ExtractAuth(ctx, r.Client, amc.Spec.Auth, amc.ObjectMeta.Namespace)
	if err != nil {
		return fmt.Errorf("failed to extract authentication settings: %w", err)
	}

	return r.deleteMimirAlertManagerConfigForTenant(ctx, auth, amc)
}

// reconcileAMConfig ensures Mimir correctly load the alert manager config
func (r *MimirAlertManagerConfigReconciler) reconcileAMConfig(ctx context.Context, amc *domain.MimirAlertManagerConfig) error {
	log.FromContext(ctx).Info("Running reconciliation of the Alert Manager Config")

	auth, err := utils.ExtractAuth(ctx, r.Client, amc.Spec.Auth, amc.ObjectMeta.Namespace)
	if err != nil {
		return fmt.Errorf("failed to extract authentication settings: %w", err)
	}

	return sendAMConfigToMimir(ctx, auth, amc.Spec.ID, amc.Spec.URL, amc.Spec.Config)
}

// setStatus updates the status of MimirAlertManagerConfig after reconciliation
// If err is not nil, the error field is populated with the error and the status is set as "Failed"
// Otherwise, status is set as "Synced"
func (r *MimirAlertManagerConfigReconciler) setStatus(ctx context.Context, amc *domain.MimirAlertManagerConfig, err error) error {
	if err != nil {
		amc.Status.Status = "Failed"
		amc.Status.Error = err.Error()

		// Also log the error in the controller for clarity
		log.FromContext(ctx).Error(err, "Failed to reconcile MimirAlertManagerConfig")
	} else {
		amc.Status.Status = "Synced"
		amc.Status.Error = ""
	}

	return r.Status().Update(context.Background(), amc)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MimirAlertManagerConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domain.MimirAlertManagerConfig{}).
		Complete(r)
}
