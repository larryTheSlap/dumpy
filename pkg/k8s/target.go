package k8s

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type T_resource struct {
	Name          string
	Namespace     string
	ContainerName string
	Type          string
	TargetPods    []*T_pod
}

type T_pod struct {
	Name          string
	Namespace     string
	ContainerID   string
	ContainerName string
	NodeName      string
}

func NewT_Pod(p_name, p_namespace, p_container string, api *ApiSettings) (*T_pod, error) {
	c_name, c_id, c_nodeName, err := api.Get_PodInfo(p_name, p_namespace, p_container)
	if err != nil {
		return nil, err
	}
	return &T_pod{
		Name:          p_name,
		Namespace:     p_namespace,
		ContainerName: c_name,
		ContainerID:   c_id,
		NodeName:      c_nodeName,
	}, nil
}

func (r *T_resource) NewT_PodList(api *ApiSettings) ([]*T_pod, error) {
	if r.Type == "pod" {
		t_pod, err := NewT_Pod(r.Name, r.Namespace, r.ContainerName, api)
		if err != nil {
			return nil, err
		}
		return []*T_pod{t_pod}, nil
	}

	t_labels, err := api.Get_matchLabels(r.Type, r.Name, r.Namespace)
	if err != nil || len(t_labels.MatchLabels) == 0 {
		return nil, errors.New(fmt.Sprintf("target resource pods not found in namespace %s", r.Namespace))
	}
	return r.Get_targetPodsFromLabels(t_labels, api)
}

func (r *T_resource) Get_targetPodsFromLabels(t_labels metav1.LabelSelector, api *ApiSettings) (t_podList []*T_pod, err error) {
	list_opt := metav1.ListOptions{LabelSelector: labels.Set(t_labels.MatchLabels).String()}

	podList, err := api.Clientset.CoreV1().Pods(r.Namespace).List(context.Background(), list_opt)
	if err != nil {
		return nil, err
	}
	for _, p := range podList.Items {
		t_pod, err := NewT_Pod(p.Name, p.Namespace, r.ContainerName, api)
		if err != nil {
			return nil, err
		}
		t_podList = append(t_podList, t_pod)
	}
	return t_podList, nil
}

func GetT_Resource(captureName, captureNamespace string, api *ApiSettings) (t *T_resource, err error) {
	d_labels := map[string]string{"dumpy-capture": captureName}
	list_opt := metav1.ListOptions{LabelSelector: labels.Set(d_labels).String()}

	podList, err := api.Clientset.CoreV1().Pods(captureNamespace).List(context.Background(), list_opt)
	if err != nil {
		return &T_resource{}, err
	}
	if len(podList.Items) == 0 {
		return &T_resource{}, errors.New(fmt.Sprintf("%s sniffers not found in namespace %s", captureName, captureNamespace))
	}
	t = &T_resource{TargetPods: *&[]*T_pod{}}
	t.Name = podList.Items[0].Labels["dumpy-target-resource"]
	t.Namespace = podList.Items[0].Labels["dumpy-target-namespace"]
	t.ContainerName = podList.Items[0].Labels["dumpy-target-container"]
	t.Type = podList.Items[0].Labels["dumpy-target-type"]
	t.TargetPods, err = t.NewT_PodList(api)
	if err != nil {
		return &T_resource{}, err
	}
	return t, nil
}
