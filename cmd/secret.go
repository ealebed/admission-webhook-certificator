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
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func createOrUpdateSecret(cs *kubernetes.Clientset, ctx context.Context, clientCert []byte, clientPrivateKeyPEM *bytes.Buffer, namespace, secret string) error {
	tlsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secret,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.key": clientPrivateKeyPEM.Bytes(),
			"tls.crt": clientCert,
		},
	}

	log.Println("Secret, status: Check if already exists...")
	secretExistsInNamespace, _ := cs.CoreV1().Secrets(namespace).Get(ctx, secret, metav1.GetOptions{})
	if secretExistsInNamespace.Name == secret {
		log.Println("Secret, status: Already exists, updating")
		if _, err := cs.CoreV1().Secrets(namespace).Update(ctx, tlsSecret, metav1.UpdateOptions{}); err != nil {
			log.Printf("Update secret - error occurred, detail: %v", err)
			return err
		}
		log.Println("Secret, status: Updated")
	} else {
		log.Println("Secret, status: Not exists, creating")
		if _, err := cs.CoreV1().Secrets(namespace).Create(ctx, tlsSecret, metav1.CreateOptions{}); err != nil {
			log.Printf("Create secret - error occurred, detail: %v", err)
			return err
		}
		log.Println("Secret, status: Created")
	}

	return nil
}
