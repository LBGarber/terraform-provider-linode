package linode

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"fmt"
	"testing"
)

func TestAccDataSourceLinodeVLANs_basic(t *testing.T) {
	t.Parallel()

	instanceName := acctest.RandomWithPrefix("tf_test")
	vlanName := "tf-test-vlan"
	resourceName := "data.linode_vlans.foolan"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceLinodeVLANsBasic(instanceName, vlanName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "vlans.0.label", vlanName),
					resource.TestCheckResourceAttr(resourceName, "vlans.0.region", "us-southeast"),
					resource.TestCheckResourceAttrSet(resourceName, "vlans.0.created"),
				),
			},
		},
	})
}

func testDataSourceLinodeVLANsInstance(instanceName, vlanName string) string {
	return fmt.Sprintf(`
resource "linode_instance" "fooinst" {
	label = "%s"
	type = "g6-standard-1"
	image = "linode/alpine3.13"
	region = "us-southeast"

	interface {
		label = "%s"
		purpose = "vlan"
	}
}
`, instanceName, vlanName)
}

func testDataSourceLinodeVLANsBasic(instanceName, vlanName string) string {
	return testDataSourceLinodeVLANsInstance(instanceName, vlanName) + fmt.Sprintf(`
data "linode_vlans" "foolan" {
	filter {
		name = "label"
		values = ["%s"]
	}
}`, vlanName)
}
