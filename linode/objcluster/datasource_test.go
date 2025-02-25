package objcluster_test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/linode/terraform-provider-linode/linode/acceptance"
	"github.com/linode/terraform-provider-linode/linode/objcluster/tmpl"

	"testing"
)

func TestAccDataSourceObjectCluster_basic(t *testing.T) {
	t.Parallel()

	objectStorageClusterID := "us-east-1"
	region := "us-east"
	resourceName := "data.linode_object_storage_cluster.foobar"
	staticSiteDomain := "website-us-east-1.linodeobjects.com"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: tmpl.DataBasic(t, objectStorageClusterID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "region", region),
					resource.TestCheckResourceAttr(resourceName, "id", objectStorageClusterID),
					resource.TestCheckResourceAttr(resourceName, "static_site_domain", staticSiteDomain),
				),
			},
		},
	})
}
