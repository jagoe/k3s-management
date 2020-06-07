# k8s introduction

This is a succinct introduction to how to use k8s to deploy an application that has a public endpoint.\
I will touch on the following topics:

* Deploying an application
* Making it publicly accessible
* Private communication between applications within the same namespace
* Persisting data
* Deploying a ready-made application with Helm

I will use a classic example application to demonstrate each step: A (very quick and dirty) todo list.

## Requirements

* docker
  * k8s does support other containerization alternatives, but we're using docker in this example
* access to a container registry, such as Docker Hub or GitLab
* kubectl

A [setup guide](#setup) can be found further below.

## Resources

I will shortly introduce each resource we're going to use.

### Pod

A pod is one of the smallest units in k8s and it's usually what's representing an instance of an application.\
A pod can contain several containers, of which one usually fulfills a primary roll.\
Any other containers in a pod are called sidecars and provide functionality to the application.

In our example, we end up with pods for the frontend and the backend. We could add a sidecar to the backend that runs a
database, since that would not be used by any other application.

### Deployments

A deployment handles maintenance of one or more pods for an application. It is responsible for managing replicas of the
application and indirectly for restarting exited pods. If you want to scale your application up or down, you change
the deployment configuraton.\
If you want to restart an application instance, you can therefore just delete the corresponding pod. If you want to
remove the application altogether, you have to delete the deployment.

In our example, we create deployments for both the frontend and the backend. Since we don't expect a heavy load, we
declare that we only need 1 instance.

### Services

A service represents the network-facing side of an application. Services handle discovery, IP and DNS and handle
load-balancing access to the application across all available pods. This way, applications can communicate with each
other via a DNS name corresponding to the service name.

Usually a service points to multiple instances of the same application (i.e. pods of a single deployment), but you can
also do A-B testing or start migrating to a new language or tool by pointing the service to pods of different
deployments. This is done by setting the `selector` to point to a shared label set by the pods.

In our example, we have a service for both the frontend and the backend, exposing the port `80` in both cases. The
frontend sets up a reverse proxy to the backend by pointing to the internal DNS name `backend`. If we had an SQL
database app, we would instead expose port `5432` (or whatever custom port we specified) for the database service.

### Ingresses

An ingress represents a public network interface for your application. By leveraging the simple ingress interface
implemented by popular services such as NGINX or Traefik, it is very easy to configure one or several public domains
for any amount of applications.\
Due to IP address shortage, you should set up one ingess for your namespace and redirect to your separate apps from there.

Setting up TLS for your application is also easily done by adding annotations used by cert-manager to automatically
request an LE certificate for your public endpoint.

In our example, we set up an ingress that uses virtual hosting to point any requests to
`https://demo.k3.infinite-turtles.dev` to the frontend service.

### Persistent Volume Claims

Persistent Volume Claims requests and reserves storage on the host for the application. The claim can then be used
by pods to mount the resulting volume into containers. That way persistent storage is quickly available for your
application.

In our example, we request 50M for our application with RW access for one mount point. The actual minimal reserved
space by our storage system is 1GB, though.

## Setup

### k3s cluster access

In order to get access to your own or an existing namespace, talk to one of the k8s admins:

* Jakob
* Fred

Once you received a `k3-<user>.kubeconfig.yaml` file from an admin, you can use it by adding the path to the
KUBECONFIG global environment variable in your `.bashrc` (or equivalent): `KUBECONFIG="$KUBECONFIG:<path/to/config>"`.

You can then set the current kubectl context by executing `kubectl config set-context k3s-<user>`.\
Now you can work on your namespace.

### Creating & pushing images

Creating and pushing docker images is not within the scope of this document.

### Setting up registry access in your namespace

Setting up registry access for your namespace is a simple one-liner, documented [here](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#create-a-secret-by-providing-credentials-on-the-command-line).

## Hooking it all up

In this section, each step to deploy the application is followed by a simple explanation of the corresponding k8s
manifest together with links to the official documentation.

### Using other images

The images used for this demo have been pushed to a public docker registry. If you want to build and use your own
images, they have to be accessible publicly or [using a secret in k8s](#setting-up-registry-access-in-your-namespace).

### Storage

Deploy the storage manifest: `kubectl apply -f ./manifests/storage.yaml`

#### Storage manifest

(Official documentation [here](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims))

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: demo-todos # This value will be used to reference the PVC
  namespace: <your namespace>
spec:
  accessModes:
    - ReadWriteOnce # `ReadWriteOnce`, `ReadOnlyMany`, `ReadWriteMany` (we only have one node, so ReadWriteMany is not necessary)
  storageClassName: longhorn # `longhorn` or `local-path` (defaults to `longhorn`, which is recommended)
  resources:
    requests:
      storage: 50M # Specifies how much storage we request
```

### Backend

The frontend sets up a reverse proxy pointing to the backend, so the backend has to be deployed first.

`kubectl apply -f ./manifests/backend.yaml` - this will deploy both deployment and service, which will be explained
in the following sections.

#### Backend deployment

(Official documentation [here](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims))

##### Backend deployment manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend # This name will be used to reference the deployment, but not necessarily the pods or service (though we use the same name)
  namespace: <your namespace> # Your namespace - you can leave this out if you set the `-n` flag
  labels:
    app: backend # Not necessary, but good form
spec:
  selector:
    matchLabels:
      app: backend # This tells the deployment, which pods it governs
  replicas: 1 # Amount of pods running concurrently, can be scaled up or down during runtime
  template:
    metadata:
      namespace: <your namespace>
      labels:
        app: backend # This has to be same value as the `matchLabel` selector above
    spec:
      containers:
        - name: backend
          image: jagoe/k3s-demo-backend # Which image to run
          ports:
            - containerPort: 80 # Ports exposed to the internal network, can be accessed if you know the pod IP
          resources:
            limits: # It is recommended you set resource limits, since otherwise a single pod might crash the whole node
              memory: "128Mi"
              cpu: "500m"
          volumeMounts:
            - name: storage # name of the volume (same as the volume name below)
              mountPath: /storage # container path the volume will be mounted into
      imagePullSecrets:
        - name: regcred # Image secret for private container registry (not necessary in this case, but this is how you reference it)
      volumes:
        - name: storage # name of the storage, to be referenced by the pod
          persistentVolumeClaim:
            claimName: demo-todos # name of the claim, same as the previously deployed one
```

#### Backend service

(Official documentation [here](https://kubernetes.io/docs/concepts/services-networking/service/))

#### Backend service manifest

```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend # This name will be used to reference the service & is the DNS name for the service, i.e. all the pods behind it
  namespace: <your namespace>
  labels:
    app: backend # Not strictly necessary, but good form
spec:
  ports:
    - port: 80 # Port exposed to the internal network, should match the exposed container port
      protocol: TCP
  selector:
    app: backend # Labels by which to select the pods traffic should be routed to
```

### Frontend

Now that the backend has been deployed and is accessible within the cluster, the frontend can be deployed as well.

`kubectl apply -f ./manifests/frontend.yaml` - this will deploy both deployment and service, which will be explained
in the following sections.

#### Frontend deployment

(Official documentation [here](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/))

#### Frontend deployment manifest

```yaml
piVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: <your namespace>
  labels:
    app: frontend
spec:
  selector:
    matchLabels:
      app: frontend
  replicas: 1
  template:
    metadata:
      namespace: <your namespace>
      labels:
        app: frontend
    spec:
      containers:
        - name: frontend
          image: jagoe/k3s-demo-frontend
          ports:
            - containerPort: 80
          resources:
            limits:
              memory: "128Mi"
              cpu: "500m"
      imagePullSecrets:
        - name: regcred
```

The frontend deployment is much like the one for the backend, just with different names, identifying labels and image.

#### Frontend service

(Official documentation [here](https://kubernetes.io/docs/concepts/services-networking/service/))

#### Frontend service manifest

```yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend
  namespace: <your namespace>
  labels:
    app: frontend
spec:
  ports:
    - port: 80
      protocol: TCP
  selector:
    app: frontend
```

The frontend service is much like the one for the backend, just with different names and identifying labels.

### Ingress

Now all that's missing is exposing our applicaton to the public:
`kubectl apply -f ./manifests/ingress.yaml`

Once that's done, all you need to do is point for domain to the public IP of the ingress, which you can find like this:
`kubectl -n <namespace> get ingress demo`. And that's it.

#### Ingress manifest

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: demo # Name you want to reference the ingress by
  namespace: <your namespace>
  annotations:
    kubernetes.io/ingress.class: traefik # Used by cert-manager to match ingresses that get served with LE certs
    cert-manager.io/cluster-issuer: letsencrypt-prod # We only have this one cluster issuer, so no other value is possible atm
    cert-manager.io/acme-challenge-type: http01 # Our cluster issuer only supports http01 challenges, so no other value is possible atm
    traefik.ingress.kubernetes.io/redirect-entry-point: https # Automatic redirect to https
spec:
  rules:
    - host: demo.k3.infinite-turtles.dev # Your host
      http:
        paths:
          - backend:
              serviceName: frontend # Which service will handle requests for the given host
              servicePort: 80 # On which port will the service handle requests for the given host
  tls:
    - hosts:
        - demo.k3.infinite-turtles.dev # The list of hosts which get TLS
      secretName: k3-demo-tls # Name of the secret used to store the cert (no rules, but maybe use something you will recognize)
```

## Making changes

If you make changes to either the application or the configuration, a simple redeployment is all that's needed.\
If nothing changed, k8s will not do anything and if there are changes, a rolling upgrade will be peformed.
