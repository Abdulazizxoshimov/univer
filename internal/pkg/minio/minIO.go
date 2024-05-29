package minio

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
)

func MinIOBucket(bucketName string, minIO *minio.Client) error {

	err := minIO.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "BucketAlreadyOwnedByYou" {

		} else {
			log.Println(err.Error())
			return err
		}
	}

	policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {
						"AWS": ["*"]
					},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucketName)

	err = minIO.SetBucketPolicy(context.Background(), bucketName, policy)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}
