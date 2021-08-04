package linode

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/linode/linodego"
	"testing"
)

func TestAccDataSourceLinodeNodeBalancerConfig_basic(t *testing.T) {
	t.Parallel()

	resName := "data.linode_nodebalancer_config.foofig"
	nodebalancerName := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeNodeBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceLinodeNodeBalancerConfigBasic(t, nodebalancerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLinodeNodeBalancerConfigExists,
					resource.TestCheckResourceAttr(resName, "port", "8080"),
					resource.TestCheckResourceAttr(resName, "protocol", string(linodego.ProtocolHTTP)),
					resource.TestCheckResourceAttr(resName, "check", string(linodego.CheckHTTP)),
					resource.TestCheckResourceAttr(resName, "check_path", "/"),

					resource.TestCheckResourceAttrSet(resName, "algorithm"),
					resource.TestCheckResourceAttrSet(resName, "stickiness"),
					resource.TestCheckResourceAttrSet(resName, "check_attempts"),
					resource.TestCheckResourceAttrSet(resName, "check_timeout"),
					resource.TestCheckResourceAttrSet(resName, "check_interval"),
					resource.TestCheckResourceAttrSet(resName, "check_passive"),
					resource.TestCheckResourceAttrSet(resName, "cipher_suite"),
					resource.TestCheckNoResourceAttr(resName, "ssl_common"),
					resource.TestCheckNoResourceAttr(resName, "ssl_ciphersuite"),
					resource.TestCheckResourceAttr(resName, "node_status.0.up", "0"),
					resource.TestCheckResourceAttr(resName, "node_status.0.down", "0"),
					resource.TestCheckNoResourceAttr(resName, "ssl_cert"),
					resource.TestCheckNoResourceAttr(resName, "ssl_key"),
				),
			},
		},
	})
}

type DataNodeBalancerConfigTemplateData struct {
	Config NodeBalancerConfigTemplateData
}

func testDataSourceLinodeNodeBalancerConfigBasic(t *testing.T, nodeBalancerName string) string {
	return testAccExecuteTemplate(t, "data_nodebalancer_config_basic",
		DataNodeBalancerConfigTemplateData{
			Config: NodeBalancerConfigTemplateData{
				NodeBalancer: NodeBalancerTemplateData{Label: nodeBalancerName}}})
}
