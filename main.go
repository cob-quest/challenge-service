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
	"sys.io/assignment-service/utils"
)

func main() {

	// set config path
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
	}

	kubeConfig, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file from kube/config: %v\n", err)
		os.Exit(1)
	}

	// TODO: retrieve secrets for gitlab to pull images and helm charts

	// generate ssh keys and convert them into strings
	pubKey, privKey, err := utils.GenerateRSAKeyPairString()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("pub: %s, priv: %s", pubKey, privKey)

	////////////////////////////////////////////////////////////////////
	// Create Challenge Deployment
	////////////////////////////////////////////////////////////////////

	//This should be customizable based on the useremail + challengeid
	var namespace = "challenge"

	helmClient, err := helmclient.NewClientFromKubeConf(&helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        namespace,
			RepositoryConfig: "",
			RepositoryCache:  "",
			Debug:            false,
			Linting:          false,
			RegistryConfig:   "",
			Output:           nil,
		},
		KubeContext: "",
		KubeConfig:  kubeConfig,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create HelmClient: %v\n", err)
		os.Exit(1)
	}

	// Set the Helm chart repository.
	chartRepo := repo.Entry{
		Name: "challenge",
		// Hardcoded url might wanna change
		URL: "https://gitlab.com/api/v4/projects/51018402/packages/helm/stable",
		// Hardcoded username
		Username: "oojingkai10",
		// Password is Gitlab token
		Password: "",
		// Since helm 3.6.1 it is necessary to pass 'PassCredentialsAll = true'.
		PassCredentialsAll: true,
	}

	err = helmClient.AddOrUpdateChartRepo(chartRepo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create HelmClient: %v\n", err)
		os.Exit(1)
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName:     "challenge",
		ChartName:       "challenge/challenge",
		Namespace:       namespace,
		CreateNamespace: true,
		// UpgradeCRDs: true,
		// Wait:        true,
		ValuesYaml: `
    image:
        repository: clitest
        pullPolicy: Never
        tag: latest
    authorized_keys: <Public Key Goes here>
    `,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		panic(err)
	}

	////////////////////////////////////////////////////////////////////
	// Get Pod IP and Port functions
	////////////////////////////////////////////////////////////////////

	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// Uncomment below to use In Cluster Config
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	// Create a Kubernetes client.
	// config, err := rest.InClusterConfig()
	// if err != nil {
	//     fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client: %v\n", err)
	//     os.Exit(1)
	// }

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
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
