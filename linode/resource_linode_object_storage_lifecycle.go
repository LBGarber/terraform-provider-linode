package linode

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceLinodeObjectStorageLifecycleExpiration() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"expired_object_delete_marker": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceLinodeObjectStorageLifecycleNoncurrentExp() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"days": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceLinodeObjectStorageLifecycleRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"abort_incomplete_multipart_upload_days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"expiration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     resourceLinodeObjectStorageLifecycleExpiration(),
			},
			"noncurrent_version_expiration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     resourceLinodeObjectStorageLifecycleNoncurrentExp(),
			},
		},
	}
}

func resourceLinodeObjectStorageLifecycle() *schema.Resource {
	return &schema.Resource{
		Create: resourceLinodeObjectStorageLifecycleCreate,
		Read:   resourceLinodeObjectStorageLifecycleRead,
		Update: resourceLinodeObjectStorageLifecycleUpdate,
		Delete: resourceLinodeObjectStorageLifecycleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secret_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"access_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lifecycle_rule": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem:     resourceLinodeObjectStorageLifecycleRule(),
			},
		},
	}
}

func resourceLinodeObjectStorageLifecycleRead(d *schema.ResourceData, meta interface{}) error {
	bucket := d.Get("bucket").(string)

	client := s3ConnFromResourceData(d)

	lifecycleConfigInput := &s3.GetBucketLifecycleConfigurationInput{
		Bucket: &bucket,
	}

	lifecycleConfig, err := client.GetBucketLifecycleConfiguration(lifecycleConfigInput)
	if err != nil {
		return fmt.Errorf("failed to get lifecycle configuration %s: %s", bucket, err)
	}

	rules := flattenLifecycleRules(lifecycleConfig.Rules)

	d.SetId(bucket)
	d.Set("lifecycle_rule", rules)

	return nil
}

func resourceLinodeObjectStorageLifecycleCreate(d *schema.ResourceData, meta interface{}) error {
	bucket := d.Get("bucket").(string)

	client := s3ConnFromResourceData(d)

	lifecycleConfig := &s3.BucketLifecycleConfiguration{}

	rules, err := expandLifecycleRules(d.Get("lifecycle_rule").([]interface{}))
	if err != nil {
		return fmt.Errorf("failed to parse lifecycle rules: %s", err)
	}

	lifecycleConfig.Rules = rules

	inputConfig := &s3.PutBucketLifecycleConfigurationInput{
		Bucket:                 &bucket,
		LifecycleConfiguration: lifecycleConfig,
	}

	if _, err = client.PutBucketLifecycleConfiguration(inputConfig); err != nil {
		return err
	}

	if err := resourceLinodeObjectStorageLifecycleRead(d, meta); err != nil {
		return err
	}

	return nil
}

func resourceLinodeObjectStorageLifecycleUpdate(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceLinodeObjectStorageLifecycleDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func expandLifecycleRules(ruleSpecs []interface{}) ([]*s3.LifecycleRule, error) {
	rules := make([]*s3.LifecycleRule, len(ruleSpecs))
	for i, ruleSpec := range ruleSpecs {
		ruleSpec := ruleSpec.(map[string]interface{})
		rule := &s3.LifecycleRule{}

		status := "Disabled"
		if ruleSpec["enabled"].(bool) {
			status = "Enabled"
		}
		rule.Status = &status

		if id, ok := ruleSpec["id"]; ok {
			id := id.(string)
			rule.ID = &id
		}

		if prefix, ok := ruleSpec["prefix"]; ok {
			prefix := prefix.(string)
			rule.Prefix = &prefix
		}

		if abortIncompleteDays, ok := ruleSpec["abort_incomplete_multipart_upload_days"].(int); ok && abortIncompleteDays > 0 {
			rule.AbortIncompleteMultipartUpload = &s3.AbortIncompleteMultipartUpload{}
			abortIncompleteDays := int64(abortIncompleteDays)

			rule.AbortIncompleteMultipartUpload.DaysAfterInitiation = &abortIncompleteDays
		}

		if expirationList := ruleSpec["expiration"].([]interface{}); len(expirationList) > 0 {
			rule.Expiration = &s3.LifecycleExpiration{}

			expirationMap := expirationList[0].(map[string]interface{})

			if dateStr, ok := expirationMap["date"].(string); ok && dateStr != "" {
				date, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", dateStr))
				if err != nil {
					return nil, err
				}

				rule.Expiration.Date = &date
			}

			if days, ok := expirationMap["days"].(int); ok && days > 0 {
				days := int64(days)

				rule.Expiration.Days = &days
			}

			if marker, ok := expirationMap["expired_object_delete_marker"].(bool); ok && marker {
				rule.Expiration.ExpiredObjectDeleteMarker = &marker
			}
		}

		if expirationList := ruleSpec["noncurrent_version_expiration"].([]interface{}); len(expirationList) > 0 {
			rule.NoncurrentVersionExpiration = &s3.NoncurrentVersionExpiration{}

			expirationMap := expirationList[0].(map[string]interface{})

			if days, ok := expirationMap["days"]; ok {
				days := int64(days.(int))
				rule.NoncurrentVersionExpiration.NoncurrentDays = &days
			}
		}

		rules[i] = rule
	}

	return rules, nil
}

func flattenLifecycleRules(rules []*s3.LifecycleRule) []map[string]interface{} {
	result := make([]map[string]interface{}, len(rules))

	for i, rule := range rules {
		ruleMap := make(map[string]interface{})

		if id := rule.ID; id != nil {
			ruleMap["id"] = *id
		}

		if prefix := rule.Prefix; prefix != nil {
			ruleMap["prefix"] = *prefix
		}

		if status := rule.Status; status != nil {
			ruleMap["enabled"] = *status == "Enabled"
		}

		if rule.AbortIncompleteMultipartUpload != nil {
			ruleMap["abort_incomplete_multipart_upload_days"] = *rule.AbortIncompleteMultipartUpload.DaysAfterInitiation
		}

		if rule.Expiration != nil {
			e := make(map[string]interface{})

			if date := rule.Expiration.Date; date != nil {
				e["date"] = rule.Expiration.Date.Format("2006-01-02")
			}

			if days := rule.Expiration.Days; days != nil {
				e["days"] = *days
			}

			if marker := rule.Expiration.ExpiredObjectDeleteMarker; marker != nil && *marker {
				e["expired_object_delete_marker"] = *marker
			}

			ruleMap["expiration"] = []interface{}{e}
		}

		if rule.NoncurrentVersionExpiration != nil {
			e := make(map[string]interface{})

			if days := rule.NoncurrentVersionExpiration.NoncurrentDays; days != nil && *days > 0 {
				e["days"] = *days
				ruleMap["noncurrent_version_expiration"] = []interface{}{e}
			}
		}

		result[i] = ruleMap
	}

	return result
}
