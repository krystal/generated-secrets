# Generated Secrets Operator

A Kubernetes operator that can generate secrets containing randomly generated values. Just create a `GeneratedSecret` resource and the operator will create a `Secret` resource with the generated values for your application to consume however you see fit.

## Getting started

1. Apply to your cluster (`kubectl apply -f https://github.com/krystal/generated-secrets/releases/latest/download/manifest.yaml`)
2. Create a `GeneratedSecret` resource
3. The operator will create a `Secret` resource with the generated values

## Example

```yaml
apiVersion: secrets.k8s.k.io/v1
kind: GeneratedSecret
metadata:
  name: my-secret
spec:
  keys:
    - name: secret-key-base
      type: Hex
      length: 128
    - name: database-password
      type: Alphanumeric
      length: 32
    - name: some-uuid
      type: UUID
```

## Supported types

Keys can use any of the following types. With the exception of UUID, they all require the `Length` attribute.

- Base64
- Base64URL
- Hex
- Alphanumeric
- Alphabetic
- Upper
- UpperNumeric
- Lower
- LowerNumeric
- Numeric
- UUID
- DNSLabel
- String
- ECDSAKey

## Developing

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster. **Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/database-provisioner:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/database-provisioner:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller from the cluster:

```sh
make undeploy
```

### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
