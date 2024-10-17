module github.com/zenmodel/zenmodel/example

go 1.19

require (
	github.com/sashabaranov/go-openai v1.29.2
	github.com/zenmodel/zenmodel v0.2.0
	github.com/zenmodel/zenmodel/community v0.1.0
)

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/invopop/jsonschema v0.12.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-sqlite3 v1.14.24 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/zenmodel/zenmodel => ./..
	github.com/zenmodel/zenmodel/community => ./../community
)
