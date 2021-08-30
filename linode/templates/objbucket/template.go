package objbucket

import "github.com/linode/terraform-provider-linode/linode/templates/objkey"

type TemplateData struct {
	ObjectKey objkey.TemplateData

	Label string

	ACL         string
	CorsEnabled bool
	Versioning  bool

	Cert       string
	PrivateKey string
}
