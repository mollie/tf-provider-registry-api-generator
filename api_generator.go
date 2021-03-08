package main

import (
	"bufio"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mollie/tf-provider-registry-api-generator/versions"
	"io/ioutil"
	"log"
	"path"
	"reflect"
	"strings"
)

func assertDiscoveryDocument(bucket *storage.BucketHandle) {
	content := make(map[string]string)
	expect := map[string]string{
		"providers.v1": "/v1/providers/",
	}

	p := path.Join(".well-known", "terraform.json")
	err := readJson(bucket, p, &content)
	if err != nil {
		log.Fatalf("could not read content of %s, %s", p, err)
	}

	if !reflect.DeepEqual(expect, content) {
		log.Printf("INFO: writing content to %s", p)
		writeJson(bucket, p, expect)
	} else {
		log.Printf("INFO: discovery document is up-to-date\n")
	}
}

func readJson(bucket *storage.BucketHandle, filename string, object interface{}) error {
	r, err := bucket.Object(filename).NewReader(context.Background())
	if err != nil {
		if err.Error() == "storage: object doesn't exist" {
			return nil
		}
		return fmt.Errorf("ERROR: failed to read file %s, %s", filename, err)
	}
	defer r.Close()
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("ERROR: failed to read content from %s, %s", filename, err)
	}
	err = json.Unmarshal(body, &object)
	if err != nil {
		return fmt.Errorf("ERROR: failed to unmarshal %s, %s", filename, err)
	}

	return nil
}

func readShasums(bucket *storage.BucketHandle, filename string, shasums map[string]string) error {
	r, err := bucket.Object(filename).NewReader(context.Background())
	if err != nil {
		if err.Error() == "storage: object doesn't exist" {
			return nil
		}
		return fmt.Errorf("ERROR: failed to read file %s, %s", filename, err)
	}
	defer r.Close()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			log.Fatalf("ERROR: expected %s to contain 2 fields on each line, found %d", filename, len(fields))
		}
		shasums[fields[1]] = fields[0]
	}
	return nil
}

func writeJson(bucket *storage.BucketHandle, filename string, content interface{}) {
	log.Printf("INFO: writing %s", filename)

	w := bucket.Object(filename).NewWriter(context.Background())
	w.ContentType = "application/json"
	w.CacheControl = "no-cache, max-age=60"

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(content)
	if err != nil {
		log.Fatalf("INFO: failed to write %s, %s", filename, err)
	}
	if err = w.Close(); err != nil {
		log.Fatalf("INFO: failed to close %s, %s", filename, err)
	}
}

func writeProviderVersions(bucket *storage.BucketHandle, directory string, newVersions *versions.ProviderVersions) {
	var existing versions.ProviderVersions
	if err := readJson(bucket, path.Join(directory, "versions"), &existing); err != nil {
		log.Fatalf("ERROR: failed to read the %s/versions, %s", directory, err)
	}
	if reflect.DeepEqual(&existing, newVersions) {
		log.Printf("INFO: %s/versions already up-to-date", directory)
		return
	}
	existing.Merge(*newVersions)
	writeJson(bucket, path.Join(directory, "versions"), existing)
}

func writeProviderVersion(bucket *storage.BucketHandle, directory string, version *versions.BinaryMetaData) {
	filename := path.Join(directory, version.Version, "download", version.Os, version.Arch)
	existing := versions.BinaryMetaData{}

	if err := readJson(bucket, filename, &existing); err != nil {
		log.Fatalf("ERROR: failed to read %s, %s", filename, err)
	}

	if existing.Equals(version) {
		log.Printf("INFO: %s is up-to-date", filename)
		return
	}
	writeJson(bucket, filename, version)
}

func WriteAPIDocuments(bucket *storage.BucketHandle, namespace string, binaries versions.BinaryMetaDataList) {
	assertDiscoveryDocument(bucket)

	providerDirectory := path.Join("v1", "providers", namespace)
	providers := binaries.ExtractVersions()

	for _, binary := range binaries {
		writeProviderVersion(bucket, path.Join(providerDirectory, binary.TypeName), &binary)
	}

	for name, versions := range providers {
		writeProviderVersions(bucket, path.Join(providerDirectory, name), versions)
	}

}
