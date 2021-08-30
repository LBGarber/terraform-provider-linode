package tmpl

import (
	"github.com/linode/terraform-provider-linode/linode/acceptance"
	"testing"
)

type TemplateData struct {
	Label string
}

func Basic(t *testing.T, nodebalancer string) string {
	return acceptance.ExecuteTemplate(t,
		"nodebalancer_basic", TemplateData{Label: nodebalancer})
}

func Updates(t *testing.T, nodebalancer string) string {
	return acceptance.ExecuteTemplate(t,
		"nodebalancer_updates", TemplateData{Label: nodebalancer})
}

func DataBasic(t *testing.T, nodebalancer string) string {
	return acceptance.ExecuteTemplate(t,
		"nodebalancer_data_basic",
		TemplateData{Label: nodebalancer})
}
