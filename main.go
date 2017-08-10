package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/michelaquino/golang_kubernetes_example/orchestrator"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	namespace   = apiv1.NamespaceDefault
	deployName  = "example-michel"
	appName     = "web-nginx"
	appPort     = 8080
	jobBaseName = "job-hello-world"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	operation := flag.String("operation", "", "Operation to perform (create, update, list, delete)")

	flag.Parse()
	if *kubeconfig == "" {
		panic("-kubeconfig not specified")
	}

	kubernetesClientSet := getKubernetesClient(*kubeconfig)

	deploymentOrchestrator := orchestrator.NewDeploymentOrchestrator(kubernetesClientSet)
	jobOrchestrator := orchestrator.NewJobOrchestrator(kubernetesClientSet)
	serviceOrchestrator := orchestrator.NewServiceOrchestrator(kubernetesClientSet)
	podOrchestrator := orchestrator.NewPodOrchestrator(kubernetesClientSet)

	deployName := "deployment-example"

	appName := "app-example"
	appPort := 8080

	serviceName := "service-example"

	switch *operation {
	case "create":
		deploymentOrchestrator.Create(deployName, appName, appPort)
	case "list":
		deploymentOrchestrator.List()
	case "delete":
		deploymentOrchestrator.Delete(deployName)
	case "create-service":
		serviceOrchestrator.Create(serviceName, appName, appPort)
	case "delete-service":
		serviceOrchestrator.Delete(serviceName)
	case "list-service":
		serviceOrchestrator.List()
	case "create-job":
		jobOrchestrator.Create()
	case "get-jobs":
		jobOrchestrator.List()
	case "get-pods":
		podOrchestrator.List()
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
