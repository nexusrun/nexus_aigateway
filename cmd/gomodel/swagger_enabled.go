//go:build swagger

package main

import swaggerdocs "github.com/nexusrun/nexus_aigateway/cmd/gomodel/docs"

func configureSwaggerDocs(basePath string) {
	swaggerdocs.SwaggerInfo.BasePath = basePath
}
