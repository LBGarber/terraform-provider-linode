package firewalldevice

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/linode/helper"
	"log"
	"strconv"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema:        resourceSchema,
		ReadContext:   readResource,
		CreateContext: createResource,
		UpdateContext: updateResource,
		DeleteContext: deleteResource,
	}
}

func readResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.Errorf("Error parsing Linode Firewall ID %s as int: %s", d.Id(), err)
	}

	firewallID := d.Get("firewall_id").(int)

	device, err := client.GetFirewallDevice(ctx, firewallID, int(id))
	if err != nil {
		if lerr, ok := err.(*linodego.Error); ok && lerr.Code == 404 {
			log.Printf("[WARN] removing Firewall Device ID %q from state because it no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error finding the specified Linode Firewall Device: %s", err)
	}

	d.Set("created", device.Created.String())
	d.Set("updated", device.Updated.String())

	return nil
}

func createResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	firewallID := d.Get("firewall_id").(int)
	entityID := d.Get("entity_id").(int)
	entityType := d.Get("entity_type").(string)

	createOpts := linodego.FirewallDeviceCreateOptions{
		ID:   entityID,
		Type: linodego.FirewallDeviceType(entityType),
	}

	config, err := client.CreateFirewallDevice(ctx, firewallID, createOpts)
	if err != nil {
		return diag.Errorf("Error creating a Linode Firewall Device: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", config.ID))

	return readResource(ctx, d, meta)
}

func updateResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return readResource(ctx, d, meta)
}

func deleteResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.Errorf("Error parsing Linode Firewall Device ID %s as int: %s", d.Id(), err)
	}

	firewallID, ok := d.Get("firewall_id").(int)
	if !ok {
		return diag.Errorf("Error parsing Linode Firewall Device ID %v as int", d.Get("nodebalancer_id"))
	}

	err = client.DeleteFirewallDevice(ctx, firewallID, int(id))
	if err != nil {
		return diag.Errorf("Error deleting Linode Firewall Device %d: %s", id, err)
	}
	return nil
}