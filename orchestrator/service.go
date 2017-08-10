package orchestrator

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type ServiceOrchestrator struct {
	KubernetesClientSet *kubernetes.Clientset
}

func NewServiceOrchestrator(kubernetesClientSet *kubernetes.Clientset) *ServiceOrchestrator {
	return &ServiceOrchestrator{
		KubernetesClientSet: kubernetesClientSet,
	}
}

func (s ServiceOrchestrator) Create(serviceName, appName string, appPort int) {
	serviceSpec := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
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
	service := s.KubernetesClientSet.Core().Services(apiv1.NamespaceDefault)
	svc, err := service.Get(serviceName, metav1.GetOptions{})
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

func (s ServiceOrchestrator) Delete(serviceName string) {
	service := s.KubernetesClientSet.Core().Services(apiv1.NamespaceDefault)

	if err := service.Delete(serviceName, &metav1.DeleteOptions{}); err != nil {
		fmt.Println("Error on delete service")
		return
	}

	fmt.Println("Service deleted")
}

func (s ServiceOrchestrator) List() {
	service := s.KubernetesClientSet.Core().Services(apiv1.NamespaceDefault)

	serviceList, err := service.List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error on list services")
		return
	}

	for _, service := range serviceList.Items {
		fmt.Printf("* %s (Cluster IP: %s)\n", service.Name, service.Spec.ClusterIP)
	}
}
