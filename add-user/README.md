# Create a k8s user with full access to their own namespace

Here you can find a manual and some command-line tools to create a user with their own namespace in k8s.

The proper method would be to [use a client certificate](#using-a-client-certificate), but since that doesn't work on my setup I've also
included tooling to leverage [service accounts]((#using-a-service-account)) for the same purpose.

* [Create a k8s user with full access to their own namespace](#create-a-k8s-user-with-full-access-to-their-own-namespace)
  * [Using a client certificate](#using-a-client-certificate)
    * [User](#user)
    * [Admin](#admin)
  * [Using a service account](#using-a-service-account)
    * [User Actions](#user-actions)
    * [Admin Actions](#admin-actions)

## Using a client certificate

This method doesn't work on my machine (k3os) for unknown reasons, but it should work in principle and is cleaner than the alternative.

### User

__Requirements:__

* bash 4+
* openssl

__Method:__

1. `K8S_USER=<user> bin/create_csr`\
2. Keep the .key file and send the .csr file to your k8s admin
3. Wait for the kubeconfig from your admin
5. `K8S_USER=<user> bin/set_credentials <path to your kubeconfig> <path to your .key file>`
6. Use the kubeconfig by setting/adding to `KUBECONFIG`

* `user`: The namespace name you want to have in k8s

### Admin

__Requirements:__

* bash 4+
* kubectl
* jq

__Method:__

1. Receive the .csr file from your user
2. `K3S_CONTEXT=<context> ./bin/add_user <path to the .csr file>`
3. Give the generated kubeconfig to your user

* `context`: The context name you use to administrate the k8s

## Using a service account

This method is easier, quicker and it also works on my setup (again, k3os). But it also feels a bit dirty.

### User Actions

1. Tell your k8s admin the namespace name you want to use in k8s
2. Wait for the kubeconfig from your admin
3. Use the kubeconfig by setting/adding to `KUBECONFIG`

### Admin Actions

#### Requirements

* bash 4+
* kubectl
* jq

#### Method

1. `K3S_CONTEXT=<context> K8S_USER=<user> ./bin/add_service_account`
2. Give the generated kubeconfig to your user

* `context`: The context name you use to administrate the k8s
* `user`: The namespace name your user wants
