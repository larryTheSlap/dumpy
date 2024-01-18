package subcmd

import (
	"dumpy/pkg/k8s"
	"errors"
	"fmt"

	"github.com/cheynewallace/tabby"

	"github.com/spf13/cobra"
)

var (
	getExample = `# get all captures in current namespace in table format
  kubectl dumpy get
# get specific capture mycap details in bar-ns
  kubectl dumpy get mycap -n bar-ns`
)

func Dumpysubcmd_get() (cmd *cobra.Command) {

	dumpy := NewDumpy()

	cmd = &cobra.Command{
		Use:          "get <captureName> [-n captureNamespace]",
		Short:        "get dumpy captures info in namespace",
		Example:      getExample,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := dumpy.Get_Complete(); err != nil {
				return err
			}
			if err := dumpy.Get_Validate(args, cmd); err != nil {
				return err
			}
			if err := dumpy.Get_Run(args); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&dumpy.Namespace, "namespace", "n", "", "dumpy capture sniffers namespace (default: current namespace)")
	return
}

func (d *Dumpy) Get_Complete() (err error) {
	if err := d.Api.Set_ClientSet(); err != nil {
		return err
	}
	if d.Namespace == "" {
		d.Namespace, err = d.Api.Get_currentNS()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dumpy) Get_Validate(args []string, cmd *cobra.Command) error {
	if len(args) > 1 {
		return errors.New("too many arguments, get command require capture name or nothing. use -h for help")
	}
	return nil
}

func (d *Dumpy) Get_Run(args []string) (err error) {
	if len(args) == 0 {
		captures, err := d.GetCaptures()
		if err != nil {
			return err
		}
		t := tabby.New()
		t.AddHeader("NAME", "NAMESPACE", "TARGET", "TARGETNAMESPACE", "TCPDUMPFILTERS", "SNIFFERS")
		for c, ns := range captures {
			d.CaptureName = c
			if err := d.NewSniffersFromExisting(); err != nil {
				return err
			}
			d.TargetResource, err = k8s.GetT_Resource(d.CaptureName, ns, d.Api)
			if err != nil {
				return err
			}
			r_count := 0
			for _, s := range d.Sniffers {
				if s.Status == "Running" {
					r_count++
				}
			}
			t.AddLine(d.CaptureName, d.Namespace,
				fmt.Sprintf("%s/%s", d.TargetResource.Type, d.TargetResource.Name),
				d.TargetResource.Namespace, d.DumpFilters,
				fmt.Sprintf("%v/%v", r_count, len(d.TargetResource.TargetPods)),
			)
		}
		t.Print()
		return nil
	}

	d.CaptureName = args[0]
	fmt.Print("Getting capture details..\n\n")
	if err := d.NewSniffersFromExisting(); err != nil {
		return err
	}
	if len(d.Sniffers) == 0 {
		return fmt.Errorf("%s sniffers not found in namespace %s", d.CaptureName, d.Namespace)
	}
	d.TargetResource, err = k8s.GetT_Resource(d.CaptureName, d.Namespace, d.Api)
	if err != nil {
		return err
	}
	podStr := ""
	for _, s := range d.Sniffers {
		s_status := s.Status
		if s.Status == "Succeeded" {
			s_status = "Stopped"
		}
		podStr += s.TargetPod.Name + "  <-----  " + s.Name + " [" + s_status + "]" + "\n        "
	}

	details := "name: %s\n" +
		"namespace: %s\n" +
		"tcpdumpfilters: %s\n" +
		"image: %s\n" +
		"targetSpec:\n" +
		"    name: %s\n" +
		"    namespace: %s\n" +
		"    type: %s\n" +
		"    container: %s\n" +
		"    items:\n" +
		"        %s\n" +
		"pvc: %s\n" +
		"pullsecret: %s\n"

	fmt.Printf(
		details,
		d.CaptureName,
		d.Namespace,
		d.DumpFilters,
		d.Image,
		d.TargetResource.Name,
		d.TargetResource.Namespace,
		d.TargetResource.Type,
		d.TargetResource.ContainerName,
		podStr,
		d.PvcName,
		d.PullSecret,
	)

	return nil
}
