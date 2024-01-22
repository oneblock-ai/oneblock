{{/*
Expand the name of the chart.
*/}}
{{- define "oneblock.name" -}}
{{- default .Chart.Name .Values.oneblock.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "oneblock.fullname" -}}
{{- if .Values.oneblock.fullnameOverride }}
{{- .Values.oneblock.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.oneblock.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "oneblock.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "oneblock.labels" -}}
helm.sh/chart: {{ include "oneblock.chart" . }}
{{ include "oneblock.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.oneblock.ai/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.oneblock.ai/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Webhook labels
*/}}
{{- define "oneblock.webhookLabels" -}}
{{ include "oneblock.labels" . }}
app.oneblock.ai/webhook: "true"
{{- end }}

{{/*
Selector labels
*/}}
{{- define "oneblock.selectorLabels" -}}
app.oneblock.ai/name: {{ include "oneblock.name" . }}
app.oneblock.ai/instance: {{ .Release.Name }}
{{- end }}

{{/*
Webhook selector labels
*/}}
{{- define "oneblock.webhookSelectorLabels" -}}
{{ include "oneblock.selectorLabels" . }}
app.oneblock.ai/webhook: "true"
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "oneblock.serviceAccountName" -}}
{{- if .Values.oneblock.serviceAccount.create }}
{{- default (include "oneblock.fullname" .) .Values.oneblock.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.oneblock.serviceAccount.name }}
{{- end }}
{{- end }}
