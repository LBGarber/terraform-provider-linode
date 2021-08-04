package linode

import (
	"io/fs"
	"log"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

// publicKeyMaterial for use while testing
var (
	privateKeyMaterial string
	publicKeyMaterial  string
	tfTemplates        *template.Template
)

func init() {
	var err error
	publicKeyMaterial, privateKeyMaterial, err = acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		log.Fatalf("Failed to generate random SSH key pair for testing: %s", err)
	}

	templateFiles := make([]string, 0)

	err = filepath.Walk("../templates", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".gotf" {
			templateFiles = append(templateFiles, path)
		}

		return nil
	})
	if err != nil {
		log.Fatalf("failed to find template files: %v", err)
	}

	tfTemplates = template.New("tf-test")
	if _, err := tfTemplates.ParseFiles(templateFiles...); err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}
}
