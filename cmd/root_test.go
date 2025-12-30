package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCmdRoot(t *testing.T) {
	out := bytes.NewBufferString("")
	cmd := NewCmdRoot(out)

	if cmd == nil {
		t.Fatal("NewCmdRoot() returned nil, expected valid command")
	}

	if !cmd.SilenceUsage {
		t.Error("Expected SilenceUsage to be true")
	}

	if !cmd.SilenceErrors {
		t.Error("Expected SilenceErrors to be true")
	}

	// Check that the certify subcommand is added
	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Fatal("Expected at least one subcommand, got none")
	}

	foundCertify := false
	for _, subcmd := range subcommands {
		if subcmd.Use == "certify" {
			foundCertify = true
			break
		}
	}

	if !foundCertify {
		t.Error("Expected 'certify' subcommand to be present")
	}
}

func TestNewCreateAndSignCertCmd(t *testing.T) {
	cmd := NewCreateAndSignCertCmd()

	if cmd == nil {
		t.Fatal("NewCreateAndSignCertCmd() returned nil, expected valid command")
	}

	if cmd.Use != "certify" {
		t.Errorf("Expected command use 'certify', got '%s'", cmd.Use)
	}

	// Check that required flags exist
	serviceFlag := cmd.Flag("service")
	if serviceFlag == nil {
		t.Error("Expected 'service' flag to be present")
	}

	namespaceFlag := cmd.Flag("namespace")
	if namespaceFlag == nil {
		t.Error("Expected 'namespace' flag to be present")
	} else {
		if namespaceFlag.DefValue != "webhook" {
			t.Errorf("Expected 'namespace' flag default value 'webhook', got '%s'", namespaceFlag.DefValue)
		}
	}

	secretFlag := cmd.Flag("secret")
	if secretFlag == nil {
		t.Error("Expected 'secret' flag to be present")
	} else {
		if secretFlag.DefValue != "webhook-certs" {
			t.Errorf("Expected 'secret' flag default value 'webhook-certs', got '%s'", secretFlag.DefValue)
		}
	}

	kubeconfigFlag := cmd.Flag("kubeconfig")
	if kubeconfigFlag == nil {
		t.Error("Expected 'kubeconfig' flag to be present")
	}

	// Check short flags (cobra uses flag shorthands as aliases)
	// Short flags are accessible through the same flag name
	// We verify they exist by checking the flag definition
	if serviceFlag.Shorthand != "s" {
		t.Errorf("Expected 's' short flag for service, got '%s'", serviceFlag.Shorthand)
	}
	if namespaceFlag.Shorthand != "n" {
		t.Errorf("Expected 'n' short flag for namespace, got '%s'", namespaceFlag.Shorthand)
	}
	if secretFlag.Shorthand != "t" {
		t.Errorf("Expected 't' short flag for secret, got '%s'", secretFlag.Shorthand)
	}
	if kubeconfigFlag.Shorthand != "k" {
		t.Errorf("Expected 'k' short flag for kubeconfig, got '%s'", kubeconfigFlag.Shorthand)
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "help command",
			args:    []string{"--help"},
			wantErr: false, // Help should not error
		},
		{
			name:    "version command",
			args:    []string{"--version"},
			wantErr: false, // Version should not error
		},
		{
			name:    "certify without required service flag",
			args:    []string{"certify"},
			wantErr: true, // Should error because service is required
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := bytes.NewBufferString("")
			cmd := NewCmdRoot(out)
			cmd.SetArgs(tt.args)
			cmd.SetOut(out)
			cmd.SetErr(out)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCmdRootCommandStructure(t *testing.T) {
	out := bytes.NewBufferString("")
	cmd := NewCmdRoot(out)

	// Test that the command has proper structure
	if cmd.Short == "" && cmd.Long == "" {
		// Root command might not have description, that's okay
	}

	// Verify subcommands are properly registered
	subcommands := cmd.Commands()
	if len(subcommands) < 1 {
		t.Error("Expected at least one subcommand")
	}

	// Verify certify command structure
	var certifyCmd *cobra.Command
	for _, subcmd := range subcommands {
		if subcmd.Use == "certify" {
			certifyCmd = subcmd
			break
		}
	}

	if certifyCmd == nil {
		t.Fatal("Could not find certify subcommand")
	}

	// Check that certify command has proper description
	if certifyCmd.Short == "" {
		t.Error("Expected certify command to have a short description")
	}

	if !strings.Contains(certifyCmd.Short, "K8S Secret") {
		t.Error("Expected certify command short description to mention 'K8S Secret'")
	}
}
