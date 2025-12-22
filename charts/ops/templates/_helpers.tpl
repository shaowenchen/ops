{{/*
Expand the name of the chart.
*/}}
{{- define "ops.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ops.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ops.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels - used for all resources
*/}}
{{- define "ops.labels" -}}
helm.sh/chart: {{ include "ops.chart" . }}
{{ include "ops.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: ops
{{- end }}

{{/*
Selector labels - core labels for resource selection
*/}}
{{- define "ops.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ops.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Controller selector labels - for controller resources
*/}}
{{- define "ops.controllerSelectorLabels" -}}
{{ include "ops.selectorLabels" . }}
app.kubernetes.io/component: controller
control-plane: controller-manager
{{- end }}

{{/*
Server selector labels - for server resources
*/}}
{{- define "ops.serverSelectorLabels" -}}
app.kubernetes.io/name: {{ include "ops.name" . }}-server
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: server
{{- end }}

{{/*
Controller labels - for controller resources metadata
*/}}
{{- define "ops.controllerLabels" -}}
{{ include "ops.labels" . }}
app.kubernetes.io/component: controller
control-plane: controller-manager
{{- end }}

{{/*
Server labels - for server resources metadata
*/}}
{{- define "ops.serverLabels" -}}
{{ include "ops.labels" . }}
app.kubernetes.io/component: server
{{- end }}

{{/*
Metrics labels - for metrics resources metadata
*/}}
{{- define "ops.metricsLabels" -}}
{{ include "ops.labels" . }}
app.kubernetes.io/component: metrics
{{- end }}

{{/*
Controller metrics labels - for controller metrics resources
*/}}
{{- define "ops.controllerMetricsLabels" -}}
{{ include "ops.labels" . }}
app.kubernetes.io/component: metrics
control-plane: controller-manager
{{- end }}

{{/*
Server metrics labels - for server metrics resources
*/}}
{{- define "ops.serverMetricsLabels" -}}
{{ include "ops.labels" . }}
app.kubernetes.io/component: metrics
{{- end }}

{{/*
Controller metrics selector labels - for ServiceMonitor to match controller metrics Service
*/}}
{{- define "ops.controllerMetricsSelectorLabels" -}}
{{ include "ops.selectorLabels" . }}
app.kubernetes.io/component: metrics
control-plane: controller-manager
{{- end }}

{{/*
Server metrics selector labels - for ServiceMonitor to match server metrics Service
*/}}
{{- define "ops.serverMetricsSelectorLabels" -}}
{{ include "ops.selectorLabels" . }}
app.kubernetes.io/component: server
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ops.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "ops.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
