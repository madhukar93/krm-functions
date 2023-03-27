package main

import (
	"fmt"
	"os"

	"github.com/bukukasio/krm-functions/pkg/pubsub"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func cmd() *cobra.Command {
	p := &framework.VersionedAPIProcessor{FilterProvider: framework.GVKFilterMap{
		"PubsubTopic": {
			"krm/v1": &pubsub.PubsubTopic{},
		},
		"PubsubSubscription": {
			"krm/v1": &pubsub.PubsubSubscription{},
		},
	},
	}
	cmd := command.Build(p, command.StandaloneEnabled, false)
	cmd.Short = ""
	cmd.Long = ""
	return cmd
}

func main() {
	if err := cmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
