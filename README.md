# Kubernetes platform lifecycle management

Keeping software up to date is always an issue and the more you have of it the harder it gets. It's not only the work of updating but also figuring out what to update and which tools have new versions. I am talking about lifecycle management. 

When you add the mix of delivering a Kubernetes platform which has a lot of components that are updated often it's hard to keep track. If you want to be up to date because of feature completeness or security and compliance then there is a lot of tedious work to do.

Luckily this project can help you out. It can track all those versions for you automatically and show you what you need to update. For implemented features and future features please see below. 

## Implemented features

* Find all the running containers and init containers in Kubernetes
* Search for new versions, works with Quay, Gcr, Docker hub, Jfrog Artifactory and probably any other Docker registry v2
* Allows you to override the registry to search in, for example, if you are using private registry but need to fetch versions from the internet
* It works with private registries and private images as well

## TODO

**Must have**

* Use Jfrog Xray to find which images contain vulnerabilities 
* Show the information regarding vulnerabilities
* Add the possibility to whitelist vulnerabilities so you only see when something changes
* Add a possibility to provide local tool versions (like terraform version and it's plugins) and find the new versions using GitHub 
* Add a possibility to find Helm versions deployed in Kubernetes and find new versions
* Provide information for Kubernetes version (for example AWS EKS)

**Bugs**

* There is an issue when using 'latest' tag it fails to compare the versions

**Nice to have**

* Specify `ALL` as an option for namespaces, it should find all namespaces and look at all of them
* Use Clair as a vulnerabilitie scanning option

## Issues

* AWS ECR "602401143452" which does not allow to list tags so it's not possible to get the latest version. (ECR uses basic auth)

## Example output

+---------------------------------------+---------+---------+---------+
|                 IMAGE                 | VERSION | LATEST  | FETCHED |
+---------------------------------------+---------+---------+---------+
| uswitch/kiam                          |  v3.3   |  v3.4   | true    |
| kubernetes-helm/tiller                | v2.13.0 | v2.16.1 | true    |
| cluster-proportional-autoscaler-amd64 |  1.1.1  |  1.7.1  | true    |
| openpolicyagent/opa                   | 0.14.1  | 0.15.1  | true    |
+---------------------------------------+---------+---------+---------+