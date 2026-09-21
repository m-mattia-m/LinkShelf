{{- define "linkshelf.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "linkshelf.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "linkshelf.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{ include "linkshelf.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "linkshelf.selectorLabels" -}}
app.kubernetes.io/name: {{ include "linkshelf.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "linkshelf.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "linkshelf.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/* Name of the Service the bundled postgres subchart creates, mirrors its own fullname logic, where the alias is the chart name. */}}
{{- define "linkshelf.postgresql.fullname" -}}
{{- if .Values.postgresql.fullnameOverride }}
{{- .Values.postgresql.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default "postgresql" .Values.postgresql.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/* http or https, depending on whether the host is listed in ingress.tls. Args: root, host. */}}
{{- define "linkshelf.scheme" -}}
{{- $tls := false }}
{{- range .root.Values.ingress.tls }}
{{- if has $.host (default (list) .hosts) }}{{ $tls = true }}{{ end }}
{{- end }}
{{- ternary "https" "http" $tls }}
{{- end }}

{{- define "linkshelf.frontendUrl" -}}
{{- if .Values.frontendUrl }}
{{- .Values.frontendUrl }}
{{- else if and .Values.ingress.enabled .Values.ingress.frontend.host }}
{{- printf "%s://%s" (include "linkshelf.scheme" (dict "root" . "host" .Values.ingress.frontend.host)) .Values.ingress.frontend.host }}
{{- end }}
{{- end }}

{{- define "linkshelf.apiUrl" -}}
{{- if .Values.apiUrl }}
{{- .Values.apiUrl }}
{{- else if and .Values.ingress.enabled .Values.ingress.backend.host }}
{{- printf "%s://%s" (include "linkshelf.scheme" (dict "root" . "host" .Values.ingress.backend.host)) .Values.ingress.backend.host }}
{{- end }}
{{- end }}

{{- define "linkshelf.oidcRedirectUrl" -}}
{{- if .Values.oidcRedirectUrl }}
{{- .Values.oidcRedirectUrl }}
{{- else if include "linkshelf.frontendUrl" . }}
{{- printf "%s/auth/callback" (include "linkshelf.frontendUrl" .) }}
{{- end }}
{{- end }}

{{/* Key inside secrets.existingSecret for one app secret, empty if it is not wired. Args: root, name. */}}
{{- define "linkshelf.secretKey" -}}
{{- if .root.Values.secrets.existingSecret.name }}
{{- get .root.Values.secrets.existingSecret.keys .name }}
{{- end }}
{{- end }}

{{- define "linkshelf.secretEnvNames" -}}
jwtSecret: APP_AUTHENTICATION_JWTSECRET
bootstrapAdminPassword: APP_AUTHENTICATION_BOOTSTRAPADMIN_PASSWORD
smtpPassword: APP_SMTP_PASSWORD
oidcClientSecret: APP_AUTHENTICATION_OIDC_CLIENTSECRET
{{- end }}

{{/*
Non-secret environment: values derived from the chart first, then .Values.env on top,
so a user can always override a derived value. Rendered as YAML.
*/}}
{{- define "linkshelf.env" -}}
{{- $env := dict }}
{{- $engine := ternary "POSTGRES" .Values.database.engine .Values.postgresql.enabled }}
{{- $env = set $env "APP_DATABASE_ENGINE" $engine }}
{{- if and .Values.database.params (not .Values.postgresql.enabled) }}
{{- $env = set $env "APP_DATABASE_PARAMS" .Values.database.params }}
{{- end }}
{{- with (include "linkshelf.frontendUrl" .) }}{{ $env = set $env "APP_APP_FRONTENDURL" . }}{{ end }}
{{- with (include "linkshelf.apiUrl" .) }}{{ $env = set $env "NUXT_PUBLIC_API_BASE" . }}{{ end }}
{{- with (include "linkshelf.oidcRedirectUrl" .) }}{{ $env = set $env "APP_AUTHENTICATION_OIDC_REDIRECTURL" . }}{{ end }}
{{- /* Strict origins need the real public addresses, which only an ingress guarantees. The backend answers on the host of apiUrl. */}}
{{- if and .Values.strictOrigins .Values.ingress.enabled (include "linkshelf.frontendUrl" .) (include "linkshelf.apiUrl" .) }}
{{- $api := urlParse (include "linkshelf.apiUrl" .) }}
{{- $env = set $env "APP_APP_STRICTORIGINS" "true" }}
{{- $env = set $env "APP_SERVER_SCHEME" $api.scheme }}
{{- $env = set $env "APP_SERVER_HOST" $api.hostname }}
{{- end }}
{{- /* The image ships a default admin account. Keep it disabled unless the operator sets one. */}}
{{- $env = set $env "APP_AUTHENTICATION_BOOTSTRAPADMIN_EMAIL" "" }}
{{- if not (include "linkshelf.secretKey" (dict "root" . "name" "bootstrapAdminPassword")) }}
{{- $env = set $env "APP_AUTHENTICATION_BOOTSTRAPADMIN_PASSWORD" "" }}
{{- end }}
{{- $env = mergeOverwrite $env (deepCopy .Values.env) }}
{{- toYaml $env }}
{{- end }}

{{- define "linkshelf.validate" -}}
{{- if not .Values.secrets.existingSecret.name }}
{{- fail "set secrets.existingSecret.name, this chart does not create Secrets" }}
{{- end }}
{{- if not (include "linkshelf.secretKey" (dict "root" . "name" "jwtSecret")) }}
{{- fail "set secrets.existingSecret.keys.jwtSecret to the key that holds the JWT secret" }}
{{- end }}
{{- $env := include "linkshelf.env" . | fromYaml }}
{{- if and (get $env "APP_AUTHENTICATION_BOOTSTRAPADMIN_EMAIL") (not (include "linkshelf.secretKey" (dict "root" . "name" "bootstrapAdminPassword"))) }}
{{- fail "APP_AUTHENTICATION_BOOTSTRAPADMIN_EMAIL is set, so also set secrets.existingSecret.keys.bootstrapAdminPassword" }}
{{- end }}
{{- if .Values.postgresql.enabled }}
{{- if .Values.database.existingSecret.name }}
{{- fail "postgresql.enabled cannot be combined with database.existingSecret" }}
{{- end }}
{{- if not (and .Values.postgresql.settings.existingSecret .Values.postgresql.userDatabase.existingSecret) }}
{{- fail "postgresql.enabled needs postgresql.settings.existingSecret and postgresql.userDatabase.existingSecret" }}
{{- end }}
{{- else if not .Values.database.existingSecret.name }}
{{- fail "configure a database: database.existingSecret.name or postgresql.enabled=true" }}
{{- end }}
{{- if .Values.ingress.enabled }}
{{- if not (and .Values.ingress.frontend.host .Values.ingress.backend.host) }}
{{- fail "ingress.enabled needs ingress.frontend.host and ingress.backend.host" }}
{{- end }}
{{- end }}
{{- end }}
