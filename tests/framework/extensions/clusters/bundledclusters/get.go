package bundledclusters

import (
	"github.com/rancher/rancher/tests/framework/clients/rancher"
)

// Get is a method of BundledCluster that uses provisioning and management clients
// to get related cluster data depending on cluster version.
func (bc *BundledCluster) Get(client *rancher.Client) (cluster *BundledCluster, err error) {
	cluster = new(BundledCluster)
	cluster.Meta = bc.Meta

	if bc.V1 != nil {
		cluster.V1, err = client.Provisioning.Cluster.ByID(cluster.Meta.ID)
		if err != nil {
			return cluster, err
		}
	} else if bc.V3 != nil {
		cluster.V3, err = client.Management.Cluster.ByID(cluster.Meta.ID)
		if err != nil {
			return cluster, err
		}
	}

	return
}
