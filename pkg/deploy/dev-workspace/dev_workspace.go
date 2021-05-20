//
// Copyright (c) 2021 Red Hat, Inc.
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
	"errors"
	"fmt"

	orgv1 "github.com/eclipse-che/che-operator/pkg/apis/org/v1"
	"github.com/eclipse-che/che-operator/pkg/deploy"
	"github.com/eclipse-che/che-operator/pkg/util"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	DevWorkspaceNamespace      = "devworkspace-controller"
	DevWorkspaceCheNamespace   = "devworkspace-che"
	DevWorkspaceWebhookName    = "controller.devfile.io"
	DevWorkspaceServiceAccount = "devworkspace-controller-serviceaccount"
	DevWorkspaceDeploymentName = "devworkspace-controller-manager"
	SubscriptionResourceName   = "subscriptions"
	CheManagerResourcename     = "chemanagers"

	OpenshiftDevWorkspaceTemplatesPath     = "/tmp/devworkspace-operator/templates/deployment/openshift/objects"
	OpenshiftDevWorkspaceCheTemplatesPath  = "/tmp/devworkspace-che-operator/templates/deployment/openshift/objects"
	KubernetesDevWorkspaceTemplatesPath    = "/tmp/devworkspace-operator/templates/deployment/kubernetes/objects"
	KubernetesDevWorkspaceCheTemplatesPath = "/tmp/devworkspace-che-operator/templates/deployment/kubernetes/objects"

	DevWorkspaceTemplates    = devWorkspaceTemplatesPath()
	DevWorkspaceCheTemplates = devWorkspaceCheTemplatesPath()

	DevWorkspaceServiceAccountFile            = DevWorkspaceTemplates + "/devworkspace-controller-serviceaccount.ServiceAccount.yaml"
	DevWorkspaceRoleFile                      = DevWorkspaceTemplates + "/devworkspace-controller-leader-election-role.Role.yaml"
	DevWorkspaceClusterRoleFile               = DevWorkspaceTemplates + "/devworkspace-controller-role.ClusterRole.yaml"
	DevWorkspaceProxyClusterRoleFile          = DevWorkspaceTemplates + "/devworkspace-controller-proxy-role.ClusterRole.yaml"
	DevWorkspaceViewWorkspacesClusterRoleFile = DevWorkspaceTemplates + "/devworkspace-controller-view-workspaces.ClusterRole.yaml"
	DevWorkspaceEditWorkspacesClusterRoleFile = DevWorkspaceTemplates + "/devworkspace-controller-edit-workspaces.ClusterRole.yaml"
	DevWorkspaceRoleBindingFile               = DevWorkspaceTemplates + "/devworkspace-controller-leader-election-rolebinding.RoleBinding.yaml"
	DevWorkspaceClusterRoleBindingFile        = DevWorkspaceTemplates + "/devworkspace-controller-rolebinding.ClusterRoleBinding.yaml"
	DevWorkspaceProxyClusterRoleBindingFile   = DevWorkspaceTemplates + "/devworkspace-controller-proxy-rolebinding.ClusterRoleBinding.yaml"
	DevWorkspaceWorkspaceRoutingCRDFile       = DevWorkspaceTemplates + "/devworkspaceroutings.controller.devfile.io.CustomResourceDefinition.yaml"
	DevWorkspaceTemplatesCRDFile              = DevWorkspaceTemplates + "/devworkspacetemplates.workspace.devfile.io.CustomResourceDefinition.yaml"
	DevWorkspaceCRDFile                       = DevWorkspaceTemplates + "/devworkspaces.workspace.devfile.io.CustomResourceDefinition.yaml"
	DevWorkspaceConfigMapFile                 = DevWorkspaceTemplates + "/devworkspace-controller-configmap.ConfigMap.yaml"
	DevWorkspaceDeploymentFile                = DevWorkspaceTemplates + "/devworkspace-controller-manager.Deployment.yaml"

	DevWorkspaceCheServiceAccountFile           = DevWorkspaceCheTemplates + "/devworkspace-che-serviceaccount.ServiceAccount.yaml"
	DevWorkspaceCheRoleFile                     = DevWorkspaceCheTemplates + "/devworkspace-che-leader-election-role.Role.yaml"
	DevWorkspaceCheClusterRoleFile              = DevWorkspaceCheTemplates + "/devworkspace-che-role.ClusterRole.yaml"
	DevWorkspaceCheProxyClusterRoleFile         = DevWorkspaceCheTemplates + "/devworkspace-che-proxy-role.ClusterRole.yaml"
	DevWorkspaceCheMetricsReaderClusterRoleFile = DevWorkspaceCheTemplates + "/devworkspace-che-metrics-reader.ClusterRole.yaml"
	DevWorkspaceCheRoleBindingFile              = DevWorkspaceCheTemplates + "/devworkspace-che-leader-election-rolebinding.RoleBinding.yaml"
	DevWorkspaceCheClusterRoleBindingFile       = DevWorkspaceCheTemplates + "/devworkspace-che-rolebinding.ClusterRoleBinding.yaml"
	DevWorkspaceCheProxyClusterRoleBindingFile  = DevWorkspaceCheTemplates + "/devworkspace-che-proxy-rolebinding.ClusterRoleBinding.yaml"
	DevWorkspaceCheManagersCRDFile              = DevWorkspaceCheTemplates + "/chemanagers.che.eclipse.org.CustomResourceDefinition.yaml"
	DevWorkspaceCheConfigMapFile                = DevWorkspaceCheTemplates + "/devworkspace-che-configmap.ConfigMap.yaml"
	DevWorkspaceCheDeploymentFile               = DevWorkspaceCheTemplates + "/devworkspace-che-manager.Deployment.yaml"
	DevWorkspaceCheMetricsServiceFile           = DevWorkspaceCheTemplates + "/devworkspace-che-controller-manager-metrics-service.Service.yaml"

	WebTerminalOperatorSubscriptionName = "web-terminal"
	WebTerminalOperatorNamespace        = "openshift-operators"
)

