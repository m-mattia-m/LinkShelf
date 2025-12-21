package controller

import (
	"backend/internal/domain"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Router(svc *domain.Service) (*gin.Engine, error) {
	if viper.GetString("app.env") == "PROD" || viper.GetString("app.env") == "prd" {
		gin.SetMode(gin.ReleaseMode)
	}

	hostWithScheme := fmt.Sprintf("%s://%s", viper.GetString("server.scheme"), viper.GetString("server.host"))
	host := viper.GetString("server.host")
	if viper.GetBool("domain.openapi.usePort") {
		host = fmt.Sprintf("%s:%d", host, viper.GetInt("server.port"))
		hostWithScheme = fmt.Sprintf("%s:%d", hostWithScheme, viper.GetInt("server.port"))
	}
	slog.Debug(fmt.Sprintf("Host: %s", hostWithScheme))

	humaConfig := huma.DefaultConfig(viper.GetString("app.name"), viper.GetString("app.version"))
	humaConfig.Info = &huma.Info{
		Title:       viper.GetString("app.name"),
		Description: viper.GetString("app.description"),
		License:     nil,
		Version:     viper.GetString("app.version"),
	}
	humaConfig.Servers = []*huma.Server{
		{URL: hostWithScheme},
		{Description: fmt.Sprintf("This is the default server of %s", viper.GetString("app.name"))},
	}
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		viper.GetString("app.name"): {
			Type: "oauth2",
			Flows: &huma.OAuthFlows{
				AuthorizationCode: &huma.OAuthFlow{
					AuthorizationURL: fmt.Sprintf("%s/oauth/v2/authorize", viper.GetString("authentication.oidc.issuer")),
					TokenURL:         fmt.Sprintf("%s/oauth/v2/token", viper.GetString("authentication.oidc.issuer")),
					RefreshURL:       fmt.Sprintf("%s/oauth/v2/token", viper.GetString("authentication.oidc.issuer")),
					Scopes: map[string]string{
						"openid":         "To return the openid basic information.",
						"profile":        "To return the profile attributes like name.",
						"email":          "To return the email address.",
						"offline_access": "?",
					},
				},
			},
		},
	}

	router := gin.Default()
	api := humagin.New(router, humaConfig)
	api.UseMiddleware(NewAuthorizationMiddleware(api))

	router.GET("/health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})
	router.GET("/health/readiness", func(c *gin.Context) {
		// You can add your readiness checks here (e.g., database connection)
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	trustedProxies, err := getTrustedProxies()
	if err != nil {
		return nil, err
	}

	err = router.SetTrustedProxies(trustedProxies)
	if err != nil {
		return nil, err
	}

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/swagger")
	})

	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		OperationID:   "post-create-user",
		Summary:       "Create user",
		Description:   "Create a new user.",
		Path:          "/v1/users",
		Tags:          []string{"User"},
		DefaultStatus: http.StatusCreated,
	}, CreateUser(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-user-by-id",
		Summary:     "Get user by ID",
		Description: "Get a user by ID.",
		Path:        "/v1/users/{userId}",
		Tags:        []string{"User"},
	}, GetUserById(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-user",
		Summary:     "Update user",
		Description: "Update an existing user. Consider that password updates are not handled here.",
		Path:        "/v1/users/{userId}",
		Tags:        []string{"User"},
	}, UpdateUser(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPatch,
		OperationID: "patch-user-password",
		Summary:     "Patch user password",
		Description: "Patch a user's password.",
		Path:        "/v1/users/{userId}/password",
		Tags:        []string{"User"},
	}, PatchUserPassword(svc))
	huma.Register(api, huma.Operation{
		Method:        http.MethodDelete,
		OperationID:   "delete-user",
		Summary:       "Delete user",
		Description:   "Delete a user by ID.",
		Path:          "/v1/users/{userId}",
		Tags:          []string{"User"},
		DefaultStatus: http.StatusNoContent,
	}, DeleteUser(svc))

	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		OperationID:   "post-create-shelf",
		Summary:       "Create shelf",
		Description:   "Create a new shelf.",
		Path:          "/v1/shelves",
		Tags:          []string{"Shelf"},
		DefaultStatus: http.StatusCreated,
	}, CreateShelf(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-shelf-by-id",
		Summary:     "Get shelf by ID",
		Description: "Get a shelf by ID.",
		Path:        "/v1/shelves/{shelfId}",
		Tags:        []string{"Shelf"},
	}, GetShelfById(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-shelf",
		Summary:     "Update shelf",
		Description: "Update an existing shelf.",
		Path:        "/v1/shelves/{shelfId}",
		Tags:        []string{"Shelf"},
	}, UpdateShelf(svc))
	huma.Register(api, huma.Operation{
		Method:        http.MethodDelete,
		OperationID:   "delete-shelf",
		Summary:       "Delete shelf",
		Description:   "Delete a shelf by ID.",
		Path:          "/v1/shelves/{shelfId}",
		Tags:          []string{"Shelf"},
		DefaultStatus: http.StatusNoContent,
	}, DeleteShelf(svc))

	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		OperationID:   "post-create-section",
		Summary:       "Create section",
		Description:   "Create a new section.",
		Path:          "/v1/sections",
		Tags:          []string{"Section"},
		DefaultStatus: http.StatusCreated,
	}, CreateSection(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-sections",
		Summary:     "Get sections by shelf ID",
		Description: "Get sections by shelf ID.",
		Path:        "/v1/sections",
		Tags:        []string{"Section"},
	}, GetSections(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-section",
		Summary:     "Update section",
		Description: "Update an existing section.",
		Path:        "/v1/sections/{sectionId}",
		Tags:        []string{"Section"},
	}, UpdateSection(svc))
	huma.Register(api, huma.Operation{
		Method:        http.MethodDelete,
		OperationID:   "delete-section",
		Summary:       "Delete section",
		Description:   "Delete a section by ID.",
		Path:          "/v1/sections/{sectionId}",
		Tags:          []string{"Section"},
		DefaultStatus: http.StatusNoContent,
	}, DeleteSection(svc))

	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		OperationID:   "post-create-link",
		Summary:       "Create link",
		Description:   "Create a new link.",
		Path:          "/v1/links",
		Tags:          []string{"Link"},
		DefaultStatus: http.StatusCreated,
	}, CreateLink(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-links",
		Summary:     "Get links by shelf ID",
		Description: "Get links by shelf ID.",
		Path:        "/v1/links",
		Tags:        []string{"Link"},
	}, GetLinks(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-link",
		Summary:     "Update link",
		Description: "Update an existing link.",
		Path:        "/v1/links/{linkId}",
		Tags:        []string{"Link"},
	}, UpdateLink(svc))
	huma.Register(api, huma.Operation{
		Method:        http.MethodDelete,
		OperationID:   "delete-link",
		Summary:       "Delete link",
		Description:   "Delete a link by ID.",
		Path:          "/v1/links/{linkId}",
		Tags:          []string{"Link"},
		DefaultStatus: http.StatusNoContent,
	}, DeleteLink(svc))

	router.GET("/swagger", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `<!DOCTYPE html>
	<html lang="en">
	<head>
	 <meta charset="utf-8" />
	 <meta name="viewport" content="width=device-width, initial-scale=1" />
	 <meta name="description" content="SwaggerUI" />
	 <title>SwaggerUI</title>
	 <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
	</head>
	<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
	<script>
	 window.onload = () => {
	   window.ui = SwaggerUIBundle({
	     url: '/openapi.json',
	     dom_id: '#swagger-ui',
	   });
	 };
	</script>
	</body>
	</html>`)
	})

	return router, nil
}

