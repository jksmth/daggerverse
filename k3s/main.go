// Runs a k3s server than can be accessed both locally and in your pipelines

package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/k-3-s/internal/dagger"
)

// defaultImageRepository is used when no image is specified.
const defaultImageRepository = "rancher/k3s"

// defaultImageRepositoryKubeconfig is used when no image is specified.
const defaultImageRepositoryKubeconfig = "alpine"

// defaultImageRepositoryKubectl is used when no image is specified.
const defaultImageRepositoryKubectl = "bitnami/kubectl"

type K3s struct {
	// +private
	Ctr *dagger.Container

	// +private
	ConfigCache *dagger.CacheVolume

	// +private
	HttpListenPort int

	// +private
	DisableServices []string

	// +private
	DisableHelmController bool

	// +private
	DisableKubeProxy bool

	// +private
	DisableNetworkPolicy bool

	// +private
	DisableCloudController bool

	// +private
	DisableScheduler bool
}

func New(
	// Name of the k3s cluster
	// +optional
	// +default="default"
	name string,

	// Version (image tag) to use from the official image repository as a base container.
	// +optional
	version string,

	// Custom image reference in "repository:tag" format to use as a base container.
	// +optional
	image string,

	// Custom container to use as a base container.
	// +optional
	container *dagger.Container,

	// HTTPS listen port
	// +optional
	// +default=6443
	port int,

	// Disable K3s default services
	// Valid values are: coredns, servicelb, traefik, local-storage, metrics-server
	// +optional
	// +default=[traefik, metrics-server]
	disableServices []string,

	// Disable helm controller
	// +optional
	// +default=false
	disableHelmController bool,

	// Disable kube-proxy
	// +optional
	// +default=false
	disableKubeProxy bool,

	// Disable network policy controller
	// +optional
	// +default=false
	disableNetworkPolicy bool,

	// Disable cloud-controller-manager
	// +optional
	// +default=false
	disableCloudController bool,

	// Disable default scheduler
	// +optional
	// +default=false
	disableScheduler bool,

	// Kubeconfig options
	//

	// Kubeconfig version (image tag) to use from the official image repository as a base container.
	// +optional
	versionKubeconfig string,

	// Kubeconfig custom image reference in "repository:tag" format to use as a base container.
	// +optional
	imageKubeconfig string,

	// Kubeconfig custom container to use as a base container.
	// +optional
	containerKubeconfig *dagger.Container,

	// Kubectl options
	//

	// Kubectl version (image tag) to use from the official image repository as a base container.
	// +optional
	versionKubectl string,

	// Kubectl custom image reference in "repository:tag" format to use as a base container.
	// +optional
	imageKubectl string,

	// Kubectl custom container to use as a base container.
	// +optional
	containerKubectl *dagger.Container,

) *K3s {
	var ctr *dagger.Container

	if version != "" {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	} else if image != "" {
		ctr = dag.Container().From(image)
	} else if container != nil {
		ctr = container
	} else {
		ctr = dag.Container().From(defaultImageRepository)
	}

	return &K3s{
		ConfigCache:            dag.CacheVolume("k3s_config_" + name),
		Ctr:                    ctr,
		HttpListenPort:         port,
		DisableServices:        disableServices,
		DisableHelmController:  disableHelmController,
		DisableKubeProxy:       disableKubeProxy,
		DisableNetworkPolicy:   disableNetworkPolicy,
		DisableCloudController: disableCloudController,
		DisableScheduler:       disableScheduler,
	}
}

// Returns a configured container for the k3s
func (m *K3s) Container() *dagger.Container {
	return m.Ctr.
		With(m.Entrypoint).
		WithMountedCache("/etc/rancher/k3s", m.ConfigCache).
		WithMountedTemp("/etc/lib/cni").
		WithMountedTemp("/var/lib/kubelet").
		WithMountedTemp("/var/lib/rancher/k3s").
		WithMountedTemp("/var/log").
		WithExposedPort(m.HttpListenPort)
}

