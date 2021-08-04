package linode

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceLinodeNodeBalancerNode_basic(t *testing.T) {
	t.Parallel()

	resName := "data.linode_nodebalancer_node.foonode"
	nodebalancerName := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeNodeBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: accTestWithProvider(testDataSourceLinodeNodeBalancerNodeBasic(t, nodebalancerName), map[string]interface{}{
					providerKeySkipInstanceReadyPoll: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeNodeBalancerNode,
					resource.TestCheckResourceAttr(resName, "label", nodebalancerName),
					resource.TestCheckResourceAttrSet(resName, "status"),
					resource.TestCheckResourceAttr(resName, "mode", "accept"),
					resource.TestCheckResourceAttr(resName, "weight", "50"),
				),
			},
		},
	})
}

type DataNodeBalancerNodeTemplateData struct {
	Node NodeBalancerNodeTemplateData
}

func testDataSourceLinodeNodeBalancerNodeBasic(t *testing.T, nodeBalancerName string) string {
	return testAccExecuteTemplate(t, "data_nodebalancer_node_basic",
		DataNodeBalancerNodeTemplateData{
			Node: NodeBalancerNodeTemplateData{
				Label: nodeBalancerName,
				Config: NodeBalancerConfigTemplateData{
					NodeBalancer: NodeBalancerTemplateData{Label: nodeBalancerName},
				},
				Instance: InstanceTemplateData{
					Label:  nodeBalancerName,
					PubKey: publicKeyMaterial,
				},
			},
		})
}
