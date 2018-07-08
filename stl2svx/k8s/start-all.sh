#!/bin/bash -ex
# -*- compile-command: "./start-all.sh"; -*-
kubectl create -f 01-master-controller.yaml
kubectl create -f 02-master-service.yaml
sleep 5
kubectl create -f 03-agent-controller.yaml
sleep 5
MASTER=$(kubectl get pods | grep master | cut -d' ' -f 1)
kubectl port-forward ${MASTER} 45326
