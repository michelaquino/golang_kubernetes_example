package main

import (
	"flag"
	"fmt"
	"os"

	appsv1beta1 "k8s.io/api/apps/v1beta1"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	typedAppsV1beta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	namespace  = apiv1.NamespaceDefault
	deployName = "example-michel"
	appName    = "web-nginx"
	appPort    = 8080
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	operation := flag.String("operation", "", "Operation to perform (create, update, list, delete)")

	flag.Parse()
	if *kubeconfig == "" {
		panic("-kubeconfig not specified")
	}

	kubernetesClientSet := getKubernetesClient(*kubeconfig)
	deploymentsClient := kubernetesClientSet.AppsV1beta1().Deployments(namespace)

	switch *operation {
	case "create":
		createDeployment(deploymentsClient)
	case "update":
	case "list":
		listDeployments(deploymentsClient)
	case "delete":
		deleteDeployment(deploymentsClient)
	case "create-service":
		createService(kubernetesClientSet)
	case "delete-service":
		deleteService(kubernetesClientSet)
	case "list-service":
		listServices(kubernetesClientSet)
	default:
		fmt.Println("Invalid operation. Must be: create | update | list | delete | create-service | delete-service")
		os.Exit(1)
	}
}

func getKubernetesClient(kubeconfig string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	kubernetesClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return kubernetesClientSet
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
							Name:  appName,
							Image: "nginx:1.13",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: appPort,
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

func listDeployments(deploymentsClient typedAppsV1beta1.DeploymentInterface) {
	fmt.Printf("Listing deployments in namespace %q:\n", namespace)
	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}
}

func createService(clientSet *kubernetes.Clientset) {
	serviceSpec := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
		Spec: apiv1.ServiceSpec{
			Type:     apiv1.ServiceTypeClusterIP,
			Selector: map[string]string{"app": appName},
			Ports: []apiv1.ServicePort{
				apiv1.ServicePort{
					Protocol: apiv1.ProtocolTCP,
					Port:     8080,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(appPort),
					},
				},
			},
		},
	}

	// Implement service update-or-create semantics.
	service := clientSet.Core().Services(namespace)
	svc, err := service.Get(appName, metav1.GetOptions{})
	switch {
	case err == nil:
		serviceSpec.ObjectMeta.ResourceVersion = svc.ObjectMeta.ResourceVersion
		serviceSpec.Spec.ClusterIP = svc.Spec.ClusterIP

		_, err = service.Update(serviceSpec)
		if err != nil {
			fmt.Printf("failed to update service: %s", err)
			return
		}

		fmt.Println("service updated")
	case errors.IsNotFound(err):
		_, err = service.Create(serviceSpec)
		if err != nil {
			fmt.Printf("failed to create service: %s", err)
			return
		}

		fmt.Println("service created")
	default:
		fmt.Printf("unexpected error: %s", err)
	}
}

func deleteService(clientSet *kubernetes.Clientset) {
	service := clientSet.Core().Services(namespace)

	if err := service.Delete(appName, &metav1.DeleteOptions{}); err != nil {
		fmt.Println("Error on delete service")
		return
	}

	fmt.Println("Service deleted")
}

func listServices(clientSet *kubernetes.Clientset) {
	service := clientSet.Core().Services(namespace)

	serviceList, err := service.List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error on list services")
		return
	}

	for _, service := range serviceList.Items {
		fmt.Printf("* %s (Cluster IP: %s)\n", service.Name, service.Spec.ClusterIP)
	}
}

func int32Ptr(i int32) *int32 { return &i }
