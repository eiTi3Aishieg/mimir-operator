package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mimirrandgenxyzv1alpha1 "mimir-operator/api/v1alpha1"
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
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlertManagerConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mimirrandgenxyzv1alpha1.AlertManagerConfig{}).
		Complete(r)
}
