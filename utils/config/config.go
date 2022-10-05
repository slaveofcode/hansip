package config

import (
	"strings"

	"github.com/spf13/viper"
)

func IsUsingS3Storage() bool {
	return strings.ToLower(viper.GetString("storage.type")) == "s3"
}

func GetS3Bucket() string {
	return viper.GetString("aws.s3.bucket_name")
}
