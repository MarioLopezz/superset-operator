podman build . --tag docker.io/kubernetesbigdataeg/superset:4.0.0-1
podman login docker.io -u kubernetesbigdataeg
podman push docker.io/kubernetesbigdataeg/superset:4.0.0-1
