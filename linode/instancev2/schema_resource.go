package instancev2

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var resourceSchema = map[string]*schema.Schema{
	"label": {
		Type:        schema.TypeString,
		Description: "The label of this Linode instance.",
		Required:    true,
	},
	"region": {
		Type:        schema.TypeString,
		Description: "The region where this Linode instance will be located.",
		Required:    true,
	},
	"type": {
		Type:        schema.TypeString,
		Description: "The type of this Linode instance.",
		Required:    true,
	},

	"backups_enabled": {
		Type:        schema.TypeBool,
		Description: "If this field is set to true, the created Linode will automatically be enrolled in the Linode Backup service.",
		Optional:    true,
		Default:     false,
	},
	"group": {
		Type:        schema.TypeString,
		Description: "A deprecated property denoting a group label for this Linode.",
		Deprecated:  "Assigning a group is deprecated.",
		Optional:    true,
	},
	"tags": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "A set of tags applied to this object.",
		Optional:    true,
	},

	"created": {
		Type:        schema.TypeString,
		Description: "When this Instance was created.",
		Computed:    true,
	},
	"ipv4": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "This Linode's IPv4 Addresses.",
		Optional:    true,
	},
	"ipv6": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "This Linode's IPv4 Addresses.",
		Optional:    true,
	},
	"status": {
		Type:        schema.TypeString,
		Description: "A brief description of this Linode’s current state.",
		Computed:    true,
	},
	"updated": {
		Type:        schema.TypeString,
		Description: "When this Instance was last updated.",
		Computed:    true,
	},
}
