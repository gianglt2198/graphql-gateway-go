package s3

import (
	"context"
	"fmt"
	"image"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gianglt2198/graphql-gateway-go/pkg/config"
	"github.com/gianglt2198/graphql-gateway-go/pkg/infra/monitoring"
)

type S3Storage struct {
	log *monitoring.AppLogger

	cfg    config.S3Config
	client *s3.S3
}

type S3StorageParams struct {
	fx.In

	Log *monitoring.AppLogger
	Cfg config.S3Config
}

type S3StorageResult struct {
	fx.Out

	S3Storage *S3Storage
}

func NewS3(params S3StorageParams) S3StorageResult {
	s, err := session.NewSession(&aws.Config{
		Region: aws.String(params.Cfg.Region)},
	)
	if err != nil {
		panic(zap.Error(err))
	}
	return S3StorageResult{
		S3Storage: &S3Storage{
			log:    params.Log,
			cfg:    params.Cfg,
			client: s3.New(s),
		},
	}
}

func (s *S3Storage) DeleteObject(ctx context.Context, key string) error {
	s.log.InfoC(ctx, "DeleteObject", zap.String("Key", key))
	_, perr := s.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if perr != nil {
		s.log.ErrorC(ctx, "DeleteObjectWithContext", zap.Error(perr))
		return perr
	}
	return nil
}

func (s *S3Storage) DeleteObjects(ctx context.Context, keys []string) (successes []string, failures []string, err error) {
	s.log.InfoC(ctx, "DeleteObjects %s", zap.Strings("Keys", keys))

	result, perr := s.client.DeleteObjectsWithContext(
		ctx,
		&s3.DeleteObjectsInput{
			Bucket: aws.String(s.cfg.Bucket),
			Delete: &s3.Delete{
				Objects: lo.Map(keys, func(item string, index int) *s3.ObjectIdentifier {
					return &s3.ObjectIdentifier{
						Key: aws.String(item),
					}
				}),
			},
		},
	)
	if perr != nil {
		s.log.ErrorC(ctx, "DeleteObjectsWithContext", zap.Error(perr))
		err = perr
	}

	successes = []string{}
	failures = []string{}

	for _, item := range result.Deleted {
		successes = append(successes, *item.Key)
	}

	for _, item := range result.Errors {
		failures = append(failures, *item.Key)
	}

	return successes, failures, err
}

func (s *S3Storage) DeleteMultiObject(ctx context.Context, files []string) error {
	s.log.InfoC(ctx, "DeleteMultiObject", zap.Strings("files", files))

	var wg sync.WaitGroup
	errorsChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			if err := s.DeleteObject(ctx, f); err != nil {
				errorsChan <- err
			}
		}(file)
	}

	wg.Wait()
	close(errorsChan)

	var errors []error
	for err := range errorsChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func (s *S3Storage) GeneratePresignedUploadURL(ctx context.Context, dto *PresignedUploadURL) (string, error) {
	s.log.InfoC(ctx, "GeneratePresignedUploadURL", zap.Any("dto", dto))
	req, _ := s.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:        aws.String(s.cfg.Bucket),
		Key:           aws.String(dto.Key),
		ContentType:   aws.String(dto.ContentType),
		ContentLength: aws.Int64(dto.ContentLength),
	})
	req.SetContext(ctx)
	if url, err := req.Presign(dto.Expiry); err != nil {
		s.log.ErrorC(ctx, "GeneratePresignedUploadURL", zap.Error(err))
		return "", err
	} else {
		return url, nil
	}
}

func (s *S3Storage) GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	s.log.InfoC(ctx, "GeneratePresignedDownloadURL", zap.String("Key", key), zap.Duration("Expiry", expiry))

	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	req.SetContext(ctx)

	if url, err := req.Presign(expiry); err != nil {
		s.log.ErrorC(ctx, "GeneratePresignedDownloadURL", zap.Error(err))
		return "", err
	} else {
		return url, nil
	}
}

func (s *S3Storage) GetContentLength(ctx context.Context, key string) (int64, error) {
	s.log.InfoC(ctx, "GetContentLength", zap.String("Key", key))
	if headObjResult, err := s.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	}); err == nil {
		return aws.Int64Value(headObjResult.ContentLength), nil
	} else {
		return 0, err
	}
}

func (s *S3Storage) GetDimensions(ctx context.Context, key string) (*image.Config, error) {
	s.log.InfoC(ctx, "GetDimensions", zap.String("Key", key))
	obj, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		s.log.ErrorC(ctx, "GetDimensions", zap.Error(err))
		return nil, err
	}
	defer obj.Body.Close()

	imgConf, err := GetImageDimensions(obj.Body)
	if err != nil {
		s.log.ErrorC(ctx, "GetDimensions", zap.Error(err))
		return nil, err
	}
	return imgConf, nil
}

func (s *S3Storage) CopyS3File(ctx context.Context, srcKey, destKey string) error {
	s.log.InfoC(ctx, "CopyS3File", zap.String("srcKey", srcKey), zap.String("destKey", destKey))

	input := &s3.CopyObjectInput{
		Bucket:     aws.String(s.cfg.Bucket),
		Key:        &destKey,
		CopySource: aws.String(fmt.Sprintf("%s/%s", s.cfg.Bucket, srcKey)),
	}

	_, copyErr := s.client.CopyObjectWithContext(ctx, input)
	if copyErr != nil {
		s.log.ErrorC(ctx, "CopyS3File", zap.Error(copyErr))
		return copyErr
	}
	return nil
}

func (s *S3Storage) MoveObject(ctx context.Context, sourceKey, destKey string) error {
	s.log.InfoC(ctx, "MoveObject", zap.String("sourceKey", sourceKey), zap.String("destKey", destKey))
	if err := s.CopyS3File(ctx, sourceKey, destKey); err != nil {
		s.log.ErrorC(ctx, "CopyS3File", zap.Error(err))
		return err
	}
	if err := s.DeleteObject(ctx, sourceKey); err != nil {
		s.log.ErrorC(ctx, "DeleteObject", zap.Error(err))
		return err
	}
	return nil
}
