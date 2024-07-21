#!/bin/bash

# CONTAINERD RUNTIME CLUSTER
minikube start --nodes=2 --container-runtime=containerd --driver=docker --cpus=2 --memory=1900MB
sleep 5
kubectl apply -f manifest/pvc.yml
kubectl apply -f manifest/secret.yml
sleep 15
bats/bin/bats . | tee -a /tmp/dumpy-test/containerd_$(date +"%Y_%m_%d_%H-%M").log
minikube delete
sleep 5

# DOCKER RUNTIME CLUSTER
minikube start --nodes=2 --container-runtime=docker --driver=docker --cpus=2 --memory=1900MB
sleep 5
kubectl apply -f manifest/pvc.yml
kubectl apply -f manifest/secret.yml
sleep 15
bats/bin/bats . | tee -a /tmp/dumpy-test/docker_$(date +"%Y_%m_%d_%H-%M").log
minikube delete
sleep 5

# CRIO RUNTIME CLUSTER
minikube start --nodes=2 --container-runtime=crio --driver=docker --cpus=2 --memory=1900MB
sleep 5
nodes=$(docker ps | awk '{print $1}' | tac | head -n 2)
cat << EOF > /tmp/reg.conf
[registries.search]
registries = ['docker.io']
EOF
for node in $(echo "$nodes")
do
    docker cp /tmp/reg.conf $node:/etc/containers/registries.conf
    docker exec -it $node bash -c "sudo systemctl restart crio"
done
sleep 15
kubectl apply -f manifest/pvc.yml
kubectl apply -f manifest/secret.yml
sleep 15
bats/bin/bats . | tee -a /tmp/dumpy-test/crio_$(date +"%Y_%m_%d_%H-%M").log
minikube delete
sleep 5