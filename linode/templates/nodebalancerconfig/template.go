package nodebalancerconfig

import "github.com/linode/terraform-provider-linode/linode/templates/nodebalancer"

type TemplateData struct {
	NodeBalancer nodebalancer.TemplateData
	SSLCert      string
	SSLKey       string
}
