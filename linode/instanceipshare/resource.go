package instanceipshare

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/linode/helper"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema:        resourceSchema,
		ReadContext:   readResource,
		CreateContext: createResource,
		UpdateContext: updateResource,
		DeleteContext: deleteResource,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func readResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("failed to parse linode id: %s", err)
	}

	r, err := client.GetInstanceIPAddresses(ctx, id)
	if err != nil {
		log.Printf("[WARN] removing instance_shared_ips %q from state because the instance no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	ips := make([]string, 0)

	for _, ipv4 := range r.IPv4.Shared {
		ips = append(ips, ipv4.Address)
	}

	for _, ipv6 := range r.IPv6.Global {
		// Shared IPs will not have a route target defined
		if len(ipv6.RouteTarget) > 0 {
			continue
		}

		ips = append(ips, stripRangePrefix(ipv6.Range))
	}

	d.Set("linode_id", id)
	d.Set("ip_addresses", ips)

	return nil
}

func createResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	linodeID := d.Get("linode_id").(int)
	newIPs := d.Get("ip_addresses").(*schema.Set)
	newIPsNormalized := normalizeIPs(helper.ExpandStringSet(newIPs))

	err := client.ShareIPAddresses(ctx, linodego.IPAddressesShareOptions{
		IPs:      newIPsNormalized,
		LinodeID: linodeID,
	})
	if err != nil {
		return diag.Errorf("failed to share ip addresses for linode %d: %s", linodeID, err)
	}

	d.SetId(strconv.Itoa(linodeID))

	return readResource(ctx, d, meta)
}


func updateResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	linodeID := d.Get("linode_id").(int)

	if d.HasChange("ip_addresses") {
		newIPs := d.Get("ip_addresses").(*schema.Set)
		newIPsNormalized := normalizeIPs(helper.ExpandStringSet(newIPs))

		err := client.ShareIPAddresses(ctx, linodego.IPAddressesShareOptions{
			IPs:      []string{},
			LinodeID: linodeID,
		})
		if err != nil {
			return diag.Errorf("failed to clear ip addresses for linode %d: %s", linodeID, err)
		}

		err = client.ShareIPAddresses(ctx, linodego.IPAddressesShareOptions{
			IPs:      newIPsNormalized,
			LinodeID: linodeID,
		})
		if err != nil {
			return diag.Errorf("failed to share ip addresses for linode %d: %s", linodeID, err)
		}
	}

	return readResource(ctx, d, meta)
}

func deleteResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	linodeID := d.Get("linode_id").(int)

	err := client.ShareIPAddresses(ctx, linodego.IPAddressesShareOptions{
		IPs:      []string{},
		LinodeID: linodeID,
	})
	if err != nil {
		return diag.Errorf("failed to share ip addresses for linode %d: %s", linodeID, err)
	}

	return nil
}

func stripRangePrefix(r string) string {
	return strings.Split(r, "/")[0]
}

func normalizeIPs(input []string) []string {
	result := make([]string, len(input))

	for i, ip := range input {
		result[i] = stripRangePrefix(ip)
	}

	return result
}
