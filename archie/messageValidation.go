package archie

import (
	"archie/event"
	"fmt"
	"golang.org/x/exp/slices"
	"strings"
)

// validate event name is allowed
func (a *Archiver) validateEventName(event event.Minio) error {
	validEvents := []string{"s3:ObjectCreated:Put", "s3:ObjectCreated:CompleteMultipartUpload", "s3:ObjectRemoved:Delete"}
	if !slices.Contains(validEvents, event.EventName) {
		return fmt.Errorf("event name not in list of valid events: [%s], terminating retries", strings.Join(validEvents, ", "))
	}
	return nil
}

// validate src bucket name in config matches event bucket name
func (a *Archiver) validateEventBucket(eventBucket string) error {
	if !a.SkipEventBucketValidation {
		if a.SrcBucket != eventBucket {
			return fmt.Errorf("event bucket (%s) doesn't match configured source bucket (%s)", eventBucket, a.SrcBucket)
		}
	}
	return nil
}
