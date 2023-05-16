package solver

import (
	"fmt"
	"strconv"

	dwo "github.com/devfile/devworkspace-operator/apis/controller/v1alpha1"
)

type EndpointStrategy interface {
	// get prefix paths for traefik config
	getPublicURLPrefix(port int32, uniqueEndpointName string) string
	getMainWorkspacePathPrefix() string
	getEndpointPath(e *dwo.Endpoint, componentName string) (routeName string, path string)
	getEndpointPathPrefix(endpointPath string) string

	// get host name for routes / ingress
	getHostname(endpointInfo *EndpointInfo, baseDomain string) string
}

// Main workspace URL is exposed on the following path:
// <CHE_DOMAIN>/<USERNAME>/<WORKSPACE_NAME>/<PORT>/

// Public endpoints defined in the devfile are exposed on the following path via route or ingress:
// <USERNAME>.<WORKSPACE_NAME>.<ENDPOINT_NAME>.<CLUSTER_INGRESS_DOMAIN>/<ENDPOINT_PATH>
type UsernameWkspName struct {
	username      string
	workspaceName string
}

func (u UsernameWkspName) getPublicURLPrefix(port int32, uniqueEndpointName string) string {
	if uniqueEndpointName == "" {
		return fmt.Sprintf(endpointURLPrefixPattern, u.username, u.workspaceName, port)
	}
	return fmt.Sprintf(uniqueEndpointURLPrefixPattern, u.username, u.workspaceName, uniqueEndpointName)
}

func (u UsernameWkspName) getMainWorkspacePathPrefix() string {
	return fmt.Sprintf("/%s/%s", u.username, u.workspaceName)
}

func (u UsernameWkspName) getEndpointPath(e *dwo.Endpoint, componentName string) (routeName string, path string) {
	if e.Attributes.GetString(uniqueEndpointAttributeName, nil) == "true" {
		// if endpoint is unique, we're exposing on /<endpoint-name>
		routeName = e.Name
	} else {
		// if endpoint is NOT unique, we're exposing on /<port-number>
		routeName = strconv.Itoa(e.TargetPort)
	}
	path = fmt.Sprintf("/%s", routeName)

	return routeName, path
}

func (u UsernameWkspName) getEndpointPathPrefix(endpointPath string) string {
	return fmt.Sprintf("/%s/%s%s", u.username, u.workspaceName, endpointPath)
}

func (u UsernameWkspName) getHostname(endpointInfo *EndpointInfo, baseDomain string) string {
	return fmt.Sprintf("%s.%s.%s.%s", u.username, u.workspaceName, endpointInfo.endpointName, baseDomain)
}

// Main workspace URL is exposed on the following path:
// <CHE_DOMAIN>/<WORKSPACE_ID>/<COMPONENT_NAME>/<PORT>/

// Public endpoints defined in the devfile are exposed on the following path via route or ingress:
// <WORKSPACE_ID>-<ENDPOINT_ORDER_NUMBER>.<CLUSTER_INGRESS_DOMAIN>/<ENDPOINT_PATH>
type Legacy struct {
	workspaceID   string
	componentName string
}

func (l Legacy) getPublicURLPrefix(port int32, uniqueEndpointName string) string {
	if uniqueEndpointName == "" {
		return fmt.Sprintf(endpointURLPrefixPattern, l.workspaceID, l.componentName, port)
	}
	return fmt.Sprintf(uniqueEndpointURLPrefixPattern, l.workspaceID, l.componentName, uniqueEndpointName)
}

func (l Legacy) getMainWorkspacePathPrefix() string {
	return fmt.Sprintf("/%s", l.workspaceID)
}

func (l Legacy) getEndpointPath(e *dwo.Endpoint, componentName string) (routeName string, path string) {
	if e.Attributes.GetString(uniqueEndpointAttributeName, nil) == "true" {
		// if endpoint is unique, we're exposing on /componentName/<endpoint-name>
		routeName = e.Name
	} else {
		// if endpoint is NOT unique, we're exposing on /componentName/<port-number>
		routeName = strconv.Itoa(e.TargetPort)
	}
	path = fmt.Sprintf("/%s/%s", componentName, routeName)

	return routeName, path
}

func (l Legacy) getEndpointPathPrefix(endpointPath string) string {
	return fmt.Sprintf("/%s%s", l.workspaceID, endpointPath)
}

func (l Legacy) getHostname(endpointInfo *EndpointInfo, baseDomain string) string {
	return fmt.Sprintf("%s-%d.%s", l.workspaceID, endpointInfo.order, baseDomain)
}
