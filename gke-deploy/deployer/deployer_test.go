/*
Copyright 2019 Google, Inc. All rights reserved.
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
package deployer

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"

	"github.com/GoogleCloudPlatform/cloud-builders/gke-deploy/services"
	"github.com/GoogleCloudPlatform/cloud-builders/gke-deploy/testservices"
)

func TestPrepare(t *testing.T) {
	ctx := context.Background()

	testDeploymentFile := "testing/deployment.yaml"
	testServiceFile := "testing/service.yaml"
	testMultiResourceFile := "testing/multi-resource.yaml"

	image := newImageWithTag(t, "my-image:1.0.0")
	appName := "my-app"
	appVersion := "b2e43cb"
	suggestedDir := "path/to/outputDir/suggested"
	expandedDir := "path/to/outputDir/expanded"
	namespace := "default"
	labels := make(map[string]string)

	configDir := "path/to/config"
	deploymentYaml := "deployment.yaml"
	multiResourceYaml := "multi-resource.yaml"
	namespaceYaml := "namespace.yaml"
	hpaYaml := "horizontalpodautoscaler.yaml"
	serviceYaml := "service.yaml"

	tests := []struct {
		name string

		image           name.Reference
		appName         string
		appVersion      string
		config          string
		suggestedOutput string
		expandedOutput  string
		namespace       string
		labels          map[string]string
		exposePort      int

		deployer *Deployer
	}{{
		name: "Config is directory",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       namespace,
		exposePort:      0,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
						expandedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    "multi-resource-deployment-test-app.yaml",
									IsDirectory: false,
								},
								&testservices.TestFileInfo{
									BaseName:    "multi-resource-service-test-app.yaml",
									IsDirectory: false,
								},
								&testservices.TestFileInfo{
									BaseName:    multiResourceYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, "multi-resource-deployment-test-app.yaml"): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
						filepath.Join(configDir, "multi-resource-service-test-app.yaml"): {
							Res: fileContents(t, testServiceFile),
							Err: nil,
						},
						filepath.Join(configDir, multiResourceYaml): {
							Res: fileContents(t, testMultiResourceFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
						expandedDir:  nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, "multi-resource-deployment-test-app.yaml"):   nil,
						filepath.Join(suggestedDir, "multi-resource-service-test-app.yaml"):      nil,
						filepath.Join(suggestedDir, "multi-resource-deployment-test-app-2.yaml"): nil,
						filepath.Join(suggestedDir, "multi-resource-service-test-app-2.yaml"):    nil,
						filepath.Join(expandedDir, "multi-resource-deployment-test-app.yaml"):    nil,
						filepath.Join(expandedDir, "multi-resource-service-test-app.yaml"):       nil,
						filepath.Join(expandedDir, "multi-resource-deployment-test-app-2.yaml"):  nil,
						filepath.Join(expandedDir, "multi-resource-service-test-app-2.yaml"):     nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "Config is file",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          multiResourceYaml,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       namespace,
		exposePort:      0,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						multiResourceYaml: {
							Res: &testservices.TestFileInfo{
								IsDirectory: false,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
						expandedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						multiResourceYaml: {
							Res: fileContents(t, testMultiResourceFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
						expandedDir:  nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, "multi-resource-deployment-test-app.yaml"): nil,
						filepath.Join(suggestedDir, "multi-resource-service-test-app.yaml"):    nil,
						filepath.Join(expandedDir, "multi-resource-deployment-test-app.yaml"):  nil,
						filepath.Join(expandedDir, "multi-resource-service-test-app.yaml"):     nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "Add custom labels",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels: map[string]string{
			"foo":         "bar",
			"hi":          "bye",
			"a/b/c.d.f.g": "h/i/j.k.l.m",
		},
		namespace:  namespace,
		exposePort: 0,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
						expandedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
						expandedDir:  nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): nil,
						filepath.Join(expandedDir, deploymentYaml):  nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "AppName and AppVersion not set",

		image:           image,
		appName:         "",
		appVersion:      "",
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       namespace,
		exposePort:      0,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
						expandedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
						expandedDir:  nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): nil,
						filepath.Join(expandedDir, deploymentYaml):  nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "Namespace is not default",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       "foobar",
		exposePort:      0,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
						expandedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
						expandedDir:  nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): nil,
						filepath.Join(suggestedDir, namespaceYaml):  nil,
						filepath.Join(expandedDir, deploymentYaml):  nil,
						filepath.Join(expandedDir, namespaceYaml):   nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "Wait for service object to be ready",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       namespace,
		exposePort:      0,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
						expandedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    serviceYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, serviceYaml): {
							Res: fileContents(t, testServiceFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
						expandedDir:  nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, serviceYaml): nil,
						filepath.Join(expandedDir, serviceYaml):  nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "No config arg",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          "",
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       namespace,
		exposePort:      0,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
						expandedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
						expandedDir:  nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): nil,
						filepath.Join(suggestedDir, hpaYaml):        nil,
						filepath.Join(expandedDir, deploymentYaml):  nil,
						filepath.Join(expandedDir, hpaYaml):         nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "Expose application",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       namespace,
		exposePort:      80,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
						expandedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
						expandedDir:  nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): nil,
						filepath.Join(suggestedDir, serviceYaml):    nil,
						filepath.Join(expandedDir, deploymentYaml):  nil,
						filepath.Join(expandedDir, serviceYaml):     nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "Namespace is empty",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       "",
		exposePort:      0,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
						expandedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
						expandedDir:  nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): nil,
						filepath.Join(expandedDir, deploymentYaml):  nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.deployer.Prepare(ctx, tc.image, tc.appName, tc.appVersion, tc.config, tc.suggestedOutput, tc.expandedOutput, tc.namespace, tc.labels, tc.exposePort); err != nil {
				t.Errorf("Prepare(ctx, %v, %s, %s, %s, %s, %s, %s, %v) = %v; want <nil>", tc.image, tc.appName, tc.appVersion, tc.config, tc.suggestedOutput, tc.expandedOutput, tc.namespace, tc.labels, err)
			}
		})
	}
}

func TestPrepareErrors(t *testing.T) {
	ctx := context.Background()

	testDeploymentFile := "testing/deployment.yaml"

	image := newImageWithTag(t, "my-image:1.0.0")
	appName := "my-app"
	appVersion := "b2e43cb"
	suggestedDir := "path/to/suggested"
	expandedDir := "path/to/expanded"
	namespace := "default"
	labels := make(map[string]string)

	configDir := "path/to/config"
	deploymentYaml := "deployment.yaml"

	tests := []struct {
		name string

		image           name.Reference
		appName         string
		appVersion      string
		config          string
		suggestedOutput string
		expandedOutput  string
		namespace       string
		labels          map[string]string

		deployer *Deployer
	}{{
		name: "Failed to parse resources",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       namespace,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{},
							Err: nil,
						},
					},
				},
			},
		},
	}, {
		name: "Failed to get image digest",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       namespace,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: nil,
					ImageErr:  fmt.Errorf("failed to get remote image"),
				},
			},
		},
	}, {
		name: "Failed to save configs",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels:          labels,
		namespace:       namespace,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): fmt.Errorf("failed to write file"),
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "Cannot set app.kubernetes.io/name label via custom labels",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels: map[string]string{
			"app.kubernetes.io/name": "foobar",
		},
		namespace: namespace,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
						//expandedDir: nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): nil,
						//filepath.Join(expandedDir, deploymentYaml): nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "Cannot set app.kubernetes.io/version label via custom labels",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels: map[string]string{
			"app.kubernetes.io/version": "foobar",
		},
		namespace: namespace,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}, {
		name: "Cannot set app.kubernetes.io/managed-by label via custom labels",

		image:           image,
		appName:         appName,
		appVersion:      appVersion,
		config:          configDir,
		suggestedOutput: suggestedDir,
		expandedOutput:  expandedDir,
		labels: map[string]string{
			"app.kubernetes.io/managed-by": "foobar",
		},
		namespace: namespace,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
						suggestedDir: {
							Res: nil,
							Err: os.ErrNotExist,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
					MkdirAllResponse: map[string]error{
						suggestedDir: nil,
					},
					WriteFileResponse: map[string]error{
						filepath.Join(suggestedDir, deploymentYaml): nil,
					},
				},
				Remote: &testservices.TestRemote{
					ImageResp: &testservices.TestImage{
						Hash: v1.Hash{
							Algorithm: "sha256",
							Hex:       "foobar",
						},
						Err: nil,
					},
					ImageErr: nil,
				},
			},
		},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.deployer.Prepare(ctx, tc.image, tc.appName, tc.appVersion, tc.config, tc.suggestedOutput, tc.expandedOutput, tc.namespace, tc.labels, 0); err == nil {
				t.Errorf("Prepare(ctx, %v, %s, %s, %s, %s, %s, %s, %v) = <nil>; want error", tc.image, tc.appName, tc.appVersion, tc.config, tc.suggestedOutput, tc.expandedOutput, tc.namespace, tc.labels)
			}
		})
	}
}

func TestApply(t *testing.T) {
	ctx := context.Background()

	testDeploymentFile := "testing/deployment.yaml"
	testServiceFile := "testing/service.yaml"
	testMultiResourceFile := "testing/multi-resource.yaml"
	testDeploymentReadyFile := "testing/deployment-ready.yaml"
	testServiceUnreadyFile := "testing/service-unready.yaml"
	testServiceReadyFile := "testing/service-ready.yaml"
	testNamespaceFile := "testing/namespace.yaml"
	testNamespace2File := "testing/namespace-2.yaml"
	testNamespaceReadyFile := "testing/namespace-ready.yaml"
	testNamespaceReady2File := "testing/namespace-ready-2.yaml"

	clusterName := "test-cluster"
	clusterLocation := "us-east1-b"
	clusterProject := "my-project"
	namespace := "default"
	waitTimeout := 5 * time.Minute

	configDir := "path/to/config"
	deploymentYaml := "deployment.yaml"
	multiResourceYaml := "multi-resource.yaml"
	namespaceYaml := "namespace.yaml"
	namespace2Yaml := "namespace-2.yaml"
	serviceYaml := "service.yaml"

	tests := []struct {
		name string

		clusterName     string
		clusterLocation string
		config          string
		namespace       string
		labels          map[string]string
		waitTimeout     time.Duration

		deployer *Deployer
	}{{
		name: "Config is directory",

		clusterName:     clusterName,
		clusterLocation: clusterLocation,
		config:          configDir,
		namespace:       namespace,
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    "multi-resource-deployment-test-app.yaml",
									IsDirectory: false,
								},
								&testservices.TestFileInfo{
									BaseName:    "multi-resource-service-test-app.yaml",
									IsDirectory: false,
								},
								&testservices.TestFileInfo{
									BaseName:    multiResourceYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, "multi-resource-deployment-test-app.yaml"): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
						filepath.Join(configDir, "multi-resource-service-test-app.yaml"): {
							Res: fileContents(t, testServiceFile),
							Err: nil,
						},
						filepath.Join(configDir, multiResourceYaml): {
							Res: fileContents(t, testMultiResourceFile),
							Err: nil,
						},
					},
				},
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
				Kubectl: &testservices.TestKubectl{
					ApplyFromStringResponse: map[string]error{
						string(fileContents(t, testDeploymentFile)): nil,
						string(fileContents(t, testServiceFile)):    nil,
					},
					GetResponse: map[string]map[string]*testservices.GetResponse{
						"Deployment": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testDeploymentReadyFile)),
									string(fileContents(t, testDeploymentReadyFile)),
								},
								Err: []error{
									nil,
									nil,
								},
							},
						},
						"Service": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testServiceReadyFile)),
									string(fileContents(t, testServiceReadyFile)),
								},
								Err: []error{
									nil,
									nil,
								},
							},
						},
					},
				},
			},
		},
	}, {
		name: "Config is file",

		clusterName:     clusterName,
		clusterLocation: clusterLocation,
		config:          multiResourceYaml,
		namespace:       namespace,
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						multiResourceYaml: {
							Res: &testservices.TestFileInfo{
								IsDirectory: false,
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						multiResourceYaml: {
							Res: fileContents(t, testMultiResourceFile),
							Err: nil,
						},
					},
				},
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
				Kubectl: &testservices.TestKubectl{
					ApplyFromStringResponse: map[string]error{
						string(fileContents(t, testDeploymentFile)): nil,
						string(fileContents(t, testServiceFile)):    nil,
					},
					GetResponse: map[string]map[string]*testservices.GetResponse{
						"Deployment": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testDeploymentReadyFile)),
								},
								Err: []error{
									nil,
								},
							},
						},
						"Service": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testServiceReadyFile)),
								},
								Err: []error{
									nil,
								},
							},
						},
					},
				},
			},
		},
	}, {
		name: "Namespace is not default",

		clusterName:     clusterName,
		clusterLocation: clusterLocation,
		config:          configDir,
		namespace:       "foobar",
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
								&testservices.TestFileInfo{
									BaseName:    namespaceYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
						filepath.Join(configDir, namespaceYaml): {
							Res: fileContents(t, testNamespaceFile),
							Err: nil,
						},
					},
				},
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
				Kubectl: &testservices.TestKubectl{
					ApplyFromStringResponse: map[string]error{
						string(fileContents(t, testNamespaceFile)):  nil,
						string(fileContents(t, testDeploymentFile)): nil,
					},
					GetResponse: map[string]map[string]*testservices.GetResponse{
						"Deployment": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testDeploymentReadyFile)),
								},
								Err: []error{
									nil,
								},
							},
						},
						"Namespace": {
							"foobar": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testNamespaceReadyFile)),
								},
								Err: []error{
									nil,
								},
							},
						},
					},
				},
			},
		},
	}, {
		name: "Wait for service object to be ready",

		clusterName:     clusterName,
		clusterLocation: clusterLocation,
		config:          configDir,
		namespace:       namespace,
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    serviceYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, serviceYaml): {
							Res: fileContents(t, testServiceFile),
							Err: nil,
						},
					},
				},
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
				Kubectl: &testservices.TestKubectl{
					ApplyFromStringResponse: map[string]error{
						string(fileContents(t, testServiceFile)): nil,
					},
					GetResponse: map[string]map[string]*testservices.GetResponse{
						"Service": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testServiceUnreadyFile)),
									string(fileContents(t, testServiceReadyFile)),
								},
								Err: []error{
									nil,
									nil,
								},
							},
						},
					},
				},
			},
		},
	}, {
		name: "Multiple namespace configs",

		clusterName:     clusterName,
		clusterLocation: clusterLocation,
		config:          configDir,
		namespace:       "foobar",
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
								&testservices.TestFileInfo{
									BaseName:    namespaceYaml,
									IsDirectory: false,
								},
								&testservices.TestFileInfo{
									BaseName:    namespace2Yaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
						filepath.Join(configDir, namespaceYaml): {
							Res: fileContents(t, testNamespaceFile),
							Err: nil,
						},
						filepath.Join(configDir, namespace2Yaml): {
							Res: fileContents(t, testNamespace2File),
							Err: nil,
						},
					},
				},
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
				Kubectl: &testservices.TestKubectl{
					ApplyFromStringResponse: map[string]error{
						string(fileContents(t, testNamespaceFile)):  nil,
						string(fileContents(t, testNamespace2File)): nil,
						string(fileContents(t, testDeploymentFile)): nil,
					},
					GetResponse: map[string]map[string]*testservices.GetResponse{
						"Deployment": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testDeploymentReadyFile)),
								},
								Err: []error{
									nil,
								},
							},
						},
						"Namespace": {
							"foobar": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testNamespaceReadyFile)),
								},
								Err: []error{
									nil,
								},
							},
							"foobar-2": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testNamespaceReady2File)),
								},
								Err: []error{
									nil,
								},
							},
						},
					},
				},
			},
		},
	}, {
		name: "No cluster name and location",

		clusterName:     "",
		clusterLocation: "",
		config:          multiResourceYaml,
		namespace:       namespace,
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						multiResourceYaml: {
							Res: &testservices.TestFileInfo{
								IsDirectory: false,
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						multiResourceYaml: {
							Res: fileContents(t, testMultiResourceFile),
							Err: nil,
						},
					},
				},
				Kubectl: &testservices.TestKubectl{
					ApplyFromStringResponse: map[string]error{
						string(fileContents(t, testDeploymentFile)): nil,
						string(fileContents(t, testServiceFile)):    nil,
					},
					GetResponse: map[string]map[string]*testservices.GetResponse{
						"Deployment": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testDeploymentReadyFile)),
								},
								Err: []error{
									nil,
								},
							},
						},
						"Service": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testServiceReadyFile)),
								},
								Err: []error{
									nil,
								},
							},
						},
					},
				},
			},
		},
	}, {
		name: "Namespace is empty",

		clusterName:     clusterName,
		clusterLocation: clusterLocation,
		config:          configDir,
		namespace:       "",
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
								&testservices.TestFileInfo{
									BaseName:    namespaceYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
						filepath.Join(configDir, namespaceYaml): {
							Res: fileContents(t, testNamespaceFile),
							Err: nil,
						},
					},
				},
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
				Kubectl: &testservices.TestKubectl{
					ApplyFromStringResponse: map[string]error{
						string(fileContents(t, testNamespaceFile)):  nil,
						string(fileContents(t, testDeploymentFile)): nil,
					},
					GetResponse: map[string]map[string]*testservices.GetResponse{
						"Deployment": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testDeploymentReadyFile)),
								},
								Err: []error{
									nil,
								},
							},
						},
						"Namespace": {
							"foobar": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testNamespaceReadyFile)),
								},
								Err: []error{
									nil,
								},
							},
						},
					},
				},
			},
		},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.deployer.Apply(ctx, tc.clusterName, tc.clusterLocation, clusterProject, tc.config, tc.namespace, tc.waitTimeout); err != nil {
				t.Errorf("Apply(ctx, %s, %s, %s, %s, %v) = %v; want <nil>", tc.clusterName, tc.clusterLocation, tc.config, tc.namespace, tc.waitTimeout, err)
			}
		})
	}
}

func TestApplyErrors(t *testing.T) {
	ctx := context.Background()

	testDeploymentFile := "testing/deployment.yaml"
	testServiceFile := "testing/service.yaml"
	testServiceUnreadyFile := "testing/service-unready.yaml"
	testNamespaceFile := "testing/namespace.yaml"

	namespace := "default"
	waitTimeout := 5 * time.Minute

	clusterName := "test-cluster"
	clusterLocation := "us-east1-b"
	clusterProject := "my-project"
	configDir := "path/to/config"
	deploymentYaml := "deployment.yaml"
	namespaceYaml := "namespace.yaml"
	serviceYaml := "service.yaml"

	tests := []struct {
		name string

		clusterName     string
		clusterLocation string
		config          string
		namespace       string
		waitTimeout     time.Duration

		deployer *Deployer
	}{{
		name: "Failed to parse resources",

		clusterName:     clusterName,
		clusterLocation: clusterLocation,
		config:          configDir,
		namespace:       namespace,
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{},
							Err: nil,
						},
					},
				},
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
			},
		},
	}, {
		name: "Failed to get deploy namespace to cluster",

		clusterName:     clusterName,
		clusterLocation: clusterLocation,
		config:          configDir,
		namespace:       "foobar",
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
								&testservices.TestFileInfo{
									BaseName:    namespaceYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
						filepath.Join(configDir, namespaceYaml): {
							Res: fileContents(t, testNamespaceFile),
							Err: nil,
						},
					},
				},
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
				Kubectl: &testservices.TestKubectl{
					ApplyFromStringResponse: map[string]error{
						string(fileContents(t, testNamespaceFile)): fmt.Errorf("failed to apply kubernetes manifests to cluster"),
					},
					GetResponse: map[string]map[string]*testservices.GetResponse{
						"Namespace": {
							"foobar": &testservices.GetResponse{
								Res: []string{
									"",
								},
								Err: []error{
									nil,
								},
							},
						},
					},
				},
			},
		},
	}, {
		name: "Failed to deploy resources to cluster",

		clusterName:     clusterName,
		clusterLocation: clusterLocation,
		config:          configDir,
		namespace:       namespace,
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    deploymentYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, deploymentYaml): {
							Res: fileContents(t, testDeploymentFile),
							Err: nil,
						},
					},
				},
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
				Kubectl: &testservices.TestKubectl{
					ApplyFromStringResponse: map[string]error{
						string(fileContents(t, testDeploymentFile)): fmt.Errorf("failed to apply kubernetes manifests to cluster"),
					},
				},
			},
		},
	}, {
		name: "Wait timeout",

		clusterName:     clusterName,
		clusterLocation: clusterLocation,
		config:          configDir,
		namespace:       namespace,
		waitTimeout:     0 * time.Minute,

		deployer: &Deployer{
			Clients: &services.Clients{
				OS: &testservices.TestOS{
					StatResponse: map[string]testservices.StatResponse{
						configDir: {
							Res: &testservices.TestFileInfo{
								IsDirectory: true,
							},
							Err: nil,
						},
					},
					ReadDirResponse: map[string]testservices.ReadDirResponse{
						configDir: {
							Res: []os.FileInfo{
								&testservices.TestFileInfo{
									BaseName:    serviceYaml,
									IsDirectory: false,
								},
							},
							Err: nil,
						},
					},
					ReadFileResponse: map[string]testservices.ReadFileResponse{
						filepath.Join(configDir, serviceYaml): {
							Res: fileContents(t, testServiceFile),
							Err: nil,
						},
					},
				},
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
				Kubectl: &testservices.TestKubectl{
					ApplyFromStringResponse: map[string]error{
						string(fileContents(t, testServiceFile)): nil,
					},
					GetResponse: map[string]map[string]*testservices.GetResponse{
						"Service": {
							"test-app": &testservices.GetResponse{
								Res: []string{
									string(fileContents(t, testServiceUnreadyFile)),
								},
								Err: []error{
									nil,
								},
							},
						},
					},
				},
			},
		},
	}, {
		name: "clusterName is provided but clusterLocation is not",

		clusterName:     clusterName,
		clusterLocation: "",
		config:          configDir,
		namespace:       namespace,
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
			},
		},
	}, {
		name: "clusterLocation is provided but clusterName is not",

		clusterName:     "",
		clusterLocation: clusterLocation,
		config:          configDir,
		namespace:       namespace,
		waitTimeout:     waitTimeout,

		deployer: &Deployer{
			Clients: &services.Clients{
				Gcloud: &testservices.TestGcloud{
					ContainerClustersGetCredentialsErr: nil,
				},
			},
		},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.deployer.Apply(ctx, tc.clusterName, tc.clusterLocation, clusterProject, tc.config, tc.namespace, tc.waitTimeout); err == nil {
				t.Errorf("Apply(ctx, %s, %s, %s, %s, %v) = <nil>; want error", tc.clusterName, tc.clusterLocation, tc.config, tc.namespace, tc.waitTimeout)
			}
		})
	}
}

func fileContents(t *testing.T, filename string) []byte {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file %s", filename)
	}
	return contents
}

func newImageWithTag(t *testing.T, image string) name.Reference {
	ref, err := name.NewTag(image)
	if err != nil {
		t.Fatalf("failed to create image with tag: %v", err)
	}
	return ref
}