type Object2Sync struct {
	obj     metav1.Object
	hash256 string
}

var (
	// cachedObjects
	cachedObj = make(map[string]*Object2Sync)
	syncItems = []func(*deploy.DeployContext) (bool, error){
		createDwNamespace,
		syncDwServiceAccount,
		syncDwClusterRole,
		syncDwProxyClusterRole,
		syncDwEditWorkspacesClusterRole,
		syncDwViewWorkspacesClusterRole,
		syncDwRole,
		syncDwRoleBinding,
		syncDwClusterRoleBinding,
		syncDwProxyClusterRoleBinding,
		syncDwCRD,
		syncDwTemplatesCRD,
		syncDwWorkspaceRoutingCRD,
		syncDwConfigMap,
		syncDwDeployment,
	}

	syncDwCheItems = []func(*deploy.DeployContext) (bool, error){
		// deleting unused objects
		deleteDwCheDeployment,
		deleteDwCheMetricsService,
		deleteDwCheServiceAccount,
		deleteDwCheClusterRoleBinding,
		deleteDwCheClusterRole,
		deleteDwCheProxyClusterRoleBinding,
		deleteDwCheProxyClusterRole,
		deleteDwCheMetricsClusterRole,
		deleteDwCheLeaderRoleBinding,
		deleteDwCheLeaderRole,

		// migrating existed objects
		migrateDwCheConfigMap,
		migrateDwCheCR,

		// syncing new objects
		syncDwCheCR,
		syncDwCheMetricsService,
	}
)

