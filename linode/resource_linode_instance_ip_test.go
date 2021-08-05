package linode

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testInstanceIPResName = "linode_instance_ip.test"

func TestAccLinodeInstanceIP_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf_test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: accTestWithProvider(testAccCheckLinodeInstanceIPBasic(t, name), map[string]interface{}{
					providerKeySkipInstanceReadyPoll: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "address"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "gateway"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "prefix"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "rdns"),
					resource.TestCheckResourceAttrSet(testInstanceIPResName, "subnet_mask"),
					resource.TestCheckResourceAttr(testInstanceIPResName, "region", "us-east"),
					resource.TestCheckResourceAttr(testInstanceIPResName, "type", "ipv4"),
				),
			},
		},
	})
}

type InstanceIPTemplateData struct {
	InstanceLabel string
	PubKey        string
}

func testAccCheckLinodeInstanceIPBasic(t *testing.T, label string) string {
	return testAccExecuteTemplate(t, "instance_ip_basic", InstanceIPTemplateData{
		InstanceLabel: label,
		PubKey:        publicKeyMaterial,
	})
}
