#!/usr/bin/env bash
set -euo pipefail
IFS='\n\t'

# We're going to use the same cluster we created in example 1 that has a local container registry
# A couple of variables to synchronize our commands without needing to copy/paste
PODMAN_MACHINE_VM_NAME="vm1"
REGISTRY_NAME='kind-registry'
REGISTRY_PORT='5001'
REGISTRY_IMAGE='docker.io/library/registry:2'

# If the specified podman machine VM does not exist, we'll need to initialize it before we can start it
if [[ "$(podman machine list --format '{{.Name}}' 2>/dev/null || false)" != ${PODMAN_MACHINE_VM_NAME} ]]; then
  echo "Creating a new podman machine ${PODMAN_MACHINE_VM_NAME}"
  podman machine init ${PODMAN_MACHINE_VM_NAME} --cpus 2 --memory 2048 --rootful --now # some additional settings are required to make enough compute room for Ingress
fi

# If it is not already running, start Podman Machine with a new VM
if [[ "$(podman machine inspect --format '{{.State}}' ${PODMAN_MACHINE_VM_NAME} 2>/dev/null || false)" != 'running' ]]; then
  echo "Starting podman machine ${PODMAN_MACHINE_VM_NAME}"
  podman machine start ${PODMAN_MACHINE_VM_NAME}
fi

# first we'll create our registry, using the specified name and port if it's not already there
if [[ "$(podman inspect -f '{{.State.Running}}' registry 2>/dev/null || echo 'true')" == 'true' ]]; then
  echo "Starting a local container registry on podman machine ${PODMAN_MACHINE_VM_NAME} as ${REGISTRY_NAME}"
  podman run -d --restart=always -p "127.0.0.1:${REGISTRY_PORT}:5000" --name "${REGISTRY_NAME}" "${REGISTRY_IMAGE}"
fi

# If the kind cluster is already running, we're going to delete and recreate it.
if [[ $(kind get clusters 2> /dev/null | grep "^kind$" | sort) == 'kind' ]]; then
  kind delete cluster
fi

# now, lets start a Kind cluster; note that we're passing config to tell it the local registry plugin is enabled in containerd
# also, unlike in our first demo we have some additional worker nodes and a config patch for kubeadm to allow the ingress to
# run on our kind nodes.
cat << EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${REGISTRY_PORT}"]
    endpoint = ["http://${REGISTRY_NAME}:5000"]
kubeadmConfigPatches:
- |
  apiVersion: kubelet.config.k8s.io/v1beta1
  kind: KubeletConfiguration
  evictionHard:
    nodefs.available: "0%"
# patch it further using a JSON 6902 patch
kubeadmConfigPatchesJSON6902:
- group: kubeadm.k8s.io
  version: v1beta3
  kind: ClusterConfiguration
  patch: |
    - op: add
      path: /apiServer/certSANs/-
      value: my-hostname
nodes:
# one control plane node
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 31017
    hostPort: 31017
    listenAddress: "0.0.0.0" # Optional, defaults to "0.0.0.0"
    protocol: tcp # Optional, defaults to tcp
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF

# now we need to connect the virtual networks between the running Kind cluster, which is really just containers, and the Registry container, which is also... just a container
if [[ "$(podman inspect --format '{{json .NetworkSettings.Networks.kind}}' ${REGISTRY_NAME} 2>/dev/null || echo 'null')" == 'null' ]]; then
  podman network connect "kind" "${REGISTRY_NAME}"
fi

# Document the local registry in the kube-public namespace
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${REGISTRY_PORT}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

# finally, we're going to install the ingress-nginx Ingress controller - this is the same one Datasite uses in its production clusters
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
