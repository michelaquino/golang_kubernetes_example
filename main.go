package main

import (
	"flag"
	"fmt"
	"os"

	appsv1beta1 "k8s.io/api/apps/v1beta1"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedAppsV1beta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	deployName = "example-michel"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	operation := flag.String("operation", "", "Operation to perform (create, update, list, delete)")

	flag.Parse()
	if *kubeconfig == "" {
		panic("-kubeconfig not specified")
	}

	deploymentsClient := getKubernetesClient(*kubeconfig)

	switch *operation {
	case "create":
		createDeployment(deploymentsClient)
	case "update":
	case "list":
	case "delete":
		deleteDeployment(deploymentsClient)
	default:
		fmt.Println("Invalid operation. Must be: create | update | list | delete")
		os.Exit(1)
	}
}

func getKubernetesClient(kubeconfig string) typedAppsV1beta1.DeploymentInterface {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	kubernetesClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentsClient := kubernetesClientSet.AppsV1beta1().Deployments(apiv1.NamespaceDefault)
	return deploymentsClient
}

func createDeployment(deploymentsClient typedAppsV1beta1.DeploymentInterface) {
	deployment := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployName,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.13",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}

func deleteDeployment(deploymentsClient typedAppsV1beta1.DeploymentInterface) {
	fmt.Println("Deleting deployment...")

	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(deployName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println("Error on delete deployment. Error: ", err.Error())
	}

	fmt.Println("Deleted deployment.")
}

func int32Ptr(i int32) *int32 { return &i }
