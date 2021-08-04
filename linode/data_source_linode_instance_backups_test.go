package linode

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/linode/linodego"
	"testing"
)

func TestAccDataSourceLinodeInstanceBackups_basic(t *testing.T) {
	t.Parallel()

	instanceName := acctest.RandomWithPrefix("tf_test")
	snapshotName := acctest.RandomWithPrefix("tf_test_cool")

	resourceName := "data.linode_instance_backups.foobar"

	var instance linodego.Instance
	var snapshot *linodego.InstanceSnapshot

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceLinodeInstanceBackupsInstance(t, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeInstanceExists("linode_instance.foobar", &instance),
				),
			},
			{
				PreConfig: func() {
					client := testAccProvider.Meta().(*ProviderMeta).Client
					newSnapshot, err := client.CreateInstanceSnapshot(context.Background(), instance.ID, snapshotName)
					if err != nil {
						t.Fatal(err)
					}

					snapshot = newSnapshot
				},
				Config: testDataSourceLinodeInstanceBackupsBasic(t, instanceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "in_progress.0.id"),
					resource.TestCheckResourceAttr(resourceName, "in_progress.0.label", snapshotName),
					resource.TestCheckResourceAttrSet(resourceName, "in_progress.0.status"),
					resource.TestCheckResourceAttrSet(resourceName, "in_progress.0.type"),
					resource.TestCheckResourceAttrSet(resourceName, "in_progress.0.created"),
				),
			},
			{
				PreConfig: func() {
					client := testAccProvider.Meta().(*ProviderMeta).Client
					if _, err := client.WaitForSnapshotStatus(context.Background(), instance.ID, snapshot.ID, linodego.SnapshotSuccessful, 600); err != nil {
						t.Fatal(err)
					}
				},
				Config: testDataSourceLinodeInstanceBackupsBasic(t, instanceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "current.0.id"),
					resource.TestCheckResourceAttr(resourceName, "current.0.label", snapshotName),
					resource.TestCheckResourceAttr(resourceName, "current.0.status", "successful"),
					resource.TestCheckResourceAttrSet(resourceName, "current.0.type"),
					resource.TestCheckResourceAttrSet(resourceName, "current.0.created"),
					resource.TestCheckResourceAttrSet(resourceName, "current.0.updated"),
					resource.TestCheckResourceAttrSet(resourceName, "current.0.finished"),
				),
			},
		},
	})
}

type DataInstanceBackupsTemplateData struct {
	Label string
}

func testDataSourceLinodeInstanceBackupsInstance(t *testing.T, instanceLabel string) string {
	return testAccExecuteTemplate(t, "data_instance_backups_instance",
		DataInstanceBackupsTemplateData{Label: instanceLabel})
}

func testDataSourceLinodeInstanceBackupsBasic(t *testing.T, instanceLabel string) string {
	return testAccExecuteTemplate(t, "data_instance_backups_basic",
		DataInstanceBackupsTemplateData{Label: instanceLabel})
}
