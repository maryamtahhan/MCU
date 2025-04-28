/*
Copyright Red Hat Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cosignimg

import (
	"fmt"

	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/options"
	"github.com/sigstore/cosign/v2/cmd/cosign/cli/sign"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sirupsen/logrus"
	"github.com/tkdk/cargohold/pkg/fetcher"
)

func SignImage(imageRef string, cosignKey string, useSigstore bool) error {
	logrus.Infof("Signing image: %s", imageRef)

	imgFetcher := fetcher.NewImgFetcher()
	img, err := imgFetcher.FetchImg(imageRef)
	if err != nil {
		return fmt.Errorf("failed to fetch image: %w", err)
	}

	digest, err := img.Digest()
	if err != nil {
		return fmt.Errorf("failed to get image digest: %w", err)
	}

	// Build digest-based reference
	resolvedRef := fmt.Sprintf("%s@%s", imageRefWithoutTag(imageRef), digest.String())

	keyOpts := options.KeyOpts{
		KeyRef:           cosignKey,
		SkipConfirmation: true,
		FulcioURL:        options.DefaultFulcioURL,
		RekorURL:         options.DefaultRekorURL,
		OIDCIssuer:       options.DefaultOIDCIssuerURL,
		PassFunc: func(confirm bool) ([]byte, error) {
			return cosign.GetPassFromTerm(confirm)
		},
	}

	if useSigstore {
		logrus.Info("Using Sigstore keyless signing (OIDC + Fulcio + Rekor)")
		keyOpts.IDToken = ""
		keyOpts.KeyRef = ""
	}

	if !useSigstore && cosignKey == "" {
		return fmt.Errorf("cosign key path is required unless using Sigstore (--use-sigstore)")
	}

	signOpts := options.SignOptions{
		Upload:           true,
		TlogUpload:       true,
		SkipConfirmation: true,
		Registry: options.RegistryOptions{
			AllowInsecure: false,
			RegistryClientOpts: []remote.Option{
				remote.WithAuthFromKeychain(authn.DefaultKeychain),
				remote.WithRetryBackoff(remote.Backoff{
					Duration: 1 * time.Second,
					Jitter:   1.0,
					Factor:   2.0,
					Steps:    5,
					Cap:      2 * time.Minute,
				}),
			},
		},
	}

	rootOpts := &options.RootOptions{
		Timeout: 2 * time.Minute,
		Verbose: false,
	}

	err = sign.SignCmd(rootOpts, keyOpts, signOpts, []string{resolvedRef})
	if err != nil {
		return fmt.Errorf("failed to sign image: %w", err)
	}

	logrus.Infof("Successfully signed image: %s", resolvedRef)
	return nil
}

func imageRefWithoutTag(imageRef string) string {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return imageRef
	}
	return ref.Context().Name()
}
