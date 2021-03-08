package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/alexflint/go-filemutex"
	"github.com/binxio/gcloudconfig"
	"github.com/mollie/tf-provider-registry-api-generator/signing_key"
	"github.com/mollie/tf-provider-registry-api-generator/versions"
	"github.com/docopt/docopt-go"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"log"
	"os"
	"regexp"
	"strings"
)

type Options struct {
	BucketName            string
	Namespace             string
	Url                   string
	Prefix                string
	Fingerprint           string
	Protocols             string
	UseDefaultCredentials bool
	Help                  bool
	Version               bool
	storage               *storage.Client
	bucket                *storage.BucketHandle
	credentials           *google.Credentials
	mutexFileName         string
	mutex                 *filemutex.FileMutex
	protocols             []string
}

var (
	version       = "dev"
	commit        = "none"
	date          = "unknown"
	builtBy       = "unknown"
	protocolRegex = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
)

func main() {
	var options Options
	usage := `generate terraform provider registry API documents.

Usage:
  tf-provider-registry-api-generator [--use-default-credentials] [--fingerprint FINGERPRINT]  --bucket-name BUCKET --url URL --namespace NAMESPACE [--protocols PROTOCOLS ] --prefix PREFIX
  tf-provider-registry-api-generator version
  tf-provider-registry-api-generator -h | --help

Options:
  --bucket-name BUCKET       - bucket containing the binaries and the website.
  --url URL                  - of the static website.
  --namespace NAMESPACE      - for the providers.
  --prefix PREFIX            - location of the released binaries in the bucket.
  --protocols PROTOCOL       - comma separated list of supported provider protocols by the provider [default: 5.0]
  --fingerprint FINGERPRINT  - of the public key used to sign, defaults to environment variable GPG_FINGERPRINT.
  --use-default-credentials  - instead of the current gcloud configuration.
  -h --help                  - shows this.
`

	arguments, err := docopt.ParseDoc(usage)
	if err != nil {
		log.Fatalf("ERROR: failed to parse command line, %s", err)
	}
	if err = arguments.Bind(&options); err != nil {
		log.Fatalf("ERROR: failed to bind arguments from command line, %s", err)
	}

	if options.Version {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}

	options.protocols = make([]string, 0)
	for _, p := range strings.Split(options.Protocols, ",") {
		if !protocolRegex.Match([]byte(p)) {
			log.Fatalf("ERROR: %s is not a version number", p)
		}
		options.protocols = append(options.protocols, p)
	}
	if len(options.protocols) == 0 {
		log.Fatalf("ERROR: no protocols specified")
	}

	if options.Fingerprint == "" {
		options.Fingerprint = os.Getenv("GPG_FINGERPRINT")
		if options.Fingerprint == "" {
			log.Fatalf("ERROR: no fingerprint specified")
		}
	}
	options.mutexFileName = fmt.Sprintf("/tmp/tf-registry-generator-%s.lck", options.BucketName)

	if options.UseDefaultCredentials || !gcloudconfig.IsGCloudOnPath() {
		log.Printf("INFO: using default credentials")
		if options.credentials, err = google.FindDefaultCredentials(context.Background(), "https://www.googleapis.com/auth/devstorage.full_control"); err != nil {
			log.Fatalf("ERROR: failed to get default credentials, %s", err)
		}
	} else {
		if options.credentials, err = gcloudconfig.GetCredentials(""); err != nil {
			log.Fatalf("ERROR: failed to get gcloud config credentials, %s", err)
		}
	}

	options.storage, err = storage.NewClient(context.Background(), option.WithCredentials(options.credentials))
	if err != nil {
		log.Fatalf("ERROR: could not create storage client, %s", err)
	}
	defer options.storage.Close()

	options.bucket = options.storage.Bucket(options.BucketName)
	options.mutex, err = filemutex.New(options.mutexFileName)
	if err != nil {
		log.Fatalf("ERROR: failed to create lock file %s, %s", options.mutexFileName, err)
	}
	defer options.mutex.Close()

	err = options.mutex.Lock()
	if err != nil {
		log.Fatalf("ERROR: failed to obtain lock, %s", err)
	}

	signingKey := signing_key.GetPublicSigningKey(options.Fingerprint)
	files := versions.LoadFromBucket(options.bucket, options.Prefix)
	if len(files) == 0 {
		log.Fatalf("ERROR: no release files found in %s at %s", options.BucketName, options.Prefix)
	}

	shasums := make(map[string]string, len(files))
	for _, filename := range files {
		if strings.HasSuffix(filename, "SHA256SUMS") {
			err = readShasums(options.bucket, filename, shasums)
			if err != nil {
				log.Fatalf("%s", err)
			}
		}
	}

	binaries := versions.CreateFromFileList(files, options.Url, signingKey, shasums, options.protocols)
	providers := binaries.ExtractVersions()
	if len(providers) == 0 {
		log.Fatalf("ERROR: no terraform provider binaries detected")
	}

	WriteAPIDocuments(options.bucket, options.Namespace, binaries)
}
