package main

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 解析config
	config, err := clientcmd.BuildConfigFromFlags("", "admin.conf")
	if err != nil {
		panic(err.Error())
	}

	// 创建连接
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 通过 watch 接口直接进行监听
	go startWatchPod(clientset, "kube-system")
	go startWatchDeployment(clientset, "kube-system")

	// 通过 client-go 的 informer 进行监听
	stopCh := make(chan struct{})
	watchAllEvents(clientset, stopCh)
	<-stopCh
}

// 监听Event变化
func watchAllEvents(client *kubernetes.Clientset, stopChannel <-chan struct{}) {
	listWatcher := cache.NewListWatchFromClient(client.Core().RESTClient(), "events", v1.NamespaceAll, fields.Everything())
	_, controller := cache.NewInformer(
		listWatcher,
		&v1.Event{},
		time.Second*5,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				fmt.Println(event.Name + "//" + event.Reason)
			},
		},
	)
	go controller.Run(stopChannel)
}

// 监听Pod变化
func startWatchPod(client *kubernetes.Clientset, ns string) {
	w, _ := client.CoreV1().Pods(ns).Watch(metav1.ListOptions{})
	for {
		select {
		case e, _ := <-w.ResultChan():
			meta, _ := meta.Accessor(e.Object)
			fmt.Println(e.Type, " Pod:", meta.GetName(), meta.GetNamespace())
		}
	}
}

// 监听Deployment变化
func startWatchDeployment(client *kubernetes.Clientset, ns string) {
	w, error := client.AppsV1beta1().Deployments(ns).Watch(metav1.ListOptions{})
	if error != nil {
		fmt.Println(error.Error())
	}
	for {
		select {
		case e, _ := <-w.ResultChan():
			meta, _ := meta.Accessor(e.Object)
			fmt.Println(e.Type, " Deploy:", meta.GetName(), meta.GetNamespace())
		}
	}
}
