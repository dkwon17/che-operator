module github.com/eclipse-che/che-operator

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/golang/mock v1.4.1
	github.com/google/go-cmp v0.5.2
	github.com/openshift/api v3.9.0+incompatible
	github.com/operator-framework/api v0.8.0
	github.com/operator-framework/operator-lifecycle-manager v0.18.1
	github.com/sirupsen/logrus v1.7.0
	go.uber.org/zap v1.13.0
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	k8s.io/api v0.20.6
	k8s.io/apiextensions-apiserver v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.2.0
	github.com/openshift/api => github.com/openshift/api v0.0.0-20190924102528-32369d4db2ad
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.3
)
