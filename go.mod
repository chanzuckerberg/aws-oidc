module github.com/chanzuckerberg/aws-oidc

go 1.24.0

toolchain go1.24.1

require (
	github.com/AlecAivazis/survey/v2 v2.3.7
	github.com/aws/aws-sdk-go v1.55.8
	github.com/blang/semver v3.5.1+incompatible
	github.com/chanzuckerberg/go-misc/oidc/v5 v5.0.0-20260226001108-30a84f160a6d
	github.com/chanzuckerberg/go-misc/ver v0.0.0-20251117200159-d2a50dbfd31c
	github.com/coreos/go-oidc v2.4.0+incompatible
	github.com/gorilla/handlers v1.5.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/okta/okta-sdk-golang/v2 v2.20.0
	github.com/peterhellberg/link v1.2.0
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
	gopkg.in/ini.v1 v1.67.0
)

require (
	al.essio.dev/pkg/shellescape v1.6.0 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/chanzuckerberg/go-misc/osutil v0.0.0-20251205003006-0acabbc1617e // indirect
	github.com/chanzuckerberg/go-misc/pidlock v0.0.0-20251117200159-d2a50dbfd31c // indirect
	github.com/coreos/go-oidc/v3 v3.17.0 // indirect
	github.com/creack/pty v1.1.24 // indirect
	github.com/danieljoos/wincred v1.2.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dustinkirkland/golang-petname v0.0.0-20260215035315-f0c533e9ce9b // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-jose/go-jose/v3 v3.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/godbus/dbus/v5 v5.2.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/nightlyone/lockfile v1.0.0 // indirect
	github.com/patrickmn/go-cache v0.0.0-20180815053127-5633e0862627 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pquerna/cachecontrol v0.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/zalando/go-keyring v0.2.6 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/oauth2 v0.33.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.37.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// breaking change for mac keychains
exclude github.com/zalando/go-keyring v0.2.0

exclude github.com/zalando/go-keyring v0.2.1