func ReconcileDevWorkspace(deployContext *deploy.DeployContext) (bool, error) {
	if util.IsOpenShift {
		if !util.IsOpenShift4 {
			// OpenShift 3.x is not supported
			return true, nil
		}
	} else if util.GetServerExposureStrategy(deployContext.CheCluster) == "single-host" {
		logrus.Warn(`DevWorkspace Che operator can't be enabled in 'single-host mode on Kubernetes cluster'.
See https://github.com/eclipse/che/issues/19714 for more details.
To enable DevWorkspace Che operator set 'spec.server.serverExposureStrategy' to 'multi-host'.`)
		return true, nil
	}

	for _, syncItem := range syncDwCheItems {
		done, err := syncItem(deployContext)
		if !util.IsTestMode() {
			if !done {
				return false, err
			}
		}
	}

	if !deployContext.CheCluster.Spec.DevWorkspace.Enable {
		return true, nil
	}

	devWorkspaceWebhookExists, err := deploy.Get(
		deployContext,
		client.ObjectKey{Name: DevWorkspaceWebhookName},
		&admissionregistrationv1.MutatingWebhookConfiguration{},
	)
	if err != nil {
		return false, err
	}

	if devWorkspaceWebhookExists {
		if err := checkWebTerminalSubscription(deployContext); err != nil {
			return false, err
		}
	}

	for _, syncItem := range syncItems {
		done, err := syncItem(deployContext)
		if !util.IsTestMode() {
			if !done {
				return false, err
			}
		}
	}

	return true, nil
}

func checkWebTerminalSubscription(deployContext *deploy.DeployContext) error {
	// If subscriptions resource doesn't exist in cluster WTO as well will not be present
	if !util.HasK8SResourceObject(deployContext.ClusterAPI.DiscoveryClient, SubscriptionResourceName) {
		return nil
	}

	subscription := &operatorsv1alpha1.Subscription{}
	if err := deployContext.ClusterAPI.NonCachedClient.Get(
		context.TODO(),
		types.NamespacedName{
			Name:      WebTerminalOperatorSubscriptionName,
			Namespace: WebTerminalOperatorNamespace,
		},
		subscription); err != nil {

		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return errors.New("A non matching version of the Dev Workspace operator is already installed")
}

func createDwNamespace(deployContext *deploy.DeployContext) (bool, error) {
	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: DevWorkspaceNamespace,
		},
		Spec: corev1.NamespaceSpec{},
	}

	return deploy.CreateIfNotExists(deployContext, namespace)
}

func syncDwServiceAccount(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceServiceAccountFile, &corev1.ServiceAccount{}, DevWorkspaceNamespace)
}

func syncDwRole(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceRoleFile, &rbacv1.Role{}, DevWorkspaceNamespace)
}

func syncDwRoleBinding(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceRoleBindingFile, &rbacv1.RoleBinding{}, DevWorkspaceNamespace)
}

func syncDwClusterRoleBinding(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceClusterRoleBindingFile, &rbacv1.ClusterRoleBinding{}, "")
}

func syncDwProxyClusterRoleBinding(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceProxyClusterRoleBindingFile, &rbacv1.ClusterRoleBinding{}, "")
}

func syncDwClusterRole(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceClusterRoleFile, &rbacv1.ClusterRole{}, "")
}

func syncDwProxyClusterRole(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceProxyClusterRoleFile, &rbacv1.ClusterRole{}, "")
}

func syncDwViewWorkspacesClusterRole(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceViewWorkspacesClusterRoleFile, &rbacv1.ClusterRole{}, "")
}

func syncDwEditWorkspacesClusterRole(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceEditWorkspacesClusterRoleFile, &rbacv1.ClusterRole{}, "")
}

func syncDwWorkspaceRoutingCRD(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceWorkspaceRoutingCRDFile, &apiextensionsv1.CustomResourceDefinition{}, "")
}

func syncDwTemplatesCRD(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceTemplatesCRDFile, &apiextensionsv1.CustomResourceDefinition{}, "")
}

func syncDwCRD(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceCRDFile, &apiextensionsv1.CustomResourceDefinition{}, "")
}

func syncDwConfigMap(deployContext *deploy.DeployContext) (bool, error) {
	obj2sync, err := readK8SObject(DevWorkspaceConfigMapFile, &corev1.ConfigMap{})
	if err != nil {
		return false, err
	}

	configMap := obj2sync.obj.(*corev1.ConfigMap)
	// Remove when DevWorkspace controller should not care about DWR base host #373 https://github.com/devfile/devworkspace-operator/issues/373
	if !util.IsOpenShift {
		if configMap.Data == nil {
			configMap.Data = make(map[string]string, 1)
		}
		configMap.Data["devworkspace.routing.cluster_host_suffix"] = deployContext.CheCluster.Spec.K8s.IngressDomain
	}

	return syncObject(deployContext, obj2sync, DevWorkspaceNamespace)
}

