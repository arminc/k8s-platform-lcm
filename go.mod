module github.com/arminc/k8s-platform-lcm

go 1.13

require (
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/blang/semver/v4 v4.0.0
	github.com/containerd/containerd v1.4.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/franela/goblin v0.0.0-20200409142057-1def193310bb // indirect
	github.com/gin-gonic/gin v1.6.3 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-github/v31 v31.0.0
	github.com/gorilla/mux v1.8.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/knadh/koanf v0.14.0
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.14.0 // indirect
	github.com/prometheus/procfs v0.2.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/target/go-arty v0.0.0-20191122155631-9967a6326524
	github.com/urfave/negroni v1.0.0
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
	golang.org/x/sys v0.0.0-20201112073958-5cba982894dd // indirect https://github.com/ory/dockertest/issues/212
	google.golang.org/appengine v1.6.6 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools/v3 v3.0.2 // indirect
	helm.sh/helm/v3 v3.3.4
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.18.8
	rsc.io/letsencrypt v0.0.3 // indirect
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
)
