package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type ApiSettings struct {
	Config    *genericclioptions.ConfigFlags
	Clientset *kubernetes.Clientset
	RestCfg   *rest.Config
}

func NewApiSettings() *ApiSettings {
	return &ApiSettings{Config: genericclioptions.NewConfigFlags(true)}
}

func (s *ApiSettings) Set_ClientSet() (err error) {
	raw := s.Config.ToRawKubeConfigLoader()

	s.RestCfg, err = raw.ClientConfig()
	if err != nil {
		return err
	}
	s.Clientset, err = kubernetes.NewForConfig(s.RestCfg)
	if err != nil {
		return err
	}
	return nil
}

func (s ApiSettings) Get_PodInfo(p_name, p_namespace, p_container string) (containerName, containerID, nodeName string, err error) {

	_pod, err := s.Clientset.CoreV1().Pods(p_namespace).Get(context.Background(), p_name, metav1.GetOptions{})
	if err != nil {
		return "", "", "", err
	}
	nodeName = _pod.Spec.NodeName
	containerName, containerID, err = s.Get_ContainerID(_pod, p_container)
	if err != nil {
		return "", "", "", err
	}
	return
}

func (s ApiSettings) Get_ContainerID(p *corev1.Pod, c_name string) (name, id string, err error) {
	if len(p.Status.ContainerStatuses) == 0 {
		return "", "", errors.New(fmt.Sprintf("target pod %s containers are down", p.Name))
	}
	if c_name == "" {
		return p.Status.ContainerStatuses[0].Name,
			strings.SplitN(p.Status.ContainerStatuses[0].ContainerID, "containerd://", 2)[1],
			nil
	}
	for _, c := range p.Status.ContainerStatuses {
		if c_name == c.Name {
			return c.Name, strings.SplitN(c.ContainerID, "containerd://", 2)[1], nil
		}
	}
	err = errors.New("specified container not found")
	return "", "", err
}
func (s ApiSettings) Exec_k8sCommand(command, p_name, p_namespace string) (string, string, error) {

	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	request := s.Clientset.CoreV1().RESTClient().
		Post().
		Namespace(p_namespace).
		Resource("pods").
		Name(p_name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: []string{"/bin/sh", "-c", command},
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     true,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(s.RestCfg, "POST", request.URL())
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})
	if err != nil {
		return "", "", fmt.Errorf("%w Failed executing command %s on %v/%v", err, command, p_namespace, p_name)
	}

	return buf.String(), errBuf.String(), nil
}

func (s ApiSettings) Get_Appv1Resource(r_t, r_n, r_ns string) (any, error) {
	switch r_t {
	case "deployment":
		return s.Clientset.AppsV1().Deployments(r_ns).Get(context.Background(), r_n, metav1.GetOptions{})
	case "daemonset":
		return s.Clientset.AppsV1().DaemonSets(r_ns).Get(context.Background(), r_n, metav1.GetOptions{})
	case "replicaset":
		return s.Clientset.AppsV1().ReplicaSets(r_ns).Get(context.Background(), r_n, metav1.GetOptions{})
	case "statefulset":
		return s.Clientset.AppsV1().StatefulSets(r_ns).Get(context.Background(), r_n, metav1.GetOptions{})
	default:
		return nil, errors.New("unknown resource type")
	}
}

func (s ApiSettings) Get_matchLabels(t, n, ns string) (metav1.LabelSelector, error) {
	k_resource, err := s.Get_Appv1Resource(t, n, ns)
	if err != nil {
		return metav1.LabelSelector{}, err
	}
	switch r := k_resource.(type) {
	case *appsv1.Deployment:
		return metav1.LabelSelector{MatchLabels: r.Spec.Selector.MatchLabels}, nil
	case *appsv1.DaemonSet:
		return metav1.LabelSelector{MatchLabels: r.Spec.Selector.MatchLabels}, nil
	case *appsv1.ReplicaSet:
		return metav1.LabelSelector{MatchLabels: r.Spec.Selector.MatchLabels}, nil
	case *appsv1.StatefulSet:
		return metav1.LabelSelector{MatchLabels: r.Spec.Selector.MatchLabels}, nil
	default:
		return metav1.LabelSelector{}, nil
	}
}

func (s ApiSettings) Get_currentNS() (ns string, err error) {
	if ns, _, err = s.Config.ToRawKubeConfigLoader().Namespace(); err != nil {
		return "", err
	}
	return ns, nil
}

func (s ApiSettings) Delete_Pods(captureLabel metav1.ListOptions, namespace string) error {
	force := int64(0)
	if err := s.Clientset.CoreV1().Pods(namespace).DeleteCollection(
		context.Background(),
		metav1.DeleteOptions{GracePeriodSeconds: &force},
		captureLabel,
	); err != nil {
		return err
	}
	fmt.Println("pods have been deleted")
	return nil
}
