package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wraps Kubernetes client functionality
type Client struct {
	// Standard clientset
	clientset *kubernetes.Clientset
	// Dynamic client
	dynamicClient dynamic.Interface
	// Discovery client
	discoveryClient *discovery.DiscoveryClient
	// REST config
	restConfig *rest.Config
	// kubeconfig path
	kubeconfigPath string
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath string) (*Client, error) {
	var kubeconfig string
	var config *rest.Config
	var err error

	// If kubeconfig path is provided, use it
	if kubeconfigPath != "" {
		kubeconfig = kubeconfigPath
	} else if home := homedir.HomeDir(); home != "" {
		// Otherwise try to use default path
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Use provided kubeconfig or try in-cluster config
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	return &Client{
		clientset:       clientset,
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		restConfig:      config,
		kubeconfigPath:  kubeconfig,
	}, nil
}

// GetAPIResources gets all API resource types in the cluster
func (c *Client) GetAPIResources(ctx context.Context, includeNamespaceScoped, includeClusterScoped bool) ([]map[string]interface{}, error) {
	// Get all API Groups and Resources in the cluster
	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil {
		// Handle partial errors, some resources may not be accessible
		if !discovery.IsGroupDiscoveryFailedError(err) {
			return nil, fmt.Errorf("failed to get API resources: %w", err)
		}
	}

	var resources []map[string]interface{}

	// Process resources in each API group
	for _, resourceList := range resourceLists {
		groupVersion := resourceList.GroupVersion
		for _, resource := range resourceList.APIResources {
			// Ignore sub-resources
			if len(resource.Group) == 0 {
				resource.Group = resourceList.GroupVersion
			}
			if len(resource.Version) == 0 {
				gv, err := schema.ParseGroupVersion(groupVersion)
				if err != nil {
					continue
				}
				resource.Version = gv.Version
			}

			// Filter by namespace scope
			if (resource.Namespaced && !includeNamespaceScoped) || (!resource.Namespaced && !includeClusterScoped) {
				continue
			}

			resources = append(resources, map[string]interface{}{
				"name":         resource.Name,
				"singularName": resource.SingularName,
				"namespaced":   resource.Namespaced,
				"kind":         resource.Kind,
				"group":        resource.Group,
				"version":      resource.Version,
				"verbs":        resource.Verbs,
			})
		}
	}

	return resources, nil
}

// GetResource gets detailed information about a specific resource
func (c *Client) GetResource(ctx context.Context, kind, name, namespace string) (map[string]interface{}, error) {
	// Get the resource's GVR
	gvr, err := c.findGroupVersionResource(kind)
	if err != nil {
		return nil, err
	}

	var obj *unstructured.Unstructured
	if namespace != "" {
		obj, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = c.dynamicClient.Resource(*gvr).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	return obj.UnstructuredContent(), nil
}

// ListResources lists all instances of a resource type
func (c *Client) ListResources(ctx context.Context, kind, namespace string, labelSelector, fieldSelector string) ([]map[string]interface{}, error) {
	// Get the resource's GVR
	gvr, err := c.findGroupVersionResource(kind)
	if err != nil {
		return nil, err
	}

	options := metav1.ListOptions{}
	if labelSelector != "" {
		options.LabelSelector = labelSelector
	}
	if fieldSelector != "" {
		options.FieldSelector = fieldSelector
	}

	var list *unstructured.UnstructuredList
	if namespace != "" {
		list, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).List(ctx, options)
	} else {
		list, err = c.dynamicClient.Resource(*gvr).List(ctx, options)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	var resources []map[string]interface{}
	for _, item := range list.Items {
		resources = append(resources, item.UnstructuredContent())
	}

	return resources, nil
}

// CreateResource creates a new resource
func (c *Client) CreateResource(ctx context.Context, kind, namespace string, manifest string) (map[string]interface{}, error) {
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal([]byte(manifest), &obj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse resource manifest: %w", err)
	}

	// Get the resource's GVR
	gvr, err := c.findGroupVersionResource(kind)
	if err != nil {
		return nil, err
	}

	var result *unstructured.Unstructured
	if namespace != "" || obj.GetNamespace() != "" {
		targetNamespace := namespace
		if targetNamespace == "" {
			targetNamespace = obj.GetNamespace()
		}
		result, err = c.dynamicClient.Resource(*gvr).Namespace(targetNamespace).Create(ctx, obj, metav1.CreateOptions{})
	} else {
		result, err = c.dynamicClient.Resource(*gvr).Create(ctx, obj, metav1.CreateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return result.UnstructuredContent(), nil
}

// UpdateResource updates an existing resource
func (c *Client) UpdateResource(ctx context.Context, kind, name, namespace string, manifest string) (map[string]interface{}, error) {
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal([]byte(manifest), &obj.Object); err != nil {
		return nil, fmt.Errorf("failed to parse resource manifest: %w", err)
	}

	// Check if name matches
	if obj.GetName() != name {
		return nil, fmt.Errorf("name in resource manifest (%s) does not match requested name (%s)", obj.GetName(), name)
	}

	// Get the resource's GVR
	gvr, err := c.findGroupVersionResource(kind)
	if err != nil {
		return nil, err
	}

	var result *unstructured.Unstructured
	if namespace != "" {
		result, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	} else {
		result, err = c.dynamicClient.Resource(*gvr).Update(ctx, obj, metav1.UpdateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update resource: %w", err)
	}

	return result.UnstructuredContent(), nil
}

// DeleteResource deletes a resource
func (c *Client) DeleteResource(ctx context.Context, kind, name, namespace string) error {
	// Get the resource's GVR
	gvr, err := c.findGroupVersionResource(kind)
	if err != nil {
		return err
	}

	var deleteErr error
	if namespace != "" {
		deleteErr = c.dynamicClient.Resource(*gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		deleteErr = c.dynamicClient.Resource(*gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}

	if deleteErr != nil {
		return fmt.Errorf("failed to delete resource: %w", deleteErr)
	}

	return nil
}

// findGroupVersionResource finds the corresponding GroupVersionResource by Kind
func (c *Client) findGroupVersionResource(kind string) (*schema.GroupVersionResource, error) {
	// Get all API Groups and Resources in the cluster
	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil {
		// Handle partial errors, some resources may not be accessible
		if !discovery.IsGroupDiscoveryFailedError(err) {
			return nil, fmt.Errorf("failed to get API resources: %w", err)
		}
	}

	// Iterate through all API groups and resources to find the specified Kind
	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range resourceList.APIResources {
			if resource.Kind == kind {
				return &schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: resource.Name,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("resource type %s not found", kind)
}

// GetKubeconfigPath returns the kubeconfig path used by the client
func (c *Client) GetKubeconfigPath() string {
	return c.kubeconfigPath
}
