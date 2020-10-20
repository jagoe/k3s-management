# Setup

This document explains how to setup a k3s cluster with Longhorn and Cert-Manager in order to start running applications on it.

## Longhorn

The [official guide](https://longhorn.io/docs/1.0.2/deploy/install/install-with-kubectl/) explains how to install Longhorn with `kubectl`.

### StorageClass

Execute the following to add a Longhorn storage class:

```bash
cat <<EOF | kubectl apply -f -
---
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: longhorn
provisioner: driver.longhorn.io
allowVolumeExpansion: true
parameters:
  numberOfReplicas: "1"
  staleReplicaTimeout: "2880"
  fromBackup: ""
#  diskSelector: "ssd,fast"
#  nodeSelector: "storage,fast"
#  recurringJobs: '[{"name":"snap", "task":"snapshot", "cron":"*/1 * * * *", "retain":1},
#                   {"name":"backup", "task":"backup", "cron":"*/2 * * * *", "retain":1,
#                    "labels": {"interval":"2m"}}]'
---
EOF
```

Uncomment the commented-out sections to enable snapshots or backups.

## Cert-Manager

The [official guide](https://cert-manager.io/docs/installation/kubernetes/#installing-with-regular-manifests) explains how to install Cert-Manager with `kubectl`.

### Cluster Issuer

In order to provide a cluster-wide Let`s Encrypt issuer, execute the following:

```bash
cat <<EOF | kubectl apply -f -
---
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    # The ACME server URL
    server: https://acme-v02.api.letsencrypt.org/directory
    # Email address used for ACME registration, update to your own. This will be used for, e.g. expiry warnings.
    email: <email>
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-prod
    # Enable the HTTP-01 challenge provider
    solvers:
      - http01:
          ingress:
            class: traefik
---
EOF
```

## Hello World

You can test the setup with the following:

```bash
cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: Pod           
metadata: 
  name: nginx
  labels:
    app: nginx
spec:               
  containers:
  - name: nginx
    image: nginx:latest
    ports:             
    - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  selector:
    app: nginx
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
--- # HTTPS endpoint
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: ingress
  annotations:
    kubernetes.io/ingress.class: traefik
    cert-manager.io/cluster-issuer: letsencrypt-prod
    cert-manager.io/acme-challenge-type: http01
    traefik.ingress.kubernetes.io/redirect-entry-point: https
spec:
  rules:
    - http:
        paths:
          - backend:
              serviceName: nginx
              servicePort: 80
      host: <host> # replace with a domain you are pointing at the cluster
  tls:
    - hosts:
        - <host> # replace with a domain you are pointing at the cluster
      secretName: pinata-tls
--- # HTTP endpoint
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: ingress-1
  annotations:
    kubernetes.io/ingress.class: traefik
spec:                        
  rules:
    - http:
        paths:
          - backend:
              serviceName: nginx
              servicePort: 80
      host: <host> # replace with a domain you are pointing at the cluster
---
EOF
```
