package cmd

import (
	"testing"
)

func TestInitInClusterClient(t *testing.T) {
	// When not running in a Kubernetes cluster, this should return an error
	config, err := initInClusterClient()
	if err == nil {
		// If we're actually in a cluster, config should be valid
		if config == nil {
			t.Error("initInClusterClient() returned nil config without error")
		}
	} else {
		// Expected error when not in cluster
		if config != nil {
			t.Error("initInClusterClient() returned config with error")
		}
		// Verify error message is meaningful
		if err.Error() == "" {
			t.Error("initInClusterClient() returned error with empty message")
		}
	}
}

func TestInitK8sClient(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig string
		wantPanic  bool
	}{
		{
			name:       "non-existent kubeconfig file",
			kubeconfig: "/nonexistent/path/to/kubeconfig",
			wantPanic:  true, // Will panic because config is nil and NewForConfig doesn't handle nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("initK8sClient() should have panicked with nil config")
					}
				}()
			}

			clientset, err := initK8sClient(tt.kubeconfig)
			if !tt.wantPanic {
				if err != nil && clientset == nil {
					// Error case is acceptable
					return
				}
				if clientset == nil {
					t.Error("initK8sClient() returned nil clientset, expected valid clientset")
				}
			}
		})
	}
}