func getTrustedProxies() ([]string, error) {
	var proxies []string

	if err := viper.UnmarshalKey("server.trustedProxies", &proxies); err != nil {
		return nil, fmt.Errorf("unable to unmarshal trusted proxies; don't forget the scheme: %v", err)
	}
	return proxies, nil
}

func NewAuthorizationMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {

		if viper.GetBool("domain.authentication.skipAuthentication") {
			next(ctx)
			return
		}

		err := huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		if err != nil {
			slog.Error("Failed to write unauthorized error", slog.String("error", err.Error()))
		}
		return
		//
		//if ctx.Operation().Security == nil {
		//	next(ctx)
		//	return
		//}
		//
		//bearer := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		//if len(bearer) == 0 {
		//	huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		//	return
		//}
		//
		//oidcClientId := viper.GetString("authentication.oidc.clientId")
		//oidcAuthority := viper.GetString("authentication.oidc.issuer")
		//runMode := viper.GetString("app.env")
		//
		////_, err := validateToken(context.Background(), bearer, oidcClientId, oidcAuthority, runMode)
		////if err != nil {
		////	huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
		////	return
		////}
		//
		//next(ctx)
	}
}

// github.com/coreos/go-oidc/v3/oidc
//func validateToken(ctx context.Context, bearer, oidcClientId, oidcAuthority, runMode string) (*oidc.Provider, error) {
//	provider, err := oidc.NewProvider(ctx, oidcAuthority)
//	if err != nil {
//		return nil, fmt.Errorf("can't create new provider -> %s", err)
//	}
//
//	insecureSkipSignatureCheck := runMode == "DEV"
//	var verifier = provider.Verifier(&oidc.Config{ClientID: oidcClientId, InsecureSkipSignatureCheck: insecureSkipSignatureCheck})
//
//	_, err = verifier.Verify(ctx, bearer)
//	if err != nil {
//		return nil, fmt.Errorf("can't verify bearer -> %s", err)
//	}
//
//	return provider, nil
//}
