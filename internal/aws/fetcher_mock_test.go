package aws

// mockEC2API and mockS3API provide in-memory implementations for testing
// without real AWS credentials.

type mockInstance struct {
	ID           string
	InstanceType string
	State        string
	AMI          string
}

type mockBucket struct {
	Name string
}

// buildResourcesFromInstances converts mock instances to Resource slices,
// mirroring the logic in FetchEC2Instances so tests can assert parity.
func buildResourcesFromInstances(instances []mockInstance) []Resource {
	resources := make([]Resource, 0, len(instances))
	for _, inst := range instances {
		resources = append(resources, Resource{
			Type: "aws_instance",
			ID:   inst.ID,
			Attributes: map[string]string{
				"instance_type": inst.InstanceType,
				"state":         inst.State,
				"ami":           inst.AMI,
			},
		})
	}
	return resources
}

// buildResourcesFromBuckets converts mock buckets to Resource slices.
func buildResourcesFromBuckets(buckets []mockBucket) []Resource {
	resources := make([]Resource, 0, len(buckets))
	for _, b := range buckets {
		resources = append(resources, Resource{
			Type:       "aws_s3_bucket",
			ID:         b.Name,
			Attributes: map[string]string{"bucket": b.Name},
		})
	}
	return resources
}
