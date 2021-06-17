package devtestlabs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurestack/azurestack/internal/tf/pluginsdk"
)

type DevTestLinuxVirtualMachineResource struct {
}

func TestAccDevTestLinuxVirtualMachine_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_dev_test_linux_virtual_machine", "test")
	r := DevTestLinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("gallery_image_reference.0.publisher").HasValue("Canonical"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
		data.ImportStep(
			// not returned from the API
			"lab_subnet_name",
			"lab_virtual_network_id",
			"password",
		),
	})
}

func TestAccDevTestLinuxVirtualMachine_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_dev_test_linux_virtual_machine", "test")
	r := DevTestLinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_dev_test_lab_linux_virtual_machine"),
		},
	})
}

func TestAccDevTestLinuxVirtualMachine_basicSSH(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_dev_test_linux_virtual_machine", "test")
	r := DevTestLinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basicSSH(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("gallery_image_reference.0.publisher").HasValue("Canonical"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
		data.ImportStep(
			// not returned from the API
			"lab_subnet_name",
			"lab_virtual_network_id",
			"password",
			"ssh_key",
		),
	})
}

func TestAccDevTestLinuxVirtualMachine_inboundNatRules(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_dev_test_linux_virtual_machine", "test")
	r := DevTestLinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.inboundNatRules(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("disallow_public_ip_address").HasValue("true"),
				check.That(data.ResourceName).Key("gallery_image_reference.0.publisher").HasValue("Canonical"),
				check.That(data.ResourceName).Key("tags.%").HasValue("1"),
				check.That(data.ResourceName).Key("tags.Acceptance").HasValue("Test"),
			),
		},
		data.ImportStep(
			// not returned from the API
			"inbound_nat_rule",
			"lab_subnet_name",
			"lab_virtual_network_id",
			"password",
		),
	})
}

func TestAccDevTestLinuxVirtualMachine_updateStorage(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_dev_test_linux_virtual_machine", "test")
	r := DevTestLinuxVirtualMachineResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.storage(data, "Standard"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("gallery_image_reference.0.publisher").HasValue("Canonical"),
				check.That(data.ResourceName).Key("storage_type").HasValue("Standard"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
		{
			Config: r.storage(data, "Premium"),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("gallery_image_reference.0.publisher").HasValue("Canonical"),
				check.That(data.ResourceName).Key("storage_type").HasValue("Premium"),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
	})
}

func (DevTestLinuxVirtualMachineResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := azure.ParseAzureResourceID(state.ID)
	if err != nil {
		return nil, err
	}
	labName := id.Path["labs"]
	name := id.Path["virtualmachines"]

	resp, err := clients.DevTestLabs.VirtualMachinesClient.Get(ctx, id.ResourceGroup, labName, name, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving DevTest Linux Virtual Machine %q (Lab %q / Resource Group: %q): %v", name, labName, id.ResourceGroup, err)
	}

	return utils.Bool(resp.LabVirtualMachineProperties != nil), nil
}

func (DevTestLinuxVirtualMachineResource) basic(data acceptance.TestData) string {
	template := DevTestLinuxVirtualMachineResource{}.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_dev_test_linux_virtual_machine" "test" {
  name                   = "acctestvm-vm%d"
  lab_name               = azurerm_dev_test_lab.test.name
  resource_group_name    = azurerm_resource_group.test.name
  location               = azurerm_resource_group.test.location
  size                   = "Standard_F2"
  username               = "acct5stU5er"
  password               = "Pa$w0rd1234!"
  lab_virtual_network_id = azurerm_dev_test_virtual_network.test.id
  lab_subnet_name        = azurerm_dev_test_virtual_network.test.subnet[0].name
  storage_type           = "Standard"

  gallery_image_reference {
    offer     = "UbuntuServer"
    publisher = "Canonical"
    sku       = "18.04-LTS"
    version   = "latest"
  }
}
`, template, data.RandomInteger)
}

func (DevTestLinuxVirtualMachineResource) requiresImport(data acceptance.TestData) string {
	template := DevTestLinuxVirtualMachineResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_dev_test_linux_virtual_machine" "import" {
  name                   = azurerm_dev_test_linux_virtual_machine.test.name
  lab_name               = azurerm_dev_test_linux_virtual_machine.test.lab_name
  resource_group_name    = azurerm_dev_test_linux_virtual_machine.test.resource_group_name
  location               = azurerm_dev_test_linux_virtual_machine.test.location
  size                   = azurerm_dev_test_linux_virtual_machine.test.size
  username               = "acct5stU5er"
  password               = "Pa$w0rd1234!"
  lab_virtual_network_id = azurerm_dev_test_virtual_network.test.id
  lab_subnet_name        = azurerm_dev_test_virtual_network.test.subnet[0].name
  storage_type           = "Standard"

  gallery_image_reference {
    offer     = "UbuntuServer"
    publisher = "Canonical"
    sku       = "18.04-LTS"
    version   = "latest"
  }
}
`, template)
}

