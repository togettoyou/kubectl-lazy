package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type client struct {
	set *kubernetes.Clientset
}

func NewClient(kubeconfig string) *client {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	klog.LogToStderr(false)
	klog.SetOutput(ioutil.Discard)

	return &client{
		set: clientset,
	}
}

func (c *client) Namespaces(ctx context.Context) ([]string, error) {
	namespaces, err := c.set.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	list := make([]string, 0)
	for _, item := range namespaces.Items {
		list = append(list, item.Name)
	}
	return list, nil
}

func (c *client) Pods(ctx context.Context, namespace string) ([]string, error) {
	pods, err := c.set.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	list := make([]string, 0)
	for _, item := range pods.Items {
		list = append(list, item.Name)
	}
	return list, nil
}

type PodInfo struct {
	Name          string
	Namespace     string
	Priority      int32
	Node          string
	StartTime     time.Time
	Labels        map[string]string
	Annotations   map[string]string
	Status        []v1.ContainerStatus
	IP            string
	IPs           []v1.PodIP
	Containers    []v1.Container
	Conditions    []v1.PodCondition
	Volumes       []v1.Volume
	QoSClass      v1.PodQOSClass
	NodeSelectors map[string]string
	Tolerations   []v1.Toleration
}

func (c *client) Infos(ctx context.Context, namespace, podName string) (*PodInfo, error) {
	pod, err := c.set.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &PodInfo{
		Name:          pod.GetName(),
		Namespace:     pod.GetNamespace(),
		Priority:      *pod.Spec.Priority,
		Node:          pod.Spec.NodeName,
		StartTime:     pod.GetCreationTimestamp().Time,
		Labels:        pod.GetLabels(),
		Annotations:   pod.GetAnnotations(),
		Status:        pod.Status.ContainerStatuses,
		IP:            pod.Status.PodIP,
		IPs:           pod.Status.PodIPs,
		Containers:    pod.Spec.Containers,
		Conditions:    pod.Status.Conditions,
		Volumes:       pod.Spec.Volumes,
		QoSClass:      pod.Status.QOSClass,
		NodeSelectors: pod.Spec.NodeSelector,
		Tolerations:   pod.Spec.Tolerations,
	}, nil
}

type PodEvents struct {
	Name         string
	Type         string
	Reason       string
	CreationTime time.Time
	Message      string
}

func (c *client) Events(ctx context.Context, namespace, podName string) ([]PodEvents, error) {
	selector, err := fields.ParseSelector(
		fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
	)
	if err != nil {
		return nil, err
	}
	eventList, err := c.set.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	podEvents := make([]PodEvents, 0)
	for _, event := range eventList.Items {
		podEvents = append(podEvents, PodEvents{
			Name:         event.Name,
			Type:         event.Type,
			Reason:       event.Reason,
			CreationTime: event.CreationTimestamp.Time,
			Message:      event.Message,
		})
	}
	return podEvents, nil
}

func (c *client) Logs(ctx context.Context, namespace, podName string) (chan string, error) {
	logs := c.set.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		Follow: true,
	})
	readCloser, err := logs.Stream(ctx)
	if err != nil {
		return nil, err
	}

	readChan := make(chan string, 1)

	go func(ctx context.Context, readCloser io.ReadCloser, readChan chan string) {
		defer func() {
			recover()
			readCloser.Close()
			close(readChan)
		}()

		r := bufio.NewReader(readCloser)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				bytes, err := r.ReadBytes('\n')
				if err != nil {
					if err != io.EOF {
						panic(err.Error())
					}
					return
				}
				readChan <- string(bytes)
			}

		}
	}(ctx, readCloser, readChan)

	return readChan, nil
}
