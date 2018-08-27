package linode

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/linode/linodego"
)

func resourceLinodeInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceLinodeInstanceCreate,
		Read:   resourceLinodeInstanceRead,
		Update: resourceLinodeInstanceUpdate,
		Delete: resourceLinodeInstanceDelete,
		Exists: resourceLinodeInstanceExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"image": &schema.Schema{
				Type:        schema.TypeString,
				Description: "An Image ID to deploy the Disk from. Official Linode Images start with linode/, while your Images start with private/. See /images for more information on the Images available for you to use.",
				Optional:    true,
				ForceNew:    true,
			},
			"backup_id": &schema.Schema{
				Type:          schema.TypeInt,
				Description:   "A Backup ID from another Linode's available backups. Your User must have read_write access to that Linode, the Backup must have a status of successful, and the Linode must be deployed to the same region as the Backup. See /linode/instances/{linodeId}/backups for a Linode's available backups. This field and the image field are mutually exclusive.",
				ConflictsWith: []string{"disk", "config"},
				Optional:      true,
				ForceNew:      true,
			},
			"stackscript_id": &schema.Schema{
				Type:          schema.TypeInt,
				Description:   "The StackScript to deploy to the newly created Linode. If provided, 'image' must also be provided, and must be an Image that is compatible with this StackScript.",
				ConflictsWith: []string{"disk"},
				Optional:      true,
				ForceNew:      true,
			},
			"stackscript_data": &schema.Schema{
				Type: schema.TypeMap,
				Elem: &schema.Schema{Type: schema.TypeString},

				Description: "An object containing responses to any User Defined Fields present in the StackScript being deployed to this Linode. Only accepted if 'stackscript_id' is given. The required values depend on the StackScript being deployed.",
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"label": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The Linode's label is for display purposes only. If no label is provided for a Linode, a default will be assigned",
				Optional:    true,
				Computed:    true,
			},
			"group": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The display group of the Linode instance.",
				Optional:    true,
			},
			"boot_config_label": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The Label of the Instance Config that should be used to boot the Linode instance.",
				Optional:    true,
				Computed:    true,
			},
			"region": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "This is the location where the Linode was deployed. This cannot be changed without opening a support ticket.",
				Required:     true,
				ForceNew:     true,
				InputDefault: "us-east",
			},
			"type": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The type of instance to be deployed, determining the price and size.",
				Optional:    true,
				Default:     "g6-standard-1",
			},
			"status": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The status of the instance, indicating the current readiness state.",
				Computed:    true,
			},
			"ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6": &schema.Schema{
				Type:        schema.TypeString,
				Description: "This Linode's IPv6 SLAAC addresses. This address is specific to a Linode, and may not be shared.",
				Computed:    true,
			},

			"ipv4": &schema.Schema{
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "This Linode's IPv4 Addresses. Each Linode is assigned a single public IPv4 address upon creation, and may get a single private IPv4 address if needed. You may need to open a support ticket to get additional IPv4 addresses.",
				Computed:    true,
			},

			"private_ip": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "If true, the created Linode will have private networking enabled, allowing use of the 192.168.0.0/17 network within the Linode's region.",
				Optional:    true,
			},
			"private_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorized_keys": &schema.Schema{
				Type:          schema.TypeList,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Description:   "A list of SSH public keys to deploy for the root user on the newly created Linode. Only accepted if 'image' is provided.",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"disk"},
				StateFunc:     sshKeyState,
				PromoteSingle: true,
			},
			"root_pass": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "The password that will be initialially assigned to the 'root' user account.",
				Sensitive:     true,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"disk"},
				StateFunc:     rootPasswordState,
			},
			"swap_size": &schema.Schema{
				Type:          schema.TypeInt,
				Description:   "When deploying from an Image, this field is optional with a Linode API default of 512mb, otherwise it is ignored. This is used to set the swap disk size for the newly-created Linode.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"disk"},
				Default:       nil,
			},
			"backups_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "If this field is set to true, the created Linode will automatically be enrolled in the Linode Backup service. This will incur an additional charge. The cost for the Backup service is dependent on the Type of Linode deployed.",
				Optional:    true,
				Computed:    true,
				Default:     nil,
			},
			"watchdog_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "The watchdog, named Lassie, is a Shutdown Watchdog that monitors your Linode and will reboot it if it powers off unexpectedly. It works by issuing a boot job when your Linode powers off without a shutdown job being responsible. To prevent a loop, Lassie will give up if there have been more than 5 boot jobs issued within 15 minutes.",
				Optional:    true,
				Default:     false,
			},
			"specs": &schema.Schema{
				Computed: true,
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"memory": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vcpus": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"transfer": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"alerts": &schema.Schema{
				Computed: true,
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"network_in": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"network_out": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"transfer_quota": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"io": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"backups": &schema.Schema{
				Computed: true,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"schedule": {
							Type: schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"day": {
										Type:     schema.TypeString,
										Required: true,
									},
									"window": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
			"config": &schema.Schema{
				Computed: true,
				Optional: true,
				Type:     schema.TypeSet,
				Set:      labelHashcode,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": {
							Type:     schema.TypeString,
							Required: true,
						},
						"helpers": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"updatedb_disabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"distro": {
										Type:        schema.TypeBool,
										Description: "Controls the behavior of the Linode Config's Distribution Helper setting.",
										Optional:    true,
										Default:     true,
									},
									"modules_dep": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"network": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Controls the behavior of the Linode Config's Network Helper setting, used to automatically configure additional IP addresses assigned to this instance.",
										Default:     true,
									},
									"devtmpfs_automount": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"devices": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sda": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_label": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"disk_id": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"volume_id": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"sdb": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_label": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"disk_id": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"volume_id": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"sdc": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_label": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"disk_id": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"volume_id": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"sdd": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_label": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"disk_id": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"volume_id": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"sde": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_label": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"disk_id": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"volume_id": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"sdf": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_label": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"disk_id": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"volume_id": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"sdg": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_label": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"disk_id": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"volume_id": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"sdh": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"disk_label": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"disk_id": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"volume_id": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"kernel": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A Kernel ID to boot a Linode with. Default is based on image choice. (examples: linode/latest-64bit, linode/grub2, linode/direct-disk)",
						},
						"run_level": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Defines the state of your Linode after booting. Defaults to default.",
							Default:     "default",
						},
						"virt_mode": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Controls the virtualization mode. Defaults to paravirt.",
							Default:     "paravirt",
						},
						"root_device": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The root device to boot. The corresponding disk must be attached.",
							Default:     "/dev/sda",
						},
						"comments": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional field for arbitrary User comments on this Config.",
						},

						"memory_limit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Defaults to the total RAM of the Linode",
						},
					},
				},
			},
			"disk": &schema.Schema{
				Computed:      true,
				PromoteSingle: true,
				Optional:      true,
				Type:          schema.TypeSet,
				Set:           labelHashcode,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"label": {
							Type:     schema.TypeString,
							Required: true,
						},
						"filesystem": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "ext4",
						},
						"read_only": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"image": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"authorized_keys": {
							Type:          schema.TypeList,
							Elem:          &schema.Schema{Type: schema.TypeString},
							Description:   "A list of SSH public keys to deploy for the root user on the newly created Linode. Only accepted if 'image' is provided.",
							Optional:      true,
							ForceNew:      true,
							StateFunc:     sshKeyState,
							PromoteSingle: true,
						},
						"stackscript_id": &schema.Schema{
							Type:        schema.TypeInt,
							Description: "The StackScript to deploy to the newly created Linode. If provided, 'image' must also be provided, and must be an Image that is compatible with this StackScript.",
							Optional:    true,
							ForceNew:    true,
						},
						"stackscript_data": &schema.Schema{
							Type:        schema.TypeMap,
							Description: "An object containing responses to any User Defined Fields present in the StackScript being deployed to this Linode. Only accepted if 'stackscript_id' is given. The required values depend on the StackScript being deployed.",
							Optional:    true,
							ForceNew:    true,
							Sensitive:   true,
						},
						"root_pass": &schema.Schema{
							Type:        schema.TypeString,
							Description: "The password that will be initialially assigned to the 'root' user account.",
							Sensitive:   true,
							Optional:    true,
							ForceNew:    true,
							StateFunc:   rootPasswordState,
						},
					},
				},
			},
		},
	}
}

func resourceLinodeInstanceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(linodego.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return false, fmt.Errorf("Error parsing Linode instance ID %s as int: %s", d.Id(), err)
	}

	_, err = client.GetInstance(context.Background(), int(id))
	if err != nil {
		if lerr, ok := err.(*linodego.Error); ok && lerr.Code == 404 {
			return false, nil
		}

		return false, fmt.Errorf("Error getting Linode Instance %s: %s", d.Id(), err)
	}
	return true, nil
}

func resourceLinodeInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(linodego.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("Error parsing Linode instance ID %s as int: %s", d.Id(), err)
	}

	instance, err := client.GetInstance(context.Background(), int(id))

	if err != nil {
		if lerr, ok := err.(linodego.Error); ok && lerr.Code == 404 {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error finding the specified Linode instance: %s", err)
	}

	instanceNetwork, err := client.GetInstanceIPAddresses(context.Background(), int(id))

	if err != nil {
		return fmt.Errorf("Error getting the IPs for Linode instance %s: %s", d.Id(), err)
	}

	var ips []string
	for _, ip := range instance.IPv4 {
		ips = append(ips, ip.String())
	}
	d.Set("ipv4", ips)
	public, private := instanceNetwork.IPv4.Public, instanceNetwork.IPv4.Private

	if len(public) > 0 {
		d.Set("ip_address", public[0].Address)

		d.SetConnInfo(map[string]string{
			"type": "ssh",
			"host": public[0].Address,
		})
		// TODO(displague) to determine 'user', need to check disk.image
		// "linode/containerlinux" is "core", else "root"
		// might be better to make this a resource field and avoid lookups
	}

	if len(private) > 0 {
		d.Set("private_ip", true)
		d.Set("private_ip_address", private[0].Address)
	} else {
		d.Set("private_ip", false)
	}

	d.Set("label", instance.Label)
	d.Set("status", instance.Status)
	d.Set("type", instance.Type)
	d.Set("region", instance.Region)

	d.Set("group", instance.Group)

	// panic: interface conversion: interface {} is map[string]int, not *schema.Set

	flatSpecs := flattenInstanceSpecs(*instance)
	flatAlerts := flattenInstanceAlerts(*instance)

	if err := d.Set("specs", flatSpecs); err != nil {
		return fmt.Errorf("Error setting Linode Instance specs: %s", err)
	}

	if err := d.Set("alerts", flatAlerts); err != nil {
		return fmt.Errorf("Error setting Linode Instance alerts: %s", err)
	}

	instanceDisks, err := client.ListInstanceDisks(context.Background(), int(id), nil)

	if err != nil {
		return fmt.Errorf("Error getting the disks for the Linode instance %d: %s", id, err)
	}

	disks, swapSize := flattenInstanceDisks(instanceDisks)

	if err := d.Set("disk", disks); err != nil {
		return fmt.Errorf("Erroring setting Linode Instance disk: %s", err)
	}

	d.Set("swap_size", swapSize)

	instanceConfigs, err := client.ListInstanceConfigs(context.Background(), int(id), nil)

	if err != nil {
		return fmt.Errorf("Error getting the config for Linode instance %d (%s): %s", instance.ID, instance.Label, err)
	} else if len(instanceConfigs) == 0 {
		return nil
	}

	configs := flattenInstanceConfigs(instanceConfigs)
	if err := d.Set("config", configs); err != nil {
		return fmt.Errorf("Erroring setting Linode Instance config: %s", err)
	}

	if len(instanceConfigs) == 1 {
		d.Set("boot_config_label", instanceConfigs[0].Label)
	}

	return nil
}

func resourceLinodeInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client, ok := meta.(linodego.Client)
	if !ok {
		return fmt.Errorf("Invalid Client when creating Linode Instance")
	}
	d.Partial(true)

	bootConfig := 0
	createOpts := linodego.InstanceCreateOptions{
		Region:         d.Get("region").(string),
		Type:           d.Get("type").(string),
		Label:          d.Get("label").(string),
		Group:          d.Get("group").(string),
		BackupsEnabled: d.Get("backups_enabled").(bool),
		PrivateIP:      d.Get("private_ip").(bool),
	}

	_, disksOk := d.GetOk("disk")
	_, configsOk := d.GetOk("config")

	// If we don't have disks and we don't have configs, use the single API call approach
	if !disksOk && !configsOk {
		for _, key := range d.Get("authorized_keys").([]interface{}) {
			createOpts.AuthorizedKeys = append(createOpts.AuthorizedKeys, key.(string))
		}
		createOpts.RootPass = d.Get("root_pass").(string)
		createOpts.Image = d.Get("image").(string)
		createOpts.BackupID = d.Get("backup_id").(int)
		if swapSize := d.Get("swap_size").(int); swapSize > 0 {
			createOpts.SwapSize = &swapSize
		}

		createOpts.StackScriptID = d.Get("stackscript_id").(int)

		for name, value := range d.Get("stackscript_data").(map[string]interface{}) {
			createOpts.StackScriptData[name] = value.(string)
		}
	} else {
		createOpts.Booted = &boolFalse // necessary to prepare disks and configs
	}

	instance, err := client.CreateInstance(context.Background(), createOpts)
	if err != nil {
		return fmt.Errorf("Error creating a Linode Instance: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", instance.ID))

	// d.Set("backups_enabled", instance.BackupsEnabled)

	d.SetPartial("private_ip")
	d.SetPartial("authorized_keys")
	d.SetPartial("root_pass")
	d.SetPartial("kernel")
	d.SetPartial("image")
	d.SetPartial("backup_id")
	d.SetPartial("stackscript_id")
	d.SetPartial("stackscript_data")
	d.SetPartial("swap_size")

	var ips []string
	for _, ip := range instance.IPv4 {
		ips = append(ips, ip.String())
	}

	d.Set("ipv4", ips)
	d.Set("ipv6", instance.IPv6)

	for _, address := range instance.IPv4 {
		if private := privateIP(*address); private {
			d.Set("private_ip_address", address.String())
		} else {
			d.Set("ip_address", address.String())
		}
	}

	/*
		if d.Get("private_networking").(bool) {
			resp, err := client.AddInstanceIPAddress(context.Background(), instance.ID, false)
			if err != nil {
				return fmt.Errorf("Error adding a private ip address to Linode instance %d: %s", instance.ID, err)
			}
			d.Set("private_ip_address", resp.Address)
			d.SetPartial("private_ip_address")
		}
	*/

	// Look up tables for any disks and configs we create
	// - so configs and initrd can reference disks by label
	// - so configs can be referenced as a boot_config_label param
	var diskIDLabelMap map[string]int
	var configIDLabelMap map[string]int
	var configDevices linodego.InstanceConfigDeviceMap

	if disksOk {
		_, err = client.WaitForEventFinished(context.Background(), instance.ID, linodego.EntityLinode, linodego.ActionLinodeCreate, *instance.Created, int(d.Timeout(schema.TimeoutCreate).Seconds()))
		if err != nil {
			return fmt.Errorf("Error waiting for Instance to finish creating")
		}

		// TODO(displague) over 8 disks is a problem
		dset := d.Get("disk").(*schema.Set)
		diskIDLabelMap = make(map[string]int, len(dset.List()))
		for index, v := range dset.List() {
			disk, ok := v.(map[string]interface{})

			if !ok {
				return fmt.Errorf("Error converting disk from Terraform Set to golang map")
			}

			diskOpts := linodego.InstanceDiskCreateOptions{
				Label:      disk["label"].(string),
				Filesystem: disk["filesystem"].(string),
				Size:       disk["size"].(int),
			}

			if image, ok := disk["image"]; ok {
				diskOpts.Image = image.(string)

				if rootPass, ok := disk["root_pass"]; ok {
					diskOpts.RootPass = rootPass.(string)
				}

				if authorizedKeys, ok := disk["authorized_keys"]; ok {
					for _, sshKey := range authorizedKeys.([]interface{}) {
						diskOpts.AuthorizedKeys = append(diskOpts.AuthorizedKeys, sshKey.(string))
					}
				}

				if stackscriptID, ok := disk["stackscript_id"]; ok {
					diskOpts.StackscriptID = stackscriptID.(int)
				}

				if stackscriptData, ok := disk["stackscript_data"]; ok {
					for name, value := range stackscriptData.(map[string]interface{}) {
						diskOpts.StackscriptData[name] = value.(string)
					}
				}

				/*
					if sshKeys, ok := d.GetOk("authorized_keys"); ok {
						if sshKeysArr, ok := sshKeys.([]interface{}); ok {
							diskOpts.AuthorizedKeys = make([]string, len(sshKeysArr))
							for k, v := range sshKeys.([]interface{}) {
								if val, ok := v.(string); ok {
									diskOpts.AuthorizedKeys[k] = val
								}
							}
						}
					}
				*/
			}

			instanceDisk, err := client.CreateInstanceDisk(context.Background(), instance.ID, diskOpts)

			if err != nil {
				return fmt.Errorf("Error creating Linode instance %d disk: %s", instance.ID, err)
			}

			_, err = client.WaitForEventFinished(context.Background(), instance.ID, linodego.EntityLinode, linodego.ActionDiskCreate, instanceDisk.Created, int(d.Timeout(schema.TimeoutCreate).Seconds()))
			if err != nil {
				return fmt.Errorf("Error waiting for Linode instance %d disk: %s", instanceDisk.ID, err)
			}

			diskIDLabelMap[diskOpts.Label] = instanceDisk.ID

			// if err := d.Set(fmt.Sprintf("disk.%d.id", index), instanceDisk.ID); err != nil {
			//	return fmt.Errorf("Error setting Linode Disk ID: %s", err)
			// }

			if index == 0 {
				configDevices.SDA = &linodego.InstanceConfigDevice{DiskID: instanceDisk.ID}
			} else if index == 1 {
				configDevices.SDB = &linodego.InstanceConfigDevice{DiskID: instanceDisk.ID}
			} else if index == 2 {
				configDevices.SDC = &linodego.InstanceConfigDevice{DiskID: instanceDisk.ID}
			} else if index == 3 {
				configDevices.SDD = &linodego.InstanceConfigDevice{DiskID: instanceDisk.ID}
			} else if index == 4 {
				configDevices.SDE = &linodego.InstanceConfigDevice{DiskID: instanceDisk.ID}
			} else if index == 5 {
				configDevices.SDF = &linodego.InstanceConfigDevice{DiskID: instanceDisk.ID}
			} else if index == 6 {
				configDevices.SDG = &linodego.InstanceConfigDevice{DiskID: instanceDisk.ID}
			} else if index == 7 {
				configDevices.SDH = &linodego.InstanceConfigDevice{DiskID: instanceDisk.ID}
			}
		}
	}

	if !configsOk {
		// TODO(displague)  should we really create a config if not specfically defined? probably not

		configOpts := linodego.InstanceConfigCreateOptions{
			Label: fmt.Sprintf("linode%d-config", instance.ID),
			// RootDevice: "/dev/sda",
			// RunLevel:   "default",
			// VirtMode:   "paravirt",
			Devices: configDevices,
		}

		instanceConfig, err := client.CreateInstanceConfig(context.Background(), instance.ID, configOpts)
		if err != nil {
			return fmt.Errorf("Error creating Linode instance %d config: %s", instance.ID, err)
		}
		bootConfig = instanceConfig.ID
		configIDLabelMap = make(map[string]int, 1)
		configIDLabelMap[configOpts.Label] = instanceConfig.ID
	} else {
		cset := d.Get("config").(*schema.Set)
		configIDLabelMap = make(map[string]int, len(cset.List()))
		for _, v := range cset.List() {
			config, ok := v.(map[string]interface{})

			if !ok {
				return fmt.Errorf("Error parsing configs  %#v ... %#v", v, cset)
			}

			configOpts := linodego.InstanceConfigCreateOptions{}
			configOpts.Kernel = config["kernel"].(string)
			configOpts.Label = config["label"].(string)
			configOpts.Comments = config["comments"].(string)
			// configOpts.InitRD = config["initrd"].(string)
			// TODO(displague) need a disk_label to initrd lookup?
			devices, _ := config["devices"].([]interface{})
			// TODO(displague) ok needed? check it
			for _, device := range devices {
				for k, rdev := range device.(map[string]interface{}) {
					devSlots := rdev.([]interface{})
					for _, rrdev := range devSlots {
						dev := rrdev.(map[string]interface{})
						if k == "sda" {
							configOpts.Devices.SDA = &linodego.InstanceConfigDevice{}
							if err := assignConfigDevice(configOpts.Devices.SDA, dev, diskIDLabelMap); err != nil {
								return err
							}
						}
						if k == "sdb" {
							configOpts.Devices.SDB = &linodego.InstanceConfigDevice{}
							if err := assignConfigDevice(configOpts.Devices.SDB, dev, diskIDLabelMap); err != nil {
								return err
							}
						}
						if k == "sdc" {
							configOpts.Devices.SDC = &linodego.InstanceConfigDevice{}
							if err := assignConfigDevice(configOpts.Devices.SDC, dev, diskIDLabelMap); err != nil {
								return err
							}
						}
						if k == "sdd" {
							configOpts.Devices.SDD = &linodego.InstanceConfigDevice{}
							if err := assignConfigDevice(configOpts.Devices.SDD, dev, diskIDLabelMap); err != nil {
								return err
							}
						}
						if k == "sde" {
							configOpts.Devices.SDE = &linodego.InstanceConfigDevice{}

							if err := assignConfigDevice(configOpts.Devices.SDE, dev, diskIDLabelMap); err != nil {
								return err
							}
						}
						if k == "sdf" {
							configOpts.Devices.SDF = &linodego.InstanceConfigDevice{}

							if err := assignConfigDevice(configOpts.Devices.SDF, dev, diskIDLabelMap); err != nil {
								return err
							}
						}
						if k == "sdg" {
							configOpts.Devices.SDG = &linodego.InstanceConfigDevice{}
							if err := assignConfigDevice(configOpts.Devices.SDG, dev, diskIDLabelMap); err != nil {
								return err
							}
						}
						if k == "sdh" {
							configOpts.Devices.SDH = &linodego.InstanceConfigDevice{}
							if err := assignConfigDevice(configOpts.Devices.SDH, dev, diskIDLabelMap); err != nil {
								return err
							}
						}
					}
				}
			}
			instanceConfig, err := client.CreateInstanceConfig(context.Background(), instance.ID, configOpts)
			if err != nil {
				return fmt.Errorf("Error creating Instance Config: %s", err)
			}
			if len(configIDLabelMap) == 1 {
				bootConfig = instanceConfig.ID
			}

			configIDLabelMap[configOpts.Label] = instanceConfig.ID
		}
	}

	d.Partial(false)

	if createOpts.Booted != nil && *createOpts.Booted {
		if err = client.BootInstance(context.Background(), instance.ID, bootConfig); err != nil {
			return fmt.Errorf("Error booting Linode instance %d: %s", instance.ID, err)
		}

		if _, err = client.WaitForInstanceStatus(context.Background(), instance.ID, linodego.InstanceRunning, int(d.Timeout(schema.TimeoutCreate).Seconds())); err != nil {
			return fmt.Errorf("Timed-out waiting for Linode instance %d to boot: %s", instance.ID, err)
		}
	}

	return resourceLinodeInstanceRead(d, meta)
}

func resourceLinodeInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(linodego.Client)
	d.Partial(true)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("Error parsing Linode Instance ID %s as int: %s", d.Id(), err)
	}

	instance, err := client.GetInstance(context.Background(), int(id))
	if err != nil {
		return fmt.Errorf("Error fetching data about the current linode: %s", err)
	}

	if d.HasChange("label") {
		if instance, err = client.RenameInstance(context.Background(), instance.ID, d.Get("label").(string)); err != nil {
			return err
		}
		d.Set("label", instance.Label)
		d.SetPartial("label")
	}

	rebootInstance := false

	if d.HasChange("type") {
		err = changeInstanceType(&client, instance, d.Get("type").(string), d)
		if err != nil {
			return err
		}
	}

	if d.HasChange("private_ip") {
		if !d.Get("private_ip").(bool) {
			return fmt.Errorf("Can't deactivate private networking for linode %s", d.Id())
		}

		resp, err := client.AddInstanceIPAddress(context.Background(), int(id), false)

		if err != nil {
			return fmt.Errorf("Error activating private networking on linode %s: %s", d.Id(), err)
		}
		d.SetPartial("private_ip")
		d.Set("private_ip_address", resp.Address)
		d.SetPartial("private_ip_address")
		rebootInstance = true
	}

	configs, err := client.ListInstanceConfigs(context.Background(), int(id), nil)
	if err != nil {
		return fmt.Errorf("Error fetching the config for linode %d: %s", id, err)
	}
	if len(configs) != 1 {
		return fmt.Errorf("Linode %d has an incorrect number of configs %d, this plugin can only handle 1", id, len(configs))
	}
	config := configs[0].GetUpdateOptions()
	updateConfig := false
	if d.HasChange("helper_distro") {
		updateConfig = true
		config.Helpers.Distro = d.Get("helper_distro").(bool)
	}
	if d.HasChange("helper_network") {
		updateConfig = true
		config.Helpers.Network = d.Get("helper_network").(bool)
	}
	if d.HasChange("kernel") {
		updateConfig = true
		config.Kernel = d.Get("kernel").(string)
	}

	if updateConfig {
		_, err := client.UpdateInstanceConfig(context.Background(), instance.ID, configs[0].ID, config)
		if err != nil {
			return fmt.Errorf("Error updating Linode %d config: %s", instance.ID, err)
		}
		d.SetPartial("helper_distro")
		d.SetPartial("helper_network")
		d.SetPartial("kernel")

		rebootInstance = true
	}

	if rebootInstance {
		err = client.RebootInstance(context.Background(), instance.ID, configs[0].ID)
		if err != nil {
			return fmt.Errorf("Error rebooting Linode instance %d: %s", instance.ID, err)
		}

		_, err = client.WaitForEventFinished(context.Background(), id, linodego.EntityLinode, linodego.ActionLinodeReboot, *instance.Created, int(d.Timeout(schema.TimeoutCreate).Seconds()))
		if err != nil {
			return fmt.Errorf("Error waiting for Linode instance %d to finish rebooting: %s", instance.ID, err)
		}
	}

	return nil // resourceLinodeInstanceRead(d, meta)
}

func resourceLinodeInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(linodego.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("Error parsing Linode Instance ID %s as int", d.Id())
	}
	minDelete := time.Now().AddDate(0, 0, -1)
	err = client.DeleteInstance(context.Background(), int(id))
	if err != nil {
		return fmt.Errorf("Error deleting Linode instance %d: %s", id, err)
	}
	// Wait for full deletion to assure volumes are detached
	client.WaitForEventFinished(context.Background(), int(id), linodego.EntityLinode, linodego.ActionLinodeDelete, minDelete, int(d.Timeout(schema.TimeoutDelete).Seconds()))

	d.SetId("")
	return nil
}
