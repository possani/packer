//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type ContentLibraryDestinationConfig
package common

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/vapi/vcenter"
)

type ContentLibraryDestinationConfig struct {
	Library      string `mapstructure:"library"`
	Name         string `mapstructure:"name"`
	Description  string `mapstructure:"description"`
	Cluster      string `mapstructure:"cluster"`
	Folder       string `mapstructure:"folder"`
	Host         string `mapstructure:"host"`
	ResourcePool string `mapstructure:"resource_pool"`
}

func (c *ContentLibraryDestinationConfig) Prepare(lc *LocationConfig) []error {
	var errs *packer.MultiError

	if c.Library == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("a library name must be provided"))
	}

	if c.Name == "" {
		// Add timestamp to the the name to differentiate from the original VM
		// otherwise vSphere won't be able to create the template which will be imported
		c.Name = lc.VMName + string(time.Now().Unix())
	}
	if c.Cluster == "" {
		c.Cluster = lc.Cluster
	}
	if c.Folder == "" {
		c.Folder = lc.Folder
	}
	if c.Host == "" {
		c.Host = lc.Host
	}
	if c.ResourcePool == "" {
		c.ResourcePool = lc.ResourcePool
	}

	if c.Description == "" {
		c.Description = fmt.Sprintf("Packer imported %s VM template", lc.VMName)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs.Errors
	}

	return nil
}

type StepImportToContentLibrary struct {
	ContentLibConfig *ContentLibraryDestinationConfig
}

func (s *StepImportToContentLibrary) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	template := vcenter.Template{
		Name:        s.ContentLibConfig.Name,
		Description: s.ContentLibConfig.Description,
		Library:     s.ContentLibConfig.Library,
		Placement: &vcenter.Placement{
			Cluster:      s.ContentLibConfig.Cluster,
			Folder:       s.ContentLibConfig.Folder,
			Host:         s.ContentLibConfig.Host,
			ResourcePool: s.ContentLibConfig.ResourcePool,
		},
	}

	ui.Say("Importing VM template to Content Library...")
	err := vm.ImportToContentLibrary(template)
	if err != nil {
		log.Printf("failed to import VM template: %s", err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepImportToContentLibrary) Cleanup(multistep.StateBag) {
}
