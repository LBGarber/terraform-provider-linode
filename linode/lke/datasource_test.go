package lke_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/linode/terraform-provider-linode/linode/acceptance"
	"github.com/linode/terraform-provider-linode/linode/lke/tmpl"
)

const dataSourceClusterName = "data.linode_lke_cluster.test"

func TestAccDataSourceLKECluster_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf_test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.TestAccProviders,
		CheckDestroy: acceptance.CheckLKEClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.DataBasic(t, clusterName, k8sVersionLatest),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceClusterName, "label", clusterName),
					resource.TestCheckResourceAttr(dataSourceClusterName, "region", "us-central"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "k8s_version", k8sVersionLatest),
					resource.TestCheckResourceAttr(dataSourceClusterName, "status", "ready"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "tags.#", "1"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.#", "1"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.type", "g6-standard-2"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.count", "3"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.nodes.#", "3"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.autoscaler.#", "0"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "control_plane.0.high_availability", "false"),
					resource.TestCheckResourceAttrSet(dataSourceClusterName, "pools.0.id"),
					resource.TestCheckResourceAttrSet(dataSourceClusterName, "kubeconfig"),
				),
			},
		},
	})
}

func TestAccDataSourceLKECluster_autoscaler(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf_test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.TestAccProviders,
		CheckDestroy: acceptance.CheckLKEClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.DataAutoscaler(t, clusterName, k8sVersionLatest),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceClusterName, "label", clusterName),
					resource.TestCheckResourceAttr(dataSourceClusterName, "region", "us-central"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "k8s_version", k8sVersionLatest),
					resource.TestCheckResourceAttr(dataSourceClusterName, "status", "ready"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "tags.#", "1"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.#", "1"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.type", "g6-standard-2"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.count", "3"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.nodes.#", "3"),
					resource.TestCheckResourceAttrSet(dataSourceClusterName, "pools.0.id"),
					resource.TestCheckResourceAttrSet(dataSourceClusterName, "kubeconfig"),

					resource.TestCheckResourceAttr(resourceClusterName, "pool.0.autoscaler.#", "1"),
					resource.TestCheckResourceAttr(resourceClusterName, "pool.0.autoscaler.0.min", "1"),
					resource.TestCheckResourceAttr(resourceClusterName, "pool.0.autoscaler.0.max", "5"),
				),
			},
		},
	})
}

func TestAccDataSourceLKECluster_controlPlane(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf_test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.TestAccProviders,
		CheckDestroy: acceptance.CheckLKEClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: tmpl.DataControlPlane(t, clusterName, k8sVersionLatest, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceClusterName, "label", clusterName),
					resource.TestCheckResourceAttr(dataSourceClusterName, "region", "us-central"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "k8s_version", k8sVersionLatest),
					resource.TestCheckResourceAttr(dataSourceClusterName, "status", "ready"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "tags.#", "1"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.#", "1"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.type", "g6-standard-2"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.count", "1"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.nodes.#", "1"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "pools.0.autoscaler.#", "0"),
					resource.TestCheckResourceAttr(dataSourceClusterName, "control_plane.0.high_availability", "true"),
					resource.TestCheckResourceAttrSet(dataSourceClusterName, "pools.0.id"),
					resource.TestCheckResourceAttrSet(dataSourceClusterName, "kubeconfig"),
				),
			},
		},
	})
}
