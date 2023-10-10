package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sys.io/assignment-service/config"
	"sys.io/assignment-service/utils"
)

func main() {

	// init env
	config.InitEnv()

	// read kubernetes config file
	kubeConfig, err := os.ReadFile(config.KUBECONFIG)
	if err != nil {
		log.Fatalf("Failed to read file from kube/config: %v\n", err)
	}

	// generate ssh keys and convert them into strings
	pubKey, privKey, err := utils.GenerateRSAKeyPairString()
	if err != nil {
		log.Fatal(err)
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
	}

	// list all the deployed releases
	releases, err := helmClient.ListDeployedReleases()
	if err != nil {
		log.Fatalf("Failed to list deployed releases %v\n", err)
	}
	log.Printf("releases are: %v", releases[0].Name)

	// specify the challenge chart
	chartSpec := helmclient.ChartSpec{
		ReleaseName:     "challenge",
		ChartName:       fmt.Sprintf("%s/%s", config.HELM_REPO_NAME, config.HELM_CHART_NAME),
		Namespace:       namespace,
		CreateNamespace: true,
		GenerateName:    true,
		ValuesYaml: `
image:
  repository: clitest
  pullPolicy: Never
  tag: latest
authorized_keys: "%s"`,
	}

	// install or upgrade a chart release
	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		panic(err)
	}

	// get pod IP and port functions

	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// Uncomment below to use In Cluster Config
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	// Create a Kubernetes client.
	// config, err := rest.InClusterConfig()
	// if err != nil {
	//     fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client: %v\n", err)
	//     os.Exit(1)
	// }

	config, err := clientcmd.BuildConfigFromFlags("", config.KUBECONFIG)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client configuration: %v\n", err)
		os.Exit(1)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Get the public IP address of the node exposing the pod.
	publicIPAddress, err := getPublicIPAddress()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get public IP address: %v\n", err)
		os.Exit(1)
	}

	// Get the NodePort port number for the `my-service` Service.
	service, err := client.CoreV1().Services(namespace).Get(context.Background(), "challenge", v1.GetOptions{
		TypeMeta: v1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get Service: %v\n", err)
		os.Exit(1)
	}

	nodePort := service.Spec.Ports[0].NodePort

	// Print the NodePort port number.
	fmt.Printf("NodePort port number for the `my-service` Service on the public IP address of the node exposing the pod: %s:%d\n", publicIPAddress, nodePort)
}

////////////////////////////////////////////////////////////////////
// Assume same node as the pod
// temporary solution to get ipaddress need a better way
////////////////////////////////////////////////////////////////////

func getPublicIPAddress() (string, error) {
	// Get the public IP address of the current node.
	resp, err := http.Get("https://api.ipify.org/?format=json")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Decode the JSON response.
	var ipAddress struct {
		IP string `json:"ip"`
	}

	err = json.NewDecoder(resp.Body).Decode(&ipAddress)
	if err != nil {
		return "", err
	}

	return ipAddress.IP, nil
}