func syncDwDeployment(deployContext *deploy.DeployContext) (bool, error) {
	obj2sync, err := readK8SObject(DevWorkspaceDeploymentFile, &appsv1.Deployment{})
	if err != nil {
		return false, err
	}

	devworkspaceControllerImage := util.GetValue(deployContext.CheCluster.Spec.DevWorkspace.ControllerImage, deploy.DefaultDevworkspaceControllerImage(deployContext.CheCluster))
	deploymentObject := obj2sync.obj.(*appsv1.Deployment)
	deploymentObject.Spec.Template.Spec.Containers[0].Image = devworkspaceControllerImage

	return syncObject(deployContext, obj2sync, DevWorkspaceNamespace)
}

func deleteDwCheServiceAccount(deployContext *deploy.DeployContext) (bool, error) {
	return readAndDeleteObject(deployContext, DevWorkspaceCheServiceAccountFile, &corev1.ServiceAccount{}, DevWorkspaceCheNamespace)
}

func deleteDwCheClusterRole(deployContext *deploy.DeployContext) (bool, error) {
	return readAndDeleteObject(deployContext, DevWorkspaceCheClusterRoleFile, &rbacv1.ClusterRole{}, "")
}

func deleteDwCheProxyClusterRole(deployContext *deploy.DeployContext) (bool, error) {
	return readAndDeleteObject(deployContext, DevWorkspaceCheProxyClusterRoleFile, &rbacv1.ClusterRole{}, "")
}

func deleteDwCheMetricsClusterRole(deployContext *deploy.DeployContext) (bool, error) {
	return readAndDeleteObject(deployContext, DevWorkspaceCheMetricsReaderClusterRoleFile, &rbacv1.ClusterRole{}, "")
}

func deleteDwCheLeaderRole(deployContext *deploy.DeployContext) (bool, error) {
	return readAndDeleteObject(deployContext, DevWorkspaceCheRoleFile, &rbacv1.Role{}, DevWorkspaceCheNamespace)
}

func deleteDwCheLeaderRoleBinding(deployContext *deploy.DeployContext) (bool, error) {
	return readAndDeleteObject(deployContext, DevWorkspaceCheRoleBindingFile, &rbacv1.RoleBinding{}, DevWorkspaceCheNamespace)
}

func deleteDwCheProxyClusterRoleBinding(deployContext *deploy.DeployContext) (bool, error) {
	return readAndDeleteObject(deployContext, DevWorkspaceCheProxyClusterRoleBindingFile, &rbacv1.ClusterRoleBinding{}, "")
}

func deleteDwCheClusterRoleBinding(deployContext *deploy.DeployContext) (bool, error) {
	return readAndDeleteObject(deployContext, DevWorkspaceCheClusterRoleBindingFile, &rbacv1.ClusterRoleBinding{}, "")
}

func deleteDwCheMetricsService(deployContext *deploy.DeployContext) (bool, error) {
	return readAndDeleteObject(deployContext, DevWorkspaceCheMetricsServiceFile, &corev1.Service{}, "")
}

func migrateDwCheConfigMap(deployContext *deploy.DeployContext) (bool, error) {
	obj2sync, err := readK8SObject(DevWorkspaceCheConfigMapFile, &corev1.ConfigMap{})
	if err != nil {
		return false, err
	}
	obj2sync.obj.SetNamespace(DevWorkspaceCheNamespace)

	actual := &corev1.ConfigMap{}
	actualObjKey := types.NamespacedName{Name: obj2sync.obj.GetName(), Namespace: obj2sync.obj.GetNamespace()}
	exists, err := deploy.Get(deployContext, actualObjKey, actual)
	if !exists {
		return err == nil, err
	}

	runtimeObject, ok := obj2sync.obj.(runtime.Object)
	if !ok {
		return false, fmt.Errorf("object %T is not a runtime.Object. Cannot sync it", runtimeObject)
	}
	newObj := runtimeObject.DeepCopyObject()
	newObj.(metav1.Object).SetNamespace(deployContext.CheCluster.Namespace)
	err = deployContext.ClusterAPI.Client.Create(context.TODO(), newObj)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			logrus.Warnf("%s config map already exists in %s namespace.", obj2sync.obj.GetName(), deployContext.CheCluster.Namespace)
		} else {
			return false, err
		}
	}

	done, err := deploy.Delete(deployContext, actualObjKey, &corev1.ConfigMap{})
	if !done {
		return false, err
	}

	return true, nil
}

