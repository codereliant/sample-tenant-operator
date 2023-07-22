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

package v1

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var tenantlog = logf.Log.WithName("tenant-resource")

// +kubebuilder:object:generate=false

type TenantValidator struct {
	client.Client
}

func (r *Tenant) SetupWebhookWithManager(mgr ctrl.Manager) error {
	validator := &TenantValidator{
		mgr.GetClient(),
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(validator).
		Complete()
}

// +kubebuilder:webhook:path=/validate-multitenancy-codereliant-io-v1-tenant,mutating=false,failurePolicy=fail,sideEffects=None,groups=multitenancy.codereliant.io,resources=tenants,verbs=create;update,versions=v1,name=vtenant.kb.io,admissionReviewVersions=v1

// var _ webhook.Validator = &tenantValidator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (v *TenantValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	tenant, ok := obj.(*Tenant)
	if !ok {
		return nil, fmt.Errorf("unexpected object type, expected Tenant")
	}

	var namespaces corev1.NamespaceList
	if err := v.List(ctx, &namespaces); err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	for _, ns := range tenant.Spec.Namespaces {
		// Check if namespace already exists
		if namespaceExists(namespaces, ns) {
			return nil, fmt.Errorf("namespace %s already exists", ns)
		}
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (v *TenantValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldTenant, ok := oldObj.(*Tenant)
	if !ok {
		return nil, fmt.Errorf("unexpected old object type, expected *Tenant")
	}

	newTenant, ok := newObj.(*Tenant)
	if !ok {
		return nil, fmt.Errorf("unexpected new object type, expected *Tenant")
	}

	var namespaces corev1.NamespaceList
	if err := v.List(ctx, &namespaces); err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	// Check new namespaces in spec
	for _, ns := range newTenant.Spec.Namespaces {
		if !contains(oldTenant.Spec.Namespaces, ns) {
			if namespaceExists(namespaces, ns) {
				return nil, fmt.Errorf("namespace %s already exists", ns)
			}
		}
	}

	return nil, nil
}

// // ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (v *TenantValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	tenantlog.Info("validate delete", "name", "test")

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

// Check if a namespace exists in a list
func namespaceExists(namespaces corev1.NamespaceList, ns string) bool {
	for _, namespace := range namespaces.Items {
		if namespace.Name == ns {
			return true
		}
	}
	return false
}

// Check if a namespace is contained in a list
func contains(namespaces []string, ns string) bool {
	for _, n := range namespaces {
		if n == ns {
			return true
		}
	}
	return false
}
