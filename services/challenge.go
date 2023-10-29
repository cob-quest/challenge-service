package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	helmclient "github.com/mittwald/go-helm-client"
	amqp "github.com/rabbitmq/amqp091-go"
	"helm.sh/helm/v3/pkg/repo"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// "sys.io/challenge-service/collections"
	"sys.io/challenge-service/collections"
	"sys.io/challenge-service/config"
	"sys.io/challenge-service/models"
	"sys.io/challenge-service/utils"
)

func StartChallenge(data map[string]interface{}) (string, string, int32, error) {

	// Unpack JSON data.
	repository, tag, release_id := data["repository"].(string), data["tag"].(string), data["release_id"].(string)

	// Create a Kubernetes client.
	var kconfig *rest.Config
	var err error

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

	// generate ssh keys and convert them into strings
	pubKey, privKey, err := utils.MakeSSHKeyPair()
	if err != nil {
		log.Fatal(err)
		return "", "", 0, err
	}
	log.Printf("pub: %s, priv: %s", pubKey, privKey)

	// create a helm client
	var namespace = "challenge"

	var helmClient helmclient.Client

	if config.ENVIRONMENT == "DEV" {

		kubeConfig, err := os.ReadFile(config.KUBECONFIG)
		if err != nil {
			log.Printf("Failed to read file from kube/config: %v\n", err)
		}

		helmClient, err = helmclient.NewClientFromKubeConf(
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
	} else {
		opt := &helmclient.RestConfClientOptions{
			Options: &helmclient.Options{
				Namespace:        namespace, // Change this to the namespace you wish the client to operate in.
				RepositoryCache:  "/tmp/.helmcache",
				RepositoryConfig: "/tmp/.helmrepo",
				Debug:            true,
				Linting:          true, // Change this to false if you don't want linting.
				DebugLog: func(format string, v ...interface{}) {
					// Change this to your own logger. Default is 'log.Printf(format, v...)'.
				},
			},
			RestConfig: kconfig,
		}

		helmClient, err = helmclient.NewClientFromRestConf(opt)
		if err != nil {
			log.Fatalf("Failed to create HelmClient: %v\n", err)
			return "", "", 0, err
		}
	}

	log.Print("helmClient Configured !!")

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

	log.Print("Helm Repository Added!!")

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

	log.Print("Helm installed or upgraded challenge!!")

	// get pod IP and port functions

	client, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client: %v\n", err)
		return "", "", 0, err
	}

	log.Print("Kubernetes Configured !!")

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

	return privKey, publicIPAddress, nodePort, nil
}

func CreateChallenge(ch *amqp.Channel, ctx context.Context, msg []byte, routingKey string) {


	// Unpack JSON data.
	var data map[string]interface{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		log.Printf("Failed to decode JSON message body: %s", err)
		utils.FailOnError(err, "Failed to decode JSON message body")
		return
	}

	var challenge models.Challenge
	err = json.Unmarshal(msg, &challenge)
	if err != nil {
		data["eventStatus"] = "challengeCreateFailed"
		log.Printf("Failed to decode JSON message body: %s", err)

		msgBody,_ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}

	//find image 
	image,err := collections.GetImage(challenge.CreatorName, challenge.ImageName, challenge.ImageTag)
	if err != nil {
		data["eventStatus"] = "challengeCreateFailed"
		log.Printf("Failed to Find image: %s", err)

		msgBody,_ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}
	challenge.ImageRegistryLink = image.ImageRegistryLink

	// Create challenge

	_, err = collections.CreateChallenge(&challenge)
	if err != nil {
		data["eventStatus"] = "challengeCreateFailed"
		log.Printf("Failed to create challenge: %s", err)

		msgBody,_ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}

	// Create attempts
	for _,v := range challenge.Participants{
		
		_, err = collections.CreateAttempt(&models.Attempt{
			Participant:       v,
			Token:             uuid.NewString(),
			Sshkey:            "",
			Result:            0,
			Ipaddress:         "",
			Port:              "",
			ChallengeName:     challenge.ChallengeName,
			CreatorName:       challenge.CreatorName,
			ImageRegistryLink: challenge.ImageRegistryLink,
		})
		if err != nil {
			data["eventStatus"] = "challengeCreateFailed"
			log.Printf("Failed to create attempt for %s: %s", v,err)

			msgBody,_ := json.Marshal(data)
			Publish(ch, ctx, msgBody, routingKey)
			return
		}
	}



	data["eventStatus"] = "challengeCreated"
	msgBody,_ := json.Marshal(data)
	Publish(ch, ctx, msgBody, routingKey)
	// return message for RMQ
}
