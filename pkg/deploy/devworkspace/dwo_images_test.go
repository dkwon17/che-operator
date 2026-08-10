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

package devworkspace

import (
	"context"
	"testing"

	"github.com/eclipse-che/che-operator/pkg/common/test"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetProjectCloneImage(t *testing.T) {
	type testCase struct {
		name          string
		deployment    *appsv1.Deployment
		dwoNamespace  string
		expectedImage string
	}

	testCases := []testCase{
		{
			name:         "project clone image found",
			dwoNamespace: "devworkspace-controller",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dwoDeploymentName,
					Namespace: "devworkspace-controller",
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
			dwoNamespace:  "devworkspace-controller",
			deployment:    nil,
			expectedImage: "",
		},
		{
			name:         "env var not present",
			dwoNamespace: "devworkspace-controller",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dwoDeploymentName,
					Namespace: "devworkspace-controller",
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
		{
			name:          "empty DWO namespace",
			dwoNamespace:  "",
			deployment:    nil,
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

			image := GetProjectCloneImage(
				context.Background(),
				ctx.ClusterAPI.NonCachingClientWrapper,
				tc.dwoNamespace,
			)

			assert.Equal(t, tc.expectedImage, image)
		})
	}
}
