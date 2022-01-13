package ipv6range

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/linode/helper"
	"log"
	"strings"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema:        resourceSchema,
		ReadContext:   readResource,
		CreateContext: createResource,
		DeleteContext: deleteResource,
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

	createOpts := linodego.IPv6RangeCreateOptions{
		PrefixLength: d.Get("prefix_length").(int),
	}

	if linodeID, ok := d.GetOk("linode_id"); ok {
		createOpts.LinodeID = linodeID.(int)
	} else {
		createOpts.RouteTarget = d.Get("route_target").(string)
	}

	r, err := client.CreateIPv6Range(ctx, createOpts)
	if err != nil {
		return diag.Errorf("failed to create ipv6 range: %s", err)
	}

	d.SetId(strings.TrimSuffix(r.Range, fmt.Sprintf("/%d", createOpts.PrefixLength)))

	return readResource(ctx, d, meta)
}

func deleteResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	if err := client.DeleteIPv6Range(ctx, d.Id()); err != nil {
		return diag.Errorf("failed to delete ipv6 range %s: %s", d.Id(), err)
	}
	return nil
}
