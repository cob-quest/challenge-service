package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	helmclient "github.com/mittwald/go-helm-client"
	amqp "github.com/rabbitmq/amqp091-go"
	"helm.sh/helm/v3/pkg/repo"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sys.io/challenge-service/collections"
	"sys.io/challenge-service/config"
	"sys.io/challenge-service/models"
	"sys.io/challenge-service/utils"
)

func StartChallenge(ch *amqp.Channel, ctx context.Context, msg []byte, routingKey string) {

	// Unpack JSON data.
	var data map[string]interface{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		log.Printf("Failed to decode JSON message body: %s", err)
		utils.FailOnError(err, "Failed to decode JSON message body")
		return
	}

	var attempt models.Attempt
	err = json.Unmarshal(msg, &attempt)
	if err != nil {
		data["eventStatus"] = "challengeCreateFailed"
		log.Printf("Failed to decode JSON message body: %s", err)

		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}

	//Get repository, tag and release_id
	url := strings.TrimPrefix(attempt.ImageRegistryLink, "https://")

	repository, tag := strings.Split(url, ":")[0], strings.Split(url, ":")[1]
	release_id := fmt.Sprintf("a%s", attempt.Token)

	// Create a Kubernetes client.
	var kconfig *rest.Config

	if config.ENVIRONMENT == "DEV" {
		kconfig, err = clientcmd.BuildConfigFromFlags("", config.KUBECONFIG)
		if err != nil {
			data["eventStatus"] = "challengeCreateFailed"
			log.Printf("Failed to create Kubernetes client: %s", err)
	
			msgBody, _ := json.Marshal(data)
			Publish(ch, ctx, msgBody, routingKey)
			return
		}
	} else {
		kconfig, err = rest.InClusterConfig()
		if err != nil {
			data["eventStatus"] = "challengeCreateFailed"
			log.Printf("Failed to create Kubernetes client: %s", err)
	
			msgBody, _ := json.Marshal(data)
			Publish(ch, ctx, msgBody, routingKey)
			return
		}
	}

	// generate ssh keys and convert them into strings
	pubKey, privKey, err := utils.MakeSSHKeyPair()
	if err != nil {
		log.Printf("%s", err)
		data["eventStatus"] = "challengeStartFailed"
		log.Printf("Challenge %s start failed ...", release_id)

		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
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
			log.Printf("Failed to create HelmClient: %v\n", err)
			data["eventStatus"] = "challengeStartFailed"
			log.Printf("Challenge %s start failed ...", release_id)

			msgBody, _ := json.Marshal(data)
			Publish(ch, ctx, msgBody, routingKey)
			return
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
			log.Printf("Failed to create HelmClient: %v\n", err)
			data["eventStatus"] = "challengeStartFailed"
			log.Printf("Challenge %s start failed ...", release_id)
			
			msgBody, _ := json.Marshal(data)
			Publish(ch, ctx, msgBody, routingKey)
			return
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
		log.Printf("Failed to add or update HelmChartRepo: %v\n", err)
		data["eventStatus"] = "challengeStartFailed"
		log.Printf("Challenge %s start failed ...", release_id)
		
		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
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
  registry: registry.gitlab.com
  repository: %s
  pullPolicy: IfNotPresent
  tag: %s
imagePullSecrets:
  - name: docker-registry-credentials
authorized_keys: %s`, repository, tag, pubKey),
	}

	// install or upgrade a chart release
	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		log.Printf("Failed to install or upgradeChart client: %v\n", err)
		data["eventStatus"] = "challengeStartFailed"
		log.Printf("Challenge %s start failed ...", release_id)
		
		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}

	log.Print("Helm installed or upgraded challenge!!")

	//Configure kubenetes client
	client, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client: %v\n", err)
		return
	}

	log.Print("Kubernetes Configured !!")

	time.Sleep(2 * time.Second)

	// check pod status
	// Get the pod list for the deployment.
	podList, err := client.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/instance=%s", release_id),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	pod := podList.Items[0]

	// Check status every 5 seconds
	for pod.Status.Phase == "Pending" {
		data["eventStatus"] = "challengeStarting"
		log.Printf("Challenge %s is starting ... ", release_id)

		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)

		time.Sleep(5 * time.Second)

		podList, err := client.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{
			LabelSelector: fmt.Sprintf("app.kubernetes.io/instance=%s", release_id),
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		pod = podList.Items[0]
	}

	if pod.Status.Phase != "Running" {
		data["eventStatus"] = "challengeStartFailed"
		log.Printf("Challenge %s start failed ...", release_id)

		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}

	// get pod IP and port functions
	// Get the external IP address of the first node.
	nodeList, err := client.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		log.Printf("%s", err)
		data["eventStatus"] = "challengeStartFailed"
		log.Printf("Challenge %s start failed ...", release_id)

		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}
	publicIPAddress := nodeList.Items[0].Status.Addresses[0].Address

	// Get the NodePort port number for the `my-service` Service.
	service, err := client.CoreV1().Services(namespace).Get(context.Background(), fmt.Sprintf("%s-challenge", release_id), v1.GetOptions{
		TypeMeta: v1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
	if err != nil {
		log.Printf("%s", err)
		data["eventStatus"] = "challengeStartFailed"
		log.Printf("Challenge %s start failed ...", release_id)

		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}

	nodePort := service.Spec.Ports[0].NodePort

	// Print the NodePort port number.
	fmt.Printf("NodePort port number for the `my-service` Service on the public IP address of the node exposing the pod: %s:%d\n", publicIPAddress, nodePort)

	// Update attempt
	attempt.Ipaddress = publicIPAddress
	attempt.Port = strconv.FormatInt(int64(nodePort),10)
	attempt.Sshkey = privKey

	_,err = collections.UpdateAttempt(&attempt)
	if err != nil {
		log.Printf("%s", err)
		data["eventStatus"] = "challengeStartFailed"
		log.Printf("Challenge %s start failed ...", release_id)

		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}


	// Successfully started
	data["eventStatus"] = "challengeStarted"
	log.Printf("Challenge %s started ...", release_id)

	msgBody, _ := json.Marshal(data)
	Publish(ch, ctx, msgBody, routingKey)

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

		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}

	//find image
	image, err := collections.GetImage(challenge.CreatorName, challenge.ImageName, challenge.ImageTag)
	if err != nil {
		data["eventStatus"] = "challengeCreateFailed"
		log.Printf("Failed to Find image: %s", err)

		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}
	challenge.ImageRegistryLink = image.ImageRegistryLink

	// Create challenge

	_, err = collections.CreateChallenge(&challenge)
	if err != nil {
		data["eventStatus"] = "challengeCreateFailed"
		log.Printf("Failed to create challenge: %s", err)

		msgBody, _ := json.Marshal(data)
		Publish(ch, ctx, msgBody, routingKey)
		return
	}

	// Create attempts
	for _, v := range challenge.Participants {

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
			log.Printf("Failed to create attempt for %s: %s", v, err)

			msgBody, _ := json.Marshal(data)
			Publish(ch, ctx, msgBody, routingKey)
			return
		}
	}

	data["eventStatus"] = "challengeCreated"
	msgBody, _ := json.Marshal(data)
	Publish(ch, ctx, msgBody, routingKey)
}
