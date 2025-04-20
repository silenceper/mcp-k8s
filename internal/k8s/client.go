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

// Client 封装了 Kubernetes 客户端功能
type Client struct {
	// 标准 clientset
	clientset *kubernetes.Clientset
	// 动态客户端
	dynamicClient dynamic.Interface
	// 发现客户端
	discoveryClient *discovery.DiscoveryClient
	// REST 配置
	restConfig *rest.Config
	// kubeconfig 路径
	kubeconfigPath string
}

// NewClient 创建一个新的 Kubernetes 客户端
func NewClient(kubeconfigPath string) (*Client, error) {
	var kubeconfig string
	var config *rest.Config
	var err error

	// 如果提供了 kubeconfig 路径，使用它
	if kubeconfigPath != "" {
		kubeconfig = kubeconfigPath
	} else if home := homedir.HomeDir(); home != "" {
		// 否则尝试使用默认路径
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// 使用提供的 kubeconfig 或尝试集群内配置
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes 配置失败: %w", err)
	}

	// 创建 clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes 客户端失败: %w", err)
	}

	// 创建动态客户端
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建动态客户端失败: %w", err)
	}

	// 创建发现客户端
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建发现客户端失败: %w", err)
	}

	return &Client{
		clientset:       clientset,
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		restConfig:      config,
		kubeconfigPath:  kubeconfig,
	}, nil
}

// GetAPIResources 获取集群中的所有 API 资源类型
func (c *Client) GetAPIResources(ctx context.Context, includeNamespaceScoped, includeClusterScoped bool) ([]map[string]interface{}, error) {
	// 获取集群中所有 API Groups 和 Resources
	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil {
		// 处理部分错误，有些资源可能无法访问
		if !discovery.IsGroupDiscoveryFailedError(err) {
			return nil, fmt.Errorf("获取 API 资源失败: %w", err)
		}
	}

	var resources []map[string]interface{}

	// 处理每个API组中的资源
	for _, resourceList := range resourceLists {
		groupVersion := resourceList.GroupVersion
		for _, resource := range resourceList.APIResources {
			// 忽略子资源
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

			// 根据命名空间范围过滤
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

// GetResource 获取特定资源的详细信息
func (c *Client) GetResource(ctx context.Context, kind, name, namespace string) (map[string]interface{}, error) {
	// 获取资源的 GVR
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
		return nil, fmt.Errorf("获取资源失败: %w", err)
	}

	return obj.UnstructuredContent(), nil
}

// ListResources 列出某类资源的所有实例
func (c *Client) ListResources(ctx context.Context, kind, namespace string, labelSelector, fieldSelector string) ([]map[string]interface{}, error) {
	// 获取资源的 GVR
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
		return nil, fmt.Errorf("列出资源失败: %w", err)
	}

	var resources []map[string]interface{}
	for _, item := range list.Items {
		resources = append(resources, item.UnstructuredContent())
	}

	return resources, nil
}

// CreateResource 创建一个新的资源
func (c *Client) CreateResource(ctx context.Context, kind, namespace string, manifest string) (map[string]interface{}, error) {
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal([]byte(manifest), &obj.Object); err != nil {
		return nil, fmt.Errorf("解析资源清单失败: %w", err)
	}

	// 获取资源的 GVR
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
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}

	return result.UnstructuredContent(), nil
}

// UpdateResource 更新现有资源
func (c *Client) UpdateResource(ctx context.Context, kind, name, namespace string, manifest string) (map[string]interface{}, error) {
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal([]byte(manifest), &obj.Object); err != nil {
		return nil, fmt.Errorf("解析资源清单失败: %w", err)
	}

	// 检查名称是否匹配
	if obj.GetName() != name {
		return nil, fmt.Errorf("资源清单中的名称 (%s) 与请求的名称 (%s) 不匹配", obj.GetName(), name)
	}

	// 获取资源的 GVR
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
		return nil, fmt.Errorf("更新资源失败: %w", err)
	}

	return result.UnstructuredContent(), nil
}

// DeleteResource 删除资源
func (c *Client) DeleteResource(ctx context.Context, kind, name, namespace string) error {
	// 获取资源的 GVR
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
		return fmt.Errorf("删除资源失败: %w", deleteErr)
	}

	return nil
}

// findGroupVersionResource 根据 Kind 查找对应的 GroupVersionResource
func (c *Client) findGroupVersionResource(kind string) (*schema.GroupVersionResource, error) {
	// 获取集群中所有 API Groups 和 Resources
	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil {
		// 处理部分错误，有些资源可能无法访问
		if !discovery.IsGroupDiscoveryFailedError(err) {
			return nil, fmt.Errorf("获取 API 资源失败: %w", err)
		}
	}

	// 遍历所有 API 组和资源，查找指定的 Kind
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

	return nil, fmt.Errorf("找不到资源类型 %s", kind)
}

// GetKubeconfigPath 返回客户端使用的 kubeconfig 路径
func (c *Client) GetKubeconfigPath() string {
	return c.kubeconfigPath
}