// migrates DWChe CR from `devworkspace-che` namespace
func migrateDwCheCR(deployContext *deploy.DeployContext) (bool, error) {
	actual := &unstructured.Unstructured{}
	actual.SetGroupVersionKind(schema.GroupVersionKind{Group: "che.eclipse.org", Version: "v1alpha1", Kind: "CheManager"})
	err := deployContext.ClusterAPI.Client.Get(context.TODO(), client.ObjectKey{Name: "devworkspace-che", Namespace: DevWorkspaceCheNamespace}, actual)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return true, nil
		} else {
			return false, err
		}
	}

	newObj := &unstructured.Unstructured{}
	newObj.SetUnstructuredContent(actual.DeepCopy().UnstructuredContent())
	newObj.SetNamespace(deployContext.CheCluster.Namespace)
	err = deployContext.ClusterAPI.Client.Create(context.TODO(), newObj)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			logrus.Warnf("CheManager custom resources already exists in %s namespace.", deployContext.CheCluster.Namespace)
		} else {
			return false, err
		}
	}

	err = deployContext.ClusterAPI.Client.Delete(context.TODO(), actual)
	if err != nil {
		return false, err
	}

	return true, nil
}

func deleteDwCheDeployment(deployContext *deploy.DeployContext) (bool, error) {
	return readAndDeleteObject(deployContext, DevWorkspaceCheDeploymentFile, &corev1.Service{}, DevWorkspaceCheNamespace)
}

func syncDwCheMetricsService(deployContext *deploy.DeployContext) (bool, error) {
	return readAndSyncObject(deployContext, DevWorkspaceCheMetricsServiceFile, &corev1.Service{}, deployContext.CheCluster.Namespace)
}

func syncDwCheCR(deployContext *deploy.DeployContext) (bool, error) {
	// We want to create a default CheManager instance to be able to configure the che-specific
	// parts of the installation, but at the same time we don't want to add a dependency on
	// devworkspace-che-operator. Note that this way of initializing will probably see changes
	// once we figure out https://github.com/eclipse/che/issues/19220

	// Wait until CRD for CheManager is created
	if !util.HasK8SResourceObject(deployContext.ClusterAPI.DiscoveryClient, CheManagerResourcename) {
		return false, nil
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "che.eclipse.org", Version: "v1alpha1", Kind: "CheManager"})
	err := deployContext.ClusterAPI.Client.Get(context.TODO(), client.ObjectKey{Name: "devworkspace-che", Namespace: deployContext.CheCluster.Namespace}, obj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			obj = nil
		} else {
			return false, err
		}
	}

	if obj == nil {
		obj := &unstructured.Unstructured{}
		if !util.IsOpenShift {
			obj.SetUnstructuredContent(map[string]interface{}{
				"spec": map[string]interface{}{
					"gatewayHost": deployContext.CheCluster.Spec.K8s.IngressDomain,
				},
			})
		}
		obj.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "che.eclipse.org",
			Version: "v1alpha1",
			Kind:    "CheManager",
		})
		obj.SetName("devworkspace-che")
		obj.SetNamespace(deployContext.CheCluster.Namespace)

		err = deployContext.ClusterAPI.Client.Create(context.TODO(), obj)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				return false, nil
			}
			return false, err
		}
	}

	return true, nil
}

func readAndSyncObject(deployContext *deploy.DeployContext, yamlFile string, obj interface{}, namespace string) (bool, error) {
	obj2sync, err := readK8SObject(yamlFile, obj)
	if err != nil {
		return false, err
	}

	return syncObject(deployContext, obj2sync, namespace)
}

