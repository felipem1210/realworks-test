# ConfigDeployment controller
A controller that will trigger a rolling restart of a deployment, when the deployment is using a specific configmap.

Powered by: [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

## Description
By default in Kubernetes, when a configmap changes, a rolling restart is not triggered in a Deployment using the configmap. That can be necessary, more than all if the configmap is storing environment variables. That's the purpose of this controller.

Will be necesary to add an annotation in the deployment that you want to track with this controller:

```yaml
annotations:
  configMapUsed: my-configmap
```

The controller is basically doing the following:

1. Detecting any change in CRDs `ConfigDeployment`.
2. Getting the configMap name that is defined in the CRD
3. Checking for Deployments with the annotation `configMapUsed` defined with the name of the configMap defined in the CRD. The configMap, ConfigDeployment and Deployment must be all in the same namespace.
4. Get the version of the `configMap`.
5. Check if the `configMapVersion` exists as an annotation called `configMapVersion` in the pod managed by the Deployment.
6. If the `configmapVersion` is different than the one of the configMap, update the annotation, that will trigger a rolling restart of the Deployment.

[Controller code](internal/controller/configdeployment_controller.go)
[CRD code](api/v1/configdeployment_types.go)

## Getting Started
You’ll need a Kubernetes cluster to run against. 

**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

Check the README of realworks-test project


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

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

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

