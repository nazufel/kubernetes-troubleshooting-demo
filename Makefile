# Makefile

# use Make to simplify and automate tasks

# --------------------------------------------------------------------------- #

# create a KinD cluster for the demo
cluster:
	$(clean_command)
	kind create cluster

# delete the cluster to finish the demo
down:
	$(clean_command)
	kind delete cluster

# build and load the demo image into the kind cluster
image:
	$(clean_command)
	docker build -f ./Dockerfile -t k8s-demo:v0 .
	kind load docker-image k8s-demo:v0
#EOF