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

{{/* Mount path of the ConfigMap rendered from .Values.configFile, see configmap.yaml. */}}
{{- define "linkshelf.configFilePath" -}}
/etc/linkshelf/config.yaml
{{- end }}

{{/*
Merged environment as a dict of NAME -> {value: ...} or {valueFrom: {...}},
i.e. one native Kubernetes EnvVar minus its "name" (which is the dict key).
Chart-derived values are set first, then every entry of .Values.env (a plain
list of native EnvVar objects) is applied on top by name, so a user can
always override or add to what the chart derives - including secrets, via
valueFrom.secretKeyRef, with no separate "secrets" section to keep in sync.
Rendered as YAML.
*/}}
{{- define "linkshelf.env" -}}
{{- $env := dict }}
{{- $engine := ternary "POSTGRES" .Values.database.engine .Values.postgresql.enabled }}
{{- $env = set $env "DATABASE_ENGINE" (dict "value" $engine) }}
{{- if and .Values.database.params (not .Values.postgresql.enabled) }}
{{- $env = set $env "DATABASE_PARAMS" (dict "value" .Values.database.params) }}
{{- end }}
{{- with (include "linkshelf.frontendUrl" .) }}{{ $env = set $env "APP_FRONTENDURL" (dict "value" .) }}{{ end }}
{{- with (include "linkshelf.apiUrl" .) }}{{ $env = set $env "NUXT_PUBLIC_API_BASE" (dict "value" .) }}{{ end }}
{{- with (include "linkshelf.oidcRedirectUrl" .) }}{{ $env = set $env "AUTHENTICATION_OIDC_REDIRECTURL" (dict "value" .) }}{{ end }}
{{- /* Optional Plausible Analytics, disabled by default - an instance owner's choice. */}}
{{- if .Values.plausible.enabled }}
{{- $env = set $env "NUXT_PUBLIC_PLAUSIBLE_ENABLED" (dict "value" "true") }}
{{- end }}
{{- with .Values.plausible.domain }}{{ $env = set $env "NUXT_PUBLIC_PLAUSIBLE_DOMAIN" (dict "value" .) }}{{ end }}
{{- with .Values.plausible.apiHost }}{{ $env = set $env "NUXT_PUBLIC_PLAUSIBLE_API_HOST" (dict "value" .) }}{{ end }}
{{- if .Values.plausible.proxy }}
{{- $env = set $env "NUXT_PUBLIC_PLAUSIBLE_PROXY" (dict "value" "true") }}
{{- end }}
{{- /* Strict origins need the real public addresses, which only an ingress guarantees. The backend answers on the host of apiUrl. */}}
{{- if and .Values.strictOrigins .Values.ingress.enabled (include "linkshelf.frontendUrl" .) (include "linkshelf.apiUrl" .) }}
{{- $api := urlParse (include "linkshelf.apiUrl" .) }}
{{- $env = set $env "APP_STRICTORIGINS" (dict "value" "true") }}
{{- $env = set $env "SERVER_SCHEME" (dict "value" $api.scheme) }}
{{- $env = set $env "SERVER_HOST" (dict "value" $api.hostname) }}
{{- end }}
{{- /* The image ships a default admin account. Keep it disabled unless the operator sets one. */}}
{{- $env = set $env "AUTHENTICATION_BOOTSTRAPADMIN_EMAIL" (dict "value" "") }}
{{- if .Values.postgresql.enabled }}
{{- $user := .Values.postgresql.userDatabase }}
{{- $dbSecret := $user.existingSecret }}
{{- $env = set $env "DATABASE_HOST" (dict "value" (include "linkshelf.postgresql.fullname" .)) }}
{{- $env = set $env "DATABASE_PORT" (dict "value" (.Values.postgresql.service.port | default 5432 | toString)) }}
{{- $env = set $env "DATABASE_USERNAME" (dict "valueFrom" (dict "secretKeyRef" (dict "name" $dbSecret "key" (default "USERDB_USER" ($user.user).secretKey)))) }}
{{- $env = set $env "DATABASE_PASSWORD" (dict "valueFrom" (dict "secretKeyRef" (dict "name" $dbSecret "key" (default "USERDB_PASSWORD" ($user.password).secretKey)))) }}
{{- $env = set $env "DATABASE_NAME" (dict "valueFrom" (dict "secretKeyRef" (dict "name" $dbSecret "key" (default "POSTGRES_DB" ($user.name).secretKey)))) }}
{{- end }}
{{- if .Values.configFile }}
{{- $env = set $env "CONFIGURATION_FILE_PATH" (dict "value" (include "linkshelf.configFilePath" .)) }}
{{- end }}
{{- range $entry := .Values.env }}
{{- $env = set $env $entry.name (omit $entry "name") }}
{{- end }}
{{- /* Only default the bootstrap admin password to disabled if the user didn't wire one up above. */}}
{{- if not (hasKey $env "AUTHENTICATION_BOOTSTRAPADMIN_PASSWORD") }}
{{- $env = set $env "AUTHENTICATION_BOOTSTRAPADMIN_PASSWORD" (dict "value" "") }}
{{- end }}
{{- toYaml $env }}
{{- end }}

