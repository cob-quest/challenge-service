package models

type Image struct {
	CorId             string `json:"corId" bson:"corId"`
	CreatorName       string `json:"creatorName" bson:"creatorName"`
	ImageName         string `json:"imageName" bson:"imageName"`
	ImageRegistryLink string `json:"imageRegistryLink" bson:"imageRegistryLink"`
	S3Path            string `json:"s3Path" bson:"s3Path"`
}

type ImageList struct {
	Images []Image `json:"images" bson:",inline"`
}
