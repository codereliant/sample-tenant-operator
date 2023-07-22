/*
Copyright 2023.

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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	multitenancyv1 "codereliant.io/tenant/api/v1"
)

const (
	finalizerName = "tenant.codereliant.io/finalizer"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=multitenancy.codereliant.io,resources=*,verbs=*
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=*

func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	tenant := &multitenancyv1.Tenant{}

	log.Info("Reconciling tenant")
	if err := r.Get(ctx, req.NamespacedName, tenant); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if tenant.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add a finalizer if not present
		if !controllerutil.ContainsFinalizer(tenant, finalizerName) {
			tenant.ObjectMeta.Finalizers = append(tenant.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, tenant); err != nil {
				log.Error(err, "unable to update Tenant")
				return ctrl.Result{}, err
			}
		}

		// Reconciliation logic for creating and managing namespaces
		for _, ns := range tenant.Spec.Namespaces {
			log.Info("Ensuring Namespace", "namespace", ns)
			if err := r.ensureNamespace(ctx, tenant, ns); err != nil {
				log.Error(err, "unable to ensure Namespace", "namespace", ns)
				return ctrl.Result{}, err
			}

			log.Info("Ensuring Admin RoleBinding", "namespace", ns)
			if err := r.ensureRoleBinding(ctx, ns, tenant.Spec.AdminGroups, "admin"); err != nil {
				log.Error(err, "unable to ensure Admin RoleBinding", "namespace", ns)
				return ctrl.Result{}, err
			}

			if err := r.ensureRoleBinding(ctx, ns, tenant.Spec.UserGroups, "edit"); err != nil {
				log.Error(err, "unable to ensure User RoleBinding", "namespace", ns)
				return ctrl.Result{}, err
			}
		}

		tenant.Status.NamespaceCount = len(tenant.Spec.Namespaces)
		tenant.Status.AdminEmail = tenant.Spec.AdminEmail
		if err := r.Status().Update(ctx, tenant); err != nil {
			log.Error(err, "unable to update Tenant status")
			return ctrl.Result{}, err
		}
	} else {
		// Check if the finalizer is present
		if controllerutil.ContainsFinalizer(tenant, finalizerName) {
			log.Info("Finalizer found, cleaning up resources")

			// Cleanup Resources
			if err := r.deleteExternalResources(ctx, tenant); err != nil {
				// retry if failed
				log.Error(err, "Failed to cleanup resources")
				return ctrl.Result{}, err
			}
			log.Info("Resource cleanup succeeded")

			// Remove the finalizer from the Tenant object once the cleanup succeded
			// This will free up tenant resource to be deleted
			controllerutil.RemoveFinalizer(tenant, finalizerName)
			if err := r.Update(ctx, tenant); err != nil {
				log.Error(err, "Unable to remove finalizer and update Tenant")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&multitenancyv1.Tenant{}).
		Complete(r)
}
