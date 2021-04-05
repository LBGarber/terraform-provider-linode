package linode

import (
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// filterTypeFunc is a function that takes in a filter name and value,
// and returns the value converted to the correct filter type.
type filterTypeFunc func(filterName string, value string) (interface{}, error)

// constructFilterString constructs a Linode filter JSON string from each filter element in the schema
func constructFilterString(d *schema.ResourceData, typeFunc filterTypeFunc) (string, error) {
	filters := d.Get("filter").([]interface{})
	resultMap := make(map[string]interface{})

	if len(filters) < 1 {
		return "{}", nil
	}

	var rootFilter []interface{}

	for _, filter := range filters {
		filter := filter.(map[string]interface{})

		name := filter["name"].(string)
		values := filter["values"].([]interface{})

		subFilter := make([]interface{}, len(values))

		for i, value := range values {
			value, err := typeFunc(name, value.(string))
			if err != nil {
				return "", err
			}

			valueFilter := make(map[string]interface{})
			valueFilter[name] = value

			subFilter[i] = valueFilter
		}

		rootFilter = append(rootFilter, map[string]interface{}{
			"+or": subFilter,
		})
	}

	resultMap["+and"] = rootFilter

	result, err := json.Marshal(resultMap)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
