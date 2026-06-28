package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var k8sClient *kubernetes.Clientset

func initK8sClient() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to get in-cluster config: %w", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}
	k8sClient = client
	return nil
}

// getDaemonPodIP finds the node-daemon pod running on the given node
// and returns its pod IP. Called fresh on every tool invocation —
// no caching, so pod restarts with new IPs are handled automatically.
func getDaemonPodIP(ctx context.Context, nodeName string) (string, error) {
	pods, err := k8sClient.CoreV1().Pods("nexus-system").List(ctx, metav1.ListOptions{
		LabelSelector: "app=nexus-node-daemon",
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list daemon pods on node %q: %w", nodeName, err)
	}
	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no nexus-node-daemon pod found on node %q", nodeName)
	}
	ip := pods.Items[0].Status.PodIP
	if ip == "" {
		return "", fmt.Errorf("nexus-node-daemon pod on node %q has no IP yet", nodeName)
	}
	return ip, nil
}
