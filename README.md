# Kubernetes platform lifecycle management

[![Build Status](https://travis-ci.org/arminc/k8s-platform-lcm.svg?branch=master)](https://travis-ci.org/arminc/k8s-platform-lcm)
[![Go Report Card](https://goreportcard.com/badge/github.com/arminc/k8s-platform-lcm)](https://goreportcard.com/report/github.com/arminc/k8s-platform-lcm)

<img src="https://github.com/arminc/k8s-platform-lcm/raw/master/assets/logo.png" width="100">

----

This project helps you keep track of all your software and tools that are used or running in and around your Kubernetes platform. It helps you with part of the lifecycle management to keep your software up to data for feature completeness, security or compliance reasons.

## Features

- [x] Keep track of versions of all the running containers (including init containers) inside the Kubernetes
- [x] Keep track of new image versions. Supporting Quay, Gcr, Docker hub, Jfrog Artifactory by default 
- [x] Works with private registries and private images
- [x] Allow overriding of the registry to search latest versions from another registry
- [x] Keep track of image vulnerabilities using Jfrog Xray
- [x] Possibility to provide local tool versions (like terraform) and find the new versions on GitHub
- [x] Keep track of Helm chart deployments and track new versions of the charts
- [x] Present the information command line

### Todo

* Run as a daemon
* Provide a web UI
* Have a helm chart to deploy the app into Kubernetes
* Use search option on hub.helm.sh instead of direct access
* Automatically fetch new versions every X time
* Add a possibility to whitelist vulnerabilities so only changes are presented
* Provide information on Kubernetes version (for example AWS EKS and it's components)
* Add tests (unit or integration)

* Use Clair as a vulnerability scanning option
* Add Slack/Teams integration
* Push changes/vulnerabilities list to a ConfigMap so anyone with kubectl access can see it

### Issues

* AWS ECR "602401143452" does not allow to list tags so it's not possible to get the latest version. (ECR uses basic auth)

## Help (how to run)

For all the configuration options please have a look at the [exampleConfig.yaml](exampleConfig.yaml). 

When running lcm you can provide certain flags which are not available in the config. The application assumes there is a config.yaml available in the same folder.

```bash
./lcm --help
usage: lcm [<flags>]

Kubernetes platform lifecycle management

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.
  --local    Run locally, default expected behavior is to run in the cluster
  --verbose  Show more information
  --debug    Show debug information, debug includes verbose
  --nok8s    Don't fetch data from kubernetes
```

## Example output

```bash
+---------------------------------------+-------------------+----------+-------+
|                 IMAGE                 |      VERSION      |  LATEST  | CVES  |
+---------------------------------------+-------------------+----------+-------+
| library/alpine                        |      3.10.1       |  3.10.3  | ERROR |
| openpolicyagent/kube-mgmt             |        0.9        |   0.10   | 0     |
| openpolicyagent/opa                   |      0.14.1       |  0.15.1  | 0     |
| velero/velero                         |      v1.1.0       |  v1.2.0  | 0     |
+---------------------------------------+-------------------+----------+-------+
+----------------------------+------------+----------+
|           CHART            |  VERSION   |  LATEST  |
+----------------------------+------------+----------+
| opa                        |   0.12.0   |  1.13.1  |
| velero                     |   2.5.0    |  2.7.0   |
+----------------------------+------------+----------+
+---------------------+---------+----------+
|        TOOL         | VERSION |  LATEST  |
+---------------------+---------+----------+
| derailed/popeye     | v0.4.1  |  v0.5.0  |
| hashicorp/terraform | 0.11.14 | v0.12.18 |
+---------------------+---------+----------+
```