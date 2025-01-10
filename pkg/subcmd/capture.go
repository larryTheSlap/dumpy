package subcmd

import (
	"dumpy/pkg/utils"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	allowed_resources = map[string]string{
		"deployment":  "deploy",
		"daemonset":   "ds",
		"replicaset":  "rs",
		"statefulset": "sts",
		"pod":         "p",
		"node":        "node",
	}
	captureExample = `# capture all traffic from foo pod in current namespace 
  kubectl dumpy capture pod foo
# capture all traffic from foo pod in foo-ns with specific capture name 
  kubectl dumpy capture pod foo -t foo-ns --name <captureName>
# capture traffic from foo pod using tcpdump filters
  kubectl dumpy capture pod foo -f "-i any host 10.0.0.1 and port 80" 
# capture traffic from foo pod specific container foo-cont
  kubectl dumpy capture pod foo -c foo-cont
# capture traffic from deployment foo-deploy in foo-ns namespace with sniffers in bar-ns
  kubectl dumpy capture deploy foo-deploy -t foo-ns -n bar-ns
# set dumpy image from private repository using docker pullSecret
  kubectl dumpy capture deploy foo-deploy -i <repository>/<path>/dumpy:latest -s <secretName>
# set pvc volume [RWX for multiple sniffers] to store tcpdump captures
  kubectl dumpy capture daemonset foo-ds -v <pvcName>
# capture traffic from node worker-node
  kubectl dumpy capture node worker-node
# capture traffic from all nodes
  kubectl dumpy capture node all
	`
)

func Dumpysubcmd_capture() (cmd *cobra.Command) {

	dumpy := NewDumpy()

	cmd = &cobra.Command{
		Use:          "capture <pod|deployment|replicaset|daemonset|statefulset> <resourceName> [-n captureNamespace] [-t targetNamespace] [-f tcpdumpFilters] [-c containerName]",
		Short:        "start tcpdump capture on target resource pods",
		Example:      captureExample,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := dumpy.Capture_Complete(cmd, args); err != nil {
				return err
			}
			if err := dumpy.Capture_Validate(cmd, args); err != nil {
				return err
			}
			if err := dumpy.Capture_Run(cmd, args); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&dumpy.TargetResource.Namespace, "target-namespace", "t", "", "target resource namespace (default: capture namespace)")
	cmd.Flags().StringVarP(&dumpy.CaptureName, "name", "", "", "dumpy capture name")
	cmd.Flags().StringVarP(&dumpy.Namespace, "namespace", "n", "", "dumpy capture sniffers namespace (default: current namespace)")
	cmd.Flags().StringVarP(&dumpy.TargetResource.ContainerName, "container", "c", "", "target resource container name (default: main pod container)")
	cmd.Flags().StringVarP(&dumpy.DumpFilters, "filter", "f", "-i any", "tcpdump filters/arguments")
	cmd.Flags().StringVarP(&dumpy.PvcName, "pvc", "v", "", "pvc name for dumpy sniffers to store network captures")
	cmd.Flags().StringVarP(&dumpy.PullSecret, "secret", "s", "", "dumpy sniffer image pull secret")
	cmd.Flags().StringVarP(&dumpy.Image, "image", "i", "larrytheslap/dumpy:latest", "dumpy sniffer docker image")

	return
}

func (d *Dumpy) Capture_Complete(cmd *cobra.Command, args []string) (err error) {
	if err := d.Api.Set_ClientSet(); err != nil {
		return err
	}
	if d.Namespace == "" {
		d.Namespace, err = d.Api.Get_currentNS()
		if err != nil {
			return err
		}
	}
	if d.TargetResource.Namespace == "" {
		d.TargetResource.Namespace = d.Namespace
	}
	if d.CaptureName == "" {
		d.CaptureName = fmt.Sprintf("dumpy-%s", utils.GenerateRandomID(8))
	}
	return nil
}

func (d *Dumpy) Capture_Validate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.Usage()
		return errors.New("no argument provided")
	}
	if len(args) < 2 {
		return errors.New("not enough arguments, use -h for help")
	}
	if len(args) > 2 {
		return errors.New("too many arguments, use -h for help")
	}
	d.TargetResource.Type = args[0]
	d.TargetResource.Name = args[1]

	var valid bool
	if d.TargetResource.Type, valid = utils.ExistIn(d.TargetResource.Type, allowed_resources); !valid {
		return errors.New("unkown resource type, use -h for help")
	}
	if err := d.NewSniffersFromExisting(); err != nil {
		return err
	}
	exSniffers := d.Sniffers
	if len(exSniffers) != 0 {
		return fmt.Errorf("%s capture already exist", d.CaptureName)
	}
	return nil
}

func (d *Dumpy) Capture_Run(cmd *cobra.Command, args []string) (err error) {

	fmt.Println("Getting target resource info..")
	if err = d.TargetResource.SetT_Items(d.Api); err != nil {
		return err
	}

	d.NewSniffers()

	fmt.Printf("Dumpy init\n\nCapture name: %s\n", d.CaptureName)
	if err := d.Sniff(); err != nil {
		return err
	}
	fmt.Println("All dumpy sniffers are Ready.")
	return
}
