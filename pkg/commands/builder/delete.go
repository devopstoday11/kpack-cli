// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pivotal/build-service-cli/pkg/k8s"
)

func NewDeleteCommand(clientSetProvider k8s.ClientSetProvider) *cobra.Command {
	var (
		namespace string
	)

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a builder",
		Long: `Delete a builder in the provided namespace.

The namespace defaults to the kubernetes current-context namespace.`,
		Example: "kp builder delete my-builder\nkp builder delete -n my-namespace other-builder",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet(namespace)
			if err != nil {
				return err
			}

			err = cs.KpackClient.KpackV1alpha1().Builders(cs.Namespace).Delete(args[0], &metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "\"%s\" deleted\n", args[0])
			return err
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace")

	return cmd
}