package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/google/uuid"
	imeta "github.com/opencontainers/image-spec/specs-go"
	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gopkg.in/resty.v1"
	"zotregistry.io/zot/errors"
	"zotregistry.io/zot/pkg/test"
)

func push(workdir string, url string, trepo string, size int,
	client *resty.Client) (map[string]string, map[string]string, map[string][]ispec.Descriptor, error) {
	var statusCode int

	// key: repo. value: manifest's name
	manifestHash := make(map[string]string)

	// key: manifest's name. value: manifest's config
	configHash := make(map[string]string)

	// key: manifest's name. value: a list of layers for the manifest
	layerHash := make(map[string][]ispec.Descriptor)

	ruid, err := uuid.NewUUID()
	if err != nil {
		return nil, nil, nil, err
	}

	var repo string

	if trepo != "" {
		repo = trepo + "/" + ruid.String()
	} else {
		repo = ruid.String()
	}

	// upload blob
	resp, err := resty.R().Post(fmt.Sprintf("%s/v2/%s/blobs/uploads/", url, repo))
	if err != nil {
		return nil, nil, nil, err
	}

	// request specific check
	statusCode = resp.StatusCode()
	if statusCode != http.StatusAccepted {
		return nil, nil, nil, errors.ErrUnknownCode
	}

	loc := test.Location(url, resp)
	blob := path.Join(workdir, fmt.Sprintf("%d.blob", size))

	fhandle, err := os.OpenFile(blob, os.O_RDONLY, defaultFilePerms)
	if err != nil {
		return nil, nil, nil, err
	}

	defer fhandle.Close()

	// stream the entire blob
	digest := blobHash[blob]

	resp, err = client.R().
		SetContentLength(true).
		SetQueryParam("digest", digest.String()).
		SetHeader("Content-Length", fmt.Sprintf("%d", size)).
		SetHeader("Content-Type", "application/octet-stream").SetBody(fhandle).Put(loc)

	if err != nil {
		return nil, nil, nil, err
	}

	// request specific check
	statusCode = resp.StatusCode()
	if statusCode != http.StatusCreated {
		return nil, nil, nil, errors.ErrUnknownCode
	}

	// upload image config blob
	resp, err = resty.R().
		Post(fmt.Sprintf("%s/v2/%s/blobs/uploads/", url, repo))

	if err != nil {
		return nil, nil, nil, err
	}

	// request specific check
	statusCode = resp.StatusCode()
	if statusCode != http.StatusAccepted {
		return nil, nil, nil, errors.ErrUnknownCode
	}

	loc = test.Location(url, resp)
	cblob, cdigest := test.GetRandomImageConfig()
	resp, err = client.R().
		SetContentLength(true).
		SetHeader("Content-Length", fmt.Sprintf("%d", len(cblob))).
		SetHeader("Content-Type", "application/octet-stream").
		SetQueryParam("digest", cdigest.String()).
		SetBody(cblob).
		Put(loc)

	if err != nil {
		return nil, nil, nil, err
	}

	// request specific check
	statusCode = resp.StatusCode()
	if statusCode != http.StatusCreated {
		return nil, nil, nil, errors.ErrUnknownCode
	}

	// create a manifest
	manifest := ispec.Manifest{
		Versioned: imeta.Versioned{
			SchemaVersion: defaultSchemaVersion,
		},
		Config: ispec.Descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    cdigest,
			Size:      int64(len(cblob)),
		},
		Layers: []ispec.Descriptor{
			{
				MediaType: "application/vnd.oci.image.layer.v1.tar",
				Digest:    digest,
				Size:      int64(size),
			},
		},
	}

	content, err := json.MarshalIndent(&manifest, "", "\t")
	if err != nil {
		return nil, nil, nil, err
	}

	manifestTag := fmt.Sprintf("tag%d", size)

	// finish upload
	resp, err = resty.R().
		SetContentLength(true).
		SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
		SetBody(content).
		Put(fmt.Sprintf("%s/v2/%s/manifests/%s", url, repo, manifestTag))

	if err != nil {
		return nil, nil, nil, err
	}

	// request specific check
	statusCode = resp.StatusCode()
	if statusCode != http.StatusCreated {
		return nil, nil, nil, errors.ErrUnknownCode
	}

	manifestHash[repo] = manifestTag
	configHash[manifestTag] = cdigest.String()
	layerHash[manifestTag] = manifest.Layers

	return manifestHash, configHash, layerHash, nil
}
