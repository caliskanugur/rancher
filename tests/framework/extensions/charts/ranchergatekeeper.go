package charts

import (
	"context"
	"fmt"

	"github.com/rancher/rancher/pkg/api/steve/catalog/types"
	catalogv1 "github.com/rancher/rancher/pkg/apis/catalog.cattle.io/v1"
	"github.com/rancher/rancher/tests/framework/clients/rancher"
	"github.com/rancher/rancher/tests/framework/pkg/wait"
	"github.com/rancher/rancher/tests/integration/pkg/defaults"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	// Namespace that rancher gatekeeper chart is installed in
	RancherGatekeeperNamespace = "cattle-gatekeeper-system"
	// Name of the rancher gatekeeper chart
	RancherGatekeeperName = "rancher-gatekeeper"
	// Name of rancher gatekeepr crd chart
	RancherGatekeeperCRDName = "rancher-gatekeeper-crd"
)

// InstallRancherGatekeeperChart installs the OPA gatekeeper chart
func InstallRancherGatekeeperChart(client *rancher.Client, installOptions *InstallOptions) error {
	hostWithProtocol := fmt.Sprintf("https:// %s", client.RancherConfig.Host)
	gatekeeperChartInstallActionPayload := &payloadOpts{
		InstallOptions: *installOptions,
		Name:           RancherGatekeeperName,
		Host:           hostWithProtocol,
		Namespace:      RancherGatekeeperNamespace,
	}

	chartInstallAction := newGatekeeperChartInstallAction(gatekeeperChartInstallActionPayload)

	catalogClient, err := client.GetClusterCatalogClient(installOptions.ClusterID)
	if err != nil {
		return err
	}

	// Cleanup registration

	// register uninstall rancher-gatekeeper as a cleanup function
	client.Session.RegisterCleanupFunc(func() error {
		// UninstallAction for when uninstalling the rancher-gatekeeper chart
		defaultChartUninstallAction := newChartUninstallAction()

		err := catalogClient.UninstallChart(RancherGatekeeperName, RancherGatekeeperNamespace, defaultChartUninstallAction)
		if err != nil {
			return err
		}

		watchAppInterface, err := catalogClient.Apps(RancherGatekeeperNamespace).Watch(context.TODO(), metav1.ListOptions{
			FieldSelector:  "metadata.name=" + RancherGatekeeperName,
			TimeoutSeconds: &defaults.WatchTimeoutSeconds,
		})
		if err != nil {
			return err
		}

		err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
			chart := event.Object.(*catalogv1.App)
			if event.Type == watch.Error {
				return false, fmt.Errorf("there was an error uninstalling rancher gatekeeper chart")
			} else if event.Type == watch.Deleted {
				return true, nil
			} else if chart == nil {
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			return err
		}

		err = catalogClient.UninstallChart(RancherGatekeeperCRDName, RancherGatekeeperNamespace, defaultChartUninstallAction)
		if err != nil {
			return err
		}

		watchAppInterface, err = catalogClient.Apps(RancherGatekeeperNamespace).Watch(context.TODO(), metav1.ListOptions{
			FieldSelector:  "metadata.name=" + RancherGatekeeperCRDName,
			TimeoutSeconds: &defaults.WatchTimeoutSeconds,
		})
		if err != nil {
			return err
		}

		return wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
			chart := event.Object.(*catalogv1.App)
			if event.Type == watch.Error {
				return false, fmt.Errorf("there was an error uninstalling rancher gatekeeper chart")
			} else if event.Type == watch.Deleted {
				return true, nil
			} else if chart == nil {
				return true, nil
			}
			return false, nil
		})

		// dynamicClient, err := client.GetDownStreamClusterClient(installOptions.ClusterID)
		// if err != nil {
		// 	return err
		// }

		// namespaceResource := dynamicClient.Resource(namespaces.NamespaceGroupVersionResource).Namespace("")

		// err = namespaceResource.Delete(context.TODO(), RancherGatekeeperNamespace, metav1.DeleteOptions{})
		// if err != nil {
		// 	return err
		// }

		// unstructuredCRDList, err := crds.ListCustomResourceDefinitions(client, installOptions.ClusterID, "")
		// if err != nil {
		// 	return err
		// }
		// CRDSlice := crds.GetCustomResourceDefinitionsListByName(unstructuredCRDList, "gatekeeper")
		// err = crds.BatchDeleteCustomResourceDefinition(client, installOptions.ClusterID, "", CRDSlice)
		// if err != nil {
		// 	return err
		// }

		// adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
		// if err != nil {
		// 	return err
		// }
		// adminDynamicClient, err := adminClient.GetDownStreamClusterClient(installOptions.ClusterID)
		// if err != nil {
		// 	return err
		// }
		// adminNamespaceResource := adminDynamicClient.Resource(namespaces.NamespaceGroupVersionResource).Namespace("")

		// watchNamespaceInterface, err := adminNamespaceResource.Watch(context.TODO(), metav1.ListOptions{
		// 	FieldSelector:  "metadata.name=" + RancherGatekeeperNamespace,
		// 	TimeoutSeconds: &defaults.WatchTimeoutSeconds,
		// })

		// if err != nil {
		// 	return err
		// }

		// return wait.WatchWait(watchNamespaceInterface, func(event watch.Event) (ready bool, err error) {
		// 	if event.Type == watch.Deleted {
		// 		return true, nil
		// 	}
		// 	return false, nil
		// })

	})

	err = catalogClient.InstallChart(chartInstallAction)
	if err != nil {
		return err
	}

	// wait for chart to be fully deployed
	watchAppInterface, err := catalogClient.Apps(RancherGatekeeperNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + RancherGatekeeperName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		app := event.Object.(*catalogv1.App)

		state := app.Status.Summary.State
		if state == string(catalogv1.StatusDeployed) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}

