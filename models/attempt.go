package models

// Generated by https://quicktype.io

type Attempt struct {
	Participant       string  `json:"participant" bson:"participant"`
	Token             string  `json:"token" bson:"token"`
	Sshkey            string  `json:"sshkey" bson:"sshkey"`
	Result            float64 `json:"result" bson:"result"`
	Ipaddress         string  `json:"ipaddress" bson:"ipaddress"`
	Port              string  `json:"port" bson:"port"`
	ChallengeName     string  `json:"challengeName" bson:"challengeName"`
	CreatorName       string  `json:"creatorName" bson:"creatorName"`
	ImageRegistryLink string  `json:"imageRegistryLink" bson:"imageRegistryLink"`
}
