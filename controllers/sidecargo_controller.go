/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/togettoyou/sidecar-go/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/togettoyou/sidecar-go/api/v1alpha1"
)

// SidecarGoReconciler reconciles a SidecarGo object
type SidecarGoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.togettoyou.com,resources=sidecargoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.togettoyou.com,resources=sidecargoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.togettoyou.com,resources=sidecargoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SidecarGo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *SidecarGoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	sidecarGo := &appsv1alpha1.SidecarGo{}
	err := r.Get(ctx, req.NamespacedName, sidecarGo)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("SidecarGo delete")
			return ctrl.Result{}, util.UpdateSidecarGoSpec(req.NamespacedName.String(), nil)
		}
		return ctrl.Result{}, err
	}

	logger.Info("SidecarGo apply")
	return ctrl.Result{}, util.UpdateSidecarGoSpec(req.NamespacedName.String(), &sidecarGo.Spec)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SidecarGoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.SidecarGo{}).
		Complete(r)
}
