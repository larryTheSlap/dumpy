package subcmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	stopExample = `# stop capture mycap sniffers in current namespace
  kubectl dumpy stop mycap`
)

func Dumpysubcmd_stop() (cmd *cobra.Command) {

	dumpy := NewDumpy()

	cmd = &cobra.Command{
		Use:          "stop <captureName> [-n captureNamespace]",
		Short:        "terminate capture sniffers",
		Example:      stopExample,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := dumpy.Stop_Complete(); err != nil {
				return err
			}
			if err := dumpy.Stop_Validate(args, cmd); err != nil {
				return err
			}
			if err := dumpy.Stop_Run(args); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&dumpy.Namespace, "namespace", "n", "", "dumpy capture sniffers namespace (default: current namespace)")
	return
}

func (d *Dumpy) Stop_Complete() (err error) {
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

func (d *Dumpy) Stop_Validate(args []string, cmd *cobra.Command) error {
	if len(args) == 0 {
		cmd.Usage()
		return errors.New("no argument provided")
	}
	if len(args) != 1 {
		return errors.New("unkown arguments, stop command require capture name. use -h for help")
	}
	return nil
}

func (d *Dumpy) Stop_Run(args []string) (err error) {
	d.CaptureName = args[0]
	if err := d.NewSniffersFromExisting(); err != nil {
		return err
	}
	if len(d.Sniffers) == 0 {
		return fmt.Errorf("%s sniffers not found in namespace %s", d.CaptureName, d.Namespace)
	}
	fmt.Printf("Stopping capture %s..\n", d.CaptureName)
	for _, s := range d.Sniffers {
		_, errBuff, err := d.Api.Exec_k8sCommand("echo true > /tmp/dumpy/termination_flag", s.Name, s.Namespace)
		if err != nil || errBuff != "" {
			return fmt.Errorf("could not interupt tcpdump capture on pod %s", s.Name)
		}
	}
	if err := d.Wait_Completed(); err != nil {
		return err
	}
	fmt.Printf("%s sniffers have been successfully stopped\n", d.CaptureName)
	return nil
}
