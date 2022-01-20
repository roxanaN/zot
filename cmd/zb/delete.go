package main

import (
	"fmt"
	"net/http"

	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gopkg.in/resty.v1"
	"zotregistry.io/zot/errors"
)

func deleteUploadedFiles(manifestHash map[string]string, configHash map[string]string,
	layerHash map[string][]ispec.Descriptor, url string, client *resty.Client) error {
	for repo, manifestTag := range manifestHash {
		layers := layerHash[manifestTag]

		for _, l := range layers {
			blobDigest := l.Digest

			resp, err := client.R().Delete((fmt.Sprintf("%s/v2/%s/blobs/%s", url, repo, blobDigest)))
			if err != nil {
				return err
			}

			// request specific check
			statusCode := resp.StatusCode()
			if statusCode != http.StatusAccepted {
				return errors.ErrUnknownCode
			}
		}

		resp, err := client.R().Delete((fmt.Sprintf("%s/v2/%s/blobs/%s", url, repo, configHash[manifestTag])))
		if err != nil {
			return err
		}

		// request specific check
		statusCode := resp.StatusCode()
		if statusCode != http.StatusAccepted {
			return errors.ErrUnknownCode
		}

		resp, err = client.R().Delete((fmt.Sprintf("%s/v2/%s/manifests/%s", url, repo, manifestTag)))
		if err != nil {
			return err
		}

		// request specific check
		statusCode = resp.StatusCode()
		if statusCode != http.StatusAccepted {
			return errors.ErrUnknownCode
		}
	}

	return nil
}
