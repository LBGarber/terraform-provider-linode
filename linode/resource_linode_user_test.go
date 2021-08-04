package linode

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/linode/linodego"
)

const testUserResName = "linode_user.test"

func testAccCheckLinodeUserDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_user" {
			continue
		}

		username := rs.Primary.ID
		_, err := client.GetUser(context.TODO(), username)

		if err == nil {
			return fmt.Errorf("should not find user %s existing after delete", username)
		}

		if apiErr, ok := err.(*linodego.Error); ok && apiErr.Code != 404 {
			return fmt.Errorf("Error getting user %s: %s", username, err)
		}
	}
	return nil
}

func TestAccLinodeUser_basic(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-test")
	email := username + "@example.com"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeUserConfigBasic(t, username, email, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testUserResName, "email", email),
					resource.TestCheckResourceAttr(testUserResName, "username", username),
					resource.TestCheckResourceAttr(testUserResName, "restricted", "true"),
					resource.TestCheckResourceAttr(testUserResName, "ssh_keys.#", "0"),
					resource.TestCheckResourceAttr(testUserResName, "tfa_enabled", "false"),
				),
			},
		},
	})
}

func TestAccLinodeUser_updates(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-test")
	updatedUsername := acctest.RandomWithPrefix("tf-test")
	email := username + "@example.com"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeUserConfigBasic(t, username, email, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testUserResName, "email", email),
					resource.TestCheckResourceAttr(testUserResName, "username", username),
					resource.TestCheckResourceAttr(testUserResName, "restricted", "false"),
					resource.TestCheckResourceAttr(testUserResName, "ssh_keys.#", "0"),
					resource.TestCheckResourceAttr(testUserResName, "tfa_enabled", "false"),
				),
			},
			{
				Config: testAccCheckLinodeUserConfigBasic(t, updatedUsername, email, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testUserResName, "email", email),
					resource.TestCheckResourceAttr(testUserResName, "username", updatedUsername),
					resource.TestCheckResourceAttr(testUserResName, "restricted", "true"),
					resource.TestCheckResourceAttr(testUserResName, "ssh_keys.#", "0"),
					resource.TestCheckResourceAttr(testUserResName, "tfa_enabled", "false"),
				),
			},
		},
	})
}

func TestAccLinodeUser_grants(t *testing.T) {
	t.Parallel()

	username := acctest.RandomWithPrefix("tf-test")
	instance := acctest.RandomWithPrefix("tf-test")

	email := username + "@example.com"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeUserConfigGrants(t, username, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.account_access", ""),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_domains", "true"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_firewalls", "true"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_images", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_linodes", "true"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_longview", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_nodebalancers", "true"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_stackscripts", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_volumes", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.cancel_account", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.longview_subscription", "false"),
					resource.TestCheckResourceAttr(testUserResName, "linode_grant.#", "0"),
				),
			},
			{
				Config: testAccCheckLinodeUserConfigGrantsUpdate(t, username, email, instance),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.account_access", "read_only"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_domains", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_firewalls", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_images", "true"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_linodes", "true"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_longview", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_nodebalancers", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_stackscripts", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.add_volumes", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.cancel_account", "false"),
					resource.TestCheckResourceAttr(testUserResName, "global_grants.0.longview_subscription", "false"),
					resource.TestCheckResourceAttr(testUserResName, "linode_grant.#", "1"),
					resource.TestCheckResourceAttr(testUserResName, "linode_grant.0.permissions", "read_write"),
				),
			},
		},
	})
}

type UserTemplateData struct {
	Username   string
	Email      string
	Restricted bool
	Instance   InstanceTemplateData
}

func testAccCheckLinodeUserConfigBasic(t *testing.T, username, email string, restricted bool) string {
	return testAccExecuteTemplate(t, "user_basic",
		UserTemplateData{
			Username:   username,
			Email:      email,
			Restricted: restricted,
		})
}

func testAccCheckLinodeUserConfigGrants(t *testing.T, username, email string) string {
	return testAccExecuteTemplate(t, "user_grants",
		UserTemplateData{
			Username: username,
			Email:    email,
		})
}

func testAccCheckLinodeUserConfigGrantsUpdate(t *testing.T, username, email, instance string) string {
	return testAccExecuteTemplate(t, "user_grants_updates",
		UserTemplateData{
			Username: username,
			Email:    email,
			Instance: InstanceTemplateData{Label: instance},
		})
}
