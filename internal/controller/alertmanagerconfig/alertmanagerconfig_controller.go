package alertmanagerconfig

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	domain "mimir-operator/api/v1alpha1"
	"mimir-operator/internal/utils"
)

const (
	alertManagerFinalizer = "mimir.randgen.xyz/finalizer"
	temporaryFiles        = "/tmp/"
)

// AlertManagerConfigReconciler reconciles a AlertManagerConfig object
type AlertManagerConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mimir.randgen.xyz,resources=alertmanagerconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mimir.randgen.xyz,resources=alertmanagerconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mimir.randgen.xyz,resources=alertmanagerconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AlertManagerConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *AlertManagerConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	amc := &domain.AlertManagerConfig{}
	err := r.Get(ctx, req.NamespacedName, amc)

	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	log.FromContext(ctx).Info("Running reconcile on AlertManagerConfig")

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

// handleCreationAndChanges handles reconciliation of MimirRules for events that are not a deletion
// This means that this function will be called for any modification in a MimirRules or for
// any creation of a new MimirRules in the API. It is also called periodically for scheduled
// reconciliation and at the startup of the controller.
func (r *AlertManagerConfigReconciler) handleCreationAndChanges(ctx context.Context, amc *domain.AlertManagerConfig) error {
	reconciliationError := r.reconcileRules(ctx, amc)
	if err := r.setStatus(ctx, amc, reconciliationError); err != nil {
		return err
	}

	return nil
}

// handleDeletion handles cleaning up after the deletion of a AlertManagerConfig
func (r *AlertManagerConfigReconciler) handleDeletion(ctx context.Context, amc *domain.AlertManagerConfig) error {
	log.FromContext(ctx).Info("Running reconciliation on deletion of a AlertManagerConfig")

	auth, err := utils.ExtractAuth(ctx, r.Client, amc.Spec.Auth, amc.ObjectMeta.Namespace)
	if err != nil {
		return fmt.Errorf("failed to extract authentication settings: %w", err)
	}

	return r.deleteAlertManagerConfigForTenant(ctx, auth, amc)
}

// reconcileRules ensures Mimir is synced with the PrometheusRules associated with a MimirRules
func (r *AlertManagerConfigReconciler) reconcileRules(ctx context.Context, amc *domain.AlertManagerConfig) error {
	log.FromContext(ctx).Info("Running reconciliation of the rules")

	auth, err := utils.ExtractAuth(ctx, r.Client, amc.Spec.Auth, amc.ObjectMeta.Namespace)
	if err != nil {
		return fmt.Errorf("failed to extract authentication settings: %w", err)
	}

	// TODO: config := r.unpack(amc)
	config := "convert config to string"
	return sendAMConfigToMimir(ctx, auth, amc.Spec.ID, amc.Spec.URL, config)
}

// setStatus updates the status of AlertManagerConfig after reconciliation
// If err is not nil, the error field is populated with the error and the status is set as "Failed"
// Otherwise, status is set as "Synced"
func (r *AlertManagerConfigReconciler) setStatus(ctx context.Context, amc *domain.AlertManagerConfig, err error) error {
	if err != nil {
		amc.Status.Status = "Failed"
		amc.Status.Error = err.Error()

		// Also log the error in the controller for clarity
		log.FromContext(ctx).Error(err, "Failed to reconcile AlertManagerConfig")
	} else {
		amc.Status.Status = "Synced"
		amc.Status.Error = ""
	}

	return r.Status().Update(context.Background(), amc)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlertManagerConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domain.AlertManagerConfig{}).
		Complete(r)
}
