# Mimir Operator
The Mimir Operator is a Kubernetes operator to control Mimir tenants using CRDs.

## Description
Currently, the operator is capable of:
- Connecting to remote Mimir instances (with optional authentication)
- Loading alerting rules for a specific Mimir tenant depending on labels
- Overriding rule parameters per tenant
- Adding external labels to the generated alerts

Read the documentation [here](docs/index.md).

## Contributing

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

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

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/mimir-operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/mimir-operator:tag
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

## Releasing a new version

- Update the documentation in ```docs/``` (check the instructions for installation and set the upcoming release as the latest release)
- Run ```make generate``` and ```make manifests``` to refresh the CRDs and the deployment files in ```config/``` (the CRDs are copied to the Helm chart)
- Bump the chart version in ```deploy/helm/mimir-operator/Chart.yml``` with the version of the upcoming release
- Run ```make helm/docs``` to regenerate the Helm README with any new documentation of the values
- Check if ```config/rbac/role.yaml``` has changed. If it did, edit the RBAC config in the Helm Chart (```deploy/helm/mimir-operator/templates/rbac.yaml```) to reflect the changes
- Change the version of the project in ```Makefile``` to the upcoming release
- Run ```git checkout -b [RELEASE]```
- Push the new branch to the Git
- Merge and create a Release on Github

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

