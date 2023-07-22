package controller

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *TenantReconciler) ensureRoleBinding(ctx context.Context, namespaceName string, groups []string, clusterRoleName string) error {
	log := log.FromContext(ctx)

	roleBindingName := fmt.Sprintf("%s-%s-rb", namespaceName, clusterRoleName)

	clusterRole := &rbacv1.ClusterRole{}
	err := r.Get(ctx, client.ObjectKey{Name: clusterRoleName}, clusterRole)
	if err != nil {
		log.Error(err, "Failed to get ClusterRole", "clusterRole", clusterRoleName)
	}

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: namespaceName,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
			APIGroup: rbacv1.GroupName,
		},
		Subjects: make([]rbacv1.Subject, len(groups)),
	}

	for i, group := range groups {
		roleBinding.Subjects[i] = rbacv1.Subject{
			Kind:     "Group",
			Name:     group,
			APIGroup: rbacv1.GroupName,
		}
	}

	err = r.Get(ctx, client.ObjectKey{Name: roleBindingName, Namespace: namespaceName}, roleBinding)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Creating RoleBinding", "roleBinding", roleBindingName, "namespace", namespaceName)

			err = r.Create(ctx, roleBinding)
			if err != nil {
				log.Error(err, "Failed to create RoleBinding", "roleBinding", roleBindingName, "namespace", namespaceName)
				return err
			}
		} else {
			log.Error(err, "Failed to get RoleBinding", "roleBinding", roleBindingName, "namespace", namespaceName)
		}
	} else {
		// Compare current and desired roleBinding
		groupsChanged := false

		existingGroups := make(map[string]bool)
		newGroups := make(map[string]bool)

		for _, subject := range roleBinding.Subjects {
			if subject.Kind == "Group" {
				existingGroups[subject.Name] = true
			}
		}

		for _, group := range groups {
			newGroups[group] = true
			if _, exists := existingGroups[group]; !exists {
				groupsChanged = true
				break
			}
		}

		if len(existingGroups) != len(newGroups) {
			groupsChanged = true
		}

		if groupsChanged {
			log.Info("Updating RoleBinding", "roleBinding", roleBindingName, "namespace", namespaceName)

			roleBinding.Subjects = make([]rbacv1.Subject, len(groups))
			for i, group := range groups {
				roleBinding.Subjects[i] = rbacv1.Subject{
					Kind:     "Group",
					Name:     group,
					APIGroup: rbacv1.GroupName,
				}
			}

			err = r.Update(ctx, roleBinding)
			if err != nil {
				log.Error(err, "Failed to update RoleBinding", "roleBinding", roleBindingName, "namespace", namespaceName)
				return err
			}
		} else {
			log.Info("RoleBinding already exists", "roleBinding", roleBindingName, "namespace", namespaceName)
		}
	}

	return nil
}
