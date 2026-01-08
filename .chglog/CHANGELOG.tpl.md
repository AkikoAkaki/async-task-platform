# {{ .Info.Title }}

{{ range .Versions }}
{{- $isUnreleased := eq .Tag.Name "Unreleased" -}}
{{- $hasPrevious := ne .Tag.Previous nil -}}
## {{ if $isUnreleased }}[Unreleased]({{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...HEAD){{ else if $hasPrevious }}[{{ .Tag.Name }}]({{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}){{ else }}{{ .Tag.Name }}{{ end }}{{ if not $isUnreleased }} - {{ datetime "2006-01-02" .Tag.Date }}{{ end }}

{{ if .CommitGroups -}}
{{ range .CommitGroups -}}
### {{ .Title }}

{{ range .Commits -}}
- {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}

{{ end -}}
{{ end -}}

{{- if .RevertCommits -}}
### Reverted

{{ range .RevertCommits -}}
- {{ .Revert.Header }}
{{ end }}

{{ end -}}

{{- if .NoteGroups -}}
### Breaking Changes

{{ range .NoteGroups -}}
{{ range .Notes -}}
- {{ .Body }}
{{ end }}
{{ end -}}

{{ end -}}

{{ if and (not .CommitGroups) (not .NoteGroups) (not .RevertCommits) }}
_No notable changes recorded._
{{ end }}

{{ end -}}
