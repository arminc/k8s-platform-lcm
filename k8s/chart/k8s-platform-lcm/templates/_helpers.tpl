{{/*
Expand the name of the chart.
*/}}
{{- define "k8s-platform-lcm.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "k8s-platform-lcm.fullname" -}}
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
{{- define "k8s-platform-lcm.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "k8s-platform-lcm.labels" -}}
helm.sh/chart: {{ include "k8s-platform-lcm.chart" . }}
{{ include "k8s-platform-lcm.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "k8s-platform-lcm.selectorLabels" -}}
app.kubernetes.io/name: {{ include "k8s-platform-lcm.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "k8s-platform-lcm.serviceAccountName" -}}
{{- default (include "k8s-platform-lcm.fullname" .) .Values.serviceAccount.name }}
{{- end }}

{{/*
Resources for ClusterRole
*/}}
{{- define "k8s-platform-lcm.clusterRole.resources" -}}
{{- if and (.Values.imageScan.enabled) (.Values.helmScan.enabled) }}
resources: ["pods", "namespaces", "secrets"]
{{- end }}
{{- if and (.Values.imageScan.enabled) (not .Values.helmScan.enabled) }}
resources: ["pods", "namespaces"]
{{- end }}
{{- if and (not .Values.imageScan.enabled) (.Values.helmScan.enabled) }}
resources: ["secrets"]
{{- end }}
{{- if and (not .Values.imageScan.enabled) (not .Values.helmScan.enabled) }}
resources: [""]
{{- end }}
{{- end }}