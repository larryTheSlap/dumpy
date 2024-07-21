package subcmd

import (
	"context"
	"dumpy/pkg/k8s"
	"errors"
	"fmt"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"dumpy/pkg/utils"
)

type Dumpy struct {
	Api            *k8s.ApiSettings
	Sniffers       []*k8s.D_pod
	Namespace      string
	TargetResource *k8s.T_resource
	CaptureName    string
	DumpFilters    string
	PvcName        string
	PullSecret     string
	Image          string
}

func NewDumpy() *Dumpy {
	return &Dumpy{
		Api:            k8s.NewApiSettings(),
		TargetResource: &k8s.T_resource{},
	}
}

func (d *Dumpy) NewSniffers() {
	d.Sniffers = nil
	for _, t := range d.TargetResource.Items {
		s_name := fmt.Sprintf("sniffer-%s-%s", d.CaptureName, utils.GenerateRandomID(4))
		newSniffer := k8s.NewD_Pod(s_name, d.CaptureName, t, d.DumpFilters, d.PvcName, d.PullSecret, d.Image)
		newSniffer.Namespace = d.Namespace
		d.Sniffers = append(d.Sniffers, newSniffer)
	}
}

func (d *Dumpy) NewSniffersFromExisting() error {
	d.Sniffers = nil
	d_labels := map[string]string{"dumpy-capture": d.CaptureName}
	list_opt := metav1.ListOptions{LabelSelector: labels.Set(d_labels).String()}

	podList, err := d.Api.Clientset.CoreV1().Pods(d.Namespace).List(context.Background(), list_opt)
	if err != nil {
		return err
	}
	if len(podList.Items) == 0 {
		return nil
	}

	d.Image = podList.Items[0].Spec.Containers[0].Image
	if podList.Items[0].Spec.Volumes[0].PersistentVolumeClaim == nil {
		d.PvcName = ""
	} else {
		d.PvcName = podList.Items[0].Spec.Volumes[0].PersistentVolumeClaim.ClaimName
	}

	if len(podList.Items[0].Spec.ImagePullSecrets) != 0 {
		d.PullSecret = podList.Items[0].Spec.ImagePullSecrets[0].Name
	}
	d.DumpFilters = podList.Items[0].Spec.Containers[0].Command[1]
	for _, p := range podList.Items {
		var t k8s.Target
		if p.Labels["dumpy-target-type"] == "node" {
			t = &k8s.T_node{Name: p.Labels["dumpy-target-node"]}
		} else {
			t = &k8s.T_pod{
				Name:          p.Labels["dumpy-target-pod"],
				Namespace:     p.Labels["dumpy-target-namespace"],
				ContainerName: p.Labels["dumpy-target-container"],
			}
		}
		newSniffer := k8s.NewD_Pod(p.Name, d.CaptureName, t, p.Spec.Containers[0].Command[1], d.PvcName, d.PullSecret, d.Image)
		newSniffer.Namespace = p.Namespace
		newSniffer.Status = string(p.Status.Phase)
		if newSniffer.Status == "Succeeded" {
			newSniffer.Status = "Stopped"
		}
		d.Sniffers = append(d.Sniffers, newSniffer)
	}
	return nil
}

func (d Dumpy) GetCaptures() (map[string]string, error) {
	d_labels := map[string]string{"component": "dumpy-sniffer"}
	list_opt := metav1.ListOptions{LabelSelector: labels.Set(d_labels).String()}

	podList, err := d.Api.Clientset.CoreV1().Pods(d.Namespace).List(context.Background(), list_opt)
	if err != nil {
		return nil, err
	}
	if len(podList.Items) == 0 {
		return nil, fmt.Errorf("no captures found in namespace %s", d.Namespace)
	}
	captures := make(map[string]string)
	var exist bool
	for _, p := range podList.Items {
		c, isSniffer := p.Labels["dumpy-capture"]
		if isSniffer {
			_, exist = captures[c]
			if !exist {
				captures[c] = p.Namespace
			}
		}

	}
	return captures, nil
}

func (d *Dumpy) Sniff() error {
	run_ := func(s *k8s.D_pod) error {
		return s.Run(d.Api, d.TargetResource)
	}
	return d.GenericConcurrentOperation("capture", run_)
}

func (d *Dumpy) Delete_Sniffers() error {
	del_ := func(s *k8s.D_pod) error {
		force := int64(0)
		fmt.Printf("deleting %s\n", s.Name)
		return d.Api.Clientset.CoreV1().Pods(s.Namespace).Delete(
			context.Background(),
			s.Name,
			metav1.DeleteOptions{GracePeriodSeconds: &force},
		)
	}
	return d.GenericConcurrentOperation("delete", del_)
}

func (d *Dumpy) Wait_Completed() error {
	wait_ := func(s *k8s.D_pod) error {
		return s.Is_Completed(d.Api)
	}
	return d.GenericConcurrentOperation("stop", wait_)
}

func (d *Dumpy) GenericConcurrentOperation(command string, operation func(*k8s.D_pod) error) error {
	var wg sync.WaitGroup
	errCh := make(chan error)
	doneCh := make(chan struct{})
	go func() {
		defer close(errCh)
		defer close(doneCh)
		wg.Wait()
	}()

	go AnimatePause(&wg, doneCh)

	for _, s := range d.Sniffers {
		wg.Add(1)
		go func(s *k8s.D_pod) {
			defer wg.Done()
			if err := operation(s); err != nil {
				errCh <- err
			}
		}(s)
	}
	errs := []error{}
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		fmt.Println("")
		errMsg := fmt.Sprintf("Dumpy %s operation failed, details:\n", command)
		for _, err := range errs {
			errMsg = errMsg + fmt.Sprintf("     %s\n", err.Error())
		}
		return errors.New(errMsg)
	}

	return nil
}

func AnimatePause(wg *sync.WaitGroup, doneCh <-chan struct{}) {
	animationChars := []rune{'\\', '|', '/', '-'}
	animationIndex := 0
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			fmt.Printf("%c\b", animationChars[animationIndex])
			animationIndex = (animationIndex + 1) % len(animationChars)
		case <-doneCh:
			fmt.Println("\b")
			return
		}
	}
}
