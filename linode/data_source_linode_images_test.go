package linode

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceLinodeImages_basic(t *testing.T) {
	t.Parallel()

	imageName := acctest.RandomWithPrefix("tf_test")
	resourceName := "data.linode_images.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceLinodeImagesBasic(t, imageName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "images.0.label", imageName),
					resource.TestCheckResourceAttr(resourceName, "images.0.description", "descriptive text"),
					resource.TestCheckResourceAttr(resourceName, "images.0.is_public", "false"),
					resource.TestCheckResourceAttr(resourceName, "images.0.type", "manual"),
					resource.TestCheckResourceAttrSet(resourceName, "images.0.created"),
					resource.TestCheckResourceAttrSet(resourceName, "images.0.created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "images.0.size"),
					resource.TestCheckResourceAttrSet(resourceName, "images.0.deprecated"),
				),
			},
		},
	})
}

func testDataSourceLinodeImagesBasic(t *testing.T, image string) string {
	return testAccExecuteTemplate(t, "data_images", ImageTemplateData{
		InstanceLabel: image,
		Label:         image,
		Description:   "descriptive text"})
}
