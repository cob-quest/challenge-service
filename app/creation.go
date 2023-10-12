package app

import (
	"context"
	"fmt"
	"log"
	"os"

	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sys.io/assignment-service/config"
	"sys.io/assignment-service/utils"
)

func CreateChallenge(repository string, tag string, release_id string) (string, string, int32, error) {
	kubeConfig, err := os.ReadFile(config.KUBECONFIG)
	if err != nil {
		log.Fatalf("Failed to read file from kube/config: %v\n", err)
	}

	// generate ssh keys and convert them into strings
	pubKey, privKey, err := utils.MakeSSHKeyPair()
	if err != nil {
		log.Fatal(err)
		return "", "", 0, err
	}
	log.Printf("pub: %s, priv: %s", pubKey, privKey)

	// create a helm client
	var namespace = "challenge"
	helmClient, err := helmclient.NewClientFromKubeConf(
		&helmclient.KubeConfClientOptions{
			Options: &helmclient.Options{
				Namespace:        namespace,
				RepositoryCache:  "/tmp/.helmcache",
				RepositoryConfig: "/tmp/.helmrepo",
				Debug:            true,
				Linting:          true,
				Output:           nil,
			},
			KubeConfig:  kubeConfig,
			KubeContext: "",
		},
	)
	if err != nil {
		log.Fatalf("Failed to create HelmClient: %v\n", err)
		return "", "", 0, err
	}

	// create a repo reference
	repo := repo.Entry{
		Name:               config.HELM_REPO_NAME,
		URL:                config.HELM_REPO_URL,
		Username:           config.HELM_REPO_USERNAME,
		Password:           config.HELM_REPO_PASSWORD,
		PassCredentialsAll: true,
	}

	// add the chart repo reference
	err = helmClient.AddOrUpdateChartRepo(repo)
	if err != nil {
		log.Fatalf("Failed to create HelmClient: %v\n", err)
		return "", "", 0, err
	}

	// list all the deployed releases
	// releases, err := helmClient.ListDeployedReleases()
	// if err != nil {
	// 	log.Fatalf("Failed to list deployed releases %v\n", err)
	// }
	// log.Printf("releases are: %v", releases[0].Name)

	// specify the challenge chart
	chartSpec := helmclient.ChartSpec{
		ReleaseName:     release_id,
		ChartName:       fmt.Sprintf("%s/%s", config.HELM_REPO_NAME, config.HELM_CHART_NAME),
		Namespace:       namespace,
		CreateNamespace: true,
		GenerateName:    true,
		ValuesYaml: fmt.Sprintf(`
image:
  repository: %s
  pullPolicy: Never
  tag: %s
imagePullSecrets:
  - name: docker-registry-credentials
authorized_keys: %s`, repository, tag, pubKey),
	}

	// install or upgrade a chart release
	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		return "", "", 0, err
	}

	// get pod IP and port functions

	// Create a Kubernetes client.
	var kconfig *rest.Config

	if config.ENVIRONMENT == "DEV" {
		kconfig, err = clientcmd.BuildConfigFromFlags("", config.KUBECONFIG)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client configuration: %v\n", err)
			return "", "", 0, err
		}
	} else {
		kconfig, err = rest.InClusterConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client: %v\n", err)
			return "", "", 0, err
		}
	}

	client, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client: %v\n", err)
		return "", "", 0, err
	}

	// Get the public IP address of the node exposing the pod.
	// publicIPAddress, err := utils.GetPublicIPAddress()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Failed to get public IP address: %v\n", err)
	// 	return "", "", 0, err
	// }

	nodeList, err := client.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Get the external IP address of the first node.
	publicIPAddress := nodeList.Items[0].Status.Addresses[0].Address

	// Print the external IP address.
	// fmt.Println(publicIPAddress)

	// Get the NodePort port number for the `my-service` Service.
	service, err := client.CoreV1().Services(namespace).Get(context.Background(), fmt.Sprintf("%s-challenge", release_id), v1.GetOptions{
		TypeMeta: v1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get Service: %v\n", err)
		return "", "", 0, err
	}

	nodePort := service.Spec.Ports[0].NodePort

	// Print the NodePort port number.
	fmt.Printf("NodePort port number for the `my-service` Service on the public IP address of the node exposing the pod: %s:%d\n", publicIPAddress, nodePort)

	return pubKey, publicIPAddress, nodePort, nil
}
