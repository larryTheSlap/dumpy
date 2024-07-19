package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type Target interface {
	ShowDetails()
	GetNodeName() string
	GetName() string
	GetEnvVar() []corev1.EnvVar
	GetManifestLabels() map[string]string
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

func (t T_pod) ShowDetails() {
	fmt.Printf("\n  PodName: %s\n  ContainerName: %s\n  ContainerID: %s\n  NodeName: %s\n\n",
		t.Name,
		t.ContainerName,
		t.ContainerID,
		t.NodeName,
	)
}

func (t T_pod) GetNodeName() string {
	return t.NodeName
}

func (t T_pod) GetName() string {
	return t.Name
}

func (t T_pod) GetEnvVar() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "TARGET_CONTAINERID",
			Value: t.ContainerID,
		},
		{
			Name:  "TARGET_POD",
			Value: t.Name,
		},
	}
}

func (t T_pod) GetManifestLabels() map[string]string {
	return map[string]string{
		"dumpy-target-pod":       t.Name,
		"dumpy-target-container": t.ContainerName,
		"dumpy-target-namespace": t.Namespace,
	}
}

type T_node struct {
	Name string
}

func NewT_node(r_name string, api *ApiSettings) (*T_node, error) {
	if err := api.Check_NodeExist(r_name); err != nil {
		return &T_node{}, err
	}
	return &T_node{
		Name: r_name,
	}, nil
}

func (n T_node) GetEnvVar() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "TARGET_NODE",
			Value: n.Name,
		},
	}
}

func (n T_node) ShowDetails() {
	fmt.Printf("\n  NodeName: %s\n", n.Name)
}

func (n T_node) GetNodeName() string {
	return n.Name
}

func (n T_node) GetName() string {
	return n.Name
}

func (n T_node) GetManifestLabels() map[string]string {
	return map[string]string{"dumpy-target-node": n.Name}
}

type T_resource struct {
	Name          string
	Namespace     string
	ContainerName string
	Type          string
	Items         []Target
}

func (r *T_resource) SetT_Items(api *ApiSettings) error {
	switch r.Type {
	case "pod":
		t_pod, err := NewT_Pod(r.Name, r.Namespace, r.ContainerName, api)
		if err != nil {
			return err
		}
		r.Items = []Target{t_pod}
	case "node":
		if r.Name == "all" {
			t_nodeList, err := api.Get_Nodes()
			if err != nil {
				return err
			}
			for _, node := range t_nodeList {
				t_node, err := NewT_node(node.Name, api)
				if err != nil {
					return err
				}
				r.Items = append(r.Items, t_node)
			}
		} else {
			t_node, err := NewT_node(r.Name, api)
			if err != nil {
				return err
			}
			r.Items = []Target{t_node}
		}

	default:
		t_labels, err := api.Get_matchLabels(r.Type, r.Name, r.Namespace)
		if err != nil || len(t_labels.MatchLabels) == 0 {
			return fmt.Errorf("target resource pods not found in namespace %s", r.Namespace)
		}
		r.Items, err = r.Get_targetPodsFromLabels(t_labels, api)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *T_resource) Get_targetPodsFromLabels(t_labels metav1.LabelSelector, api *ApiSettings) (t_pods []Target, err error) {
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
		t_pods = append(t_pods, t_pod)
	}
	return t_pods, nil
}
