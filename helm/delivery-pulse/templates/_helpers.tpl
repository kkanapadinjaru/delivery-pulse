{{/*
Expand the name of the chart.
*/}}
{{- define "delivery-pulse.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "delivery-pulse.fullname" -}}
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
Common labels
*/}}
{{- define "delivery-pulse.labels" -}}
helm.sh/chart: {{ include "delivery-pulse.name" . }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: {{ include "delivery-pulse.name" . }}
{{- end }}

{{/*
Server labels
*/}}
{{- define "delivery-pulse.server.labels" -}}
{{ include "delivery-pulse.labels" . }}
app.kubernetes.io/name: {{ include "delivery-pulse.name" . }}-server
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: server
{{- end }}

{{/*
Server selector labels
*/}}
{{- define "delivery-pulse.server.selectorLabels" -}}
app.kubernetes.io/name: {{ include "delivery-pulse.name" . }}-server
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
UI labels
*/}}
{{- define "delivery-pulse.ui.labels" -}}
{{ include "delivery-pulse.labels" . }}
app.kubernetes.io/name: {{ include "delivery-pulse.name" . }}-ui
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: ui
{{- end }}

{{/*
UI selector labels
*/}}
{{- define "delivery-pulse.ui.selectorLabels" -}}
app.kubernetes.io/name: {{ include "delivery-pulse.name" . }}-ui
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Image name helper — defaults tag to Chart.AppVersion if not specified
*/}}
{{- define "delivery-pulse.image" -}}
{{- $tag := .tag | default .appVersion -}}
{{- if .registry }}
{{- printf "%s/%s:%s" .registry .repository $tag }}
{{- else }}
{{- printf "%s:%s" .repository $tag }}
{{- end }}
{{- end }}
