/*
Copyright © 2024 Yevhen Lebid ealebed@gmail.com

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

package cmd

import (
	"fmt"
	"io"

	"github.com/ealebed/admission-webhook-certificator/cmd/version"
	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately
func Execute(out io.Writer) error {
	cmd := NewCmdRoot(out)
	return cmd.Execute()
}

// NewCmdRoot returns new root command
func NewCmdRoot(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.String(),
	}

	// create subcommands
	cmd.AddCommand(NewCreateAndSignCertCmd())

	return cmd
}

// CreateAndSignCertOptions represents options for create and sign certificate command
type CreateAndSignCertOptions struct {
	service    string
	namespace  string
	secret     string
	kubeconfig string
}

// NewDockerhubDeleteRepositoryCmd returns new docker delete repository command
func NewCreateAndSignCertCmd() *cobra.Command {
	options := CreateAndSignCertOptions{}

	cmd := &cobra.Command{
		Use:     "certify",
		Short:   "Create K8S Secret with TLS type which includes private key and corresponding client certificates signed by K8S CA.",
		Long:    "This tool generates a certificate for usage with a admission webhook service.\nCertificate is signed by k8s CA using CertificateSigningRequest API",
		Example: "certify [--service=webhook-svc --namespace=webhook --secret=webhook-certs]",
		RunE: func(cmd *cobra.Command, args []string) error {
			return createAndSignCert(options.service, options.namespace, options.secret, options.kubeconfig)
		},
	}

	cmd.Flags().StringVarP(&options.service, "service", "s", "", "Webhook service name.")
	cmd.Flags().StringVarP(&options.namespace, "namespace", "n", "webhook", "Namespace where webhook service and secret reside.")
	cmd.Flags().StringVarP(&options.secret, "secret", "t", "webhook-certs", "Secret name for CA certificate and server certificate/key pair.")
	cmd.Flags().StringVarP(&options.kubeconfig, "kubeconfig", "k", "", "kubeconfig path")

	if err := cmd.MarkFlagRequired("service"); err != nil {
		fmt.Println("`service` flag is required")
	}

	return cmd
}
