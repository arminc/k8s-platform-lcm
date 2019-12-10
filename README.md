# Kubernetes platform lifecycle management

This project helps you keep track of all your software and tools used and running in and around your Kubernetes platform. It helps you with part of the lifecycle management to keep your software up to data for feature completeness, security or compliance reasons. 

## Features

- [x] Keep track of versions of all the running containers (inclusive init containers) inside the Kubernetes
- [x] Keep track of new versions. Supporting Quay, Gcr, Docker hub, Jfrog Artifactory and probably any other Docker registry v2
- [x] Present the information command line
- [x] Allow overriding of the registry to search in, for example, if you are using private registry but need to fetch versions from the internet
- [x] Works with private registries and private images
- [ ] Keep track of image vulnerabilities using Jfrog Xray
- [ ] Present the vulnerabilities command line
- [ ] Possibility to whitelist vulnerabilities so only changes are presented
- [ ] Possibility to provide local tool versions (like terraform version and it's plugins) and find the new versions using GitHub
- [ ] Keep track of Helm chart deployments and track new versions of the charts
- [ ] Provide information for Kubernetes version (for example AWS EKS)

### Nice to have

* Scan image with Xray instead of only getting the report
* Specify `ALL` as an option for namespaces, it should find all namespaces and look at all of them
* Use Clair as a vulnerabilitie scanning option

### Issues

* AWS ECR "602401143452" which does not allow to list tags so it's not possible to get the latest version. (ECR uses basic auth)

## Example output


|                 IMAGE                 | VERSION | LATEST  | FETCHED |
| --------------------------------------|---------|---------|---------|
| uswitch/kiam                          |  v3.3   |  v3.4   | true    |
| kubernetes-helm/tiller                | v2.13.0 | v2.16.1 | true    |
| cluster-proportional-autoscaler-amd64 |  1.1.1  |  1.7.1  | true    |
| openpolicyagent/opa                   | 0.14.1  | 0.15.1  | true    |
