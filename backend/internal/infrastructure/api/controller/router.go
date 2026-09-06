package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Router(svc *domain.Service) (*gin.Engine, error) {
	if config.String("app.environment") == "production" || config.String("app.environment") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	hostWithScheme := fmt.Sprintf("%s://%s", config.String("server.scheme"), config.String("server.host"))
	host := config.String("server.host")
	if config.Bool("domain.openapi.usePort") {
		host = fmt.Sprintf("%s:%s", host, config.String("server.port"))
		hostWithScheme = fmt.Sprintf("%s:%s", hostWithScheme, config.String("server.port"))
	}
	zap.L().Debug(fmt.Sprintf("Host: %s", hostWithScheme))

	humaConfig := huma.DefaultConfig(config.String("app.name"), config.String("app.version"))
	humaConfig.Info = &huma.Info{
		Title:       config.String("app.name"),
		Description: config.String("app.description"),
		License:     nil,
		Version:     config.String("app.version"),
	}
	humaConfig.Servers = []*huma.Server{
		{URL: hostWithScheme},
		{Description: fmt.Sprintf("This is the default server of %s", config.String("app.name"))},
	}
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Obtain a token via POST /v1/auth/login (LOCAL) or POST /v1/auth/oidc/callback (OIDC).",
		},
	}

	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	router.Use(cors.New(corsConfig))
	api := humagin.New(router, humaConfig)
	api.UseMiddleware(NewAuthenticationMiddleware(api))

	router.GET("/health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})
	router.GET("/health/readiness", func(c *gin.Context) {
		// You can add your readiness checks here (e.g., database connection)
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	trustedProxies := config.Strings("server.trustedProxies")
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		return nil, err
	}

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/swagger")
	})

	// --- Auth -----------------------------------------------------------
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		OperationID: "post-login",
		Summary:     "Login",
		Description: "Log in with a local username and password, returning an access/refresh token pair.",
		Path:        "/v1/auth/login",
		Tags:        []string{"Auth"},
	}, Login(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		OperationID: "post-refresh",
		Summary:     "Refresh token",
		Description: "Exchange a refresh token for a new access/refresh token pair. The refresh token used is invalidated (single-use rotation).",
		Path:        "/v1/auth/refresh",
		Tags:        []string{"Auth"},
	}, Refresh(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		OperationID: "post-logout",
		Summary:     "Logout",
		Description: "Invalidate a refresh token. The current access token simply expires on its own shortly after.",
		Path:        "/v1/auth/logout",
		Tags:        []string{"Auth"},
	}, Logout(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-oidc-login",
		Summary:     "Start OIDC login",
		Description: "Returns the authorization URL (PKCE-protected) the frontend should redirect the browser to, and the state to send back to the callback.",
		Path:        "/v1/auth/oidc/login",
		Tags:        []string{"Auth"},
	}, OidcLogin(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		OperationID: "post-oidc-callback",
		Summary:     "Complete OIDC login",
		Description: "Exchanges an authorization code for tokens. Behaves as login-or-auto-provision when called anonymously, or as link-to-my-account when called with a valid Bearer token.",
		Path:        "/v1/auth/oidc/callback",
		Tags:        []string{"Auth"},
	}, OidcCallback(svc))

	// --- Users (admin only, except self-registration, own profile, and own password) ---
	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		OperationID:   "post-create-user",
		Summary:       "Create user",
		Description:   "Register a new local user. When called with a valid admin Bearer token, the role field may also be set - ignored otherwise, so self-registration always creates a 'user' role account.",
		Path:          "/v1/users",
		Tags:          []string{"User"},
		DefaultStatus: http.StatusCreated,
	}, CreateUser(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-current-user",
		Summary:     "Get current user",
		Description: "Get the authenticated caller's own profile.",
		Path:        "/v1/users/me",
		Tags:        []string{"User"},
		Security:    bearerSecurity(),
	}, GetCurrentUser(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "list-users",
		Summary:     "List users",
		Description: "List all users.",
		Path:        "/v1/users",
		Tags:        []string{"User"},
		Security:    bearerSecurity(),
		Metadata:    requireAdmin(),
	}, ListUsers(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-user-by-id",
		Summary:     "Get user by ID",
		Description: "Get a user by ID.",
		Path:        "/v1/users/{userId}",
		Tags:        []string{"User"},
		Security:    bearerSecurity(),
		Metadata:    requireAdmin(),
	}, GetUserById(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-user",
		Summary:     "Update user",
		Description: "Update your own profile, or (as an admin) anyone's. Only an admin caller may change the role field. Password updates are not handled here.",
		Path:        "/v1/users/{userId}",
		Tags:        []string{"User"},
		Security:    bearerSecurity(),
	}, UpdateUser(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPatch,
		OperationID: "patch-user-password",
		Summary:     "Patch user password",
		Description: "Patch your own password. Only the account owner may do this - not even an admin can patch another user's password through this endpoint.",
		Path:        "/v1/users/{userId}/password",
		Tags:        []string{"User"},
		Security:    bearerSecurity(),
	}, PatchUserPassword(svc))
	huma.Register(api, huma.Operation{
		Method:        http.MethodDelete,
		OperationID:   "delete-user",
		Summary:       "Delete user",
		Description:   "Delete a user by ID.",
		Path:          "/v1/users/{userId}",
		Tags:          []string{"User"},
		DefaultStatus: http.StatusNoContent,
		Security:      bearerSecurity(),
		Metadata:      requireAdmin(),
	}, DeleteUser(svc))

	// --- Shelves (owner or admin) -----------------------------------------
	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		OperationID:   "post-create-shelf",
		Summary:       "Create shelf",
		Description:   "Create a new shelf, owned by the authenticated caller.",
		Path:          "/v1/shelves",
		Tags:          []string{"Shelf"},
		DefaultStatus: http.StatusCreated,
		Security:      bearerSecurity(),
	}, CreateShelf(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "list-shelves",
		Summary:     "List shelves",
		Description: "List shelves - the caller's own, or every shelf for an admin.",
		Path:        "/v1/shelves",
		Tags:        []string{"Shelf"},
		Security:    bearerSecurity(),
	}, ListShelf(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-public-shelf-by-path",
		Summary:     "Get public shelf by path",
		Description: "Get the public-safe view of a shelf by its path. Used to render a shelf's public link page and requires no authentication.",
		Path:        "/v1/shelves/by-path/{path}",
		Tags:        []string{"Shelf"},
	}, GetPublicShelfByPath(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-shelf-by-id",
		Summary:     "Get shelf by ID",
		Description: "Get a shelf by ID. Only its owner or an admin may access it.",
		Path:        "/v1/shelves/{shelfId}",
		Tags:        []string{"Shelf"},
		Security:    bearerSecurity(),
	}, GetShelfById(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-shelf",
		Summary:     "Update shelf",
		Description: "Update an existing shelf. Only its owner or an admin may update it.",
		Path:        "/v1/shelves/{shelfId}",
		Tags:        []string{"Shelf"},
		Security:    bearerSecurity(),
	}, UpdateShelf(svc))
	huma.Register(api, huma.Operation{
		Method:        http.MethodDelete,
		OperationID:   "delete-shelf",
		Summary:       "Delete shelf",
		Description:   "Delete a shelf by ID. Only its owner or an admin may delete it.",
		Path:          "/v1/shelves/{shelfId}",
		Tags:          []string{"Shelf"},
		DefaultStatus: http.StatusNoContent,
		Security:      bearerSecurity(),
	}, DeleteShelf(svc))

	// --- Sections (owner-of-shelf or admin to write; public to read) -----
	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		OperationID:   "post-create-section",
		Summary:       "Create section",
		Description:   "Create a new section on a shelf owned by the caller.",
		Path:          "/v1/sections",
		Tags:          []string{"Section"},
		DefaultStatus: http.StatusCreated,
		Security:      bearerSecurity(),
	}, CreateSection(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-sections",
		Summary:     "Get sections by shelf ID",
		Description: "Get sections by shelf ID. Used to render a shelf's public link page and requires no authentication.",
		Path:        "/v1/sections",
		Tags:        []string{"Section"},
	}, GetSections(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-section",
		Summary:     "Update section",
		Description: "Update an existing section. Only the owner of its shelf or an admin may update it.",
		Path:        "/v1/sections/{sectionId}",
		Tags:        []string{"Section"},
		Security:    bearerSecurity(),
	}, UpdateSection(svc))
	huma.Register(api, huma.Operation{
		Method:        http.MethodDelete,
		OperationID:   "delete-section",
		Summary:       "Delete section",
		Description:   "Delete a section by ID. Only the owner of its shelf or an admin may delete it.",
		Path:          "/v1/sections/{sectionId}",
		Tags:          []string{"Section"},
		DefaultStatus: http.StatusNoContent,
		Security:      bearerSecurity(),
	}, DeleteSection(svc))

	// --- Links (owner-of-shelf or admin to write; public to read) --------
	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		OperationID:   "post-create-link",
		Summary:       "Create link",
		Description:   "Create a new link in a section owned by the caller.",
		Path:          "/v1/links",
		Tags:          []string{"Link"},
		DefaultStatus: http.StatusCreated,
		Security:      bearerSecurity(),
	}, CreateLink(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-links",
		Summary:     "Get links by shelf ID",
		Description: "Get links by shelf ID. Used to render a shelf's public link page and requires no authentication.",
		Path:        "/v1/links",
		Tags:        []string{"Link"},
	}, GetLinks(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-link",
		Summary:     "Update link",
		Description: "Update an existing link. Only the owner of its shelf or an admin may update it.",
		Path:        "/v1/links/{linkId}",
		Tags:        []string{"Link"},
		Security:    bearerSecurity(),
	}, UpdateLink(svc))
	huma.Register(api, huma.Operation{
		Method:        http.MethodDelete,
		OperationID:   "delete-link",
		Summary:       "Delete link",
		Description:   "Delete a link by ID. Only the owner of its shelf or an admin may delete it.",
		Path:          "/v1/links/{linkId}",
		Tags:          []string{"Link"},
		DefaultStatus: http.StatusNoContent,
		Security:      bearerSecurity(),
	}, DeleteLink(svc))

	// --- Settings (read is public - it renders the public site shell; write is admin only) ---
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-page-settings",
		Summary:     "Get page settings",
		Description: "Get page settings by language code. Used to render the public site shell (title, contact info, legal pages, ...) and requires no authentication.",
		Path:        "/v1/settings",
		Tags:        []string{"Setting"},
	}, GetPageSettings(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-setting",
		Summary:     "Update setting",
		Description: "Update page settings.",
		Path:        "/v1/settings",
		Tags:        []string{"Setting"},
		Security:    bearerSecurity(),
		Metadata:    requireAdmin(),
	}, UpdateSetting(svc))

	// --- Statistics (any authenticated user, scoped to their own data) ---
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-statistic",
		Summary:     "Get statistic",
		Description: "Get usage statistics for the current user.",
		Path:        "/v1/statistics",
		Tags:        []string{"Statistic"},
		Security:    bearerSecurity(),
	}, GetStatistic(svc))

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