// newGatekeeperChartInstallAction is a helper function that returns an array of newChartInstallActions for installing the gatekeeper and gatekeepr-crd charts
func newGatekeeperChartInstallAction(p *payloadOpts) *types.ChartInstallAction {
	gatekeeperValues := map[string]interface{}{}

	chartInstall := newChartInstall(p.Name, p.InstallOptions.Version, p.InstallOptions.ClusterID, p.InstallOptions.ClusterName, p.Host, gatekeeperValues)
	chartInstallCRD := newChartInstall(p.Name+"-crd", p.InstallOptions.Version, p.InstallOptions.ClusterID, p.InstallOptions.ClusterName, p.Host, gatekeeperValues)
	chartInstalls := []types.ChartInstall{*chartInstallCRD, *chartInstall}

	chartInstallAction := newChartInstallAction(p.Namespace, p.ProjectID, chartInstalls)

	return chartInstallAction
}

// UpgradeRanchergatekeeperChart is a helper function that upgrades the rancher-gatekeeper chart.
func UpgradeRancherGatekeeperChart(client *rancher.Client, installOptions *InstallOptions) error {
	gatekeeperChartUpgradeActionPayload := &payloadOpts{
		InstallOptions: *installOptions,
		Name:           RancherGatekeeperName,
		Host:           client.RancherConfig.Host,
		Namespace:      RancherGatekeeperNamespace,
	}

	chartUpgradeAction := newGatekeeperChartUpgradeAction(gatekeeperChartUpgradeActionPayload)

	catalogClient, err := client.GetClusterCatalogClient(installOptions.ClusterID)
	if err != nil {
		return err
	}

	err = catalogClient.UpgradeChart(chartUpgradeAction)
	if err != nil {
		return err
	}

	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	if err != nil {
		return err
	}
	adminCatalogClient, err := adminClient.GetClusterCatalogClient(installOptions.ClusterID)
	if err != nil {
		return err
	}

	// wait for chart to be in status pending upgrade
	watchAppInterface, err := adminCatalogClient.Apps(RancherGatekeeperNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + RancherGatekeeperName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		app := event.Object.(*catalogv1.App)

		state := app.Status.Summary.State
		if state == string(catalogv1.StatusPendingUpgrade) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	// wait for chart to be full deployed
	watchAppInterface, err = adminCatalogClient.Apps(RancherGatekeeperNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + RancherGatekeeperName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		app := event.Object.(*catalogv1.App)

		state := app.Status.Summary.State
		if state == string(catalogv1.StatusDeployed) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// newGatekeeperChartUpgradeAction is a private helper function that returns chart upgrade action.
func newGatekeeperChartUpgradeAction(p *payloadOpts) *types.ChartUpgradeAction {
	gatekeeperValues := map[string]interface{}{}

	chartUpgrade := newChartUpgrade(p.Name, p.InstallOptions.Version, p.InstallOptions.ClusterID, p.InstallOptions.ClusterName, p.Host, gatekeeperValues)
	chartUpgradeCRD := newChartUpgrade(p.Name+"-crd", p.InstallOptions.Version, p.InstallOptions.ClusterID, p.InstallOptions.ClusterName, p.Host, gatekeeperValues)
	chartUpgrades := []types.ChartUpgrade{*chartUpgradeCRD, *chartUpgrade}

	chartUpgradeAction := newChartUpgradeAction(p.Namespace, chartUpgrades)

	return chartUpgradeAction
}
