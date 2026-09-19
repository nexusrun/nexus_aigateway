//go:build swagger

package main

import swaggerdocs "github.com/nexusrun/nexus_aigateway/cmd/aigateway/docs"

func configureSwaggerDocs(basePath string) {
	swaggerdocs.SwaggerInfo.Title = "NEXUS AI Gateway"
	swaggerdocs.SwaggerInfo.BasePath = basePath
}
