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

There's a lot of output that has been minified for this demo. The relevant information has been preserved. The `Events` list shows what the problem is. The second event says, `Warning  Failed     30s (x12 over 2m43s)  kubelet            Error: configmap "commonn" not found`. The error shows that the deployment is looking for a configmap that doesn't exist. The `kubectl get all` command doesn't show ConfigMaps. Let's take a look at them.

```sh
kubectl get cm
NAME               DATA   AGE
common             1      2m13s
kube-root-ca.crt   1      2m13s
```

We can see there are two ConfigMaps. The first one is the one we care about, `common`. Looking at the error, the Deployment is looking for a ConfigMap named `commonn`. Looks like there's a typo. Let's update the [overlay](kustomize/overlays/scenario-1/envFrom.yaml) to use the existing ConfigMap and redeploy.

```sh
kubectl apply -k kustomize/overlays/scenario-1
namespace/scenario-1 unchanged
configmap/common unchanged
service/demo unchanged
deployment.apps/demo configured
poddisruptionbudget.policy/demo configured
horizontalpodautoscaler.autoscaling/demo configured


kubectl get pods
NAME                   READY   STATUS    RESTARTS   AGE
demo-c79f576bd-4rfm4   1/1     Running   0          47s
demo-c79f576bd-f6z4b   1/1     Running   0          48s
demo-c79f576bd-rcbmk   1/1     Running   0          46s
```

Now the Pods are running since they have the correct ConfigMap. Let's check the app is working inside.

```sh
kubectl exec -ti deployment/demo -- curl demo:8080
{"message":"hello K8s toubleshooting demo","year":"2022"}
```

All is working. Finished wih this scenario.

## Scenario 2

Deploy the scenario.

```sh
kubectl apply -k kustomize/overlays/scenario-2
namespace/scenario-2 created
configmap/common created
configmap/common2-dtg2cgmhb6 created
service/demo created
deployment.apps/demo created
poddisruptionbudget.policy/demo created
horizontalpodautoscaler.autoscaling/demo created
```

Switch to the `scenario-2` namespace and check on the cluster.

```sh
kn scenario-1

Context "kind-kind" modified.
Active namespace is "scenario-2".

kubectl get all
NAME                        READY   STATUS             RESTARTS   AGE
pod/demo-6bb6f576db-4gqq2   0/1     CrashLoopBackOff   3          49s
pod/demo-6bb6f576db-6ztfs   0/1     CrashLoopBackOff   4          64s
pod/demo-6bb6f576db-rrwxt   0/1     CrashLoopBackOff   3          49s

NAME           TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/demo   ClusterIP   10.96.92.61   <none>        8080/TCP   65s

NAME                   READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/demo   0/3     3            0           64s

NAME                              DESIRED   CURRENT   READY   AGE
replicaset.apps/demo-6bb6f576db   3         3         0       64s

NAME                                       REFERENCE         TARGETS                        MINPODS   MAXPODS   REPLICAS   AGE
horizontalpodautoscaler.autoscaling/demo   Deployment/demo   <unknown>/80%, <unknown>/80%   3         4         3          64s
```

The pods in this scenario are in a `CrashLoopBackoff`. Let's see what's going on.

```sh
kubectl describe pod/demo-6bb6f576db-4gqq2
Name:         demo-6bb6f576db-4gqq2
Namespace:    scenario-2
# ...
Events:
  Type     Reason     Age                   From               Message
  ----     ------     ----                  ----               -------
  Normal   Scheduled  2m22s                 default-scheduler  Successfully assigned scenario-2/demo-6bb6f576db-4gqq2 to kind-control-plane
  Normal   Pulled     2m1s (x4 over 2m22s)  kubelet            Container image "k8s-demo:v0" already present on machine
  Normal   Created    2m1s (x4 over 2m22s)  kubelet            Created container demo
  Normal   Started    2m1s (x4 over 2m22s)  kubelet            Started container demo
  Warning  Unhealthy  2m (x4 over 2m20s)    kubelet            Liveness probe failed: HTTP probe failed with statuscode: 500
  Normal   Killing    2m (x4 over 2m20s)    kubelet            Container demo failed liveness probe, will be restarted
  Warning  BackOff    119s (x5 over 2m16s)  kubelet            Back-off restarting failed container
```

Looks like the LivenessProbe is failing. Maybe the logs will say why.

```sh
kubectl logs -f deploy/demo

Found 3 pods, using pod/demo-6bb6f576db-6ztfs
2022/08/17 15:59:04 DEMO_YEAR env var did not match expected: 2022, got: 2021
```

The logs say the `DEMO_YEAR` environment variable did not match what was expected. The application is in a bad state and failed the LivenessProbe. Since we know that environment variable is passed via a ConfigMap, let's check the configmap. 

```sh
cat kustomize/overlays/scenario-2/envFrom.yaml
- op: "replace"
  path: "/spec/template/spec/containers/0/envFrom/0/configMapRef/name"
  value: "common2"
```

That's interesting. The overlay has the deployment using `common2` ConfigMap. Let's take a look at its contents.

```sh
kubectl describe cm common2
Name:         common2-dtg2cgmhb6
Namespace:    scenario-2
Labels:       app=demo
              env=scenario-2
Annotations:  <none>

Data
====
DEMO_YEAR:
----
2021

BinaryData
====

Events:  <none>
```

This ConfigMap does that the expected environment variable, but it's wrong. `common2` is a stange name for a ConfigMap and there should be just a `common`. 

