/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package container

import (
	"encoding/json"
	"fmt"

	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

// NOTE(random-liu):
// 1) Metadata is immutable after created.
// 2) Metadata is checkpointed as containerd container label.

// metadataVersion  is current version of container metadata.
const metadataVersion = "v1" // nolint

// versionedMetadata is the internal versioned container metadata.
// nolint
type versionedMetadata struct {
	// Version indicates the version of the versioned container metadata.
	Version string
	// Metadata's type is metadataInternal. If not there will be a recursive call in MarshalJSON.
	Metadata metadataInternal
}

// metadataInternal is for internal use.
type metadataInternal Metadata

// Metadata is the unversioned container metadata.
type Metadata struct {
	// ID is the container id.
	ID string
	// Name is the container name.
	Name string
	// SandboxID is the sandbox id the container belongs to.
	SandboxID string
	// Config is the CRI container config.
	// NOTE(random-liu): Resource limits are updatable, the source
	// of truth for resource limits are in containerd.
	Config *runtime.ContainerConfig
	// ImageRef is the reference of image used by the container.
	ImageRef string
	// LogPath is the container log path.
	LogPath string
}

// MarshalJSON encodes Metadata into bytes in json format.
func (c *Metadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(&versionedMetadata{
		Version:  metadataVersion,
		Metadata: metadataInternal(*c),
	})
}

// UnmarshalJSON decodes Metadata from bytes.
func (c *Metadata) UnmarshalJSON(data []byte) error {
	versioned := &versionedMetadata{}
	if err := json.Unmarshal(data, versioned); err != nil {
		return err
	}
	// Handle old version after upgrade.
	switch versioned.Version {
	case metadataVersion:
		*c = Metadata(versioned.Metadata)
		return nil
	}
	return fmt.Errorf("unsupported version: %q", versioned.Version)
}
