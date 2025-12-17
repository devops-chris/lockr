package ssm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// Secret represents a secret from SSM Parameter Store
type Secret struct {
	Name        string
	Value       string
	Type        string
	Version     int64
	Description string
	Tags        map[string]string
}

// SecretMetadata represents secret metadata without the value
type SecretMetadata struct {
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Version      int64      `json:"version"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	Description  string     `json:"description,omitempty"`
	Tier         string     `json:"tier,omitempty"`
}

// Client wraps the SSM client
type Client struct {
	ssm *ssm.Client
}

// NewClient creates a new SSM client
func NewClient(region string) (*Client, error) {
	ctx := context.Background()

	var opts []func(*config.LoadOptions) error
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		ssm: ssm.NewFromConfig(cfg),
	}, nil
}

// WriteSecret writes a secret to SSM Parameter Store
// Handles the AWS limitation where tags can't be set with overwrite
func (c *Client) WriteSecret(path, value string, tags map[string]string, overwrite bool, kmsKey string) error {
	ctx := context.Background()

	input := &ssm.PutParameterInput{
		Name:  aws.String(path),
		Value: aws.String(value),
		Type:  types.ParameterTypeSecureString,
	}

	// Set KMS key if provided
	if kmsKey != "" {
		input.KeyId = aws.String(kmsKey)
	}

	// AWS doesn't allow tags with overwrite, so we handle this in two steps:
	// 1. Try to create/update the parameter
	// 2. If tags provided, add them separately

	if len(tags) > 0 {
		// First, try without overwrite (new parameter)
		_, err := c.ssm.PutParameter(ctx, input)
		if err != nil {
			// Check if it's a parameter already exists error
			var pae *types.ParameterAlreadyExists
			if errors.As(err, &pae) && overwrite {
				// Parameter exists, update with overwrite (no tags)
				input.Overwrite = aws.Bool(true)
				_, err = c.ssm.PutParameter(ctx, input)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// Now add/update tags separately
		return c.SetTags(path, tags)
	}

	// No tags - simple path
	input.Overwrite = aws.Bool(overwrite)
	_, err := c.ssm.PutParameter(ctx, input)
	return err
}

// SetTags sets tags on a parameter (replaces existing tags with same keys)
func (c *Client) SetTags(path string, tags map[string]string) error {
	ctx := context.Background()

	var ssmTags []types.Tag
	for k, v := range tags {
		ssmTags = append(ssmTags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	_, err := c.ssm.AddTagsToResource(ctx, &ssm.AddTagsToResourceInput{
		ResourceType: types.ResourceTypeForTaggingParameter,
		ResourceId:   aws.String(path),
		Tags:         ssmTags,
	})
	return err
}

// ReadSecret reads a secret from SSM Parameter Store
func (c *Client) ReadSecret(path string) (*Secret, error) {
	ctx := context.Background()

	// Get parameter value
	result, err := c.ssm.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(path),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	secret := &Secret{
		Name:    aws.ToString(result.Parameter.Name),
		Value:   aws.ToString(result.Parameter.Value),
		Type:    string(result.Parameter.Type),
		Version: result.Parameter.Version,
	}

	// Get tags
	tagsResult, err := c.ssm.ListTagsForResource(ctx, &ssm.ListTagsForResourceInput{
		ResourceType: types.ResourceTypeForTaggingParameter,
		ResourceId:   aws.String(path),
	})
	if err == nil && len(tagsResult.TagList) > 0 {
		secret.Tags = make(map[string]string)
		for _, tag := range tagsResult.TagList {
			secret.Tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
		}
	}

	return secret, nil
}

// ListSecrets lists secrets at a path
func (c *Client) ListSecrets(path string, recursive bool) ([]SecretMetadata, error) {
	ctx := context.Background()

	input := &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		Recursive:      aws.Bool(recursive),
		WithDecryption: aws.Bool(false), // Don't decrypt for listing
	}

	var secrets []SecretMetadata
	paginator := ssm.NewGetParametersByPathPaginator(c.ssm, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, p := range page.Parameters {
			meta := SecretMetadata{
				Name:    aws.ToString(p.Name),
				Type:    string(p.Type),
				Version: p.Version,
			}
			if p.LastModifiedDate != nil {
				meta.LastModified = p.LastModifiedDate
			}
			secrets = append(secrets, meta)
		}
	}

	return secrets, nil
}

// DeleteSecret deletes a secret from SSM Parameter Store
func (c *Client) DeleteSecret(path string) error {
	ctx := context.Background()

	_, err := c.ssm.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(path),
	})
	return err
}

// Exists checks if a parameter exists
func (c *Client) Exists(path string) (bool, error) {
	ctx := context.Background()

	_, err := c.ssm.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(path),
		WithDecryption: aws.Bool(false),
	})
	if err != nil {
		var pnf *types.ParameterNotFound
		if errors.As(err, &pnf) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
