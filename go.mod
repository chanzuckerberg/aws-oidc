module github.com/chanzuckerberg/aws-oidc

go 1.23.0

toolchain go1.24.1

require (
	github.com/AlecAivazis/survey/v2 v2.3.7
	github.com/aws/aws-sdk-go v1.55.7
	github.com/blang/semver v3.5.1+incompatible
	github.com/chanzuckerberg/go-misc/oidc_cli v0.0.0-20241218181938-e245ce8d3ba5
	github.com/chanzuckerberg/go-misc/sets v0.0.0-20250423213358-0dc5b698c7d3
	github.com/chanzuckerberg/go-misc/ver v0.0.0-20250423213358-0dc5b698c7d3
	github.com/coreos/go-oidc v2.3.0+incompatible
	github.com/gorilla/handlers v1.5.2
	github.com/honeycombio/beeline-go v1.19.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/okta/okta-sdk-golang/v2 v2.20.0
	github.com/peterhellberg/link v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.9.1
	github.com/stretchr/testify v1.10.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/chanzuckerberg/go-misc/osutil v0.0.0-20250423213358-0dc5b698c7d3 // indirect
	github.com/chanzuckerberg/go-misc/pidlock v0.0.0-20250423213358-0dc5b698c7d3 // indirect
	github.com/creack/pty v1.1.24 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/facebookgo/limitgroup v0.0.0-20150612190941-6abd8d71ec01 // indirect
	github.com/facebookgo/muster v0.0.0-20150708232844-fd3d7953fd52 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-jose/go-jose/v3 v3.0.4 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/honeycombio/libhoney-go v1.25.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/nightlyone/lockfile v1.0.0 // indirect
	github.com/patrickmn/go-cache v0.0.0-20180815053127-5633e0862627 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pquerna/cachecontrol v0.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/zalando/go-keyring v0.2.6 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/oauth2 v0.29.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/term v0.31.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/grpc v1.72.0 // indirect
	gopkg.in/alexcesaro/statsd.v2 v2.0.0 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// breaking change for mac keychains
exclude github.com/zalando/go-keyring v0.2.0

exclude github.com/zalando/go-keyring v0.2.1

replace github.com/zalando/go-keyring => github.com/zalando/go-keyring v0.1.1
