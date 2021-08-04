package linode

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLinodeImage_basic(t *testing.T) {
	t.Parallel()

	imageID := "linode/debian8"
	resourceName := "data.linode_image.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceLinodeImageBasic(t, imageID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", imageID),
					resource.TestCheckResourceAttr(resourceName, "label", "Debian 8"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "is_public", "true"),
					resource.TestCheckResourceAttr(resourceName, "type", "manual"),
					resource.TestCheckResourceAttr(resourceName, "size", "1300"),
					resource.TestCheckResourceAttr(resourceName, "vendor", "Debian"),
				),
			},
		},
	})
}

type DataImageTemplateData struct {
	ID string
}

func testDataSourceLinodeImageBasic(t *testing.T, imageID string) string {
	return testAccExecuteTemplate(t, "data_image", DataImageTemplateData{ID: imageID})
}
