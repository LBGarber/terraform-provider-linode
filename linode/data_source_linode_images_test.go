package linode

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"fmt"
	"strconv"
	"testing"
)

func TestAccDataSourceLinodeImages_basic(t *testing.T) {
	t.Parallel()

	imageID := "linode/debian8"
	resourceName := "data.linode_images.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceLinodeImagesBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "image.0.id", imageID),
					resource.TestCheckResourceAttr(resourceName, "image.0.label", "Debian 8"),
					resource.TestCheckResourceAttr(resourceName, "image.0.description", ""),
					resource.TestCheckResourceAttr(resourceName, "image.0.is_public", "true"),
					resource.TestCheckResourceAttr(resourceName, "image.0.type", "manual"),
					resource.TestCheckResourceAttr(resourceName, "image.0.size", "1300"),
					resource.TestCheckResourceAttr(resourceName, "image.0.vendor", "Debian"),
				),
			},
		},
	})
}

func TestAccDataSourceLinodeImages_noFilters(t *testing.T) {
	t.Parallel()

	resourceName := "data.linode_images.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceLinodeImagesNoFilters(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceCheckLinodeAllImagesNotEmpty(resourceName),
				),
			},
		},
	})
}

func testAccDataSourceCheckLinodeAllImagesNotEmpty(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		linodeCount, err := strconv.Atoi(rs.Primary.Attributes["image.#"])
		if err != nil {
			return fmt.Errorf("failed to parse: %s", err)
		}

		if linodeCount < 1 {
			return fmt.Errorf("expected at least 1 linode image")
		}

		return nil
	}
}

func testDataSourceLinodeImagesBasic() string {
	return `
data "linode_images" "foobar" {
	filter {
		name = "label"
		values = ["Debian 8"]
	}

	filter {
		name = "is_public"
		values = ["true"]
	}

	filter {
		name = "size"
		values = ["1300"]
	}
}`
}

func testDataSourceLinodeImagesNoFilters() string {
	return `
data "linode_images" "foobar" {}`
}
