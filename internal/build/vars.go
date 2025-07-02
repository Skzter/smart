//nolint:gochecknoglobals
package build

// Compile-time variables injected via ldflags
var (
	Version            string
	OpenAIKey          string
	AwsAccessKey       string
	AwsSecretAccessKey string
)
