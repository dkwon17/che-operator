//
// Copyright (c) 2019-2026 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package imagepuller

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eclipse-che/che-operator/pkg/common/test"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExternalImages(t *testing.T) {
	type testCase struct {
		name           string
		editorsFile    string
		samplesFile    string
		expectedImages []string
	}

	testCases := []testCase{
		{
			name:           "both editors and samples images",
			editorsFile:    "image-puller-resources-test/editors.json",
			samplesFile:    "image-puller-resources-test/samples.json",
			expectedImages: []string{"image_1", "image_2", "image_3"},
		},
		{
			name:           "no external images",
			editorsFile:    "image-puller-resources-test/empty.json",
			samplesFile:    "image-puller-resources-test/empty.json",
			expectedImages: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := test.NewCtxBuilder().Build()

			editorsEndpointUrl := getDashboardEditorsInternalAPIUrl(ctx)
			samplesEndpointUrl := getDashboardSamplesInternalAPIUrl(ctx)

			imagesProvider := &ExternalImagesProvider{
				imagesFilePath: filepath.Join(os.TempDir(), externalImagesStoreFileName),
				fetchRawDataFunc: func(url string) ([]byte, error) {
					switch url {
					case editorsEndpointUrl:
						return os.ReadFile(tc.editorsFile)
					case samplesEndpointUrl:
						return os.ReadFile(tc.samplesFile)
					case "sample_1_url":
						return os.ReadFile("image-puller-resources-test/sample_1.yaml")
					case "sample_2_url":
						return os.ReadFile("image-puller-resources-test/sample_2.yaml")
					default:
						return []byte{}, fmt.Errorf("unexpected url: %s", url)
					}
				},
			}

			images, err := imagesProvider.Get(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedImages, images)

			data, err := os.ReadFile(filepath.Join(os.TempDir(), externalImagesStoreFileName))
			assert.NoError(t, err)

			expectedFileContent := strings.Join(tc.expectedImages, "\n")
			assert.Equal(t, expectedFileContent, string(data))
		})
	}
}

func TestFetchDWOProjectCloneImage(t *testing.T) {
	type testCase struct {
		name          string
		deployment    *appsv1.Deployment
		expectedImage string
	}

	testCases := []testCase{
		{
			name: "project clone image found",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dwoDeploymentName,
					Namespace: "eclipse-che",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "devworkspace-controller",
									Env: []corev1.EnvVar{
										{Name: "RELATED_IMAGE_project_clone", Value: "quay.io/devspaces/project-clone:latest"},
									},
								},
							},
						},
					},
				},
			},
			expectedImage: "quay.io/devspaces/project-clone:latest",
		},
		{
			name:          "deployment not found",
			deployment:    nil,
			expectedImage: "",
		},
		{
			name: "env var not present",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dwoDeploymentName,
					Namespace: "eclipse-che",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "devworkspace-controller",
									Env:  []corev1.EnvVar{},
								},
							},
						},
					},
				},
			},
			expectedImage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := test.NewCtxBuilder()
			if tc.deployment != nil {
				builder.WithObjects(tc.deployment)
			}
			ctx := builder.Build()

			imagesProvider := NewExternalImagesProvider()
			image := imagesProvider.fetchDWOProjectCloneImage(ctx)

			assert.Equal(t, tc.expectedImage, image)
		})
	}
}
