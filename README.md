# realworks-test
Test for realworks

## Local machine (MacOS)

### Setup minikube

1. Install minikube `brew install minikube`
2. Create k8s cluster `minikube start --driver=docker --kubernetes-version=v1.27.3`
3. Build image of the app: `docker build -t configmap-app:latest .`
4. Load the image inside minikube cluster `minikube image load configmap-app:latest`

### Install Istio

1. Download latest release of istioctl: `curl -L https://istio.io/downloadIstio | sh -`
2. Install istio: `istio-1.18.2/bin/istioctl install --set profile=default -y`. The version may change
3. Init istio-operator in cluster: `istio-1.18.2/bin/istioctl operator init`

### Build controller

1. Build the image: `docker build -f controller/Dockerfile -t config-deployment:latest controller/`
2. Load image in minikube cluster: `minikube image load config-deployment:latest`

For more info about how the controller works check its [README file](controller/README.md)

### Deploy manifests

1. Run to deploy controller, istio Ingress gateway and application:

```sh
kubectl apply -k infra/configdeploy-controller/default
kubectl apply -k infra/istio
kubectl apply -k infra/application/overlays/test
kubectl apply -k infra/application/overlays/prod
```

### Testing application

1. Use the command `minikube tunnel` to create a virtual load balancer to access the service.
2. Run this commands to get the IP address assigned for your virtual load balancer:

```sh
export INGRESS_NAME=istio-ingressgateway
export INGRESS_NS=istio-system
export INGRESS_HOST=$(kubectl -n "$INGRESS_NS" get service "$INGRESS_NAME" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

3. Access the service using curl:

* For the test environment: 

```sh
curl  -H Host:test.example.com "http://$INGRESS_HOST"
```

* For the prod environment:

```sh
curl  -H Host:prod.example.com "http://$INGRESS_HOST"
```

### Testing controller working

1. Go to the file `infra/application/overlays/prod/configmap.yaml` or `infra/application/overlays/test/configmap.yaml` and edit the data of the configmap.
2. Apply the changes with `kubectl apply -k`
3. A new pod will be created automatically.
4. Perform the test of application again, and you will see the new message displayed.

### Development considerations

If you make some changes to your code after uploading the images to minikube is necessary to unload the image with tag latest and add it again, after building the new version of the image. Otherwise, you can do your own local tagging strategy and edit the tag in `kustomization.yaml` in folders `infra/application/base` for app and `infra/configdeploy-controller/manager` for the controller.

## Cloud considerations

There are several things to consider in order to deploy this app in a cloud environment:

1. There must be a Load Balancer created in the cloud to route traffic to the Kubernetes cluster. This is the entrypoint to the cluster and replaces `minikube tunnel`. There are several ways to do it.
2. Push docker images to a docker image registry with proper image tagging strategy.
3. Automate the use of image in kubernetes manifests. Kustomize provide a way to do it with `kustomize edit` command
4. Change imagePullPolicy to Always to be able to download the image from the registry.
5. Create the secret with docker credentials and use them in deployment in `imagePullSecrets` section.