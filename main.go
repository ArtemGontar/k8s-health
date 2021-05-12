package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	podRunningCount, podSucceededCount := 0, 0
	var unhealthyPods [100]v1.Pod
	for i := 0; i < len(pods.Items); i++ {
		pod := pods.Items[i]
		if pod.Status.Phase == v1.PodRunning { // Step 1: If any pods in running status
			podRunningCount++
			fmt.Printf("Pod %s is Running\n", pod.Name)
		} else if pod.Status.Phase == v1.PodSucceeded { // Step 1: If any pods in succeeded status
			podSucceededCount++
			fmt.Printf("Pod %s is Succeeded\n", pod.Name)
		} else {
			unhealthyPods[i] = pod
		}
	}

	fmt.Printf("Pods in running: %d\n", podRunningCount)
	fmt.Printf("Pods in ready: %d\n", podSucceededCount)

	podPendingCount, podFailedCount, podUnknownCount := 0, 0, 0
	for i := 0; i < len(unhealthyPods); i++ {
		unhealthyPod := unhealthyPods[i]
		if unhealthyPod.Status.Phase == v1.PodPending { // Step 3: If any pods in pending status
			podPendingCount++
			fmt.Printf("Pod %s is Pending\n", unhealthyPod.Name)
		} else if unhealthyPod.Status.Phase == v1.PodFailed { // Step 4: If any pods in failed status
			podFailedCount++
			fmt.Printf("Pod %s is Failed\n", unhealthyPod.Name)
		} else if unhealthyPod.Status.Phase == v1.PodUnknown { // Step 5: If any pods in unknown status
			podUnknownCount++
			fmt.Printf("Pod %s is Unknown\n", unhealthyPod.Name)
		}
	}

	fmt.Printf("Pods in pending: %d\n", podPendingCount)
	fmt.Printf("Pods in failed: %d\n", podFailedCount)
	fmt.Printf("Pods in unknown: %d\n", podUnknownCount)
}
