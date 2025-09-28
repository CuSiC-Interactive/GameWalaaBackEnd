package services

import (
	"GameWala-Arcade/config"
	"GameWala-Arcade/models"
	"GameWala-Arcade/repositories"
	"GameWala-Arcade/utils"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3BucketInfo struct {
	Client        *s3.Client
	BucketName    string
	Prefix        string
	Region        string
	PublicURLBase string
}

type MarketPlaceService interface {
	FetchProducts(productType models.ProductType) ([]models.Product, error)
}

type marketPlaceService struct {
	marketPlaceRepository repositories.MarketPlaceRepository
}

func NewMarketPlaceService(marketPlaceRepository repositories.MarketPlaceRepository) *marketPlaceService {
	return &marketPlaceService{marketPlaceRepository: marketPlaceRepository}
}

func (s *marketPlaceService) FetchProducts(productType models.ProductType) ([]models.Product, error) {
	products, err := s.marketPlaceRepository.FetchProducts(productType)

	if err != nil {
		utils.LogError("Some error occured while fetching data from DB: %v", err)
		return nil, err
	}

	// aws stuff
	supabaseURL := fmt.Sprintf("https://%s.storage.supabase.co/storage/v1/s3", config.GetString("supabaseProjectID"))

	creds := credentials.NewStaticCredentialsProvider("ac72365485137b2b1d4e423fed8bd915", "137a90eb7b0a6555457ce3696c98b1febb047f81933e70239ea819da1c032b14", "")

	// Load the configuration, providing the custom endpoint resolver and static credentials.
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(config.GetString("region")),
		awsconfig.WithCredentialsProvider(creds),
		awsconfig.WithClientLogMode(aws.LogRequest|aws.LogRequestWithBody),
	)

	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// Set the BaseEndpoint directly to the custom URL
		o.BaseEndpoint = &supabaseURL
		o.UsePathStyle = true
		o.Region = config.GetString("region")
	})

	// This is the base URL for constructing the final public links.
	publicURLBase := fmt.Sprintf("https://%s.supabase.co/storage/v1/object/public", config.GetString("supabaseProjectID"))

	storageInfo := S3BucketInfo{
		Client:        s3Client,
		BucketName:    config.GetString("bucketName"),
		Prefix:        config.GetString("prefix"),
		PublicURLBase: publicURLBase,
	}

	for i := range products {
		// cover image
		coverImageDir := fmt.Sprintf("%s%d/", storageInfo.Prefix, products[i].ProductId)
		storageInfo.Prefix = coverImageDir
		images, err := getImageLinks(context.TODO(), storageInfo)
		if err == nil {
			products[i].CoverImage = images[0]
			products[i].Images = images[1:]
		}
	}
	for _, product := range products {
		fmt.Print(product.CoverImage)
		fmt.Print(product.Images)
	}
	return products, nil
}

func getImageLinks(ctx context.Context, info S3BucketInfo) ([]string, error) {
	var imageLinks []string

	// The S3 paginator works with Supabase because its API is S3-compatible.
	paginator := s3.NewListObjectsV2Paginator(info.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(info.BucketName),
		Prefix: aws.String(info.Prefix),
	})

	// Loop through each page of results
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			utils.LogError("Some error occured accessing bucket %v", err)
			return nil, fmt.Errorf("failed to get page of objects: %w", err)
		}

		// Loop through the objects in the current page
		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)

			if key == info.Prefix {
				continue
			}

			// Check if the object key has one of the desired image extensions.
			isImage := strings.HasSuffix(strings.ToLower(key), ".jpg") ||
				strings.HasSuffix(strings.ToLower(key), ".jpeg") ||
				strings.HasSuffix(strings.ToLower(key), ".png")

			if isImage {
				// Construct the public Supabase URL for the object.
				// Format: https://<project-id>.supabase.co/storage/v1/object/public/<bucketname>/<object-key>
				url := fmt.Sprintf("%s/%s/%s", info.PublicURLBase, info.BucketName, key)
				imageLinks = append(imageLinks, url)
			}
		}
	}

	return imageLinks, nil
}
