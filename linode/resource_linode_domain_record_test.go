package linode

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/linode/linodego"
)

func TestAccLinodeDomainRecord_basic(t *testing.T) {
	t.Parallel()

	resName := "linode_domain_record.foobar"
	domainRecordName := acctest.RandomWithPrefix("tf-test-")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeDomainRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeDomainRecordConfigBasic(t, domainRecordName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeDomainRecordExists,
					resource.TestCheckResourceAttr(resName, "name", domainRecordName),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccStateIDDomainRecord,
			},
		},
	})
}

func TestAccLinodeDomainRecord_roundedTTLSec(t *testing.T) {
	t.Parallel()

	resName := "linode_domain_record.foobar"
	domainRecordName := acctest.RandomWithPrefix("tf-test-")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeDomainRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeDomainRecordConfigWithTTL(t, domainRecordName, 299),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeDomainRecordExists,
					resource.TestCheckResourceAttr(resName, "name", domainRecordName),
					resource.TestCheckResourceAttr(resName, "ttl_sec", "300"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccStateIDDomainRecord,
			},
		},
	})
}

func TestAccLinodeDomainRecord_ANoName(t *testing.T) {
	t.Parallel()

	resName := "linode_domain_record.foobar"
	domainName := acctest.RandomWithPrefix("tf-test-") + ".example"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeDomainRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeDomainRecordConfigANoName(t, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "record_type", "A"),
					resource.TestCheckResourceAttr(resName, "name", ""),
				),
			},
		},
	})
}

func TestAccLinodeDomainRecord_AAAANoName(t *testing.T) {
	t.Parallel()

	resName := "linode_domain_record.foobar"
	domainName := acctest.RandomWithPrefix("tf-test-") + ".example"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeDomainRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeDomainRecordConfigAAAANoName(t, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "record_type", "AAAA"),
					resource.TestCheckResourceAttr(resName, "name", ""),
				),
			},
		},
	})
}

func TestAccLinodeDomainRecord_CAANoName(t *testing.T) {
	t.Parallel()

	resName := "linode_domain_record.foobar"
	domainName := acctest.RandomWithPrefix("tf-test-") + ".example"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeDomainRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeDomainRecordConfigCAANoName(t, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "record_type", "CAA"),
					resource.TestCheckResourceAttr(resName, "name", ""),
				),
			},
		},
	})
}

func TestAccLinodeDomainRecord_SRV(t *testing.T) {
	t.Parallel()

	resName := "linode_domain_record.foobar"
	domainName := acctest.RandomWithPrefix("tftest") + ".example"
	expectedName := "_myservice._tcp"
	expectedTarget := "mysubdomain." + domainName
	expectedTargetExternal := "subdomain.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeDomainRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeDomainRecordConfigSRV(t, domainName, expectedTarget),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", expectedName),
					resource.TestCheckResourceAttr(resName, "target", expectedTarget),
					resource.TestCheckResourceAttr(resName, "record_type", "SRV"),
					resource.TestCheckResourceAttr(resName, "protocol", "tcp"),
				),
			},
			{
				Config: testAccCheckLinodeDomainRecordConfigSRV(t, domainName, expectedTargetExternal),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", expectedName),
					resource.TestCheckResourceAttr(resName, "target", expectedTargetExternal),
					resource.TestCheckResourceAttr(resName, "record_type", "SRV"),
					resource.TestCheckResourceAttr(resName, "protocol", "tcp"),
				),
			},
		},
	})
}

func TestAccLinodeDomainRecord_SRVNoFQDN(t *testing.T) {
	t.Parallel()

	resName := "linode_domain_record.foobar"
	domainName := acctest.RandomWithPrefix("tftest") + ".example"
	expectedName := "_myservice._tcp"
	expectedTarget := "mysubdomain." + domainName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeDomainRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeDomainRecordConfigSRV(t, domainName, "mysubdomain"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", expectedName),
					resource.TestCheckResourceAttr(resName, "target", expectedTarget),
					resource.TestCheckResourceAttr(resName, "record_type", "SRV"),
					resource.TestCheckResourceAttr(resName, "protocol", "tcp"),
				),
			},
			{
				Config: testAccCheckLinodeDomainRecordConfigSRV(t, domainName, "mysubdomainbutnew"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resName, "name", expectedName),
					resource.TestCheckResourceAttr(resName, "target", "mysubdomainbutnew."+domainName),
					resource.TestCheckResourceAttr(resName, "record_type", "SRV"),
					resource.TestCheckResourceAttr(resName, "protocol", "tcp"),
				),
			},
		},
	})
}

