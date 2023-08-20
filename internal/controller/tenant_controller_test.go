package controller

import (
	"context"
	"fmt"

	multitenancyv1 "codereliant.io/tenant/api/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Tests related to the Tenant controller.
var _ = Describe("Tenant controller", func() {
	const (
		TenantName = "test-tenant"
	)

	ctx := context.Background()

	Context("When reconciling a Tenant", func() {
		// Tests the tenant creation process.
		It("should create corresponding namespaces and rolebindgings", func() {
			tenant := createTestTenant(TenantName)
			Expect(k8sClient.Create(ctx, tenant)).Should(Succeed())

			reconciler := &TenantReconciler{
				Client: k8sClient,
			}

			_, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: client.ObjectKey{Name: TenantName},
			})
			Expect(err).ToNot(HaveOccurred())

			for _, ns := range tenant.Spec.Namespaces {
				// Checking the annotations of the namespace.
				namespace := fetchNamespace(ctx, ns, k8sClient)
				Expect(namespace.Annotations["adminEmail"]).To(Equal(tenant.Spec.AdminEmail), "Expected adminEmail annotation to match")

				// Verifying the admin RoleBinding exists.
				adminRoleBinding := fetchRoleBinding(ctx, ns, fmt.Sprintf("%s-admin-rb", ns), k8sClient)
				Expect(adminRoleBinding).NotTo(BeNil(), "Expected admin RoleBinding to exist")

				// Verifying the user RoleBinding exists.
				userRoleBinding := fetchRoleBinding(ctx, ns, fmt.Sprintf("%s-edit-rb", ns), k8sClient)
				Expect(userRoleBinding).NotTo(BeNil(), "Expected user RoleBinding to exist")
			}
		})
	})
})

func createTestTenant(name string) *multitenancyv1.Tenant {
	return &multitenancyv1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: multitenancyv1.TenantSpec{
			AdminEmail:  "test@example.com",
			Namespaces:  []string{"test-namespace"},
			AdminGroups: []string{"test-admin-group"},
			UserGroups:  []string{"test-user-group"},
		},
	}
}

func fetchNamespace(ctx context.Context, name string, k8sClient client.Client) *corev1.Namespace {
	ns := &corev1.Namespace{}
	err := k8sClient.Get(ctx, client.ObjectKey{Name: name}, ns)
	Expect(err).ToNot(HaveOccurred(), "Failed to fetch namespace: %s", name)
	return ns
}

func fetchRoleBinding(ctx context.Context, nsName, roleName string, k8sClient client.Client) *rbacv1.RoleBinding {
	rb := &rbacv1.RoleBinding{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: nsName, Name: roleName}, rb)
	Expect(err).ToNot(HaveOccurred(), "Failed to fetch RoleBinding in namespace %s with name %s", nsName, roleName)
	return rb
}
