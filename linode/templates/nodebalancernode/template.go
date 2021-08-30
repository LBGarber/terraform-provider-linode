package nodebalancernode

import "github.com/linode/terraform-provider-linode/linode/templates/nodebalancerconfig"

type TemplateData struct {
	Config nodebalancerconfig.TemplateData
	Label  string
}
