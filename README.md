# admission-webhook-certificator
Tool for creating K8S Secret (with TLS type) which contains private key and signed by K8S CA client certificate.

### Description
Generate a certificate suitable for use with a admission webhook services.

**NOTE:** This tool was initially created for usage with admission webhook service described [here](https://github.com/ealebed/token-injector).

This cli tool uses k8s' CertificateSigningRequest API to generate a certificate signed by k8s CA suitable for use with sidecar-injector webhook services. This requires permissions to create and approve CSR. See [Kubernetes TLS management](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/) for detailed explanation and additional instructions.

### Understanding the problem 
Kubernetes Admission Webhook has a requirement that apiserver and admission webhook server must connect via TLS with each other, see [contacting the webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#contacting-the-webhook).
To ensure that we need a CA (Certificate Authority) and a client certificate which is signed by this CA. 

There are many alternative ways to do that like creating a scripts that create CA and a client itself using `openssl` cli or using Kubernetes TLS management which is create client certificates by approving CSR's. 

### Solution 
This cli tool helps to create CSR (CertificateSigningRequest) with a client certificate which is approved by this CSR with CA which is belongs to Kubernetes cluster itself and then creating a Kubernetes Secret which includes private key and a client certificate.
The whole process could be completed by calling this cli tool in Kubernetes Job.
