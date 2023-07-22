package controller

import (
	"context"

	multitenancyv1 "codereliant.io/tenant/api/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	tenantOperatorAnnotation = "tenant-operator"
)

func (r *TenantReconciler) ensureNamespace(ctx context.Context, tenant *multitenancyv1.Tenant, namespaceName string) error {
	log := log.FromContext(ctx)

	// Define a namespace object
	namespace := &corev1.Namespace{}

	// Attempt to get the namespace with the provided name
	err := r.Get(ctx, client.ObjectKey{Name: namespaceName}, namespace)
	if err != nil {
		// If the namespace doesn't exist, create it
		if apierrors.IsNotFound(err) {
			log.Info("Creating Namespace", "namespace", namespaceName)
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
					Annotations: map[string]string{
						"adminEmail": tenant.Spec.AdminEmail,
						"managed-by": tenantOperatorAnnotation,
					},
				},
			}

			// Attempt to create the namespace
			if err = r.Create(ctx, namespace); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// If the namespace already exists, check for required annotations
		log.Info("Namespace already exists", "namespace", namespaceName)

		// If the namespace does not have any annotations, initialize the annotations map
		if namespace.Annotations == nil {
			namespace.Annotations = map[string]string{}
		}

		// Define required annotations and their desired values
		requiredAnnotations := map[string]string{
			"adminEmail": tenant.Spec.AdminEmail,
			"managed-by": tenantOperatorAnnotation,
		}

		// Iterate over the required annotations and update if necessary
		for annotationKey, desiredValue := range requiredAnnotations {
			existingValue, ok := namespace.Annotations[annotationKey]
			if !ok || existingValue != desiredValue {
				log.Info("Updating namespace annotation", "namespace", namespaceName, "annotation", annotationKey)
				namespace.Annotations[annotationKey] = desiredValue
				if err = r.Update(ctx, namespace); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
