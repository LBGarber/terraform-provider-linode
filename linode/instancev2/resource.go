package instancev2

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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func readResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Error parsing Linode Instance ID %s as int: %s", d.Id(), err)
	}

	instance, err := client.GetInstance(ctx, id)
	if err != nil {
		if lerr, ok := err.(*linodego.Error); ok && lerr.Code == 404 {
			log.Printf("[WARN] removing Linode Instance ID %q from state because it no longer exists", d.Id())
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error finding the specified Linode Instance: %s", err)
	}

	instanceNetworking, err := client.GetInstanceIPAddresses(ctx, instance.ID)
	if err != nil {
		return diag.Errorf("failed to get networking for instance %d: %s", instance.ID, err)
	}

	d.Set("label", instance.Label)
	d.Set("region", instance.Region)
	d.Set("type", instance.Type)
	d.Set("backups_enabled", instance.Backups.Enabled)
	d.Set("group", instance.Group)
	d.Set("tags", instance.Tags)

	return nil
}

func createResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client
	label := d.Get("label").(string)
	clientConnThrottle := d.Get("client_conn_throttle").(int)

	createOpts := linodego.NodeBalancerCreateOptions{
		Region:             d.Get("region").(string),
		Label:              &label,
		ClientConnThrottle: &clientConnThrottle,
	}

	if tagsRaw, tagsOk := d.GetOk("tags"); tagsOk {
		for _, tag := range tagsRaw.(*schema.Set).List() {
			createOpts.Tags = append(createOpts.Tags, tag.(string))
		}
	}

	nodebalancer, err := client.CreateNodeBalancer(ctx, createOpts)
	if err != nil {
		return diag.Errorf("Error creating a Linode NodeBalancer: %s", err)
	}
	d.SetId(fmt.Sprintf("%d", nodebalancer.ID))

	return readResource(ctx, d, meta)
}

func updateResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.Errorf("Error parsing Linode NodeBalancer id %s as int: %s", d.Id(), err)
	}

	nodebalancer, err := client.GetNodeBalancer(ctx, int(id))
	if err != nil {
		return diag.Errorf("Error fetching data about the current NodeBalancer: %s", err)
	}

	if d.HasChanges("label", "client_conn_throttle", "tags") {
		label := d.Get("label").(string)
		clientConnThrottle := d.Get("client_conn_throttle").(int)

		// @TODO nodebalancer.GetUpdateOptions, avoid clobbering client_conn_throttle
		updateOpts := linodego.NodeBalancerUpdateOptions{
			Label:              &label,
			ClientConnThrottle: &clientConnThrottle,
		}

		tags := []string{}
		for _, tag := range d.Get("tags").(*schema.Set).List() {
			tags = append(tags, tag.(string))
		}

		updateOpts.Tags = &tags

		if nodebalancer, err = client.UpdateNodeBalancer(ctx, nodebalancer.ID, updateOpts); err != nil {
			return diag.FromErr(err)
		}
	}

	return readResource(ctx, d, meta)
}

func deleteResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*helper.ProviderMeta).Client
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.Errorf("Error parsing Linode NodeBalancer id %s as int", d.Id())
	}
	err = client.DeleteNodeBalancer(ctx, int(id))
	if err != nil {
		return diag.Errorf("Error deleting Linode NodeBalancer %d: %s", id, err)
	}
	return nil
}

func ResourceNodeBalancerV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"transfer": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func ResourceNodeBalancerV0Upgrade(ctx context.Context,
	rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	oldTransfer, ok := rawState["transfer"].(map[string]interface{})
	newTransfer := []map[string]interface{}{
		{
			"in":    0.0,
			"out":   0.0,
			"total": 0.0,
		},
	}
	rawState["transfer"] = newTransfer

	if !ok {
		// The transfer key does not exist; this is a computed map so it will be populated with the next
		// state refresh.
		return rawState, nil
	}

	for key, val := range oldTransfer {
		val := val.(string)

		// This is necessary because it is possible old versions of the state have empty transfer fields
		// that must default to zero.
		if val == "" {
			continue
		}

		result, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to upgrade state: %v", err)
		}

		newTransfer[0][key] = result
	}

	return rawState, nil
}