```sh
kubectl get cm
NAME                 DATA   AGE
common               1      7m43s
common2-dtg2cgmhb6   1      7m43s
kube-root-ca.crt     1      7m43s

kubectl describe cm common
Name:         common
Namespace:    scenario-2
Labels:       app=demo
              env=scenario-2
Annotations:  <none>

Data
====
DEMO_YEAR:
----
2022

BinaryData
====

Events:  <none>
```
There is a `common` ConfigMap and it has the correct value. Let's update the deployment to use this one and redeploy.

```sh
kubectl apply -k kustomize/overlays/scenario-2
namespace/scenario-2 unchanged
configmap/common unchanged
configmap/common2-dtg2cgmhb6 unchanged
service/demo unchanged
deployment.apps/demo configured
poddisruptionbudget.policy/demo configured
horizontalpodautoscaler.autoscaling/demo configured

kubectl get pods
NAME                   READY   STATUS    RESTARTS   AGE
demo-fff6d757c-bvwt7   1/1     Running   0          8s
demo-fff6d757c-wfcq2   1/1     Running   0          6s
demo-fff6d757c-xqnvt   1/1     Running   0          7s
```

The Pods have the correct ConfigMap with the correct value. Let's check the application.

```sh
kubectl exec -ti deployment/demo -- curl demo:8080
{"message":"hello K8s toubleshooting demo","year":"2022"}

kubectl logs -f deploy/demo
Found 3 pods, using pod/demo-fff6d757c-bvwt7
2022/08/17 16:04:31 hit healthckeck endpoint
2022/08/17 16:04:32 hit healthckeck endpoint
2022/08/17 16:04:33 hit healthckeck endpoint
2022/08/17 16:04:34 hit healthckeck endpoint
2022/08/17 16:04:35 hit healthckeck endpoint
2022/08/17 16:04:36 hit healthckeck endpoint
2022/08/17 16:04:37 hit healthckeck endpoint
2022/08/17 16:04:38 hit healthckeck endpoint
2022/08/17 16:04:39 hit healthckeck endpoint
2022/08/17 16:04:40 hit healthckeck endpoint
2022/08/17 16:04:41 hit healthckeck endpoint

Ctrl+C
```

The application works as well as it is logging each health check endpoint hit. The LivenessProbe is working as well.

This concludes the scenario.

## Scenario 3

Deploy the scenario.

```sh
kubectl apply -k kustomize/overlays/scenario-3
namespace/scenario-3 created
configmap/common created
service/demo created
deployment.apps/demo created
poddisruptionbudget.policy/demo created
horizontalpodautoscaler.autoscaling/demo created
```

Change to the `scenario-3` namespace and get all resources.

```sh
kn scenario-3
Context "kind-kind" modified.
Active namespace is "scenario-3".

kubectl get all
NAME                        READY   STATUS    RESTARTS   AGE
pod/demo-76956b6dcd-5vxj9   1/1     Running   0          34s
pod/demo-76956b6dcd-ktcbx   1/1     Running   0          34s
pod/demo-76956b6dcd-s7dtd   1/1     Running   0          49s

NAME           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/demo   ClusterIP   10.96.225.254   <none>        8080/TCP   49s

NAME                   READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/demo   3/3     3            3           49s

NAME                              DESIRED   CURRENT   READY   AGE
replicaset.apps/demo-76956b6dcd   3         3         3       49s

NAME                                       REFERENCE         TARGETS                        MINPODS   MAXPODS   REPLICAS   AGE
horizontalpodautoscaler.autoscaling/demo   Deployment/demo   <unknown>/80%, <unknown>/80%   3         4         3          49s

```

Everything looks good. Let's check the application.

```sh
kubectl exec -ti deployment/demo -- curl demo:8080
curl: (7) Failed to connect to demo port 8080 after 1 ms: Connection refused
command terminated with exit code 7
```

Looks like there's an issue with networking. The above output shows there's a service named `demo`. Let's look there.

```sh
kubectl describe service demo
Name:              demo
Namespace:         scenario-3
Labels:            app=demo
                   env=scenario-3
                   name=demo
Annotations:       <none>
Selector:          app=demo,env=scenario-3,name=demo
Type:              ClusterIP
IP Family Policy:  SingleStack
IP Families:       IPv4
IP:                10.96.225.254
IPs:               10.96.225.254
Port:              <unset>  8080/TCP
TargetPort:        9998/TCP
Endpoints:         10.244.0.17:9998,10.244.0.18:9998,10.244.0.19:9998
Session Affinity:  None
Events:            <none>
```

The service has the correct `Selector`s as well as showing three endpoints in the `Endpoints` list. It's listening on the proper port of `8080`. However, that `Endpoints` list shows the traffic is being sent to the pods on port `9998`. Let's check the deployment to see which port is exposed for the Pods and to see if this is correct.

```sh
 kubectl describe deploy/demo
Name:                   demo
# ...
Pod Template:
  Labels:  app=demo
           env=scenario-3
           name=demo
  Containers:
   demo:
    Image:      k8s-demo:v0
    Port:       9999/TCP
    Host Port:  0/TCP
#...
```

The Deployment output shows us what we need. The Pods are listening on port `9999`, while the service is forwarding traffic on port `9998`. Let's update the Service overlay to use the proper port and redeploy.

```sh
kubectl apply -k kustomize/overlays/scenario-3
namespace/scenario-3 unchanged
configmap/common unchanged
service/demo configured
deployment.apps/demo configured
poddisruptionbudget.policy/demo configured
horizontalpodautoscaler.autoscaling/demo configured

kubectl exec -ti deployment/demo -- curl demo:8080
{"message":"hello K8s toubleshooting demo","year":"2022"}
```

Success. How the LoadBalancer Service is forwarding to the same port that the Pods are listening on.

This concludes the scenario.

## Scenario 4

Deploy the scenario.

```sh

``