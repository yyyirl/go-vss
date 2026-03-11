// package icloud implements Go bindings for the CloudKit Web Services API.
//
// Usage:
//
//	import "github.com/lukasmalkmus/icloud-go/icloud"
//
// Construct a new iCloud client, then use the various services on the client to
// access different parts of the CloudKit Web Services API. A valid container
// identifier and the environment to use must be passed:
//
//	client, err := icloud.NewClient("iCloud.com.lukasmalkmus.Example-App", icloud.Development)
//
// Get the version of the configured deployment:
//
//	version, err := client.Version.Get(ctx)
//
// Some API methods have additional parameters that can be passed:
//
//	dashboards, err := client.Dashboards.List(ctx, icloud.ListOptions{
//		Limit: 5,
//	})
//
// NOTE: Every client method mapping to an API method takes a context.Context
// (https://godoc.org/context) as its first parameter to pass cancelation
// signals and deadlines to requests. In case there is no context available,
// then context.Background() can be used as a starting point.
//
// For more code samples, check out the https://github.com/lukasmalkmus/icloud-go/tree/main/examples directory.
package icloud
