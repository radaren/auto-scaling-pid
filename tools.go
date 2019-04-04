package main
// refer to out-cluster example

import (
	"os"
	"fmt"
	"time"
	"flag"
	"bufio"
	"path/filepath"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/tools/clientcmd"
)


func int32Ptr(i int32) *int32 { return &i }

func int64Ptr(i int64) *int64 { return &i }

// api from metric server
type PodMetricsList struct {
    Kind       		string 	`json:"kind"`
    APIVersion 		string 	`json:"apiVersion"`
    Metadata   struct {
        SelfLink string 	`json:"selfLink"`
    } 						`json:"metadata"`
    Items []struct {
        Metadata struct {
            Name              string    `json:"name"`
            Namespace         string    `json:"namespace"`
            SelfLink          string    `json:"selfLink"`
            CreationTimestamp time.Time `json:"creationTimestamp"`
        } 								`json:"metadata"`
        Timestamp  time.Time 			`json:"timestamp"`
        Window     string    			`json:"window"`
        Containers []struct {
            Name  string 				`json:"name"`
            Usage struct {
                CPU    string 			`json:"cpu"`
                Memory string 			`json:"memory"`
            } 							`json:"usage"`
        } 								`json:"containers"`
    } 									`json:"items"`
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func initClientSet()*kubernetes.Clientset{
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
	return clientset
}
