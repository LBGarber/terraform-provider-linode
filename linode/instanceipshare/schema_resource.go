package instanceipshare

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var resourceSchema = map[string]*schema.Schema{
	"linode_id": {
		Type:          schema.TypeInt,
		Description:   "The ID of the Linode to assign this range to.",
		Optional:      true,
		ForceNew:      true,
	},
	"ip_addresses": {
		Type:          schema.TypeSet,
		Description:   "A set of IP addresses to share with this Linode. This can include IPv4 and IPv6 addresses.",
		Required: true,
		Elem: &schema.Schema{Type: schema.TypeString},
	},
}
