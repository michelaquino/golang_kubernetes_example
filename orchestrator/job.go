package orchestrator

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/oklog/ulid"
	apiBatchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type JobOrchestrator struct {
	KubernetesClientSet *kubernetes.Clientset
}

func NewJobOrchestrator(kubernetesClientSet *kubernetes.Clientset) *JobOrchestrator {
	return &JobOrchestrator{
		KubernetesClientSet: kubernetesClientSet,
	}
}

func (j JobOrchestrator) Create() error {
	ulid := ulid.MustNew(ulid.Now(), rand.Reader)
	jobName := strings.ToLower(fmt.Sprintf("%s-%s", ulid, "job-example"))

	job := &apiBatchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: apiBatchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "job-demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    "docker-hello-world-ubuntu",
							Image:   "ubuntu:latest",
							Command: []string{"echo", "Hello World!"},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

	jobInterface := j.KubernetesClientSet.Jobs(apiv1.NamespaceDefault)
	jobCreated, err := jobInterface.Create(job)
	if err != nil {
		fmt.Printf("Error on create %s Job. Error: %s", jobName, err.Error())
		return err
	}

	fmt.Printf("Job %s created with success\n", jobCreated.Name)
	jobOutput, err := j.getJobOutput(jobCreated.Name)
	if err != nil {
		fmt.Println("Error on get logs: ", err.Error())
		return err
	}

	fmt.Println("Job output: \n", jobOutput)
	return nil
}

func (j JobOrchestrator) List() {
	jobList, err := j.KubernetesClientSet.Jobs(apiv1.NamespaceDefault).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error on get Jobs")
		return
	}

	for _, job := range jobList.Items {
		fmt.Println("Job: ", job.Name)
	}
}

func (j JobOrchestrator) getJobOutput(jobName string) (string, error) {
	jobInterface := j.KubernetesClientSet.Jobs(apiv1.NamespaceDefault)
	watch, err := jobInterface.Watch(metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})

	if err != nil {
		fmt.Println("Error on watch Jobs")
		return "", err
	}

	for result := range watch.ResultChan() {
		fmt.Println("result.Type: ", result.Type)

		jobWatched, parsed := result.Object.(*apiBatchv1.Job)
		if !parsed {
			fmt.Println("Error on parse object")
			continue
		}

		statusJSON, _ := json.Marshal(jobWatched.Status)
		fmt.Println("statusJSON: ", string(statusJSON))

		if jobWatched.Status.CompletionTime == nil {
			fmt.Println("Job not finished yet!")
			continue
		}

		if jobWatched.Status.Succeeded == 1 {
			fmt.Println("Job succeeded!")
		} else {
			fmt.Println("Job failed!")
		}

		watch.Stop()
	}

	podOutput, err := j.getPodOutput(jobName)
	return podOutput, err
}

func (j JobOrchestrator) getPodOutput(jobName string) (string, error) {
	podInterface := j.KubernetesClientSet.Pods(apiv1.NamespaceDefault)
	podList, err := podInterface.List(metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})

	if err != nil {
		fmt.Println("Error on get pods")
		return "", err
	}

	if len(podList.Items) <= 0 {
		return "", errors.New("Pod not found")
	}

	pod := podList.Items[0]
	logRequest := podInterface.GetLogs(pod.Name, &apiv1.PodLogOptions{})

	logByteArray, err := logRequest.DoRaw()
	if err != nil {
		return "", err
	}

	return string(logByteArray), nil
}