func (DevTestLinuxVirtualMachineResource) basicSSH(data acceptance.TestData) string {
	template := DevTestLinuxVirtualMachineResource{}.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_dev_test_linux_virtual_machine" "test" {
  name                   = "acctestvm-vm%d"
  lab_name               = azurerm_dev_test_lab.test.name
  resource_group_name    = azurerm_resource_group.test.name
  location               = azurerm_resource_group.test.location
  size                   = "Standard_F2"
  username               = "acct5stU5er"
  ssh_key                = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDCsTcryUl51Q2VSEHqDRNmceUFo55ZtcIwxl2QITbN1RREti5ml/VTytC0yeBOvnZA4x4CFpdw/lCDPk0yrH9Ei5vVkXmOrExdTlT3qI7YaAzj1tUVlBd4S6LX1F7y6VLActvdHuDDuXZXzCDd/97420jrDfWZqJMlUK/EmCE5ParCeHIRIvmBxcEnGfFIsw8xQZl0HphxWOtJil8qsUWSdMyCiJYYQpMoMliO99X40AUc4/AlsyPyT5ddbKk08YrZ+rKDVHF7o29rh4vi5MmHkVgVQHKiKybWlHq+b71gIAUQk9wrJxD+dqt4igrmDSpIjfjwnd+l5UIn5fJSO5DYV4YT/4hwK7OKmuo7OFHD0WyY5YnkYEMtFgzemnRBdE8ulcT60DQpVgRMXFWHvhyCWy0L6sgj1QWDZlLpvsIvNfHsyhKFMG1frLnMt/nP0+YCcfg+v1JYeCKjeoJxB8DWcRBsjzItY0CGmzP8UYZiYKl/2u+2TgFS5r7NWH11bxoUzjKdaa1NLw+ieA8GlBFfCbfWe6YVB9ggUte4VtYFMZGxOjS2bAiYtfgTKFJv+XqORAwExG6+G2eDxIDyo80/OA9IG7Xv/jwQr7D6KDjDuULFcN/iTxuttoKrHeYz1hf5ZQlBdllwJHYx6fK2g8kha6r2JIQKocvsAXiiONqSfw== hello@world.com"
  lab_virtual_network_id = azurerm_dev_test_virtual_network.test.id
  lab_subnet_name        = azurerm_dev_test_virtual_network.test.subnet[0].name
  storage_type           = "Standard"

  gallery_image_reference {
    offer     = "UbuntuServer"
    publisher = "Canonical"
    sku       = "18.04-LTS"
    version   = "latest"
  }
}
`, template, data.RandomInteger)
}

func (DevTestLinuxVirtualMachineResource) inboundNatRules(data acceptance.TestData) string {
	template := DevTestLinuxVirtualMachineResource{}.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_dev_test_linux_virtual_machine" "test" {
  name                       = "acctestvm-vm%d"
  lab_name                   = azurerm_dev_test_lab.test.name
  resource_group_name        = azurerm_resource_group.test.name
  location                   = azurerm_resource_group.test.location
  size                       = "Standard_F2"
  username                   = "acct5stU5er"
  password                   = "Pa$w0rd1234!"
  disallow_public_ip_address = true
  lab_virtual_network_id     = azurerm_dev_test_virtual_network.test.id
  lab_subnet_name            = azurerm_dev_test_virtual_network.test.subnet[0].name
  storage_type               = "Standard"

  gallery_image_reference {
    offer     = "UbuntuServer"
    publisher = "Canonical"
    sku       = "18.04-LTS"
    version   = "latest"
  }

  inbound_nat_rule {
    protocol     = "Tcp"
    backend_port = 22
  }

  inbound_nat_rule {
    protocol     = "Tcp"
    backend_port = 3389
  }

  tags = {
    "Acceptance" = "Test"
  }
}
`, template, data.RandomInteger)
}

func (DevTestLinuxVirtualMachineResource) storage(data acceptance.TestData, storageType string) string {
	template := DevTestLinuxVirtualMachineResource{}.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_dev_test_linux_virtual_machine" "test" {
  name                   = "acctestvm-vm%d"
  lab_name               = azurerm_dev_test_lab.test.name
  resource_group_name    = azurerm_resource_group.test.name
  location               = azurerm_resource_group.test.location
  size                   = "Standard_B1ms"
  username               = "acct5stU5er"
  password               = "Pa$w0rd1234!"
  lab_virtual_network_id = azurerm_dev_test_virtual_network.test.id
  lab_subnet_name        = azurerm_dev_test_virtual_network.test.subnet[0].name
  storage_type           = "%s"

  gallery_image_reference {
    offer     = "UbuntuServer"
    publisher = "Canonical"
    sku       = "18.04-LTS"
    version   = "latest"
  }
}
`, template, data.RandomInteger, storageType)
}

func (DevTestLinuxVirtualMachineResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_dev_test_lab" "test" {
  name                = "acctestdtl%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_dev_test_virtual_network" "test" {
  name                = "acctestdtvn%d"
  lab_name            = azurerm_dev_test_lab.test.name
  resource_group_name = azurerm_resource_group.test.name

  subnet {
    use_public_ip_address           = "Allow"
    use_in_virtual_machine_creation = "Allow"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}