{{/* True if the named entry of a merged env dict (see linkshelf.env) exists and has a non-empty value or any valueFrom. Args: env, name. */}}
{{- define "linkshelf.envConfigured" -}}
{{- if hasKey .env .name }}
{{- $entry := get .env .name }}
{{- if or $entry.valueFrom $entry.value }}true{{ end }}
{{- end }}
{{- end }}

{{- define "linkshelf.validate" -}}
{{- $env := include "linkshelf.env" . | fromYaml }}
{{- if ne "true" (include "linkshelf.envConfigured" (dict "env" $env "name" "AUTHENTICATION_JWTSECRET")) }}
{{- fail "set AUTHENTICATION_JWTSECRET in env (a literal value, or valueFrom.secretKeyRef) - this chart does not create Secrets" }}
{{- end }}
{{- if and (eq "true" (include "linkshelf.envConfigured" (dict "env" $env "name" "AUTHENTICATION_BOOTSTRAPADMIN_EMAIL"))) (ne "true" (include "linkshelf.envConfigured" (dict "env" $env "name" "AUTHENTICATION_BOOTSTRAPADMIN_PASSWORD"))) }}
{{- fail "AUTHENTICATION_BOOTSTRAPADMIN_EMAIL is set in env, so also set AUTHENTICATION_BOOTSTRAPADMIN_PASSWORD" }}
{{- end }}
{{- if .Values.postgresql.enabled }}
{{- range $name := (list "DATABASE_HOST" "DATABASE_PORT" "DATABASE_USERNAME" "DATABASE_PASSWORD" "DATABASE_NAME") }}
{{- range $entry := $.Values.env }}
{{- if eq $entry.name $name }}
{{- fail (printf "postgresql.enabled cannot be combined with setting %s in env" $name) }}
{{- end }}
{{- end }}
{{- end }}
{{- if not (and .Values.postgresql.settings.existingSecret .Values.postgresql.userDatabase.existingSecret) }}
{{- fail "postgresql.enabled needs postgresql.settings.existingSecret and postgresql.userDatabase.existingSecret" }}
{{- end }}
{{- else }}
{{- range $name := (list "DATABASE_HOST" "DATABASE_PORT" "DATABASE_USERNAME" "DATABASE_PASSWORD" "DATABASE_NAME") }}
{{- if ne "true" (include "linkshelf.envConfigured" (dict "env" $env "name" $name)) }}
{{- fail (printf "configure a database: set %s in env, or postgresql.enabled=true" $name) }}
{{- end }}
{{- end }}
{{- end }}
{{- if .Values.ingress.enabled }}
{{- if not (and .Values.ingress.frontend.host .Values.ingress.backend.host) }}
{{- fail "ingress.enabled needs ingress.frontend.host and ingress.backend.host" }}
{{- end }}
{{- end }}
{{- end }}
