package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type D_pod struct {
	Name        string
	Namespace   string
	Target      Target
	CaptureName string
	DumpFilters string
	Status      string
	PvcName     string
	PullSecret  string
	Image       string
}

func NewD_Pod(n, c string, t Target, f, v, s, i string) *D_pod {
	return &D_pod{
		Name:        n,
		Target:      t,
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

	d.Target.ShowDetails()

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
	return fmt.Errorf("retry count exceeded, %s not ready, use (kubectl describe pod %s -n %s) to troubleshoot", d.Name, d.Name, d.Namespace)
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
	return fmt.Errorf("retry count exceeded, %s could not be stopped", d.Name)
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

	d_labels := map[string]string{
		"app":                   "dumpy-kubectl-plugin",
		"component":             "dumpy-sniffer",
		"dumpy-capture":         d.CaptureName,
		"dumpy-target-resource": t.Name,
		"dumpy-target-type":     t.Type,
	}
	for key, value := range d.Target.GetManifestLabels() {
		d_labels[key] = value
	}
	objectMetadata := metav1.ObjectMeta{
		Name:      d.Name,
		Namespace: d.Namespace,
		Labels:    d_labels,
	}

	privileged_flag := new(bool)
	pathType := new(corev1.HostPathType)

	*privileged_flag = true
	*pathType = corev1.HostPathType("Directory")

	shared_env := []corev1.EnvVar{
		{
			Name:  "CAPTURE_NAME",
			Value: d.CaptureName,
		},
		{
			Name:  "DUMPY_TARGET_TYPE",
			Value: t.Type,
		},
	}

	podSpec := corev1.PodSpec{
		RestartPolicy:    "Never",
		NodeName:         d.Target.GetNodeName(),
		HostPID:          true,
		Volumes:          vol,
		ImagePullSecrets: secret,
		Containers: []corev1.Container{{
			Image:           d.Image,
			ImagePullPolicy: "IfNotPresent",
			Name:            "dumpy-container",
			SecurityContext: &corev1.SecurityContext{Privileged: privileged_flag},
			Env:             append(shared_env, d.Target.GetEnvVar()...),
			Command:         []string{"./dumpy_sniff.sh", d.DumpFilters},
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
