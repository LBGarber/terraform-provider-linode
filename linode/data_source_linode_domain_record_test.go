package linode

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLinodeDomainRecord_basic(t *testing.T) {
	datasourceName := "data.linode_domain_record.record"
	domain := acctest.RandomWithPrefix("recordtest") + ".com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLinodeDomainRecordConfigBasic(t, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", "www"),
					resource.TestCheckResourceAttr(datasourceName, "type", "CNAME"),
					resource.TestCheckResourceAttr(datasourceName, "ttl_sec", "7200"),
					resource.TestCheckResourceAttr(datasourceName, "target", domain),
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					resource.TestCheckResourceAttrSet(datasourceName, "domain_id"),
				),
			},
		},
	})
}

func TestAccDataSourceLinodeDomainRecord_idLookup(t *testing.T) {
	datasourceName := "data.linode_domain_record.record"
	domain := acctest.RandomWithPrefix("idloikup") + ".com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLinodeDomainRecordConfigIDLookup(t, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", "www"),
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					resource.TestCheckResourceAttrSet(datasourceName, "domain_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "type"),
				),
			},
		},
	})
}

func TestAccDataSourceLinodeDomainRecord_srv(t *testing.T) {
	datasourceName := "data.linode_domain_record.record"
	domain := acctest.RandomWithPrefix("srv") + ".com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLinodeDomainRecordConfigSRV(t, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "type", "SRV"),
					resource.TestCheckResourceAttr(datasourceName, "port", "80"),
					resource.TestCheckResourceAttr(datasourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(datasourceName, "service", "sip"),
					resource.TestCheckResourceAttr(datasourceName, "weight", "5"),
					resource.TestCheckResourceAttr(datasourceName, "priority", "10"),
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					resource.TestCheckResourceAttrSet(datasourceName, "domain_id"),
				),
			},
		},
	})
}

func TestAccDataSourceLinodeDomainRecord_caa(t *testing.T) {
	datasourceName := "data.linode_domain_record.record"
	domain := acctest.RandomWithPrefix("caa") + ".com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLinodeDomainRecordConfigCAA(t, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", "caa_test"),
					resource.TestCheckResourceAttr(datasourceName, "type", "CAA"),
					resource.TestCheckResourceAttr(datasourceName, "tag", "issue"),
					resource.TestCheckResourceAttr(datasourceName, "target", "test"),
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					resource.TestCheckResourceAttrSet(datasourceName, "domain_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "type"),
				),
			},
		},
	})
}

type DataDomainRecordTemplateData struct {
	Domain string
}

func testAccDataSourceLinodeDomainRecordConfigBasic(t *testing.T, domain string) string {
	return testAccExecuteTemplate(t, "data_domain_record_basic",
		DataDomainRecordTemplateData{Domain: domain})
}

func testAccDataSourceLinodeDomainRecordConfigIDLookup(t *testing.T, domain string) string {
	return testAccExecuteTemplate(t, "data_domain_record_id",
		DataDomainRecordTemplateData{Domain: domain})
}

func testAccDataSourceLinodeDomainRecordConfigSRV(t *testing.T, domain string) string {
	return testAccExecuteTemplate(t, "data_domain_record_srv",
		DataDomainRecordTemplateData{Domain: domain})
}

func testAccDataSourceLinodeDomainRecordConfigCAA(t *testing.T, domain string) string {
	return testAccExecuteTemplate(t, "data_domain_record_caa",
		DataDomainRecordTemplateData{Domain: domain})
}
