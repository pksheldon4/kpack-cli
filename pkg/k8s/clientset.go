package k8s

import (
	"os"

	"github.com/pkg/errors"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	kpack "github.com/pivotal/kpack/pkg/client/clientset/versioned"
)

type ClientSet struct {
	KpackClient kpack.Interface
	K8sClient   k8s.Interface
	Namespace   string
}

type ClientSetProvider interface {
	GetClientSet(namespace string) (ClientSet, error)
}

type DefaultClientSetProvider struct {
	context ClientSet
}

func (d DefaultClientSetProvider) GetClientSet(namespace string) (ClientSet, error) {
	var err error

	if namespace == "" {
		if d.context.Namespace, err = d.getDefaultNamespace(); err != nil {
			return d.context, err
		}
	} else {
		d.context.Namespace = namespace
	}

	if d.context.KpackClient, err = d.getKpackClient(); err != nil {
		return d.context, err
	}

	d.context.K8sClient, err = d.getK8sClient()
	return d.context, err
}

func (d DefaultClientSetProvider) getKpackClient() (*kpack.Clientset, error) {
	restConfig, err := d.restConfig()
	if err != nil {
		return nil, err
	}

	return kpack.NewForConfig(restConfig)
}

func (d DefaultClientSetProvider) getK8sClient() (*k8s.Clientset, error) {
	restConfig, err := d.restConfig()
	if err != nil {
		return nil, err
	}

	return k8s.NewForConfig(restConfig)
}

func (d DefaultClientSetProvider) restConfig() (*rest.Config, error) {
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
		os.Stdin,
	)

	restConfig, err := clientConfig.ClientConfig()
	return restConfig, err
}

func (d DefaultClientSetProvider) getDefaultNamespace() (string, error) {
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
		os.Stdin,
	)

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return "", err
	}

	if _, ok := rawConfig.Contexts[rawConfig.CurrentContext]; !ok {
		return "", errors.New("Kubernetes current context is not set")
	}

	defaultNamespace := rawConfig.Contexts[rawConfig.CurrentContext].Namespace
	if defaultNamespace == "" {
		defaultNamespace = "default"
	}

	return defaultNamespace, nil
}
