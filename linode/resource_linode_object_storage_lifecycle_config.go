package linode

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceLinodeObjectStorageLifecycleConfigExpiration() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"date": {
				Type: schema.TypeString,
				Optional: true,
			},
			"days": {
				Type: schema.TypeInt,
				Optional: true,
			},
			"expired_object_delete_marker": {
				Type: schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceLinodeObjectStorageLifecycleConfigNoncurrentExp() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"days": {
				Type: schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceLinodeObjectStorageLifecycleConfigRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type: schema.TypeString,
				Optional: true,
			},
			"prefix": {
				Type: schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type: schema.TypeBool,
				Required: true,
			},
			"abort_incomplete_multipart_upload_days": {
				Type: schema.TypeInt,
				Optional: true,
			},
			"expiration": {
				Type: schema.TypeMap,
				Optional: true,
				Elem: resourceLinodeObjectStorageLifecycleConfigExpiration(),
			},
			"noncurrent_version_expiration": {
				Type: schema.TypeMap,
				Optional: true,
				Elem: resourceLinodeObjectStorageLifecycleConfigNoncurrentExp(),
			},
		},
	}
}

func resourceLinodeObjectStorageLifecycleConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceLinodeObjectStorageLifecycleConfigCreate,
		Read:   resourceLinodeObjectStorageLifecycleConfigRead,
		Update: resourceLinodeObjectStorageLifecycleConfigUpdate,
		Delete: resourceLinodeObjectStorageLifecycleConfigDelete,
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
				Elem:     resourceLinodeObjectStorageLifecycleConfigRule(),
			},
		},
	}
}

func resourceLinodeObjectStorageLifecycleConfigRead(d *schema.ResourceData, meta interface{}) error {
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

	d.Set("lifecycle_rule", rules)

	return nil
}

func resourceLinodeObjectStorageLifecycleConfigCreate(d *schema.ResourceData, meta interface{}) error {
	bucket := d.Get("bucket").(string)

	client := s3ConnFromResourceData(d)

	lifecycleConfig := &s3.BucketLifecycleConfiguration{}

	rules, err := expandLifecycleRules(d.Get("lifecycle_rule").([]interface{}))
	if err != nil {
		return fmt.Errorf("Failed to parse lifecycle rules: %s", err)
	}

	lifecycleConfig.Rules = rules

	inputConfig := &s3.PutBucketLifecycleConfigurationInput{
		Bucket:                 &bucket,
		LifecycleConfiguration: lifecycleConfig,
	}
	
	client.PutBucketLifecycleConfiguration(inputConfig)
	
	return nil
}

func resourceLinodeObjectStorageLifecycleConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceLinodeObjectStorageLifecycleConfigDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func expandLifecycleRules(ruleSpecs []interface{}) ([]*s3.LifecycleRule, error) {
	rules := make([]*s3.LifecycleRule, len(ruleSpecs))
	for i, ruleSpec := range ruleSpecs {
		ruleSpec := ruleSpec.(map[string]interface{})

		id := ruleSpec["id"].(string)
		prefix := ruleSpec["prefix"].(string)

		status := "Disabled"
		if ruleSpec["enabled"].(bool) {
			status = "Enabled"
		}

		abortIncompleteDays := ruleSpec["abort_incomplete_multipart_upload_days"].(int64)

		expiration := ruleSpec["expiration"].(map[string]interface{})
		expirationDate, err := time.Parse(time.RFC3339, expiration["date"].(string))
		if err != nil {
			return nil, err
		}
		expirationDays := expiration["days"].(int64)
		expirationMarker := expiration["expired_object_delete_marker"].(bool)

		ncVersionExpiration := ruleSpec["noncurrent_version_expiration"].(map[string]interface{})
		ncVersionExpirationDays := ncVersionExpiration["days"].(int64)

		rules[i] = &s3.LifecycleRule{
			ID: &id,
			Prefix: &prefix,
			Status: &status,
			AbortIncompleteMultipartUpload: &s3.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: &abortIncompleteDays,
			},
			Expiration: &s3.LifecycleExpiration{
				Date:                      &expirationDate,
				Days:                      &expirationDays,
				ExpiredObjectDeleteMarker: &expirationMarker,
			},
			NoncurrentVersionExpiration: &s3.NoncurrentVersionExpiration{
				NoncurrentDays: &ncVersionExpirationDays,
			},
		}
	}

	return rules, nil
}

func flattenLifecycleRules(rules []*s3.LifecycleRule) []map[string]interface{} {
	result := make([]map[string]interface{}, len(rules))

	for i, rule := range rules {
		result[i] = map[string]interface{}{
			"id": rule.ID,
			"prefix": rule.Prefix,
			"enabled": *rule.Status == "Enabled",
			"abort_incomplete_multipart_upload_days": rule.AbortIncompleteMultipartUpload.DaysAfterInitiation,
			"expiration": map[string]interface{}{
				"date": rule.Expiration.Date.Format(time.RFC3339),
				"days": rule.Expiration.Days,
				"expired_object_delete_marker": rule.Expiration.ExpiredObjectDeleteMarker,
			},
			"noncurrent_version_expiration": map[string]interface{}{
				"days": rule.NoncurrentVersionExpiration.NoncurrentDays,
			},
		}
	}

	return result
}