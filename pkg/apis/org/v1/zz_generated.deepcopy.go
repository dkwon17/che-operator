// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1

import (
	labels "k8s.io/apimachinery/pkg/labels"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheCluster) DeepCopyInto(out *CheCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheCluster.
func (in *CheCluster) DeepCopy() *CheCluster {
	if in == nil {
		return nil
	}
	out := new(CheCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CheCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterList) DeepCopyInto(out *CheClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CheCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterList.
func (in *CheClusterList) DeepCopy() *CheClusterList {
	if in == nil {
		return nil
	}
	out := new(CheClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CheClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpec) DeepCopyInto(out *CheClusterSpec) {
	*out = *in
	in.Server.DeepCopyInto(&out.Server)
	out.Database = in.Database
	in.Auth.DeepCopyInto(&out.Auth)
	out.Storage = in.Storage
	out.Metrics = in.Metrics
	out.K8s = in.K8s
	out.ImagePuller = in.ImagePuller
	out.DevWorkspace = in.DevWorkspace
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpec.
func (in *CheClusterSpec) DeepCopy() *CheClusterSpec {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecAuth) DeepCopyInto(out *CheClusterSpecAuth) {
	*out = *in
	if in.InitialOpenShiftOAuthUser != nil {
		in, out := &in.InitialOpenShiftOAuthUser, &out.InitialOpenShiftOAuthUser
		*out = new(bool)
		**out = **in
	}
	if in.OpenShiftoAuth != nil {
		in, out := &in.OpenShiftoAuth, &out.OpenShiftoAuth
		*out = new(bool)
		**out = **in
	}
	in.IdentityProviderIngress.DeepCopyInto(&out.IdentityProviderIngress)
	in.IdentityProviderRoute.DeepCopyInto(&out.IdentityProviderRoute)
	out.IdentityProviderContainerResources = in.IdentityProviderContainerResources
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecAuth.
func (in *CheClusterSpecAuth) DeepCopy() *CheClusterSpecAuth {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecDB) DeepCopyInto(out *CheClusterSpecDB) {
	*out = *in
	out.ChePostgresContainerResources = in.ChePostgresContainerResources
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecDB.
func (in *CheClusterSpecDB) DeepCopy() *CheClusterSpecDB {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecDB)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecDevWorkspace) DeepCopyInto(out *CheClusterSpecDevWorkspace) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecDevWorkspace.
func (in *CheClusterSpecDevWorkspace) DeepCopy() *CheClusterSpecDevWorkspace {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecDevWorkspace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecImagePuller) DeepCopyInto(out *CheClusterSpecImagePuller) {
	*out = *in
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecImagePuller.
func (in *CheClusterSpecImagePuller) DeepCopy() *CheClusterSpecImagePuller {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecImagePuller)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecK8SOnly) DeepCopyInto(out *CheClusterSpecK8SOnly) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecK8SOnly.
func (in *CheClusterSpecK8SOnly) DeepCopy() *CheClusterSpecK8SOnly {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecK8SOnly)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecMetrics) DeepCopyInto(out *CheClusterSpecMetrics) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecMetrics.
func (in *CheClusterSpecMetrics) DeepCopy() *CheClusterSpecMetrics {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecMetrics)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecServer) DeepCopyInto(out *CheClusterSpecServer) {
	*out = *in
	in.DashboardIngress.DeepCopyInto(&out.DashboardIngress)
	in.DashboardRoute.DeepCopyInto(&out.DashboardRoute)
	in.DevfileRegistryIngress.DeepCopyInto(&out.DevfileRegistryIngress)
	in.DevfileRegistryRoute.DeepCopyInto(&out.DevfileRegistryRoute)
	if in.ExternalDevfileRegistries != nil {
		in, out := &in.ExternalDevfileRegistries, &out.ExternalDevfileRegistries
		*out = make([]ExternalDevfileRegistries, len(*in))
		copy(*out, *in)
	}
	in.PluginRegistryIngress.DeepCopyInto(&out.PluginRegistryIngress)
	in.PluginRegistryRoute.DeepCopyInto(&out.PluginRegistryRoute)
	if in.CustomCheProperties != nil {
		in, out := &in.CustomCheProperties, &out.CustomCheProperties
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.SingleHostGatewayConfigMapLabels != nil {
		in, out := &in.SingleHostGatewayConfigMapLabels, &out.SingleHostGatewayConfigMapLabels
		*out = make(labels.Set, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.CheServerIngress.DeepCopyInto(&out.CheServerIngress)
	in.CheServerRoute.DeepCopyInto(&out.CheServerRoute)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecServer.
func (in *CheClusterSpecServer) DeepCopy() *CheClusterSpecServer {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecStorage) DeepCopyInto(out *CheClusterSpecStorage) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecStorage.
func (in *CheClusterSpecStorage) DeepCopy() *CheClusterSpecStorage {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterStatus) DeepCopyInto(out *CheClusterStatus) {
	*out = *in
	out.DevworkspaceStatus = in.DevworkspaceStatus
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterStatus.
func (in *CheClusterStatus) DeepCopy() *CheClusterStatus {
	if in == nil {
		return nil
	}
	out := new(CheClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalDevfileRegistries) DeepCopyInto(out *ExternalDevfileRegistries) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalDevfileRegistries.
func (in *ExternalDevfileRegistries) DeepCopy() *ExternalDevfileRegistries {
	if in == nil {
		return nil
	}
	out := new(ExternalDevfileRegistries)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IngressCustomSettings) DeepCopyInto(out *IngressCustomSettings) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IngressCustomSettings.
func (in *IngressCustomSettings) DeepCopy() *IngressCustomSettings {
	if in == nil {
		return nil
	}
	out := new(IngressCustomSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Resources) DeepCopyInto(out *Resources) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Resources.
func (in *Resources) DeepCopy() *Resources {
	if in == nil {
		return nil
	}
	out := new(Resources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourcesCustomSettings) DeepCopyInto(out *ResourcesCustomSettings) {
	*out = *in
	out.Requests = in.Requests
	out.Limits = in.Limits
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourcesCustomSettings.
func (in *ResourcesCustomSettings) DeepCopy() *ResourcesCustomSettings {
	if in == nil {
		return nil
	}
	out := new(ResourcesCustomSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteCustomSettings) DeepCopyInto(out *RouteCustomSettings) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteCustomSettings.
func (in *RouteCustomSettings) DeepCopy() *RouteCustomSettings {
	if in == nil {
		return nil
	}
	out := new(RouteCustomSettings)
	in.DeepCopyInto(out)
	return out
}
