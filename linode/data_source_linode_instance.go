package linode

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceLinodeInstanceFilter() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type: schema.TypeInt,
				Description: "The unique ID of the Linode.",
				Optional: true,
			},
			"group": {
				Type:        schema.TypeString,
				Description: "A deprecated property denoting a group label for the Linode.",
				Optional:    true,
			},
			"image": {
				Type: schema.TypeString,
				Description: "The image the Linode instance was deployed from.",
				Optional: true,
			},
			"label": {
				Type: schema.TypeString,
				Description: "The label assigned to the Linode instance.",
				Optional: true,
			},
			"region": {
				Type: schema.TypeString,
				Description: "The region the Linode instance is located in.",
				Optional: true,
			},
			"tags": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "An array of tags given to the Linode.",
			},
		},
	}
}

func dataSourceLinodeInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceLinodeInstanceTypeRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type: schema.TypeList,
				MinItems: 1,
				Required: true,
				Elem: dataSourceLinodeInstanceFilter(),
			},
		},
	}
}
