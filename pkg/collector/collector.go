// Package collector connects to Kubernetes and captures resource state.
package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/raghu-007/GitOps-Time-Machine/pkg/config"
	"github.com/raghu-007/GitOps-Time-Machine/pkg/types"
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// resourceMapping maps friendly names to GVR (GroupVersionResource).
var resourceMapping = map[string]schema.GroupVersionResource{
	"deployments":            {Group: "apps", Version: "v1", Resource: "deployments"},
	"statefulsets":           {Group: "apps", Version: "v1", Resource: "statefulsets"},
	"daemonsets":             {Group: "apps", Version: "v1", Resource: "daemonsets"},
	"services":               {Group: "", Version: "v1", Resource: "services"},
	"configmaps":             {Group: "", Version: "v1", Resource: "configmaps"},
	"secrets":                {Group: "", Version: "v1", Resource: "secrets"},
	"persistentvolumeclaims": {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	"serviceaccounts":        {Group: "", Version: "v1", Resource: "serviceaccounts"},
	"ingresses":              {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	"networkpolicies":        {Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	"cronjobs":               {Group: "batch", Version: "v1", Resource: "cronjobs"},
	"roles":                  {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
	"rolebindings":           {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
	"clusterroles":           {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
	"clusterrolebindings":    {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
}

// Collector connects to a Kubernetes cluster and captures resource state.
type Collector struct {
	dynamicClient   dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
	config          *config.Config
}

// New creates a new Collector from the given configuration.
func New(cfg *config.Config) (*Collector, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = cfg.Kubeconfig

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules,
		&clientcmd.ConfigOverrides{CurrentContext: cfg.Context},
	)

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	discoClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	return &Collector{
		dynamicClient:   dynClient,
		discoveryClient: discoClient,
		config:          cfg,
	}, nil
}

// Collect captures the current state of all configured resources.
func (c *Collector) Collect(ctx context.Context) (*types.ResourceSnapshot, error) {
	snapshot := &types.ResourceSnapshot{
		Metadata: types.SnapshotMetadata{
			Timestamp:   time.Now().UTC(),
			ClusterName: c.getClusterName(),
			Context:     c.config.Context,
		},
	}

	namespacesSet := make(map[string]bool)

	for _, resType := range c.config.Snapshot.ResourceTypes {
		gvr, ok := resourceMapping[resType]
		if !ok {
			log.WithField("resource", resType).Warn("unknown resource type, skipping")
			continue
		}

		resources, err := c.collectResource(ctx, gvr)
		if err != nil {
			log.WithError(err).WithField("resource", resType).Warn("failed to collect resource")
			continue
		}

		for _, res := range resources {
			if c.shouldExcludeNamespace(res.Namespace) {
				continue
			}
			if len(c.config.Snapshot.Namespaces) > 0 && !c.shouldIncludeNamespace(res.Namespace) {
				continue
			}
			snapshot.Resources = append(snapshot.Resources, res)
			if res.Namespace != "" {
				namespacesSet[res.Namespace] = true
			}
		}

		log.WithFields(log.Fields{
			"resource": resType,
			"count":    len(resources),
		}).Debug("collected resources")
	}

	// Build namespace list
	for ns := range namespacesSet {
		snapshot.Metadata.Namespaces = append(snapshot.Metadata.Namespaces, ns)
	}
	snapshot.Metadata.ResourceCount = len(snapshot.Resources)

	log.WithFields(log.Fields{
		"totalResources": snapshot.Metadata.ResourceCount,
		"namespaces":     len(snapshot.Metadata.Namespaces),
	}).Info("snapshot collection completed")

	return snapshot, nil
}

// collectResource fetches all instances of a specific resource type.
func (c *Collector) collectResource(ctx context.Context, gvr schema.GroupVersionResource) ([]types.Resource, error) {
	var resources []types.Resource

	list, err := c.dynamicClient.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", gvr.Resource, err)
	}

	for _, item := range list.Items {
		obj := item.Object

		// Strip configured fields
		c.stripFields(obj)

		res := types.Resource{
			APIVersion: item.GetAPIVersion(),
			Kind:       item.GetKind(),
			Namespace:  item.GetNamespace(),
			Name:       item.GetName(),
			Labels:     item.GetLabels(),
			Annotations: cleanAnnotations(item.GetAnnotations()),
			Raw:        obj,
		}

		// Extract spec and data if present
		if spec, ok := obj["spec"].(map[string]interface{}); ok {
			res.Spec = spec
		}
		if data, ok := obj["data"].(map[string]interface{}); ok {
			res.Data = data
		}

		resources = append(resources, res)
	}

	return resources, nil
}

// stripFields removes configured fields from the resource object.
func (c *Collector) stripFields(obj map[string]interface{}) {
	for _, field := range c.config.Snapshot.StripFields {
		switch field {
		case ".metadata.managedFields":
			if metadata, ok := obj["metadata"].(map[string]interface{}); ok {
				delete(metadata, "managedFields")
			}
		case ".metadata.resourceVersion":
			if metadata, ok := obj["metadata"].(map[string]interface{}); ok {
				delete(metadata, "resourceVersion")
			}
		case ".metadata.uid":
			if metadata, ok := obj["metadata"].(map[string]interface{}); ok {
				delete(metadata, "uid")
			}
		case ".metadata.generation":
			if metadata, ok := obj["metadata"].(map[string]interface{}); ok {
				delete(metadata, "generation")
			}
		case ".status":
			delete(obj, "status")
		}
	}
}

// shouldExcludeNamespace checks if a namespace is in the exclusion list.
func (c *Collector) shouldExcludeNamespace(ns string) bool {
	for _, excluded := range c.config.Snapshot.ExcludeNamespaces {
		if ns == excluded {
			return true
		}
	}
	return false
}

// shouldIncludeNamespace checks if a namespace is in the inclusion list.
func (c *Collector) shouldIncludeNamespace(ns string) bool {
	for _, included := range c.config.Snapshot.Namespaces {
		if ns == included {
			return true
		}
	}
	return false
}

// cleanAnnotations removes noisy annotations from resources.
func cleanAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return nil
	}
	noisy := []string{
		"kubectl.kubernetes.io/last-applied-configuration",
		"deployment.kubernetes.io/revision",
	}
	cleaned := make(map[string]string)
	for k, v := range annotations {
		skip := false
		for _, n := range noisy {
			if k == n {
				skip = true
				break
			}
		}
		if !skip {
			cleaned[k] = v
		}
	}
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}

// getClusterName extracts the cluster name from the kubeconfig context.
func (c *Collector) getClusterName() string {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = c.config.Kubeconfig
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return "unknown"
	}
	ctx := c.config.Context
	if ctx == "" {
		ctx = rawConfig.CurrentContext
	}
	if ctxObj, ok := rawConfig.Contexts[ctx]; ok {
		return ctxObj.Cluster
	}
	return "unknown"
}
