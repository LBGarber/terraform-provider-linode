package ipv6range

import (
	"context"
	"fmt"
	"log"
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
		DeleteContext: deleteResource,
		UpdateContext: updateResource,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func readResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	r, err := client.GetIPv6Range(ctx, d.Id())
	if err != nil {
		log.Printf("[WARN] removing ipv6 range %q from state because it no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("prefix_length", r.Prefix)
	d.Set("is_bgp", r.IsBGP)
	d.Set("linodes", r.Linodes)
	d.Set("range", r.Range)
	d.Set("region", r.Region)

	return nil
}

func createResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	linodeID, linodeIDExists := d.GetOk("linode_id")
	routeTarget, routeTargetExists := d.GetOk("route_target")
	_, shouldShare := d.GetOk("shared_linodes")

	createOpts := linodego.IPv6RangeCreateOptions{
		PrefixLength: d.Get("prefix_length").(int),
	}

	if linodeIDExists {
		createOpts.LinodeID = linodeID.(int)
	} else if routeTargetExists {
		// Strip the prefix if provided
		createOpts.RouteTarget = strings.Split(routeTarget.(string), "/")[0]
	} else {
		return diag.Errorf("either linode_id or route_target must be specified")
	}

	r, err := client.CreateIPv6Range(ctx, createOpts)
	if err != nil {
		return diag.Errorf("failed to create ipv6 range: %s", err)
	}

	d.SetId(stripRangePrefix(r.Range))

	if shouldShare {
		if err := validateSharingConfig(ctx, d, meta); err != nil {
			return diag.Errorf("invalid sharing configuration: %s", err)
		}

		if err := shareResourceIPToLinodes(ctx, d, meta); err != nil {
			return diag.Errorf("failed to share ip: %s", err)
		}
	}

	return readResource(ctx, d, meta)
}

func updateResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	_, shouldShare := d.GetOk("shared_linodes")
	sharedLinodesOld, sharedLinodesNew := d.GetChange("shared_linodes")

	if d.HasChange("shared_linodes") && shouldShare {
		if err := validateSharingConfig(ctx, d, meta); err != nil {
			return diag.Errorf("invalid sharing configuration: %s", err)
		}

		for _, oldLinode := range sharedLinodesOld.(*schema.Set).List() {
			if sharedLinodesNew.(*schema.Set).Contains(oldLinode) {
				continue
			}

			if err := unshareInstanceIP(ctx, client, oldLinode.(int), d.Id()); err != nil {
				return diag.FromErr(err)
			}
		}

		if err := shareResourceIPToLinodes(ctx, d, meta); err != nil {
			return diag.Errorf("failed to share ip: %s", err)
		}
	}

	return readResource(ctx, d, meta)
}

func deleteResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	if err := client.DeleteIPv6Range(ctx, d.Id()); err != nil {
		return diag.Errorf("failed to delete ipv6 range %s: %s", d.Id(), err)
	}
	return nil
}

func stripRangePrefix(r string) string {
	return strings.Split(r, "/")[0]
}

func getLinodeSharedIPs(ctx context.Context, client linodego.Client, linodeID int) ([]string, error) {
	nwInfo, err := client.GetInstanceIPAddresses(ctx, linodeID)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, ip := range nwInfo.IPv4.Shared {
		result = append(result, ip.Address)
	}

	for _, ip := range nwInfo.IPv6.Global {
		if ip.RouteTarget != "" {
			continue
		}

		result = append(result, stripRangePrefix(ip.Range))
	}

	return result, nil
}

func filterSourceIP(ips []string, sourceIP string) []string {
	result := make([]string, 0)
	for _, ip := range ips {
		if ip == sourceIP {
			continue
		}

		result = append(result, ip)
	}

	return result
}

func unshareInstanceIP(ctx context.Context, client linodego.Client, instanceID int, ip string) error {
	sharedIPs, err := getLinodeSharedIPs(ctx, client, instanceID)
	if err != nil{
		return fmt.Errorf("failed to get shared ips for linode %d: %s", instanceID, err)
	}

	// Clear the shared IPs
	err = client.ShareIPAddresses(ctx, linodego.IPAddressesShareOptions{
		IPs:      []string{},
		LinodeID: instanceID,
	})
	if err != nil {
		return fmt.Errorf("failed to share ip addresses for linode %d: %s", instanceID, err)
	}

	// Restore other shared IPs
	err = client.ShareIPAddresses(ctx, linodego.IPAddressesShareOptions{
		IPs:      filterSourceIP(sharedIPs, ip),
		LinodeID: instanceID,
	})
	if err != nil {
		return fmt.Errorf("failed to share ip addresses for linode %d: %s", instanceID, err)
	}

	return nil
}

func shareResourceIPToLinodes(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	client := meta.(*helper.ProviderMeta).Client

	sharedLinodes := d.Get("shared_linodes").(*schema.Set).List()
	targetLinodeId := d.Get("linode_id").(int)

	err := client.ShareIPAddresses(ctx, linodego.IPAddressesShareOptions{
		IPs:      []string{stripRangePrefix(d.Id())},
		LinodeID: targetLinodeId,
	})
	if err != nil {
		return fmt.Errorf("failed to share ip addresses for target linode %d: %s", targetLinodeId, err)
	}

	for _, instanceID := range sharedLinodes {
		if instanceID == targetLinodeId {
			continue
		}

		err := client.ShareIPAddresses(ctx, linodego.IPAddressesShareOptions{
			IPs:      []string{stripRangePrefix(d.Id())},
			LinodeID: instanceID.(int),
		})
		if err != nil {
			return fmt.Errorf("failed to share ip addresses for linode %d: %s", instanceID, err)
		}
	}

	return nil
}

func validateSharingConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	lid, lidOk := d.GetOk("linode_id")
	sharedLinodes := d.Get("shared_linodes").(*schema.Set)

	if !lidOk {
		return fmt.Errorf("linode_id must be defined to share an ipv6 range")
	}

	if !sharedLinodes.Contains(lid) {
		return fmt.Errorf("shared_linodes must contain the target linode id")
	}

	return nil
}