// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package customclusterbuilder_test

import (
	"testing"

	expv1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	kpackfakes "github.com/pivotal/kpack/pkg/client/clientset/versioned/fake"
	"github.com/sclevine/spec"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfakes "k8s.io/client-go/kubernetes/fake"

	"github.com/pivotal/build-service-cli/pkg/commands/customclusterbuilder"
	"github.com/pivotal/build-service-cli/pkg/testhelpers"
)

func TestCustomClusterBuilderCreateCommand(t *testing.T) {
	spec.Run(t, "TestCustomBuilderCreateCommand", testCustomClusterBuilderCreateCommand)
}

func testCustomClusterBuilderCreateCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		config = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{
				"canonical.repository":                "some-registry/some-project",
				"canonical.repository.serviceaccount": "some-serviceaccount",
			},
		}

		expectedBuilder = &expv1alpha1.CustomClusterBuilder{
			TypeMeta: metav1.TypeMeta{
				Kind:       expv1alpha1.CustomClusterBuilderKind,
				APIVersion: "experimental.kpack.pivotal.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "test-builder",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"some-stack"},"store":{"kind":"ClusterStore","name":"some-store"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`,
				},
			},
			Spec: expv1alpha1.CustomClusterBuilderSpec{
				CustomBuilderSpec: expv1alpha1.CustomBuilderSpec{
					Tag: "some-registry/some-project/test-builder",
					Stack: corev1.ObjectReference{
						Name: "some-stack",
						Kind: expv1alpha1.ClusterStackKind,
					},
					Store: corev1.ObjectReference{
						Name: "some-store",
						Kind: expv1alpha1.ClusterStoreKind,
					},
					Order: []expv1alpha1.OrderEntry{
						{
							Group: []expv1alpha1.BuildpackRef{
								{
									BuildpackInfo: expv1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.nodejs",
									},
								},
							},
						},
						{
							Group: []expv1alpha1.BuildpackRef{
								{
									BuildpackInfo: expv1alpha1.BuildpackInfo{
										Id: "org.cloudfoundry.go",
									},
								},
							},
						},
					},
				},
				ServiceAccountRef: corev1.ObjectReference{
					Namespace: "kpack",
					Name:      "some-serviceaccount",
				},
			},
		}
	)

	cmdFunc := func(k8sClientSet *k8sfakes.Clientset, kpackClientSet *kpackfakes.Clientset) *cobra.Command {
		clientSetProvider := testhelpers.GetFakeClusterProvider(k8sClientSet, kpackClientSet)
		return customclusterbuilder.NewCreateCommand(clientSetProvider)
	}

	it("creates a CustomClusterBuilder", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			Args: []string{
				expectedBuilder.Name,
				"--tag", expectedBuilder.Spec.Tag,
				"--stack", expectedBuilder.Spec.Stack.Name,
				"--store", expectedBuilder.Spec.Store.Name,
				"--order", "./testdata/order.yaml",
			},
			ExpectedOutput: `"test-builder" created
`,
			ExpectCreates: []runtime.Object{
				expectedBuilder,
			},
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("creates a CustomClusterBuilder with the default stack", func() {
		expectedBuilder.Spec.Stack.Name = "default"
		expectedBuilder.Spec.Store.Name = "default"
		expectedBuilder.Annotations["kubectl.kubernetes.io/last-applied-configuration"] = `{"kind":"CustomClusterBuilder","apiVersion":"experimental.kpack.pivotal.io/v1alpha1","metadata":{"name":"test-builder","creationTimestamp":null},"spec":{"tag":"some-registry/some-project/test-builder","stack":{"kind":"ClusterStack","name":"default"},"store":{"kind":"ClusterStore","name":"default"},"order":[{"group":[{"id":"org.cloudfoundry.nodejs"}]},{"group":[{"id":"org.cloudfoundry.go"}]}],"serviceAccountRef":{"namespace":"kpack","name":"some-serviceaccount"}},"status":{"stack":{}}}`

		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			Args: []string{
				expectedBuilder.Name,
				"--tag", expectedBuilder.Spec.Tag,
				"--order", "./testdata/order.yaml",
			},
			ExpectedOutput: "\"test-builder\" created\n",
			ExpectCreates: []runtime.Object{
				expectedBuilder,
			},
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("creates a CustomClusterBuilder with the canonical tag when tag is not specified", func() {
		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				config,
			},
			Args: []string{
				expectedBuilder.Name,
				"--stack", expectedBuilder.Spec.Stack.Name,
				"--store", expectedBuilder.Spec.Store.Name,
				"--order", "./testdata/order.yaml",
			},
			ExpectedOutput: `"test-builder" created
`,
			ExpectCreates: []runtime.Object{
				expectedBuilder,
			},
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("fails when kp-config map is not found", func() {
		testhelpers.CommandTest{
			Args: []string{
				expectedBuilder.Name,
				"--tag", expectedBuilder.Spec.Tag,
				"--stack", expectedBuilder.Spec.Stack.Name,
				"--store", expectedBuilder.Spec.Store.Name,
				"--order", "./testdata/order.yaml",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical service account: configmaps "kp-config" not found
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("fails when canonical.repository.serviceaccount key is not found in kp-config configmap", func() {
		badConfig := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{},
		}

		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				badConfig,
			},
			Args: []string{
				expectedBuilder.Name,
				"--tag", expectedBuilder.Spec.Tag,
				"--stack", expectedBuilder.Spec.Stack.Name,
				"--store", expectedBuilder.Spec.Store.Name,
				"--order", "./testdata/order.yaml",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical service account: key "canonical.repository.serviceaccount" not found in configmap "kp-config"
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})

	it("fails when tag is not specified and canonical.repository key is not found in kp-config configmap", func() {
		badConfig := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kp-config",
				Namespace: "kpack",
			},
			Data: map[string]string{
				"canonical.repository.serviceaccount": "some-serviceaccount",
			},
		}

		testhelpers.CommandTest{
			K8sObjects: []runtime.Object{
				badConfig,
			},
			Args: []string{
				expectedBuilder.Name,
				"--stack", expectedBuilder.Spec.Stack.Name,
				"--store", expectedBuilder.Spec.Store.Name,
				"--order", "./testdata/order.yaml",
			},
			ExpectErr: true,
			ExpectedOutput: `Error: failed to get canonical repository: key "canonical.repository" not found in configmap "kp-config"
`,
		}.TestK8sAndKpack(t, cmdFunc)
	})
}
