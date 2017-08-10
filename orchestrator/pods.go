package orchestrator

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodOrchestrator struct {
	KubernetesClientSet *kubernetes.Clientset
}

func NewPodOrchestrator(kubernetesClientSet *kubernetes.Clientset) *PodOrchestrator {
	return &PodOrchestrator{
		KubernetesClientSet: kubernetesClientSet,
	}
}

func (p PodOrchestrator) List() {
	podInterface := p.KubernetesClientSet.Pods(apiv1.NamespaceDefault)

	podList, err := podInterface.List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error on get pods")
		return
	}

	for _, pod := range podList.Items {
		fmt.Println("pod.UID: ", pod.UID)
		fmt.Println("pod.Name: ", pod.Name)
		fmt.Println("pod.Labels: ", pod.Labels)
	}
}
