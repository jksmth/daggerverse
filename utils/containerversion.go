package main

import (
	"dagger/utils/internal/dagger"
	"fmt"
)

func (m *Utils) WithContainerVersion(
	// Container
	ctr *dagger.Container,

	// Version (image tag) to use from the official image repository as a base container.
	// +optional
	// +default=""
	version string,

	// Custom image reference in "repository:tag" format to use as a base container.
	// +optional
	// +default=""
	image string,

	// Custom container to use as a base container.
	// +optional
	// +default=nil
	container *dagger.Container,

	// Default image repository to use as a base container.
	// +optional
	// +default="alpine"
	defaultImageRepository string,

) *dagger.Container {
	if version != "" {
		ctr = ctr.From(fmt.Sprintf("%s:%s", defaultImageRepository, version))
	} else if image != "" {
		ctr = ctr.From(image)
	} else if container != nil {
		ctr = container
	} else {
		ctr = ctr.From(defaultImageRepository)
	}

	return ctr
}