func syncObject(deployContext *deploy.DeployContext, obj2sync *Object2Sync, namespace string) (bool, error) {
	obj2sync.obj.SetNamespace(namespace)

	runtimeObject, ok := obj2sync.obj.(runtime.Object)
	if !ok {
		return false, fmt.Errorf("object %T is not a runtime.Object. Cannot sync it", runtimeObject)
	}

	actual := runtimeObject.DeepCopyObject()
	key := types.NamespacedName{Namespace: obj2sync.obj.GetNamespace(), Name: obj2sync.obj.GetName()}
	exists, err := deploy.Get(deployContext, key, actual.(metav1.Object))
	if err != nil {
		return false, err
	}

	isOnlyOneOperatorManagesDWResources, err := isOnlyOneOperatorManagesDWResources(deployContext)
	if err != nil {
		return false, err
	}

	// sync objects if it has been created by same operator
	// or it is the only operator with the `spec.devWorkspace.enable: true`
	if !exists ||
		(actual.(metav1.Object).GetAnnotations()[deploy.CheEclipseOrgHash256] != obj2sync.hash256 &&
			(actual.(metav1.Object).GetAnnotations()[deploy.CheEclipseOrgNamespace] == deployContext.CheCluster.Namespace || isOnlyOneOperatorManagesDWResources)) {

		setAnnotations(deployContext, obj2sync)
		return deploy.Sync(deployContext, obj2sync.obj)
	}

	return true, nil
}

func readAndDeleteObject(deployContext *deploy.DeployContext, yamlFile string, obj interface{}, namespace string) (bool, error) {
	obj2sync, err := readK8SObject(yamlFile, obj)
	if err != nil {
		return false, err
	}
	return deleteObject(deployContext, obj2sync, namespace)
}

func deleteObject(deployContext *deploy.DeployContext, obj2sync *Object2Sync, namespace string) (bool, error) {
	obj2sync.obj.SetNamespace(namespace)
	key := types.NamespacedName{Name: obj2sync.obj.GetName(), Namespace: obj2sync.obj.GetNamespace()}
	return deploy.Delete(deployContext, key, obj2sync.obj.(metav1.Object))
}

func setAnnotations(deployContext *deploy.DeployContext, obj2sync *Object2Sync) {
	annotations := obj2sync.obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[deploy.CheEclipseOrgNamespace] = deployContext.CheCluster.Namespace
	annotations[deploy.CheEclipseOrgHash256] = obj2sync.hash256
	obj2sync.obj.SetAnnotations(annotations)
}

func isOnlyOneOperatorManagesDWResources(deployContext *deploy.DeployContext) (bool, error) {
	cheClusters := &orgv1.CheClusterList{}
	err := deployContext.ClusterAPI.NonCachedClient.List(context.TODO(), cheClusters)
	if err != nil {
		return false, err
	}

	devWorkspaceEnabledNum := 0
	for _, cheCluster := range cheClusters.Items {
		if cheCluster.Spec.DevWorkspace.Enable {
			devWorkspaceEnabledNum++
		}
	}

	return devWorkspaceEnabledNum == 1, nil
}

func readK8SObject(yamlFile string, obj interface{}) (*Object2Sync, error) {
	_, exists := cachedObj[yamlFile]
	if !exists {
		if err := util.ReadObject(yamlFile, obj); err != nil {
			return nil, err
		}

		hash256, err := util.ComputeHash256(yamlFile)
		if err != nil {
			return nil, err
		}

		cachedObj[yamlFile] = &Object2Sync{
			obj.(metav1.Object),
			hash256,
		}
	}

	return cachedObj[yamlFile], nil
}

func devWorkspaceTemplatesPath() string {
	if util.IsOpenShift {
		return OpenshiftDevWorkspaceTemplatesPath
	}
	return KubernetesDevWorkspaceTemplatesPath
}

func devWorkspaceCheTemplatesPath() string {
	if util.IsOpenShift {
		return OpenshiftDevWorkspaceCheTemplatesPath
	}
	return KubernetesDevWorkspaceCheTemplatesPath
}
