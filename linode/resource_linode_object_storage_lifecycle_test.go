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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testObjectStorageLifecycleResName, "bucket", bucketName),
					resource.TestCheckResourceAttr(testObjectStorageLifecycleResName, "cluster", "us-east-1"),
					resource.TestCheckResourceAttr(testObjectStorageLifecycleResName, "lifecycle_rule.#", "1"),
					resource.TestCheckResourceAttr(testObjectStorageLifecycleResName, "lifecycle_rule.0.id", "test-rule"),
					resource.TestCheckResourceAttr(testObjectStorageLifecycleResName, "lifecycle_rule.0.prefix", "tf"),
					resource.TestCheckResourceAttr(testObjectStorageLifecycleResName, "lifecycle_rule.0.enabled", "true"),
					resource.TestCheckResourceAttr(testObjectStorageLifecycleResName, "lifecycle_rule.0.expiration.#", "1"),
					//resource.TestCheckResourceAttr(testObjectStorageLifecycleResName, "lifecycle_rule.0.expiration.0.days", "7"),
					//resource.TestCheckResourceAttr(testObjectStorageLifecycleResName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", "true"),
					resource.TestCheckResourceAttrSet(testObjectStorageLifecycleResName, "lifecycle_rule.0.expiration.0.date"),
				),
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
			date = "2021-06-21"
		}
	}
}`)
}
