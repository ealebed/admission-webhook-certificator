package cmd

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"testing"

	certv1 "k8s.io/api/certificates/v1"
)

func TestGenerateCertificateRequest(t *testing.T) {
	tests := []struct {
		name      string
		service   string
		namespace string
		wantErr   bool
		validate  func(t *testing.T, csrPEM, keyPEM *bytes.Buffer, csrName string)
	}{
		{
			name:      "valid service and namespace",
			service:   "webhook-svc",
			namespace: "webhook",
			wantErr:   false,
			validate: func(t *testing.T, csrPEM, keyPEM *bytes.Buffer, csrName string) {
				if csrPEM == nil || csrPEM.Len() == 0 {
					t.Error("CSR PEM should not be empty")
				}
				if keyPEM == nil || keyPEM.Len() == 0 {
					t.Error("Private key PEM should not be empty")
				}
				if csrName != "webhook-svc.webhook" {
					t.Errorf("Expected CSR name 'webhook-svc.webhook', got '%s'", csrName)
				}

				// Validate CSR PEM format
				block, _ := pem.Decode(csrPEM.Bytes())
				if block == nil {
					t.Error("Failed to decode CSR PEM")
				}
				if block.Type != "CERTIFICATE REQUEST" {
					t.Errorf("Expected block type 'CERTIFICATE REQUEST', got '%s'", block.Type)
				}

				// Validate CSR content
				csr, err := x509.ParseCertificateRequest(block.Bytes)
				if err != nil {
					t.Errorf("Failed to parse certificate request: %v", err)
				}
				if csr.Subject.CommonName != "webhook-svc.webhook" {
					t.Errorf("Expected CommonName 'webhook-svc.webhook', got '%s'", csr.Subject.CommonName)
				}

				// Validate DNS names
				expectedDNSNames := []string{"webhook-svc", "webhook-svc.webhook", "webhook-svc.webhook.svc"}
				if len(csr.DNSNames) != len(expectedDNSNames) {
					t.Errorf("Expected %d DNS names, got %d", len(expectedDNSNames), len(csr.DNSNames))
				}
				for _, expected := range expectedDNSNames {
					found := false
					for _, dns := range csr.DNSNames {
						if dns == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected DNS name '%s' not found in %v", expected, csr.DNSNames)
					}
				}

				// Validate private key PEM format
				keyBlock, _ := pem.Decode(keyPEM.Bytes())
				if keyBlock == nil {
					t.Error("Failed to decode private key PEM")
				}
				if keyBlock.Type != "RSA PRIVATE KEY" {
					t.Errorf("Expected block type 'RSA PRIVATE KEY', got '%s'", keyBlock.Type)
				}
			},
		},
		{
			name:      "service with special characters",
			service:   "webhook-svc-v2",
			namespace: "default",
			wantErr:   false,
			validate: func(t *testing.T, csrPEM, keyPEM *bytes.Buffer, csrName string) {
				if csrName != "webhook-svc-v2.default" {
					t.Errorf("Expected CSR name 'webhook-svc-v2.default', got '%s'", csrName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csrPEM, keyPEM, csrName, err := generateCertificateRequest(tt.service, tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateCertificateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, csrPEM, keyPEM, csrName)
			}
		})
	}
}

func TestCreateCSRObject(t *testing.T) {
	tests := []struct {
		name      string
		csrName   string
		csrPEM    *bytes.Buffer
		wantErr   bool
		validate  func(t *testing.T, csr *certv1.CertificateSigningRequest)
	}{
		{
			name:    "valid CSR object",
			csrName: "webhook-svc.webhook",
			csrPEM:  bytes.NewBufferString("test-csr-data"),
			wantErr: false,
			validate: func(t *testing.T, csr *certv1.CertificateSigningRequest) {
				if csr == nil {
					t.Fatal("CSR object should not be nil")
				}
				if csr.Name != "webhook-svc.webhook" {
					t.Errorf("Expected CSR name 'webhook-svc.webhook', got '%s'", csr.Name)
				}
				if len(csr.Spec.Request) == 0 {
					t.Error("CSR request should not be empty")
				}
				if len(csr.Spec.Usages) == 0 {
					t.Error("CSR usages should not be empty")
				}
				// Validate required usages
				expectedUsages := []certv1.KeyUsage{
					certv1.UsageDigitalSignature,
					certv1.UsageKeyEncipherment,
					certv1.UsageServerAuth,
				}
				if len(csr.Spec.Usages) != len(expectedUsages) {
					t.Errorf("Expected %d usages, got %d", len(expectedUsages), len(csr.Spec.Usages))
				}
				for _, expected := range expectedUsages {
					found := false
					for _, usage := range csr.Spec.Usages {
						if usage == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected usage '%s' not found", expected)
					}
				}
				if csr.Spec.SignerName != "kubernetes.io/kube-apiserver-client" {
					t.Errorf("Expected signer name 'kubernetes.io/kube-apiserver-client', got '%s'", csr.Spec.SignerName)
				}
				if len(csr.Spec.Groups) == 0 || csr.Spec.Groups[0] != "system:authenticated" {
					t.Errorf("Expected groups to contain 'system:authenticated', got %v", csr.Spec.Groups)
				}
			},
		},
		{
			name:    "CSR with different name",
			csrName: "test-service.default",
			csrPEM:  bytes.NewBufferString("test-csr-data"),
			wantErr: false,
			validate: func(t *testing.T, csr *certv1.CertificateSigningRequest) {
				if csr.Name != "test-service.default" {
					t.Errorf("Expected CSR name 'test-service.default', got '%s'", csr.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csr := createCSRObject(tt.csrName, tt.csrPEM)
			if csr == nil && !tt.wantErr {
				t.Error("createCSRObject() returned nil, expected valid CSR object")
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, csr)
			}
		})
	}
}

