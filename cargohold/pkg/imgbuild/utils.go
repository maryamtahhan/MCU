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
package imgbuild

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	logging "github.com/sirupsen/logrus"
)

const DockerfileTemplate = `FROM scratch
LABEL org.opencontainers.image.title={{ .ImageTitle }}
COPY io.triton.cache/ io.triton.cache/
`

type DockerfileData struct {
	ImageTitle string
	CacheDir   string
}

func generateDockerfile(imageName, cacheDir, outputPath string) error {
	parts := strings.Split(imageName, "/")
	fullImageName := parts[len(parts)-1]
	imageTitle := strings.Split(fullImageName, ":")[0]

	data := DockerfileData{
		ImageTitle: imageTitle,
		CacheDir:   cacheDir,
	}

	tmpl, err := template.New("dockerfile").Parse(DockerfileTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating Dockerfile: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	logging.Infof("Dockerfile generated successfully at %s", outputPath)
	return nil
}
