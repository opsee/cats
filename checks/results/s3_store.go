package results

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gogo/protobuf/proto"
	"github.com/opsee/basic/schema"
)

// S3Store stores CheckResult objects in S3 by ResultId (check_id:bastion_id).
type S3Store struct {
	BucketName string
	S3Client   *s3.S3
}

// GetResultsByCheckId gets the latest CheckResults for a Check from persistent storage.
func (s *S3Store) GetResultsByCheckId(checkId string) ([]*schema.CheckResult, error) {
	resultsPath := fmt.Sprintf("latest/%s", checkId)
	resp, err := s.S3Client.ListObjects(&s3.ListObjectsInput{
		Bucket:    aws.String(s.BucketName),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(resultsPath),
	})
	if err != nil {
		return nil, err
	}

	results := make([]*schema.CheckResult, len(resp.Contents))
	for i, obj := range resp.Contents {
		getObjResp, err := s.S3Client.GetObject(&s3.GetObjectInput{
			Bucket:              aws.String(s.BucketName),
			Key:                 obj.Key,
			ResponseContentType: aws.String("application/octet-stream"),
		})
		if err != nil {
			return nil, err
		}

		bodyBytes, err := ioutil.ReadAll(getObjResp.Body)
		getObjResp.Body.Close()
		if err != nil {
			return nil, err
		}

		result := &schema.CheckResult{}
		if err := proto.Unmarshal(bodyBytes, result); err != nil {
			return nil, err
		}
		results[i] = result
	}

	return results, nil
}

// PutResult puts a CheckResult to persistent storage.
func (s *S3Store) PutResult(result *schema.CheckResult) error {
	timestamp := time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos)).UTC()
	resultPath := fmt.Sprintf("latest/%s/%s/%d.pb", result.CheckId, result.BastionId, timestamp)

	resultBytes, err := proto.Marshal(result)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(resultBytes)

	_, err = s.S3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(resultPath),
		Body:   reader,
	})

	if err != nil {
		return err
	}

	return nil
}
