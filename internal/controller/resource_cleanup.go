package controller

import (
	"context"

	multitenancyv1 "codereliant.io/tenant/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *TenantReconciler) deleteExternalResources(ctx context.Context, tenant *multitenancyv1.Tenant) error {
	// Delete any external resources created for this tenant
	log := log.FromContext(ctx)
	for _, ns := range tenant.Spec.Namespaces {
		// Delete Namespace
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}
		if err := r.Delete(ctx, namespace); client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to delete Namespace", "namespace", ns)
			return err
		}
		log.Info("Namespace deleted", "namespace", ns)
	}
	log.Info("All resources deleted for tenant", "tenant", tenant.Name)
	return nil
}