func testAccStateIDDomainRecord(s *terraform.State) (string, error) {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_domain_record" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return "", fmt.Errorf("Error parsing ID %v to int", rs.Primary.ID)
		}
		domainID, err := strconv.Atoi(rs.Primary.Attributes["domain_id"])
		if err != nil {
			return "", fmt.Errorf("Error parsing domain_id %v to int", rs.Primary.Attributes["domain_id"])
		}
		return fmt.Sprintf("%d,%d", domainID, id), nil
	}

	return "", fmt.Errorf("Error finding linode_domain_record")
}

func TestAccLinodeDomainRecord_update(t *testing.T) {
	t.Parallel()

	domainRecordName := acctest.RandomWithPrefix("tf-test-")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeDomainRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeDomainRecordConfigBasic(t, domainRecordName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeDomainRecordExists,
					resource.TestCheckResourceAttr("linode_domain_record.foobar", "name", domainRecordName),
				),
			},
			{
				Config: testAccCheckLinodeDomainRecordConfigUpdates(t, domainRecordName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeDomainRecordExists,
					resource.TestCheckResourceAttr("linode_domain_record.foobar", "name", fmt.Sprintf("renamed-%s", domainRecordName)),
				),
			},
		},
	})
}

func testAccCheckLinodeDomainRecordExists(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_domain_record" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error parsing %v to int", rs.Primary.ID)
		}
		domainID, err := strconv.Atoi(rs.Primary.Attributes["domain_id"])
		if err != nil {
			return fmt.Errorf("Error parsing %v to int", rs.Primary.Attributes["domain_id"])
		}
		_, err = client.GetDomainRecord(context.Background(), domainID, id)
		if err != nil {
			return fmt.Errorf("Error retrieving state of DomainRecord %s: %s", rs.Primary.Attributes["name"], err)
		}
	}

	return nil
}

func testAccCheckLinodeDomainRecordDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderMeta).Client
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_domain_record" {
			continue
		}
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error parsing %v to int", rs.Primary.ID)
		}
		domainID, err := strconv.Atoi(rs.Primary.Attributes["domain_id"])
		if err != nil {
			return fmt.Errorf("Error parsing domain_id %v to int", rs.Primary.Attributes["domain_id"])
		}

		if id == 0 {
			return fmt.Errorf("Would have considered %v as %d", rs.Primary.ID, id)

		}

		_, err = client.GetDomainRecord(context.Background(), domainID, id)

		if err == nil {
			return fmt.Errorf("Linode DomainRecord with id %d still exists", id)
		}

		if apiErr, ok := err.(*linodego.Error); ok && apiErr.Code != 404 {
			return fmt.Errorf("Error requesting Linode DomainRecord with id %d", id)
		}
	}

	return nil
}

type DomainRecordTemplateData struct {
	Domain     string
	RecordName string
	TTLSec     int
	Target     string
}

func testAccCheckLinodeDomainRecordConfigBasic(t *testing.T, domainRecord string) string {
	return testAccExecuteTemplate(t, "domain_record_basic", DomainRecordTemplateData{
		Domain: domainRecord + ".example", RecordName: domainRecord})
}

func testAccCheckLinodeDomainRecordConfigWithTTL(t *testing.T, domainRecord string, ttlSec int) string {
	return testAccExecuteTemplate(t, "domain_record_ttl", DomainRecordTemplateData{
		RecordName: domainRecord, Domain: domainRecord + ".example", TTLSec: ttlSec})
}

func testAccCheckLinodeDomainRecordConfigUpdates(t *testing.T, domainRecord string) string {
	return testAccExecuteTemplate(t, "domain_record_updates", DomainRecordTemplateData{
		Domain: domainRecord + ".example", RecordName: domainRecord})
}

func testAccCheckLinodeDomainRecordConfigANoName(t *testing.T, domainName string) string {
	return testAccExecuteTemplate(t, "domain_record_a_no_name", DomainRecordTemplateData{Domain: domainName})
}

func testAccCheckLinodeDomainRecordConfigAAAANoName(t *testing.T, domainName string) string {
	return testAccExecuteTemplate(t, "domain_record_aaaa_no_name", DomainRecordTemplateData{
		Domain: domainName})
}

func testAccCheckLinodeDomainRecordConfigCAANoName(t *testing.T, domainName string) string {
	return testAccExecuteTemplate(t, "domain_record_caa_no_name", DomainRecordTemplateData{
		Domain: domainName})
}

func testAccCheckLinodeDomainRecordConfigSRV(t *testing.T, domainName string, target string) string {
	return testAccExecuteTemplate(t, "domain_record_srv", DomainRecordTemplateData{
		Domain: domainName, Target: target})
}
