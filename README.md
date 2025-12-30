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

## Pre-commit hooks

Git pre-commit hooks are scripts that run automatically before a commit is finalized. They are used to enforce code quality, style, or other checks before changes are saved to the repository.

### Installation and usage

1. Install pre-commit Python package:

```bash
pip install pre-commit
```

or

```bash
brew install pre-commit
```

2. In the root of Git repository, a file named `.pre-commit-config.yaml` is already created with Go-specific hooks.

3. Install the hooks:

```bash
pre-commit install
```

This command will set up the necessary Git hook scripts in `.git/hooks` to run the hooks defined in your `.pre-commit-config.yaml`.

4. Manually run hooks:

```bash
pre-commit run --all-files
```

5. Depending on hooks configured, you might need to install additional packages/dependencies:

```bash
# golangci-lint (required for linting)
brew install golangci-lint
# or download from https://golangci-lint.run/usage/install/
```

The pre-commit hooks will automatically:
- Format code with `go fmt`
- Run `go vet` for static analysis
- Run `golangci-lint` for comprehensive linting
- Run `go mod tidy` to ensure dependencies are clean
- Run tests with race detector
- Check for common issues (trailing whitespace, large files, merge conflicts, etc.)
