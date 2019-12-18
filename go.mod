module github.com/arminc/k8s-platform-lcm

go 1.13

require (
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/docker/distribution v2.7.1+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/heroku/docker-registry-client v0.0.0-20190909225348-afc9e1acc3d5
	github.com/jfrog/jfrog-client-go v0.5.9
	github.com/knadh/koanf v0.6.0
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/prometheus/common v0.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/target/go-arty v0.0.0-20191122155631-9967a6326524
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
	google.golang.org/appengine v1.6.5
	helm.sh/helm/v3 v3.0.1
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
