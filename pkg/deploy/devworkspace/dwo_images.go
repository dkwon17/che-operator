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

	"github.com/eclipse-che/che-operator/pkg/common/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	k8sclient "github.com/eclipse-che/che-operator/pkg/common/k8s-client"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	dwoDeploymentName          = "devworkspace-controller-manager"
	dwoProjectCloneImageEnvVar = "RELATED_IMAGE_project_clone"
)

var logger = ctrl.Log.WithName("devworkspace")

func GetProjectCloneImage(
	ctx context.Context,
	clientWrapper *k8sclient.K8sClientWrapper,
	dwoNamespace string,
) string {
	if dwoNamespace == "" {
		logger.Info("DevWorkspace operator namespace is not resolved, skipping project clone image")
		return ""
	}

	deployment := &appsv1.Deployment{}
	key := types.NamespacedName{
		Name:      dwoDeploymentName,
		Namespace: dwoNamespace,
	}

	exists, err := clientWrapper.GetIgnoreNotFound(ctx, key, deployment)
	if err != nil {
		logger.Info("Failed to get DevWorkspace controller deployment", "error", err)
		return ""
	}

	if !exists {
		logger.Info("DevWorkspace controller deployment not found, skipping project clone image")
		return ""
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		image := utils.GetEnvByName(dwoProjectCloneImageEnvVar, container.Env)
		if image != "" {
			return image
		}
	}

	logger.Info("Project clone image env var not found in DevWorkspace controller deployment")
	return ""
}
