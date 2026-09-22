package controller

import (
	"backend/internal/config"
	"backend/internal/domain"
	"fmt"
	"net/http"
	"strings"

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
	if config.Bool("app.strictOrigins") {
		router.Use(hostGuard())
		corsConfig.AllowAllOrigins = false
		corsConfig.AllowOriginFunc = newOriginPolicy(svc).allowed
		zap.L().Info("strict origins are enabled",
			zap.String("frontendUrl", config.String("app.frontendUrl")),
			zap.String("serverHost", config.String("server.host")),
			zap.Bool("serverPortIsChecked", config.Bool("domain.openapi.usePort")))
	}
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

	// Serves an instance admin's mounted assets directory (e.g. theme
	// background images) at assets.basePath, e.g. a file "my-dog.webp"
	// becomes "{basePath}/my-dog.webp". A no-op if assets.directory isn't set.
	if assetsDir := strings.TrimSpace(config.String("assets.directory")); assetsDir != "" {
		router.Static(config.String("assets.basePath"), assetsDir)
	}

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
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		OperationID: "post-resend-verification",
		Summary:     "Resend verification email",
		Description: "Re-sends whatever verification/invite link is still pending for the given email, rate-limited. Always responds the same way regardless of whether the address exists, is already verified, or was rate-limited.",
		Path:        "/v1/auth/resend-verification",
		Tags:        []string{"Auth"},
	}, ResendVerification(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		OperationID: "post-verify-email",
		Summary:     "Verify email",
		Description: "Completes email verification for an account that already has a password, using the token from the emailed link.",
		Path:        "/v1/auth/verify-email",
		Tags:        []string{"Auth"},
	}, VerifyEmail(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		OperationID: "post-set-password",
		Summary:     "Set password",
		Description: "Completes the admin-invite flow: sets an account's first password and marks it verified, using the token from the emailed link.",
		Path:        "/v1/auth/set-password",
		Tags:        []string{"Auth"},
	}, SetPassword(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		OperationID: "post-forgot-password",
		Summary:     "Request a password reset",
		Description: "Emails a password reset link to the given address. Always responds the same way regardless of whether the address belongs to an account or the request was rate-limited. Responds 403 while password reset is disabled.",
		Path:        "/v1/auth/forgot-password",
		Tags:        []string{"Auth"},
	}, ForgotPassword(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		OperationID: "post-reset-password",
		Summary:     "Reset password",
		Description: "Sets a new password using the token from the emailed reset link, marks the address verified and signs the account out everywhere.",
		Path:        "/v1/auth/reset-password",
		Tags:        []string{"Auth"},
	}, ResetPassword(svc))

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
		Method:      http.MethodPatch,
		OperationID: "patch-user-verify",
		Summary:     "Mark user verified",
		Description: "Admin override: forces a user's email to verified without requiring the emailed link - a safety valve for when SMTP delivery is broken.",
		Path:        "/v1/users/{userId}/verify",
		Tags:        []string{"User"},
		Security:    bearerSecurity(),
		Metadata:    requireAdmin(),
	}, MarkUserVerified(svc))
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
		OperationID: "get-public-shelf-by-domain",
		Summary:     "Get public shelf by domain",
		Description: "Get the public-safe view of the shelf that is served on a domain of its own, for example profile.example.com or profile.example.com:9443. The domain is normalized first (lowercased, without a trailing dot or slash, without :80 or :443). Used to render that shelf when the frontend is reached on the domain, and requires no authentication. Unlike the lookups by path it doesn't depend on app.userBasedPaths.",
		Path:        "/v1/shelves/by-domain/{domain}",
		Tags:        []string{"Shelf"},
	}, GetPublicShelfByDomain(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-public-shelf-by-username-and-path",
		Summary:     "Get public shelf by username and path",
		Description: "Get the public-safe view of a shelf by its owner's username and its path, used while app.userBasedPaths is enabled. Requires no authentication. While the setting is disabled, use the lookup by path alone instead - the two never answer at the same time.",
		Path:        "/v1/shelves/by-user/{username}/{path}",
		Tags:        []string{"Shelf"},
	}, GetPublicShelfByUsernameAndPath(svc))
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
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-sections-order",
		Summary:     "Update sections order in batch",
		Description: "Update the order of many sections in a single request. Invalid items are rejected individually (reported in the response's failures list) without aborting the rest of the batch.",
		Path:        "/v1/sections/reorder",
		Tags:        []string{"Section"},
		Security:    bearerSecurity(),
	}, UpdateSectionsOrder(svc))

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
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-links-order",
		Summary:     "Update links order in batch",
		Description: "Update the order of many links in a single request. A link can only be reordered within the section it already belongs to. Invalid items are rejected individually (reported in the response's failures list) without aborting the rest of the batch.",
		Path:        "/v1/links/reorder",
		Tags:        []string{"Link"},
		Security:    bearerSecurity(),
	}, UpdateLinksOrder(svc))

	// --- Themes (instance themes are read-only, seeded from a mounted directory; every user may manage their own) ---
	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		OperationID:   "post-create-theme",
		Summary:       "Create theme",
		Description:   "Create a new theme, owned by the authenticated caller.",
		Path:          "/v1/themes",
		Tags:          []string{"Theme"},
		DefaultStatus: http.StatusCreated,
		Security:      bearerSecurity(),
	}, CreateTheme(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "list-themes",
		Summary:     "List themes",
		Description: "List themes available to the caller, grouped into instance-provided and their own.",
		Path:        "/v1/themes",
		Tags:        []string{"Theme"},
		Security:    bearerSecurity(),
	}, ListThemes(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "list-all-user-themes",
		Summary:     "List all user themes",
		Description: "List every user-created theme across the instance, for moderation. Admin only.",
		Path:        "/v1/themes/admin",
		Tags:        []string{"Theme"},
		Security:    bearerSecurity(),
		Metadata:    requireAdmin(),
	}, ListAllUserThemes(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-theme-by-id",
		Summary:     "Get theme by ID",
		Description: "Get a theme by ID. Instance themes are readable by any authenticated caller; user themes only by their owner or an admin.",
		Path:        "/v1/themes/{themeId}",
		Tags:        []string{"Theme"},
		Security:    bearerSecurity(),
	}, GetTheme(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-theme",
		Summary:     "Update theme",
		Description: "Update an existing theme. Only its owner may update it - not even an admin, and never an instance theme (managed via the mounted directory instead).",
		Path:        "/v1/themes/{themeId}",
		Tags:        []string{"Theme"},
		Security:    bearerSecurity(),
	}, UpdateTheme(svc))
	huma.Register(api, huma.Operation{
		Method:        http.MethodDelete,
		OperationID:   "delete-theme",
		Summary:       "Delete theme",
		Description:   "Delete a theme by ID. Its owner or an admin may delete a user theme; instance themes can never be deleted through the API.",
		Path:          "/v1/themes/{themeId}",
		Tags:          []string{"Theme"},
		DefaultStatus: http.StatusNoContent,
		Security:      bearerSecurity(),
	}, DeleteTheme(svc))

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
	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		OperationID: "put-update-settings-batch",
		Summary:     "Update settings in batch",
		Description: "Update many page settings in a single request. Invalid items are rejected individually (reported in the response's failures list) without aborting the rest of the batch.",
		Path:        "/v1/settings/batch",
		Tags:        []string{"Setting"},
		Security:    bearerSecurity(),
		Metadata:    requireAdmin(),
	}, UpdateSettingsBatch(svc))
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		OperationID: "get-email-delivery-info",
		Summary:     "Get email delivery info",
		Description: "Admin-only, read-only: the SMTP host and from-address currently configured. SMTP itself is configured exclusively via config, not through this API.",
		Path:        "/v1/settings/email-delivery",
		Tags:        []string{"Setting"},
		Security:    bearerSecurity(),
		Metadata:    requireAdmin(),
	}, GetEmailDeliveryInfo(svc))

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