// helper function configure the k3s server command execution
func (m *K3s) K3s(ctr *dagger.Container) *dagger.Container {
	// k3s server -- options
	opts := []string{"k3s", "server"}

	// HTTPS listen port (default: 6443)
	opts = append(opts, "--https-listen-port", fmt.Sprintf("%d", m.HttpListenPort))

	// Do not deploy packaged components and delete any deployed components
	for _, service := range m.DisableServices {
		opts = append(opts, "--disable", service)
	}

	// Disable Helm controller
	if m.DisableHelmController {
		opts = append(opts, "--disable-helm-controller")
	}

	// Disable running kube-proxy
	if m.DisableKubeProxy {
		opts = append(opts, "--disable-kube-proxy")
	}

	// Disable k3s default network policy controller
	if m.DisableNetworkPolicy {
		opts = append(opts, "--disable-network-policy")
	}

	// Disable k3s default cloud controller manager
	if m.DisableCloudController {
		opts = append(opts, "--disable-cloud-controller")
	}

	// Disable Kubernetes default scheduler
	if m.DisableScheduler {
		opts = append(opts, "--disable-scheduler")

	}

	return ctr.WithExec(
		[]string{"sh", "-c", strings.Join(opts, " ")},
		dagger.ContainerWithExecOpts{InsecureRootCapabilities: true},
	)
}

// Launch K3s as a service
func (m *K3s) Service() *dagger.Service {
	return m.Container().With(m.K3s).AsService()
}

// Use a new container
func (m *K3s) WithContainer(ctr *dagger.Container) *K3s {
	m.Ctr = ctr

	return m
}

// Returns the kubeconfig file from the k3s container
func (m *K3s) Kubeconfig(ctx context.Context,
	// Indicates that the kubeconfig should be use localhost instead of the container IP. This is useful when running k3s as service
	// +optional
	// +default=false
	local bool,
) *dagger.File {
	var ctr *dagger.Container

	if versionKubeconfig != "" {
		ctr = dag.Container().From(fmt.Sprintf("%s:%s", defaultImageRepositoryKubeconfig, versionKubeconfig))
	} else if imageKubeconfig != "" {
		ctr = dag.Container().From(imageKubeconfig)
	} else if containerKubeconfig != nil {
		ctr = containerKubeconfig
	} else {
		ctr = dag.Container().From(defaultImageRepositoryKubeconfig)
	}

	return ctr.
		WithCacheBurster(). // cache buster to force the copy of the k3s.yaml
		WithMountedCache("/cache/k3s", m.ConfigCache).
		WithExec([]string{"cp", "/cache/k3s/k3s.yaml", "k3s.yaml"}).
		With(func(ctr *dagger.Container) *dagger.Container {
			if !local {
				return ctr
			}

			return ctr.WithExec([]string{"sed", "-i", `s/https:.*:/https:\/\/localhost:/g`, "k3s.yaml"})
		}).
		File("k3s.yaml")
}

// runs kubectl on the target k3s cluster
func (m *K3s) Kubectl(ctx context.Context, args string) (string, error) {
	return dag.Container().
		From("bitnami/kubectl").
		WithoutEntrypoint().
		WithMountedCache("/cache/k3s", m.ConfigCache).
		WithCacheBurster(). // cache buster to force the copy of the k3s.yaml
		WithFile("/.kube/config", m.Kubeconfig(ctx, false), dagger.ContainerWithFileOpts{Permissions: 1001}).
		WithUser("1001").
		WithExec([]string{"sh", "-c", "kubectl " + args}).Stdout(ctx)
}

// Helper functions used to configure the k3s container

// a helper function to add the entrypoint to the container
func (_ *K3s) Entrypoint(ctr *dagger.Container) *dagger.Container {
	var (
		file = dag.CurrentModule().Source().File("hack/entrypoint.sh")
		opts = dagger.ContainerWithFileOpts{Permissions: 0o755}
	)

	return ctr.WithFile("/usr/bin/entrypoint.sh", file, opts).WithEntrypoint([]string{"entrypoint.sh"})
}

// helper function to add a cache buster to the container. This will force
// the container execute follow-up steps instead of using the cache
func (m *K3s) WithCacheBurster(
	// Define if the cache burster level is done per day ('daily'), per hour ('hour'), per minute ('minute'), per second ('default') or no cache buster ('none')
	// +optional
	cacheBusterLevel string,
) *K3s {
	// return dag.Utils().WithCacheBurster(m.Ctr, "none")
	return m.WithContainer(dag.Utils().WithCacheBurster(m.Ctr, dagger.UtilsWithCacheBursterOpts{CacheBursterLevel: cacheBusterLevel}))
}
