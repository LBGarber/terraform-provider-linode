package linode

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testObjectStorageLifecycleResName = "linode_object_storage_lifecycle.foocycle"

func TestAccLinodeObjectStorageLifecycle_basic(t *testing.T) {
	t.Parallel()

	bucketName := acctest.RandomWithPrefix("tf-test")
	keyName := acctest.RandomWithPrefix("tf_test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeObjectStorageKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeObjectStorageBucketLifecycleConfigBasic(bucketName, keyName),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func testAccCheckLinodeObjectStorageBucketLifecycleConfigBasic(name, keyName string) string {
	return testAccCheckLinodeObjectStorageBucketConfigBasic(name) + testAccCheckLinodeObjectStorageKeyConfigBasic(keyName) + fmt.Sprintf(`
resource "linode_object_storage_lifecycle" "foocycle" {
	bucket     = linode_object_storage_bucket.foobar.label
	cluster    = "us-east-1"
	access_key = linode_object_storage_key.foobar.access_key
	secret_key = linode_object_storage_key.foobar.secret_key
	
	lifecycle_rule {
		id = "test-rule"
		prefix = "tf"
		enabled = true

		expiration {
			days = 7
		}
	}
}`)
}
