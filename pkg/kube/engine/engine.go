package engine

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Engine reconciles Kubernetes resources onto the target cluster using server-side apply.
type Engine struct {
	client dynamic.Interface
	mapper meta.RESTMapper
}

// Client returns the Kubernetes client used by the engine.
func (m *Engine) Client() dynamic.Interface {
	return m.client
}

// NewDefaultEngine creates a new Engine with default configuration using the specified config flags.
func NewDefaultEngine() (*Engine, error) {
	return NewFromClientGetter(genericclioptions.NewConfigFlags(true))
}

// NewFromKubeConfig creates a new Engine from a kubeconfig file path, initializing the Kubernetes client and necessary configurations.
func NewFromKubeConfig(kubeConfigPath string) (*Engine, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig failed: %w", err)
	}
	return NewFromRestConfig(cfg)
}

// NewFromClientGetter creates a new Engine from a REST client getter, initializing the Kubernetes client and necessary configurations.
func NewFromClientGetter(getter genericclioptions.RESTClientGetter) (*Engine, error) {
	cfg, err := getter.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig failed: %w", err)
	}
	return NewFromRestConfig(cfg)
}

// NewFromRestConfig creates a new Engine from a rest.Config, initializing the Kubernetes client and necessary configurations.
func NewFromRestConfig(cfg *rest.Config) (*Engine, error) {
	client, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}
	mapper, err := apiutil.NewDynamicRESTMapper(cfg, client)
	if err != nil {
		return nil, err
	}
	// Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Engine{
		mapper: mapper,
		client: dyn,
	}, nil
}
