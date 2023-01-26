podman build . --tag docker.io/kubernetesbigdataeg/superset:2.0.1-1
podman login docker.io -u kubernetesbigdataeg
podman push docker.io/kubernetesbigdataeg/superset:2.0.1-1
