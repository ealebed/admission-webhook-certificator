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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"strings"
	"time"

	certv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/certificates/v1"
)

const (
	csrNameTemplate0 = "${service}"
	csrNameTemplate1 = "${service}.${namespace}"
	csrNameTemplate2 = "${service}.${namespace}.svc"
)

func createAndSignCert(service, namespace, secret, kubeconfig string) error {
	start := time.Now()

	ctx := context.TODO()
	cs, _ := initK8sClient(kubeconfig)

	r := strings.NewReplacer("${service}", service, "${namespace}", namespace)

	clientPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("rsa.GenerateKey - error occurred, detail: %v", err)
	}

	csrNameWithService := r.Replace(csrNameTemplate0)
	csrNameWithServiceAndNamespace := r.Replace(csrNameTemplate1)
	csrNameFull := r.Replace(csrNameTemplate2)

	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: csrNameWithServiceAndNamespace,
		},
		DNSNames: []string{csrNameWithService, csrNameWithServiceAndNamespace, csrNameFull},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, clientPrivateKey)
	if err != nil {
		log.Fatalf("x509.CreateCertificateRequest - error occurred, detail: %v", err)
	}

	clientCSRPEM := new(bytes.Buffer)
	_ = pem.Encode(clientCSRPEM, &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	clientPrivateKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(clientPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientPrivateKey),
	})

	csrClient := cs.CertificatesV1().CertificateSigningRequests()

	csr := &certv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: csrNameWithServiceAndNamespace,
		},
		Spec: certv1.CertificateSigningRequestSpec{
			Request:    clientCSRPEM.Bytes(),
			Usages:     []certv1.KeyUsage{certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment, certv1.UsageServerAuth},
			Groups:     []string{"system:authenticated"},
			SignerName: "kubernetes.io/kube-apiserver-client",
		},
	}

	if err := createCSR(csrClient, ctx, csr, csrNameWithServiceAndNamespace); err != nil {
		log.Fatalf("Create CertificateSigningRequest - error occurred, detail: %v", err)
	}

	if err := approveCSR(csrClient, ctx, csr); err != nil {
		log.Fatalf("Approve CertificateSigningRequest - error occurred, detail: %v", err)
	}

	updatedCsr, err := retrieveUpdatedCSR(csrClient, ctx, csrNameWithServiceAndNamespace)
	if err != nil {
		log.Fatalf("Retrieve updated CertificateSigningRequest - error occurred, detail: %v", err)
	}

	clientCert := updatedCsr.Status.Certificate
	if err := createOrUpdateSecret(cs, ctx, clientCert, clientPrivateKeyPEM, namespace, secret); err != nil {
		log.Fatalf("Secret, status: Error occurred, detail: %v", err)
	}

	log.Printf("Done in %d milliseconds", time.Since(start).Milliseconds())

	return nil
}

func createCSR(csrClient v1.CertificateSigningRequestInterface, ctx context.Context,
	csr *certv1.CertificateSigningRequest, csrNameWithServiceAndNamespace string) error {

	log.Println("Certificate signing request, status: Check if already exists")
	csExistInCluster, _ := csrClient.Get(ctx, csrNameWithServiceAndNamespace, metav1.GetOptions{})
	if csExistInCluster.Status.Certificate != nil {
		log.Println("Certificate signing request, status: Already exists, deleting")
		if err := csrClient.Delete(ctx, csrNameWithServiceAndNamespace, metav1.DeleteOptions{}); err != nil {
			log.Printf("Delete CertificateSigningRequest - error occurred, detail: %v, but ignored", err)
			return err
		}
		log.Println("Certificate signing request, status: Deleted")
	}

	log.Println("Certificate signing request, status: Not exists, creating")
	if _, err := csrClient.Create(ctx, csr, metav1.CreateOptions{}); err != nil {
		log.Printf("Create CertificateSigningRequest - error occurred, detail: %v", err)
		return err
	}
	log.Println("Certificate signing request, status: Created")

	return nil
}

func approveCSR(csrClient v1.CertificateSigningRequestInterface, ctx context.Context,
	csr *certv1.CertificateSigningRequest) error {
	log.Println("Certificate signing request, status: Approving")

	csr.Status.Conditions = append(csr.Status.Conditions, certv1.CertificateSigningRequestCondition{
		Type:           certv1.CertificateApproved,
		Status:         corev1.ConditionTrue,
		Reason:         "Self-generated and auto-approved by certificator",
		Message:        "This CSR was approved by certificator cli",
		LastUpdateTime: metav1.Now(),
	})

	if _, err := csrClient.UpdateApproval(ctx, csr.Name, csr, metav1.UpdateOptions{}); err != nil {
		log.Printf("UpdateApproval - error occurred, detail: %v", err)
		return err
	}
	log.Println("Certificate signing request, status: Approved")

	return nil
}

func retrieveUpdatedCSR(csrClient v1.CertificateSigningRequestInterface, ctx context.Context,
	csrNameWithServiceAndNamespace string) (*certv1.CertificateSigningRequest, error) {
	log.Println("Certificate signing request, status: Retrieving updated CSR")

	var updatedCsr *certv1.CertificateSigningRequest
	var attempt = 0

	for {
		if attempt < 5 {
			res, err := csrClient.Get(ctx, csrNameWithServiceAndNamespace, metav1.GetOptions{})
			if err != nil {
				log.Fatalf("Get CertificateSigningRequest - error occurred, detail: %v", err)
			}
			updatedCsr = res
			if updatedCsr.Status.Certificate != nil {
				log.Println("Certificate signing request, status: Certificate Found")
				break
			}
			log.Println("Certificate signing request, status: No certificate found trying after 1 sec")
			time.Sleep(1 * time.Second)
		} else {
			log.Printf("Certificate signing request, status: No certificate found, backed off after 5 attempt")
			return nil, fmt.Errorf("certificate signing request, status: No certificate found")
		}
		attempt += 1
	}

	log.Println("Certificate signing request, status: Retrieved")

	return updatedCsr, nil
}
