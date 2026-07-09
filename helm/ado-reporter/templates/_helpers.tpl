{{/*
Expand the name of the chart.
*/}}
{{- define "ado-reporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "ado-reporter.fullname" -}}
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
{{- define "ado-reporter.labels" -}}
helm.sh/chart: {{ include "ado-reporter.name" . }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: {{ include "ado-reporter.name" . }}
{{- end }}

{{/*
Backend labels
*/}}
{{- define "ado-reporter.backend.labels" -}}
{{ include "ado-reporter.labels" . }}
app.kubernetes.io/name: {{ include "ado-reporter.name" . }}-backend
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Backend selector labels
*/}}
{{- define "ado-reporter.backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ado-reporter.name" . }}-backend
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Frontend labels
*/}}
{{- define "ado-reporter.frontend.labels" -}}
{{ include "ado-reporter.labels" . }}
app.kubernetes.io/name: {{ include "ado-reporter.name" . }}-frontend
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "ado-reporter.frontend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ado-reporter.name" . }}-frontend
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Image name helper
*/}}
{{- define "ado-reporter.image" -}}
{{- if .registry }}
{{- printf "%s/%s:%s" .registry .repository .tag }}
{{- else }}
{{- printf "%s:%s" .repository .tag }}
{{- end }}
{{- end }}
