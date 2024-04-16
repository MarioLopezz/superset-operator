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

package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bigdatav1alpha1 "github.com/kubernetesbigdataeg/superset-operator/api/v1alpha1"
)

var _ = Describe("Superset controller", func() {
	Context("Superset controller test", func() {

		const SupersetName = "test-superset"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      SupersetName,
				Namespace: SupersetName,
			},
		}

		typeNamespaceName := types.NamespacedName{Name: SupersetName, Namespace: SupersetName}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

			By("Setting the Image ENV VAR which stores the Operand image")
			err = os.Setenv("SUPERSET_IMAGE", "example.com/image:test")
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			// TODO(user): Attention if you improve this code by adding other context test you MUST
			// be aware of the current delete namespace limitations. More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)

			By("Removing the Image ENV VAR which stores the Operand image")
			_ = os.Unsetenv("SUPERSET_IMAGE")
		})

		It("should successfully reconcile a custom resource for Superset", func() {
			By("Creating the custom resource for the Kind Superset")
			superset := &bigdatav1alpha1.Superset{}
			err := k8sClient.Get(ctx, typeNamespaceName, superset)
			if err != nil && errors.IsNotFound(err) {
				// Let's mock our custom resource at the same way that we would
				// apply on the cluster the manifest under config/samples
				superset := &bigdatav1alpha1.Superset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      SupersetName,
						Namespace: namespace.Name,
					},
					Spec: bigdatav1alpha1.SupersetSpec{
						MasterSize: 1,
						WorkerSize: 1,
						Service: bigdatav1alpha1.ServiceSpec{
							UiNodePort: 30213,
						},
						DbConnection: bigdatav1alpha1.DbConnection{
							RedisHost: "superset-redis-headless.superset.svc.cluster.local",
							RedisPort: "6379",
							DbHost:    "postgresql-svc.kudu.svc.cluster.local",
							DbPort:    "5432",
							DbUser:    "postgres",
							DbPass:    "postgres",
							DbName:    "superset",
						},
					},
				}

				err = k8sClient.Create(ctx, superset)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &bigdatav1alpha1.Superset{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			supersetReconciler := &SupersetReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = supersetReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			deployments := []string{"test-superset-master", "test-superset-worker"}
			for _, deploymentName := range deployments {
				By(fmt.Sprintf("Checking if %s StatefulSet was successfully created in the reconciliation", deploymentName))
				deploymentNamespaceName := types.NamespacedName{Name: deploymentName, Namespace: SupersetName}
				Eventually(func() error {
					found := &appsv1.Deployment{}
					return k8sClient.Get(ctx, deploymentNamespaceName, found)
				}, time.Minute, time.Second).Should(Succeed())
			}

			By("Checking the latest Status Condition added to the Superset instance")
			Eventually(func() error {
				if superset.Status.Conditions != nil && len(superset.Status.Conditions) != 0 {
					latestStatusCondition := superset.Status.Conditions[len(superset.Status.Conditions)-1]
					expectedLatestStatusCondition := metav1.Condition{Type: typeAvailableSuperset,
						Status: metav1.ConditionTrue, Reason: "Reconciling",
						Message: fmt.Sprintf("Deployment for custom resource (%s) with %d replicas created successfully", superset.Name, superset.Spec.MasterSize)}
					if latestStatusCondition != expectedLatestStatusCondition {
						return fmt.Errorf("The latest status condition added to the superset instance is not as expected")
					}
				}
				return nil
			}, time.Minute, time.Second).Should(Succeed())
		})
	})
})
