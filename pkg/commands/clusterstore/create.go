// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package clusterstore

import (
	"github.com/spf13/cobra"

	"github.com/pivotal/build-service-cli/pkg/buildpackage"
	"github.com/pivotal/build-service-cli/pkg/clusterstore"
	"github.com/pivotal/build-service-cli/pkg/commands"

	"github.com/pivotal/build-service-cli/pkg/k8s"
	"github.com/pivotal/build-service-cli/pkg/registry"
)

func NewCreateCommand(clientSetProvider k8s.ClientSetProvider, rup registry.UtilProvider) *cobra.Command {
	var (
		buildpackages []string
		tlsCfg        registry.TLSConfig
	)

	cmd := &cobra.Command{
		Use:   "create <store> -b <buildpackage> [-b <buildpackage>...]",
		Short: "Create a cluster store",
		Long: `Create a cluster-scoped buildpack store by providing command line arguments.

Buildpackages will be uploaded to the canonical repository.
Therefore, you must have credentials to access the registry on your machine.

This clusterstore will be created only if it does not exist.
The canonical repository is read from the "canonical.repository" key in the "kp-config" ConfigMap within "kpack" namespace.
`,
		Example: `kp clusterstore create my-store -b my-registry.com/my-buildpackage
kp clusterstore create my-store -b my-registry.com/my-buildpackage -b my-registry.com/my-other-buildpackage
kp clusterstore create my-store -b ../path/to/my-local-buildpackage.cnb`,
		Args:         commands.ExactArgsWithUsage(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := clientSetProvider.GetClientSet("")
			if err != nil {
				return err
			}

			ch, err := commands.NewCommandHelper(cmd)
			if err != nil {
				return err
			}

			factory, err := newClusterStoreFactory(cs, ch, rup, tlsCfg)
			if err != nil {
				return err
			}

			name := args[0]
			return create(name, buildpackages, factory, ch, cs)
		},
	}

	cmd.Flags().StringArrayVarP(&buildpackages, "buildpackage", "b", []string{}, "location of the buildpackage")
	commands.SetImgUploadDryRunOutputFlags(cmd)
	commands.SetTLSFlags(cmd, &tlsCfg)
	return cmd
}

func newClusterStoreFactory(cs k8s.ClientSet, ch *commands.CommandHelper, rup registry.UtilProvider, tlsCfg registry.TLSConfig) (*clusterstore.Factory, error) {
	repo, err := k8s.DefaultConfigHelper(cs).GetCanonicalRepository()
	if err != nil {
		return nil, err
	}

	return &clusterstore.Factory{
		Uploader: &buildpackage.Uploader{
			Fetcher:   rup.Fetcher(),
			Relocator: rup.Relocator(ch.CanChangeState()),
		},
		TLSConfig:  tlsCfg,
		Repository: repo,
		Printer:    ch,
	}, nil
}

func create(name string, buildpackages []string, factory *clusterstore.Factory, ch *commands.CommandHelper, cs k8s.ClientSet) (err error) {
	if err = ch.PrintStatus("Creating ClusterStore..."); err != nil {
		return err
	}

	newStore, err := factory.MakeStore(name, buildpackages...)
	if err != nil {
		return err
	}

	if !ch.IsDryRun() {
		newStore, err = cs.KpackClient.KpackV1alpha1().ClusterStores().Create(newStore)
		if err != nil {
			return err
		}
	}

	if err = ch.PrintObj(newStore); err != nil {
		return err
	}

	return ch.PrintResult("ClusterStore %q created", name)
}
