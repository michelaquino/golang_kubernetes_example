package orchestrator

import (
	"fmt"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DeploymentOrchestrator struct {
	KubernetesClientSet *kubernetes.Clientset
}

func NewDeploymentOrchestrator(kubernetesClientSet *kubernetes.Clientset) *DeploymentOrchestrator {
	return &DeploymentOrchestrator{
		KubernetesClientSet: kubernetesClientSet,
	}
}

func (d DeploymentOrchestrator) Create(deployName, appName string, appPort int) {
	var replicas int32 = 1

	deployment := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployName,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: &replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  appName,
							Image: "nginx:1.13",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: int32(appPort),
								},
							},
						},
					},
				},
			},
		},
	}

	deploymentsClient := d.KubernetesClientSet.AppsV1beta1().Deployments(apiv1.NamespaceDefault)

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}

func (d DeploymentOrchestrator) Delete(deployName string) {
	deploymentsClient := d.KubernetesClientSet.AppsV1beta1().Deployments(apiv1.NamespaceDefault)

	fmt.Println("Deleting deployment...")

	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(deployName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println("Error on delete deployment. Error: ", err.Error())
	}

	fmt.Println("Deleted deployment.")
}

func (d DeploymentOrchestrator) List() {
	deploymentsClient := d.KubernetesClientSet.AppsV1beta1().Deployments(apiv1.NamespaceDefault)

	fmt.Printf("Listing deployments in namespace %q:\n", apiv1.NamespaceDefault)
	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}
}
