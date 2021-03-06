# kubernetes-troubleshooting-demo

This repository holds several different troubleshooting scenarios for Kubernetes.

## Limitations

This demo will be using KinD (see below) as the Kubernets runtime and control plane. Kubernetes running anywhere should suffice for these scenarios. The rest of this demo will assume KinD is being used. If not, then you will have to make the appropiate translations, which should be few.

However, I have not at the time of this writing, gotten any sort of Ingress or LoadBalancer so that traffic outside of the cluster can get in to use the demo app. I tried using [metallb](https://kind.sigs.k8s.io/docs/user/loadbalancer/) since I have use it successfully before with Kind. I have been at a lost as to why it doesn't work. The only thing I can think of is that I have developed this demo series within [WSL2](https://docs.microsoft.com/en-us/windows/wsl/install). Maybe there's something squirrely going on with that. Perhaps with more time I could work it out. The last time I used metallb, I was using [Fedora](https://getfedora.org/) Linux. I spent too much time trying to get this to work and it's not critical to the demo. I have a way around this as you'll see in the scenarios. Hitting the app's ReST endpoints shouldn't be a blocker for troubleshooting these scenarios. The focus is Kubernetes, not a [Go](https://go.dev/) server.
## Prerequisites

Here is a list of required prerequisites:

* [Docker](https://docs.docker.com/engine/install/)
* [KinD](https://kind.sigs.k8s.io/)
* [Kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
* [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/)

Here is a list of reccomended prerequisites:

* [GNU Make](https://www.gnu.org/software/make/)
* [Kubectx and Kubens](https://github.com/ahmetb/kubectx#installation)

## Set Up the Cluster

The KinD cluster must be set up first. There is a Make command for this.

```sh
make cluster
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.21.1) 🖼 
 ✓ Preparing nodes 📦  
 ✓ Writing configuration 📜 
 ✓ Starting control-plane 🕹️ 
 ✓ Installing CNI 🔌 
 ✓ Installing StorageClass 💾 
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Have a question, bug, or feature request? Let us know! https://kind.sigs.k8s.io/#community 🙂
```

The next thing is to build the demo app and load it into KinD (if not using KinD, then make sure your cluster can pull the Docker Image). There is a Make command for this as well

```sh
make image
docker build -f ./Dockerfile -t k8s-demo:v0 .
# ...
 => => naming to docker.io/library/k8s-demo:v0       
kind load docker-image k8s-demo:v0
Image: "k8s-demo:v0" with ID "sha256:cfabef7128cd2aa736f9764fdbc1b55aec8db55f640bb41b2772cb4f248be266" not yet present on node "kind-control-plane", loading...                                  
```

Cluster set up is complete. Ensure you are on using the `kind` context.

```sh
kubectl config current-context
kind-kind
```

Verify you can access the KinD control plane.
```sh
kubectl get all
NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   33m
```


Now it is time to go through the scenarios.

## Scenario 0

This scenario is a completely working scenario. The point is to demonstrate what the end state should look like. The remaining scenarios should not work out of the box.

Deploy the first scenario. 

```sh
kubectl apply -k kustomize/overlays/scenario-0
namespace/scenario-0 created
configmap/common created
service/demo created
deployment.apps/demo created
poddisruptionbudget.policy/demo created
horizontalpodautoscaler.autoscaling/demo created
```

Switch to the `scenario-0` namespace an check to see what's there.
```sh
kubens scenario-0
Context "kind-kind" modified.

kubectl get all
NAME                        READY   STATUS    RESTARTS   AGE
pod/demo-85b5c5684c-7ftdc   1/1     Running   0          66s
pod/demo-85b5c5684c-cwt2n   1/1     Running   0          66s
pod/demo-85b5c5684c-hftlx   1/1     Running   0          81s

NAME           TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/demo   ClusterIP   10.96.89.95   <none>        8080/TCP   81s

NAME                   READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/demo   3/3     3            3           81s

NAME                              DESIRED   CURRENT   READY   AGE
replicaset.apps/demo-85b5c5684c   3         3         3       81s

NAME                                       REFERENCE         TARGETS                        MINPODS   MAXPODS   REPLICAS   AGE
horizontalpodautoscaler.autoscaling/demo   Deployment/demo   <unknown>/80%, <unknown>/80%   3         4         3          81s
```

The stack is running. There should be:

* 3 Pods
* 1 Service of type `ClusterIP`
* 1 Deployment
* 1 ReplicaSet
* 1 HorizontalPodAutoScaler

Make a query to the app to see what's inside (this is the part where an ingress would be helpful. However, running a `curl` command from inside a pod will suffice.)

```sh
kubectl exec -ti deployment/demo -- curl demo:8080
{"message":"hello K8s toubleshooting demo","year":"2022"}
```

We can use `kubectl` to run a command from any pod within the `deployment/demo` resource. The command is to use `curl` (already baked into the image) and send a reqeust to the `service/demo`. The response should be `json` with a `message` and `year` fields. 

Great! The app works. Now, let's move onto some failure scenarios.

## Scenario 1

Deploy the scenario.

```sh
kubectl apply -k kustomize/overlays/scenario-1
namespace/scenario-1 created
configmap/common created
service/demo created
deployment.apps/demo created
poddisruptionbudget.policy/demo created
horizontalpodautoscaler.autoscaling/demo created
```

Switch to the `scenario-1` namespace to see the resources.

```sh
kubens scenario-1
Context "kind-kind" modified.
Active namespace is "scenario-1".
```

Now inspect the resource.

```sh
kubectl get all
NAME                        READY   STATUS                       RESTARTS   AGE
pod/demo-8658f75c46-5lnhd   0/1     CreateContainerConfigError   0          57s
pod/demo-8658f75c46-s4b2g   0/1     CreateContainerConfigError   0          72s
pod/demo-8658f75c46-wh68m   0/1     CreateContainerConfigError   0          57s

NAME           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/demo   ClusterIP   10.96.218.214   <none>        8080/TCP   72s

NAME                   READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/demo   0/3     3            0           72s

NAME                              DESIRED   CURRENT   READY   AGE
replicaset.apps/demo-8658f75c46   3         3         0       72s

NAME                                       REFERENCE         TARGETS                        MINPODS   MAXPODS   REPLICAS   AGE
horizontalpodautoscaler.autoscaling/demo   Deployment/demo   <unknown>/80%, <unknown>/80%   3         4         3          72s
```

The pods are showing `CreateContainerConfigError`. Let's investigate what's going on. 

We'll use the `describe` subcommand on a pod.
```sh
kubectl describe pod demo-8658f75c46-5lnhd
Name:         demo-8658f75c46-5lnhd
Namespace:    scenario-1

# ---

Events:
  Type     Reason     Age                   From               Message
  ----     ------     ----                  ----               -------
  Normal   Scheduled  2m43s                 default-scheduler  Successfully assigned scenario-1/demo-8658f75c46-5lnhd to kind-control-plane
  Warning  Failed     30s (x12 over 2m43s)  kubelet            Error: configmap "commonn" not found
  Normal   Pulled     19s (x13 over 2m43s)  kubelet            Container image "k8s-demo:v0" already present on machine
```

There's a lot of output that has been minified for this demo. The relevant information has been preserved. The `Events` list shows what the problem is. 