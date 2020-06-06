module github.com/arminc/k8s-platform-lcm

go 1.13

require (
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/docker/distribution v2.7.1+incompatible
	github.com/franela/goblin v0.0.0-20200409142057-1def193310bb // indirect
	github.com/gin-gonic/gin v1.6.3 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/google/go-github/v31 v31.0.0
	github.com/gorilla/mux v1.7.4
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/knadh/koanf v0.10.0
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/prometheus/common v0.9.1 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.0
	github.com/target/go-arty v0.0.0-20191122155631-9967a6326524
	github.com/urfave/negroni v1.0.0
	github.com/xenolf/lego v2.7.2+incompatible
	go.etcd.io/etcd v3.3.22+incompatible
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
	google.golang.org/appengine v1.6.6 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.0.1
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	github.com/blang/semver/v4 v4.0.0
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
