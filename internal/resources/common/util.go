package resource

import (
	"os"
	"strings"
)

var (
	ImageList = []string{"MCSP_UTILS", "ACCOUNT_IAM_APP", "IM_CONFIG_JOB", "ACCOUNT_IAM_UI_INSTANCE_SERVICE", "ACCOUNT_IAM_UI_API_SERVICE"}
)

func ReplaceImages(resource string) (result string) {
	result = resource
	for _, imageName := range ImageList {
		result = strings.ReplaceAll(result, imageName, GetImage(imageName))
	}
	return result
}

func GetImage(imageName string) string {
	image, _ := os.LookupEnv(imageName)
	return image
}
