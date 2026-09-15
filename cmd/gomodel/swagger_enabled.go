//go:build swagger

package main

import swaggerdocs "github.com/nexusrun/nexus_aigateway/cmd/gomodel/docs"

func configureSwaggerDocs(basePath string) {
	swaggerdocs.SwaggerInfo.Title = "NEXUS AI Gateway"
	swaggerdocs.SwaggerInfo.BasePath = basePath
}
