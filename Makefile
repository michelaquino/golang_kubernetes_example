.PHONY: run
run: go run main.go -kubeconfig=$HOME/.kube/config -operation=create-job