package k8s

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type D_pod struct {
	Name        string
	Namespace   string
	TargetPod   *T_pod
	CaptureName string
	DumpFilters string
	Status      string
	PvcName     string
	PullSecret  string
	Image       string
}

func NewD_Pod(n, c string, t *T_pod, f, v, s, i string) *D_pod {
	return &D_pod{
		Name:        n,
		TargetPod:   t,
		CaptureName: c,
		DumpFilters: f,
		Status:      "Init",
		PvcName:     v,
		PullSecret:  s,
		Image:       i,
	}
}

func (d D_pod) Run(api *ApiSettings, t *T_resource) error {

	podManifest := d.GeneratePodManifest(t)

	fmt.Printf("\n  PodName: %s\n  ContainerName: %s\n  ContainerID: %s\n  NodeName: %s\n\n",
		d.TargetPod.Name,
		d.TargetPod.ContainerName,
		d.TargetPod.ContainerID,
		d.TargetPod.NodeName,
	)

	_, err := api.Clientset.CoreV1().Pods(d.Namespace).Create(context.Background(), &podManifest, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	if err := d.Is_Ready(api); err != nil {
		return err
	}
	return nil
}

func (d *D_pod) Is_Ready(api *ApiSettings) error {
	retry_count := 20
	for c := 0; c < retry_count; c++ {
		p, err := api.Clientset.CoreV1().Pods(d.Namespace).Get(context.Background(), d.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if len(p.Status.ContainerStatuses) != 0 && p.Status.ContainerStatuses[0].Ready {
			fmt.Printf("%s started sniffing\n", d.Name)
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return errors.New(fmt.Sprintf("Retry count exceeded, %s not ready, use (kubectl describe pod %s -n %s) to troubleshoot", d.Name, d.Name, d.Namespace))
}

func (d D_pod) Is_Completed(api *ApiSettings) error {
	retry_count := 20
	for c := 0; c < retry_count; c++ {
		p, err := api.Clientset.CoreV1().Pods(d.Namespace).Get(context.Background(), d.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if p.Status.Phase != "Running" {
			fmt.Printf("%s stopped\n", d.Name)
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return errors.New(fmt.Sprintf("Retry count exceeded, %s could not be stopped", d.Name))
}
func (d D_pod) GeneratePodManifest(t *T_resource) corev1.Pod {
	var vol []corev1.Volume
	var secret []corev1.LocalObjectReference
	if d.PvcName != "" {
		vol = []corev1.Volume{
			{
				Name: "dumpy-tmp-vol",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: d.PvcName,
					},
				},
			},
		}
	} else {
		vol = []corev1.Volume{
			{
				Name: "dumpy-tmp-vol",
			},
		}
	}

	if d.PullSecret != "" {
		secret = []corev1.LocalObjectReference{{Name: d.PullSecret}}
	}

	typeMetadata := metav1.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
	objectMetadata := metav1.ObjectMeta{
		Name:      d.Name,
		Namespace: d.Namespace,
		Labels: map[string]string{
			"app":                    "dumpy-kubectl-plugin",
			"component":              "dumpy-sniffer",
			"dumpy-target-pod":       d.TargetPod.Name,
			"dumpy-target-container": d.TargetPod.ContainerName,
			"dumpy-target-namespace": d.TargetPod.Namespace,
			"dumpy-capture":          d.CaptureName,
			"dumpy-target-resource":  t.Name,
			"dumpy-target-type":      t.Type,
		},
	}

	privileged_flag := new(bool)
	pathType := new(corev1.HostPathType)

	*privileged_flag = true
	*pathType = corev1.HostPathType("Directory")

	podSpec := corev1.PodSpec{
		RestartPolicy:    "Never",
		NodeName:         d.TargetPod.NodeName,
		HostPID:          true,
		Volumes:          vol,
		ImagePullSecrets: secret,
		Containers: []corev1.Container{{
			Image:           d.Image,
			Name:            "dumpy-container",
			SecurityContext: &corev1.SecurityContext{Privileged: privileged_flag},
			Env: []corev1.EnvVar{
				{
					Name:  "TARGET_CONTAINERID",
					Value: d.TargetPod.ContainerID,
				},
				{
					Name:  "TARGET_POD",
					Value: d.TargetPod.Name,
				},
				{
					Name:  "CAPTURE_NAME",
					Value: d.CaptureName,
				},
			},
			Command: []string{"./dumpy_sniff.sh", d.DumpFilters},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "dumpy-tmp-vol",
					MountPath: "/tmp/dumpy",
				},
			},
			StartupProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{
						Command: []string{"cat", "/tmp/dumpy/healthy"},
					},
				},
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"cpu":    resource.MustParse("100m"),
					"memory": resource.MustParse("64Mi"),
				},
				Limits: corev1.ResourceList{
					"cpu":    resource.MustParse("250m"),
					"memory": resource.MustParse("128Mi"),
				},
			},
		}},
	}
	return corev1.Pod{
		TypeMeta:   typeMetadata,
		ObjectMeta: objectMetadata,
		Spec:       podSpec,
	}
}
