# Changelog

## [0.35.0](https://github.com/chanzuckerberg/aws-oidc/compare/v0.34.2...v0.35.0) (2026-02-27)


### Features

* **pixi:** simplify pixi recipe and release workflow ([#1170](https://github.com/chanzuckerberg/aws-oidc/issues/1170)) ([cd59c69](https://github.com/chanzuckerberg/aws-oidc/commit/cd59c6911867d0caf1154cfc02441631c4e01327))

## [0.34.2](https://github.com/chanzuckerberg/aws-oidc/compare/v0.34.1...v0.34.2) (2026-02-27)


### Bug Fixes

* wrong tag for install gh ([#1168](https://github.com/chanzuckerberg/aws-oidc/issues/1168)) ([538cd5d](https://github.com/chanzuckerberg/aws-oidc/commit/538cd5d9159f7645224c325a3821d8771f739d60))

## [0.34.1](https://github.com/chanzuckerberg/aws-oidc/compare/v0.34.0...v0.34.1) (2026-02-26)


### Bug Fixes

* install gh in runner ([#1166](https://github.com/chanzuckerberg/aws-oidc/issues/1166)) ([c6e77ee](https://github.com/chanzuckerberg/aws-oidc/commit/c6e77eeab41f8af1a8f8a52e36385926daa7a7ca))

## [0.34.0](https://github.com/chanzuckerberg/aws-oidc/compare/v0.33.0...v0.34.0) (2026-02-26)


### Features

* allow users to install with pixi ([#1164](https://github.com/chanzuckerberg/aws-oidc/issues/1164)) ([fefb6e6](https://github.com/chanzuckerberg/aws-oidc/commit/fefb6e6137e46a9672c5536b3ac4895352e4a6c6))

## [0.33.0](https://github.com/chanzuckerberg/aws-oidc/compare/v0.32.3...v0.33.0) (2026-02-26)


### Features

* add node-local cache, token inspection subcommands, and per-host logging ([#1162](https://github.com/chanzuckerberg/aws-oidc/issues/1162)) ([672f3c3](https://github.com/chanzuckerberg/aws-oidc/commit/672f3c33203716bb2399bc3b1c16d730428aa74c))
* sync root token if newer refresh token ([#1163](https://github.com/chanzuckerberg/aws-oidc/issues/1163)) ([ed27104](https://github.com/chanzuckerberg/aws-oidc/commit/ed27104974acf4e972ed698b1360e71ca44fd26e))


### Misc

* bump goreleaser/goreleaser-action from 6 to 7 ([#1160](https://github.com/chanzuckerberg/aws-oidc/issues/1160)) ([825ec2e](https://github.com/chanzuckerberg/aws-oidc/commit/825ec2e559ec18718759b3264005d482d0d4a4dd))

## [0.32.3](https://github.com/chanzuckerberg/aws-oidc/compare/v0.32.2...v0.32.3) (2026-02-19)


### Bug Fixes

* bump go-misc/oidc to v5.1.4 for atomic file writes ([#1158](https://github.com/chanzuckerberg/aws-oidc/issues/1158)) ([ac00b31](https://github.com/chanzuckerberg/aws-oidc/commit/ac00b31348e01c80ea9f4388175133377fa67016))

## [0.32.2](https://github.com/chanzuckerberg/aws-oidc/compare/v0.32.1...v0.32.2) (2026-02-10)


### Bug Fixes

* consolidate the token expiration checks ([#1156](https://github.com/chanzuckerberg/aws-oidc/issues/1156)) ([de6cb81](https://github.com/chanzuckerberg/aws-oidc/commit/de6cb81b6416887943922b277f67bfd4741ebfb6))

## [0.32.1](https://github.com/chanzuckerberg/aws-oidc/compare/v0.32.0...v0.32.1) (2026-02-05)


### Bug Fixes

* TOCOTU ([#1154](https://github.com/chanzuckerberg/aws-oidc/issues/1154)) ([2725df5](https://github.com/chanzuckerberg/aws-oidc/commit/2725df53ddb0a64b741c192dc5ccf6e2f79e16e0))

## [0.32.0](https://github.com/chanzuckerberg/aws-oidc/compare/v0.31.2...v0.32.0) (2026-02-04)


### Features

* better retry logic ([#1152](https://github.com/chanzuckerberg/aws-oidc/issues/1152)) ([919b5c6](https://github.com/chanzuckerberg/aws-oidc/commit/919b5c66a6eb217a20a39d6e059acb7d4c9e5c43))

## [0.31.2](https://github.com/chanzuckerberg/aws-oidc/compare/v0.31.1...v0.31.2) (2026-02-04)


### Bug Fixes

* add logging configuration and gracefully handle cases without an id_token  ([#1151](https://github.com/chanzuckerberg/aws-oidc/issues/1151)) ([e3ad7e0](https://github.com/chanzuckerberg/aws-oidc/commit/e3ad7e022b246210f8cf139968132b693c36eb8a))


### Misc

* configure helm dependency updater workflow ([#1149](https://github.com/chanzuckerberg/aws-oidc/issues/1149)) ([dc9d181](https://github.com/chanzuckerberg/aws-oidc/commit/dc9d181a89a817be123adb70fae65744106a0ee0))

## [0.31.1](https://github.com/chanzuckerberg/aws-oidc/compare/v0.31.0...v0.31.1) (2025-12-15)


### Bug Fixes

* bump go-misc oidc package which keeps the refresh token for file storage caches ([#1147](https://github.com/chanzuckerberg/aws-oidc/issues/1147)) ([f4b9dbe](https://github.com/chanzuckerberg/aws-oidc/commit/f4b9dbe47ca840083add64c08cf59fd0cd572d96))

## [0.31.0](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.11...v0.31.0) (2025-12-05)


### Features

* add device auth flow to aws-oidc ([#1143](https://github.com/chanzuckerberg/aws-oidc/issues/1143)) ([5084f7a](https://github.com/chanzuckerberg/aws-oidc/commit/5084f7ae25f674810f7281675681cbcdac2f1481))


### Misc

* bump golang.org/x/crypto from 0.37.0 to 0.45.0 ([#1140](https://github.com/chanzuckerberg/aws-oidc/issues/1140)) ([32e1fd2](https://github.com/chanzuckerberg/aws-oidc/commit/32e1fd28394f523c72efa924973e717282d0ed05))
* Remove imaging-dev from rolemap ([#1142](https://github.com/chanzuckerberg/aws-oidc/issues/1142)) ([22dc2d9](https://github.com/chanzuckerberg/aws-oidc/commit/22dc2d9d731503b53a3baa67a8bd64c7e2771578))

## [0.30.11](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.10...v0.30.11) (2025-10-29)


### Bug Fixes

* update roles ([#1138](https://github.com/chanzuckerberg/aws-oidc/issues/1138)) ([24a50e8](https://github.com/chanzuckerberg/aws-oidc/commit/24a50e880df976abd3a8892ba8b18cf01a9783ba))

## [0.30.10](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.9...v0.30.10) (2025-10-07)


### Misc

* CCIE-4984 conform to open sourcing guidelines ([#1133](https://github.com/chanzuckerberg/aws-oidc/issues/1133)) ([1588193](https://github.com/chanzuckerberg/aws-oidc/commit/1588193a42ea8985ae7a899702a07f9c46520559))

## [0.30.9](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.8...v0.30.9) (2025-06-03)


### Misc

* CCIE-4332 use GH_ACTIONS_HELPER_* ([#1128](https://github.com/chanzuckerberg/aws-oidc/issues/1128)) ([6df7c68](https://github.com/chanzuckerberg/aws-oidc/commit/6df7c6886bb0f640432e82a0b730331aa910f8d7))
* update stack helm chart version for prod ([#1123](https://github.com/chanzuckerberg/aws-oidc/issues/1123)) ([dbf8093](https://github.com/chanzuckerberg/aws-oidc/commit/dbf809358210569324dc036acd2100b56a55ea9b))

## [0.30.8](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.7...v0.30.8) (2025-05-08)


### Bug Fixes

* have 3 pods running in prod ([#1125](https://github.com/chanzuckerberg/aws-oidc/issues/1125)) ([b6d01d0](https://github.com/chanzuckerberg/aws-oidc/commit/b6d01d041f73e37b39e17c8e537714c0b0e5111d))


### Misc

* update stack helm chart version for rdev ([#1124](https://github.com/chanzuckerberg/aws-oidc/issues/1124)) ([a128166](https://github.com/chanzuckerberg/aws-oidc/commit/a128166a1f0ccc3ab360ab3a02f8164e245f336d))

## [0.30.7](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.6...v0.30.7) (2025-05-07)


### Bug Fixes

* nil error check ([eade713](https://github.com/chanzuckerberg/aws-oidc/commit/eade713c5a15e9ce9f044c385441c2d5852e8a5c))

## [0.30.6](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.5...v0.30.6) (2025-05-07)


### Bug Fixes

* permissions to write to homebrew repo ([bd50070](https://github.com/chanzuckerberg/aws-oidc/commit/bd5007023fdf64993ad4cd55dc0bf6359832b1e1))

## [0.30.5](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.4...v0.30.5) (2025-05-07)


### Bug Fixes

* wrong goreleaser file ([b579e26](https://github.com/chanzuckerberg/aws-oidc/commit/b579e26b3533d43c9d414d03841c9a6141abc016))


### Misc

* update resources; remove outdated sentry ([647e5ad](https://github.com/chanzuckerberg/aws-oidc/commit/647e5adf913575291b96adab15031743560b28c8))

## [0.30.4](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.3...v0.30.4) (2025-05-07)


### Bug Fixes

* don't build docker as part of release-ci ([8fe4ed8](https://github.com/chanzuckerberg/aws-oidc/commit/8fe4ed8da0f327f927bc8c32461c91451ae089e3))

## [0.30.3](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.2...v0.30.3) (2025-05-07)


### Bug Fixes

* need a privileged runner ([a3a9bc1](https://github.com/chanzuckerberg/aws-oidc/commit/a3a9bc1bc333edb1d70901cf583094fdc107c72a))

## [0.30.2](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.1...v0.30.2) (2025-05-07)


### Bug Fixes

* missing permissions ([39d034c](https://github.com/chanzuckerberg/aws-oidc/commit/39d034c0a29027a01e8e3620a6f087272f77aea9))

## [0.30.1](https://github.com/chanzuckerberg/aws-oidc/compare/v0.30.0...v0.30.1) (2025-05-07)


### Bug Fixes

* release-cli needs aws creds to build ([#1115](https://github.com/chanzuckerberg/aws-oidc/issues/1115)) ([1d4c798](https://github.com/chanzuckerberg/aws-oidc/commit/1d4c798be0614d20a1aebc2a713622ec69ff59c3))

## [0.30.0](https://github.com/chanzuckerberg/aws-oidc/compare/v0.29.0...v0.30.0) (2025-05-07)


### Features

* allow dependabot to automerge ([#737](https://github.com/chanzuckerberg/aws-oidc/issues/737)) ([a20bce2](https://github.com/chanzuckerberg/aws-oidc/commit/a20bce2169ae6b93673e78341e7ff1b9570d2dbc))
* Downgrade go to 1.20 ([#792](https://github.com/chanzuckerberg/aws-oidc/issues/792)) ([d57d4e1](https://github.com/chanzuckerberg/aws-oidc/commit/d57d4e184366e0a340218605cf68a370b218d615))
* migrate aws OIDC to argus deployment prod central ([#1108](https://github.com/chanzuckerberg/aws-oidc/issues/1108)) ([bad5b05](https://github.com/chanzuckerberg/aws-oidc/commit/bad5b0519e2a8dc500ed8fa761fa78b23e7983e3))
* Upgrade to go 1.21 ([#782](https://github.com/chanzuckerberg/aws-oidc/issues/782)) ([5783aff](https://github.com/chanzuckerberg/aws-oidc/commit/5783aff5ffd80b3e0b9ebb43e8405ac5e01ec710))


### Bug Fixes

* helper scripts to update deps ([#1086](https://github.com/chanzuckerberg/aws-oidc/issues/1086)) ([6cba7e1](https://github.com/chanzuckerberg/aws-oidc/commit/6cba7e1c11c9cebcd4c4d2698471c5c216e89096))
* release-please version ([47df96f](https://github.com/chanzuckerberg/aws-oidc/commit/47df96ff0b790a8330201c9d7c846c54734f4a30))
* remove toolchain directive ([#1101](https://github.com/chanzuckerberg/aws-oidc/issues/1101)) ([107923e](https://github.com/chanzuckerberg/aws-oidc/commit/107923e8b457f957a5248d490b07ccc0495ca237))
* Update runs-on labels in GitHub Actions workflows ([#1095](https://github.com/chanzuckerberg/aws-oidc/issues/1095)) ([fe0228b](https://github.com/chanzuckerberg/aws-oidc/commit/fe0228b88d9b3c92fee5b6d7a59c9f185e8cd5d9))
* Update runs-on to use ARM64 or X64 ([fe0228b](https://github.com/chanzuckerberg/aws-oidc/commit/fe0228b88d9b3c92fee5b6d7a59c9f185e8cd5d9))


### Misc

* bump actions/checkout from 3 to 4 ([#1112](https://github.com/chanzuckerberg/aws-oidc/issues/1112)) ([13a0959](https://github.com/chanzuckerberg/aws-oidc/commit/13a0959c3bdebad8d699928cf5c51c835ef9eec3))
* bump actions/create-github-app-token from 1 to 2 ([#1113](https://github.com/chanzuckerberg/aws-oidc/issues/1113)) ([761e2fe](https://github.com/chanzuckerberg/aws-oidc/commit/761e2fe4c7fecc75cb15b0a91c4be6f97bde80a2))
* bump actions/setup-go from 4 to 5 ([#1110](https://github.com/chanzuckerberg/aws-oidc/issues/1110)) ([af5cdc7](https://github.com/chanzuckerberg/aws-oidc/commit/af5cdc72d8929aa12c73ea127cea6d0b2400c14b))
* bump dependabot/fetch-metadata from 1 to 2 ([#1111](https://github.com/chanzuckerberg/aws-oidc/issues/1111)) ([23c3a5e](https://github.com/chanzuckerberg/aws-oidc/commit/23c3a5e7226a1d4e029a3dd444e67a955c644453))
* bump github.com/AlecAivazis/survey/v2 from 2.3.6 to 2.3.7 ([#682](https://github.com/chanzuckerberg/aws-oidc/issues/682)) ([eb639d0](https://github.com/chanzuckerberg/aws-oidc/commit/eb639d098f87499b8cd9b350ea3c88dcc4873c3e))
* bump github.com/aws/aws-sdk-go from 1.44.220 to 1.44.221 ([#584](https://github.com/chanzuckerberg/aws-oidc/issues/584)) ([3998a45](https://github.com/chanzuckerberg/aws-oidc/commit/3998a45ef7230fe18bd81dbb072c264fdcef70e4))
* bump github.com/aws/aws-sdk-go from 1.44.221 to 1.44.222 ([#586](https://github.com/chanzuckerberg/aws-oidc/issues/586)) ([d89221e](https://github.com/chanzuckerberg/aws-oidc/commit/d89221ee1f073ea8918765bc9f468f77f05676bd))
* bump github.com/aws/aws-sdk-go from 1.44.222 to 1.44.223 ([#588](https://github.com/chanzuckerberg/aws-oidc/issues/588)) ([64a8cd2](https://github.com/chanzuckerberg/aws-oidc/commit/64a8cd2ec6cad4b30b20f4a1235068399a856377))
* bump github.com/aws/aws-sdk-go from 1.44.223 to 1.44.229 ([#595](https://github.com/chanzuckerberg/aws-oidc/issues/595)) ([78416dd](https://github.com/chanzuckerberg/aws-oidc/commit/78416dd47123b5e612ef63d6f8dfc4f630ee2a7b))
* bump github.com/aws/aws-sdk-go from 1.44.229 to 1.44.230 ([#597](https://github.com/chanzuckerberg/aws-oidc/issues/597)) ([5473ad0](https://github.com/chanzuckerberg/aws-oidc/commit/5473ad0d37fd2d7ba977f23fbc1f9f722467ff03))
* bump github.com/aws/aws-sdk-go from 1.44.230 to 1.44.231 ([#598](https://github.com/chanzuckerberg/aws-oidc/issues/598)) ([34cf4ea](https://github.com/chanzuckerberg/aws-oidc/commit/34cf4eaaa938a14668b74b4531677bdc9e1d0a93))
* bump github.com/aws/aws-sdk-go from 1.44.231 to 1.44.232 ([#599](https://github.com/chanzuckerberg/aws-oidc/issues/599)) ([d48a89f](https://github.com/chanzuckerberg/aws-oidc/commit/d48a89f8a9cd22106c0bfa29df2f7229a1e5cb7a))
* bump github.com/aws/aws-sdk-go from 1.44.232 to 1.44.234 ([#603](https://github.com/chanzuckerberg/aws-oidc/issues/603)) ([5b10b25](https://github.com/chanzuckerberg/aws-oidc/commit/5b10b254e10122aa871df9d26090559a2203b5d6))
* bump github.com/aws/aws-sdk-go from 1.44.234 to 1.44.235 ([#606](https://github.com/chanzuckerberg/aws-oidc/issues/606)) ([a48a02d](https://github.com/chanzuckerberg/aws-oidc/commit/a48a02df0e45e9d9d7d50233f2080e4cfef57a8f))
* bump github.com/aws/aws-sdk-go from 1.44.235 to 1.44.236 ([#608](https://github.com/chanzuckerberg/aws-oidc/issues/608)) ([44891bf](https://github.com/chanzuckerberg/aws-oidc/commit/44891bf88d5022c9fde19570dee25bdabaf297f7))
* bump github.com/aws/aws-sdk-go from 1.44.236 to 1.44.237 ([#611](https://github.com/chanzuckerberg/aws-oidc/issues/611)) ([f8f55fd](https://github.com/chanzuckerberg/aws-oidc/commit/f8f55fdb596a1296b7ca140da7d7c63428c9119e))
* bump github.com/aws/aws-sdk-go from 1.44.237 to 1.44.238 ([#613](https://github.com/chanzuckerberg/aws-oidc/issues/613)) ([dc1ef13](https://github.com/chanzuckerberg/aws-oidc/commit/dc1ef13221e3a73f4331fdd2d2855c06126f93ed))
* bump github.com/aws/aws-sdk-go from 1.44.238 to 1.44.241 ([#618](https://github.com/chanzuckerberg/aws-oidc/issues/618)) ([0d6fafd](https://github.com/chanzuckerberg/aws-oidc/commit/0d6fafd1e8497b99a5edda2f0a5b10bbec8b19e2))
* bump github.com/aws/aws-sdk-go from 1.44.241 to 1.44.242 ([#620](https://github.com/chanzuckerberg/aws-oidc/issues/620)) ([9f41fc1](https://github.com/chanzuckerberg/aws-oidc/commit/9f41fc15a9d100a1ac55ad21693b3ebb65c930e1))
* bump github.com/aws/aws-sdk-go from 1.44.242 to 1.44.243 ([#621](https://github.com/chanzuckerberg/aws-oidc/issues/621)) ([e7e58f6](https://github.com/chanzuckerberg/aws-oidc/commit/e7e58f6be724e3153a6c7dffaf325625f8d81a9e))
* bump github.com/aws/aws-sdk-go from 1.44.243 to 1.44.244 ([#623](https://github.com/chanzuckerberg/aws-oidc/issues/623)) ([9bdd071](https://github.com/chanzuckerberg/aws-oidc/commit/9bdd071d5543e7a15a9f107d5b2ffb1e1d3df8a9))
* bump github.com/aws/aws-sdk-go from 1.44.244 to 1.44.245 ([#626](https://github.com/chanzuckerberg/aws-oidc/issues/626)) ([215b869](https://github.com/chanzuckerberg/aws-oidc/commit/215b8692a85a32f4eb41bef9ca0c91aea958a5e0))
* bump github.com/aws/aws-sdk-go from 1.44.245 to 1.44.246 ([#627](https://github.com/chanzuckerberg/aws-oidc/issues/627)) ([36eeb5a](https://github.com/chanzuckerberg/aws-oidc/commit/36eeb5af391e3bd3dfadd4d20206b0aee5693985))
* bump github.com/aws/aws-sdk-go from 1.44.246 to 1.44.247 ([#629](https://github.com/chanzuckerberg/aws-oidc/issues/629)) ([5c61e60](https://github.com/chanzuckerberg/aws-oidc/commit/5c61e6053aa778a9834bac9e45df12fe0e33afb4))
* bump github.com/aws/aws-sdk-go from 1.44.247 to 1.44.248 ([#631](https://github.com/chanzuckerberg/aws-oidc/issues/631)) ([51e2f37](https://github.com/chanzuckerberg/aws-oidc/commit/51e2f37ce4ddba41ce71b6d57379ab44614f0aa0))
* bump github.com/aws/aws-sdk-go from 1.44.248 to 1.44.249 ([#632](https://github.com/chanzuckerberg/aws-oidc/issues/632)) ([fe948c0](https://github.com/chanzuckerberg/aws-oidc/commit/fe948c024542a70f0e44694865c75e22ef8827e6))
* bump github.com/aws/aws-sdk-go from 1.44.249 to 1.44.250 ([#633](https://github.com/chanzuckerberg/aws-oidc/issues/633)) ([a60275f](https://github.com/chanzuckerberg/aws-oidc/commit/a60275fa0aea00dec9766c6f50b1c9e98676929d))
* bump github.com/aws/aws-sdk-go from 1.44.250 to 1.44.251 ([#634](https://github.com/chanzuckerberg/aws-oidc/issues/634)) ([f5496e5](https://github.com/chanzuckerberg/aws-oidc/commit/f5496e54a7c1461dc5eefd83e630201311ee5cf5))
* bump github.com/aws/aws-sdk-go from 1.44.251 to 1.44.253 ([#637](https://github.com/chanzuckerberg/aws-oidc/issues/637)) ([3ba839f](https://github.com/chanzuckerberg/aws-oidc/commit/3ba839fbcf07db479caf6b000ae79cca7ddaa325))
* bump github.com/aws/aws-sdk-go from 1.44.253 to 1.44.254 ([#638](https://github.com/chanzuckerberg/aws-oidc/issues/638)) ([9e5dd6d](https://github.com/chanzuckerberg/aws-oidc/commit/9e5dd6df77409107cedcfa1af5d3eab84edc7692))
* bump github.com/aws/aws-sdk-go from 1.44.254 to 1.44.255 ([#640](https://github.com/chanzuckerberg/aws-oidc/issues/640)) ([70c45e7](https://github.com/chanzuckerberg/aws-oidc/commit/70c45e726befbb45d7991e6dc5d5b3dcdd544822))
* bump github.com/aws/aws-sdk-go from 1.44.255 to 1.44.257 ([#644](https://github.com/chanzuckerberg/aws-oidc/issues/644)) ([0f7462c](https://github.com/chanzuckerberg/aws-oidc/commit/0f7462c809b792fb0f8efd6f699e2c5ca90afb8c))
* bump github.com/aws/aws-sdk-go from 1.44.257 to 1.44.258 ([#646](https://github.com/chanzuckerberg/aws-oidc/issues/646)) ([f32c1b6](https://github.com/chanzuckerberg/aws-oidc/commit/f32c1b6965fc0051b099344225ef5257c556b87d))
* bump github.com/aws/aws-sdk-go from 1.44.258 to 1.44.259 ([#647](https://github.com/chanzuckerberg/aws-oidc/issues/647)) ([3120d7f](https://github.com/chanzuckerberg/aws-oidc/commit/3120d7f598425d3175a151929b02c025c00b173b))
* bump github.com/aws/aws-sdk-go from 1.44.259 to 1.44.260 ([#649](https://github.com/chanzuckerberg/aws-oidc/issues/649)) ([a69ed11](https://github.com/chanzuckerberg/aws-oidc/commit/a69ed11ed5b3a1d34e427e0724d2ec7333c7adef))
* bump github.com/aws/aws-sdk-go from 1.44.260 to 1.44.261 ([#650](https://github.com/chanzuckerberg/aws-oidc/issues/650)) ([cb28a4f](https://github.com/chanzuckerberg/aws-oidc/commit/cb28a4f377a881ad58386604a2600857b5c3a3e5))
* bump github.com/aws/aws-sdk-go from 1.44.261 to 1.44.267 ([#661](https://github.com/chanzuckerberg/aws-oidc/issues/661)) ([9464cb8](https://github.com/chanzuckerberg/aws-oidc/commit/9464cb8a5d6ee7b7352cd208ca7e4ca0c4330aeb))
* bump github.com/aws/aws-sdk-go from 1.44.269 to 1.44.271 ([#665](https://github.com/chanzuckerberg/aws-oidc/issues/665)) ([1ee6eb5](https://github.com/chanzuckerberg/aws-oidc/commit/1ee6eb5025ab8bec585031aea7400d6725767194))
* bump github.com/aws/aws-sdk-go from 1.44.271 to 1.44.272 ([#668](https://github.com/chanzuckerberg/aws-oidc/issues/668)) ([48e51fd](https://github.com/chanzuckerberg/aws-oidc/commit/48e51fd170485007cadef3fbd4d88d8fe5bf4fb6))
* bump github.com/aws/aws-sdk-go from 1.44.272 to 1.44.273 ([#670](https://github.com/chanzuckerberg/aws-oidc/issues/670)) ([a068c94](https://github.com/chanzuckerberg/aws-oidc/commit/a068c949f0f80223cbf0d7f9096e0df69b5021d6))
* bump github.com/aws/aws-sdk-go from 1.44.273 to 1.44.274 ([#671](https://github.com/chanzuckerberg/aws-oidc/issues/671)) ([410ac0c](https://github.com/chanzuckerberg/aws-oidc/commit/410ac0cc4fa6e80fe39731ac87162e125e732a27))
* bump github.com/aws/aws-sdk-go from 1.44.274 to 1.44.275 ([#674](https://github.com/chanzuckerberg/aws-oidc/issues/674)) ([c12f1e0](https://github.com/chanzuckerberg/aws-oidc/commit/c12f1e0a13e2b32de7f6bc2b9a9e8096a923ca44))
* bump github.com/aws/aws-sdk-go from 1.44.275 to 1.44.277 ([#676](https://github.com/chanzuckerberg/aws-oidc/issues/676)) ([1e4ac15](https://github.com/chanzuckerberg/aws-oidc/commit/1e4ac15c561adf23357817819d7ba1935fdb5464))
* bump github.com/aws/aws-sdk-go from 1.44.277 to 1.44.280 ([#679](https://github.com/chanzuckerberg/aws-oidc/issues/679)) ([5fb236f](https://github.com/chanzuckerberg/aws-oidc/commit/5fb236fcf38923cba34bf9c7f2abddcb7a09fa79))
* bump github.com/aws/aws-sdk-go from 1.44.282 to 1.44.283 ([#686](https://github.com/chanzuckerberg/aws-oidc/issues/686)) ([035dd6e](https://github.com/chanzuckerberg/aws-oidc/commit/035dd6e61bb76492b5953aa63e8a4acfec561e5c))
* bump github.com/aws/aws-sdk-go from 1.44.283 to 1.44.284 ([#688](https://github.com/chanzuckerberg/aws-oidc/issues/688)) ([5c3aae4](https://github.com/chanzuckerberg/aws-oidc/commit/5c3aae498947383963e52fea14a98d03cb480be6))
* bump github.com/aws/aws-sdk-go from 1.44.284 to 1.44.285 ([#690](https://github.com/chanzuckerberg/aws-oidc/issues/690)) ([e424ecb](https://github.com/chanzuckerberg/aws-oidc/commit/e424ecb9d9d732b88e91fce98e4706c3ca75f803))
* bump github.com/aws/aws-sdk-go from 1.44.285 to 1.44.286 ([#691](https://github.com/chanzuckerberg/aws-oidc/issues/691)) ([da5e2f1](https://github.com/chanzuckerberg/aws-oidc/commit/da5e2f119ed330df66c88d36f24a9962fecd239d))
* bump github.com/aws/aws-sdk-go from 1.44.286 to 1.44.289 ([#695](https://github.com/chanzuckerberg/aws-oidc/issues/695)) ([05b6872](https://github.com/chanzuckerberg/aws-oidc/commit/05b6872731301cdecc72513c60c577cfa3f5b791))
* bump github.com/aws/aws-sdk-go from 1.44.289 to 1.44.291 ([#699](https://github.com/chanzuckerberg/aws-oidc/issues/699)) ([68f5c51](https://github.com/chanzuckerberg/aws-oidc/commit/68f5c51c59fd1f4458856846e80b22777ae87eef))
* bump github.com/aws/aws-sdk-go from 1.44.291 to 1.44.294 ([#703](https://github.com/chanzuckerberg/aws-oidc/issues/703)) ([fc7eba2](https://github.com/chanzuckerberg/aws-oidc/commit/fc7eba276514bc7cf8c1d8dbb7776d5593016c39))
* bump github.com/aws/aws-sdk-go from 1.44.294 to 1.44.295 ([#705](https://github.com/chanzuckerberg/aws-oidc/issues/705)) ([ee1e4ea](https://github.com/chanzuckerberg/aws-oidc/commit/ee1e4ea5beb5d4552179f1caa4306df02d5789d2))
* bump github.com/aws/aws-sdk-go from 1.44.295 to 1.44.298 ([#707](https://github.com/chanzuckerberg/aws-oidc/issues/707)) ([a6e5079](https://github.com/chanzuckerberg/aws-oidc/commit/a6e5079f1491a17e12aa8061f5eb8e08c555c370))
* bump github.com/aws/aws-sdk-go from 1.44.298 to 1.44.299 ([#708](https://github.com/chanzuckerberg/aws-oidc/issues/708)) ([d3d0b02](https://github.com/chanzuckerberg/aws-oidc/commit/d3d0b024df7967222124adf690dc9ea4b389c29f))
* bump github.com/aws/aws-sdk-go from 1.44.299 to 1.44.300 ([#709](https://github.com/chanzuckerberg/aws-oidc/issues/709)) ([fdc0ca4](https://github.com/chanzuckerberg/aws-oidc/commit/fdc0ca47c051abf9c21eefd5b5dfe42b9522ccad))
* bump github.com/aws/aws-sdk-go from 1.44.300 to 1.44.306 ([#716](https://github.com/chanzuckerberg/aws-oidc/issues/716)) ([2b548b8](https://github.com/chanzuckerberg/aws-oidc/commit/2b548b821ea04458aa9fbf0ffc9c52cfe1d647b5))
* bump github.com/aws/aws-sdk-go from 1.44.306 to 1.44.307 ([#718](https://github.com/chanzuckerberg/aws-oidc/issues/718)) ([ddc9508](https://github.com/chanzuckerberg/aws-oidc/commit/ddc95082ce6d97486a7cc096f59e30feee3a384b))
* bump github.com/aws/aws-sdk-go from 1.44.307 to 1.44.308 ([#720](https://github.com/chanzuckerberg/aws-oidc/issues/720)) ([e968ddb](https://github.com/chanzuckerberg/aws-oidc/commit/e968ddb8bc120bc7ecb84c9eab924aaa8fafdda0))
* bump github.com/aws/aws-sdk-go from 1.44.308 to 1.44.309 ([#722](https://github.com/chanzuckerberg/aws-oidc/issues/722)) ([bc82803](https://github.com/chanzuckerberg/aws-oidc/commit/bc82803409c9c98f0ab03a356b72b043be9d0151))
* bump github.com/aws/aws-sdk-go from 1.44.309 to 1.44.312 ([#725](https://github.com/chanzuckerberg/aws-oidc/issues/725)) ([a2dc901](https://github.com/chanzuckerberg/aws-oidc/commit/a2dc901e0050dde66d1403806786efbc338dd664))
* bump github.com/aws/aws-sdk-go from 1.44.312 to 1.44.313 ([#730](https://github.com/chanzuckerberg/aws-oidc/issues/730)) ([360fcb0](https://github.com/chanzuckerberg/aws-oidc/commit/360fcb0368b8d464ec8ef6f4d02fe9a431454504))
* bump github.com/aws/aws-sdk-go from 1.44.313 to 1.44.314 ([#731](https://github.com/chanzuckerberg/aws-oidc/issues/731)) ([2d2602b](https://github.com/chanzuckerberg/aws-oidc/commit/2d2602b3d858ce48f0ff93afcad09290468d3264))
* bump github.com/aws/aws-sdk-go from 1.44.314 to 1.44.315 ([#732](https://github.com/chanzuckerberg/aws-oidc/issues/732)) ([9384956](https://github.com/chanzuckerberg/aws-oidc/commit/9384956644239515bf53a0d7f18fcb2308f847fa))
* bump github.com/aws/aws-sdk-go from 1.44.315 to 1.44.316 ([#734](https://github.com/chanzuckerberg/aws-oidc/issues/734)) ([49f5c3e](https://github.com/chanzuckerberg/aws-oidc/commit/49f5c3edeafddbc923abdb5eee025d190f54f8b0))
* bump github.com/aws/aws-sdk-go from 1.44.316 to 1.44.317 ([#736](https://github.com/chanzuckerberg/aws-oidc/issues/736)) ([a3832c5](https://github.com/chanzuckerberg/aws-oidc/commit/a3832c58f3d9f56a4768c05d368ea96d259bbc48))
* bump github.com/aws/aws-sdk-go from 1.44.317 to 1.44.318 ([#739](https://github.com/chanzuckerberg/aws-oidc/issues/739)) ([42eb492](https://github.com/chanzuckerberg/aws-oidc/commit/42eb49221ed81ae24b805e225f9777bc951feb66))
* bump github.com/aws/aws-sdk-go from 1.44.318 to 1.44.319 ([#740](https://github.com/chanzuckerberg/aws-oidc/issues/740)) ([dc62272](https://github.com/chanzuckerberg/aws-oidc/commit/dc622720478d1c181e2b85287e51e19f68e7cd92))
* bump github.com/aws/aws-sdk-go from 1.44.319 to 1.44.320 ([#742](https://github.com/chanzuckerberg/aws-oidc/issues/742)) ([3cffbcc](https://github.com/chanzuckerberg/aws-oidc/commit/3cffbcc1fc179ce6626253f1d98567dc84891173))
* bump github.com/aws/aws-sdk-go from 1.44.320 to 1.44.321 ([#744](https://github.com/chanzuckerberg/aws-oidc/issues/744)) ([a4e27b2](https://github.com/chanzuckerberg/aws-oidc/commit/a4e27b2b4c8d8660f0cdaa9b237ef00159a3967a))
* bump github.com/aws/aws-sdk-go from 1.44.321 to 1.44.322 ([#746](https://github.com/chanzuckerberg/aws-oidc/issues/746)) ([042fbd9](https://github.com/chanzuckerberg/aws-oidc/commit/042fbd9cf0035836e87fba36507ba7b10f0f2f03))
* bump github.com/aws/aws-sdk-go from 1.44.322 to 1.44.323 ([#748](https://github.com/chanzuckerberg/aws-oidc/issues/748)) ([922380c](https://github.com/chanzuckerberg/aws-oidc/commit/922380cf5f4261169940723b34f8c369cb85f1ae))
* bump github.com/aws/aws-sdk-go from 1.44.323 to 1.44.324 ([#751](https://github.com/chanzuckerberg/aws-oidc/issues/751)) ([5396e68](https://github.com/chanzuckerberg/aws-oidc/commit/5396e68f6d22f7c626a3c92121c5cff17622790b))
* bump github.com/aws/aws-sdk-go from 1.44.324 to 1.44.325 ([#753](https://github.com/chanzuckerberg/aws-oidc/issues/753)) ([70a7c03](https://github.com/chanzuckerberg/aws-oidc/commit/70a7c03431339b2646e8866339d7c515a2b3aa8c))
* bump github.com/aws/aws-sdk-go from 1.44.325 to 1.44.326 ([#755](https://github.com/chanzuckerberg/aws-oidc/issues/755)) ([816e61d](https://github.com/chanzuckerberg/aws-oidc/commit/816e61d1a0b0e46547146c9f150051f04672077a))
* bump github.com/aws/aws-sdk-go from 1.44.326 to 1.44.327 ([#757](https://github.com/chanzuckerberg/aws-oidc/issues/757)) ([5a194e9](https://github.com/chanzuckerberg/aws-oidc/commit/5a194e973c044be0f2ed7c561abda90e39444a72))
* bump github.com/aws/aws-sdk-go from 1.44.327 to 1.44.328 ([#758](https://github.com/chanzuckerberg/aws-oidc/issues/758)) ([0d28f89](https://github.com/chanzuckerberg/aws-oidc/commit/0d28f89cea711d21b5b72fc8abe17ad682301daa))
* bump github.com/aws/aws-sdk-go from 1.44.328 to 1.44.329 ([#760](https://github.com/chanzuckerberg/aws-oidc/issues/760)) ([6f0adcc](https://github.com/chanzuckerberg/aws-oidc/commit/6f0adcc828398818a1a37fe54fa34312a2b35d8f))
* bump github.com/aws/aws-sdk-go from 1.44.329 to 1.44.330 ([#762](https://github.com/chanzuckerberg/aws-oidc/issues/762)) ([6d0d5aa](https://github.com/chanzuckerberg/aws-oidc/commit/6d0d5aa57762ba577ddf9994e0a071c420bb9e25))
* bump github.com/aws/aws-sdk-go from 1.44.330 to 1.44.331 ([#764](https://github.com/chanzuckerberg/aws-oidc/issues/764)) ([44edfd2](https://github.com/chanzuckerberg/aws-oidc/commit/44edfd2d313851dbc2543bf9b996183bf9455382))
* bump github.com/aws/aws-sdk-go from 1.44.331 to 1.44.332 ([#766](https://github.com/chanzuckerberg/aws-oidc/issues/766)) ([da235c0](https://github.com/chanzuckerberg/aws-oidc/commit/da235c0f10a3a4df3d9a75675507715f860e2e99))
* bump github.com/aws/aws-sdk-go from 1.44.332 to 1.44.333 ([#767](https://github.com/chanzuckerberg/aws-oidc/issues/767)) ([5f06eff](https://github.com/chanzuckerberg/aws-oidc/commit/5f06eff05666c4efaf7a76f50835322298ef0fbe))
* bump github.com/aws/aws-sdk-go from 1.44.333 to 1.44.334 ([#769](https://github.com/chanzuckerberg/aws-oidc/issues/769)) ([4e110f3](https://github.com/chanzuckerberg/aws-oidc/commit/4e110f31d653a1ec6889b3bcd309c098c50b2bef))
* bump github.com/aws/aws-sdk-go from 1.44.334 to 1.45.0 ([#771](https://github.com/chanzuckerberg/aws-oidc/issues/771)) ([0dea440](https://github.com/chanzuckerberg/aws-oidc/commit/0dea4406df86ab77863786886a207492a5187a6d))
* bump github.com/aws/aws-sdk-go from 1.45.0 to 1.45.1 ([#773](https://github.com/chanzuckerberg/aws-oidc/issues/773)) ([e37ab65](https://github.com/chanzuckerberg/aws-oidc/commit/e37ab657fc6076384e4ebbb4bb4e29fea82fd68d))
* bump github.com/aws/aws-sdk-go from 1.45.1 to 1.45.2 ([#774](https://github.com/chanzuckerberg/aws-oidc/issues/774)) ([4b2ff31](https://github.com/chanzuckerberg/aws-oidc/commit/4b2ff31660cbe0dfd1cd008ea4c0b2680c477470))
* bump github.com/aws/aws-sdk-go from 1.45.12 to 1.45.13 ([#787](https://github.com/chanzuckerberg/aws-oidc/issues/787)) ([9d42752](https://github.com/chanzuckerberg/aws-oidc/commit/9d42752bf1be7b3bac580d99d981be8622a644a1))
* bump github.com/aws/aws-sdk-go from 1.45.13 to 1.45.14 ([#788](https://github.com/chanzuckerberg/aws-oidc/issues/788)) ([77c2d5e](https://github.com/chanzuckerberg/aws-oidc/commit/77c2d5e39bb4ad0f0f32bbabea4bba8df9eb300d))
* bump github.com/aws/aws-sdk-go from 1.45.14 to 1.45.15 ([#790](https://github.com/chanzuckerberg/aws-oidc/issues/790)) ([0ab7df4](https://github.com/chanzuckerberg/aws-oidc/commit/0ab7df4b92566a042856b61658d3a7b0b86239de))
* bump github.com/aws/aws-sdk-go from 1.45.15 to 1.45.16 ([#791](https://github.com/chanzuckerberg/aws-oidc/issues/791)) ([c255f29](https://github.com/chanzuckerberg/aws-oidc/commit/c255f29848da0e80b932a43fd8ef13cce9407b3f))
* bump github.com/aws/aws-sdk-go from 1.45.16 to 1.45.17 ([#793](https://github.com/chanzuckerberg/aws-oidc/issues/793)) ([b098e18](https://github.com/chanzuckerberg/aws-oidc/commit/b098e18803a9cfeb036b39df080be8ae3471b146))
* bump github.com/aws/aws-sdk-go from 1.45.17 to 1.45.18 ([#794](https://github.com/chanzuckerberg/aws-oidc/issues/794)) ([d0b2889](https://github.com/chanzuckerberg/aws-oidc/commit/d0b2889272e69f0f9c6a76a1fbbad044f54a61a9))
* bump github.com/aws/aws-sdk-go from 1.45.18 to 1.45.19 ([#796](https://github.com/chanzuckerberg/aws-oidc/issues/796)) ([f617607](https://github.com/chanzuckerberg/aws-oidc/commit/f61760799ded9849e4f936206174551ed72d7919))
* bump github.com/aws/aws-sdk-go from 1.45.19 to 1.45.20 ([#798](https://github.com/chanzuckerberg/aws-oidc/issues/798)) ([7477b22](https://github.com/chanzuckerberg/aws-oidc/commit/7477b22607715ccadc5f961a1299d3737ce14c6b))
* bump github.com/aws/aws-sdk-go from 1.45.2 to 1.45.3 ([#776](https://github.com/chanzuckerberg/aws-oidc/issues/776)) ([ff8ab5b](https://github.com/chanzuckerberg/aws-oidc/commit/ff8ab5b72030a624d2a2b0072dcfa50401e4f550))
* bump github.com/aws/aws-sdk-go from 1.45.20 to 1.45.21 ([#800](https://github.com/chanzuckerberg/aws-oidc/issues/800)) ([9824bf1](https://github.com/chanzuckerberg/aws-oidc/commit/9824bf17159fe6caf44bedb373ca0cdd79493dba))
* bump github.com/aws/aws-sdk-go from 1.45.21 to 1.45.22 ([#802](https://github.com/chanzuckerberg/aws-oidc/issues/802)) ([4e28dad](https://github.com/chanzuckerberg/aws-oidc/commit/4e28dad6e1d3987f8de7963c0d59a6bc0769bf62))
* bump github.com/aws/aws-sdk-go from 1.45.22 to 1.45.23 ([#804](https://github.com/chanzuckerberg/aws-oidc/issues/804)) ([f519943](https://github.com/chanzuckerberg/aws-oidc/commit/f519943eabf72da894c7b9142a9cfe9580b9f133))
* bump github.com/aws/aws-sdk-go from 1.45.23 to 1.45.24 ([#805](https://github.com/chanzuckerberg/aws-oidc/issues/805)) ([2e9572a](https://github.com/chanzuckerberg/aws-oidc/commit/2e9572acc3aea10fcaca5a42cbcdb2dde9f1c370))
* bump github.com/aws/aws-sdk-go from 1.45.24 to 1.45.25 ([#808](https://github.com/chanzuckerberg/aws-oidc/issues/808)) ([ec54b0e](https://github.com/chanzuckerberg/aws-oidc/commit/ec54b0edce908c4704b969d5d23f193ace669d8f))
* bump github.com/aws/aws-sdk-go from 1.45.25 to 1.45.26 ([#810](https://github.com/chanzuckerberg/aws-oidc/issues/810)) ([30f2291](https://github.com/chanzuckerberg/aws-oidc/commit/30f2291a6f7ea973d94ba6d5ee290314cb6871fd))
* bump github.com/aws/aws-sdk-go from 1.45.26 to 1.45.27 ([#812](https://github.com/chanzuckerberg/aws-oidc/issues/812)) ([5068557](https://github.com/chanzuckerberg/aws-oidc/commit/5068557313ccec30e3c2cf46ab985e8828b53370))
* bump github.com/aws/aws-sdk-go from 1.45.27 to 1.45.28 ([#813](https://github.com/chanzuckerberg/aws-oidc/issues/813)) ([15a39e7](https://github.com/chanzuckerberg/aws-oidc/commit/15a39e7f6ea759346a65a98119018d9de65e73e5))
* bump github.com/aws/aws-sdk-go from 1.45.28 to 1.46.0 ([#815](https://github.com/chanzuckerberg/aws-oidc/issues/815)) ([077a83b](https://github.com/chanzuckerberg/aws-oidc/commit/077a83b71f2a9dd711c0976253907015271183eb))
* bump github.com/aws/aws-sdk-go from 1.45.3 to 1.45.4 ([#778](https://github.com/chanzuckerberg/aws-oidc/issues/778)) ([b1f9d1d](https://github.com/chanzuckerberg/aws-oidc/commit/b1f9d1d3e6ee1601fbfd1ac544409fc538a359d0))
* bump github.com/aws/aws-sdk-go from 1.45.4 to 1.45.5 ([#781](https://github.com/chanzuckerberg/aws-oidc/issues/781)) ([b48f389](https://github.com/chanzuckerberg/aws-oidc/commit/b48f3891608a7db5a7b2e27b81d089dbabb586f5))
* bump github.com/aws/aws-sdk-go from 1.45.5 to 1.45.12 ([#783](https://github.com/chanzuckerberg/aws-oidc/issues/783)) ([adfb562](https://github.com/chanzuckerberg/aws-oidc/commit/adfb56234bf139fa043a61879df8a21bb3284b2c))
* bump github.com/aws/aws-sdk-go from 1.46.0 to 1.46.2 ([#817](https://github.com/chanzuckerberg/aws-oidc/issues/817)) ([20c5914](https://github.com/chanzuckerberg/aws-oidc/commit/20c5914bf335039875f6a2a9aa06b3f3b0a3f748))
* bump github.com/aws/aws-sdk-go from 1.46.2 to 1.46.3 ([#819](https://github.com/chanzuckerberg/aws-oidc/issues/819)) ([2a7cb24](https://github.com/chanzuckerberg/aws-oidc/commit/2a7cb24a93a233e0a4063b655d7ae41ad78a8c70))
* bump github.com/aws/aws-sdk-go from 1.46.3 to 1.46.4 ([#823](https://github.com/chanzuckerberg/aws-oidc/issues/823)) ([afbe61f](https://github.com/chanzuckerberg/aws-oidc/commit/afbe61f0fa9d1a72e1a52b09ac7d01163d767f2b))
* bump github.com/aws/aws-sdk-go from 1.46.4 to 1.46.5 ([#824](https://github.com/chanzuckerberg/aws-oidc/issues/824)) ([2d9613c](https://github.com/chanzuckerberg/aws-oidc/commit/2d9613c1738b65a31ec791a5107d8a09b53e5f76))
* bump github.com/aws/aws-sdk-go from 1.46.5 to 1.46.6 ([#826](https://github.com/chanzuckerberg/aws-oidc/issues/826)) ([5dd5a70](https://github.com/chanzuckerberg/aws-oidc/commit/5dd5a703dfdbfc8cd6ff0e0cfa445849d9596a74))
* bump github.com/aws/aws-sdk-go from 1.46.6 to 1.46.7 ([#828](https://github.com/chanzuckerberg/aws-oidc/issues/828)) ([bd12246](https://github.com/chanzuckerberg/aws-oidc/commit/bd122469f373b68ee78bb024eeb84a533a186916))
* bump github.com/aws/aws-sdk-go from 1.46.7 to 1.47.0 ([#829](https://github.com/chanzuckerberg/aws-oidc/issues/829)) ([b65fc40](https://github.com/chanzuckerberg/aws-oidc/commit/b65fc407737a46ae979cbb574208b340233b6687))
* bump github.com/aws/aws-sdk-go from 1.47.0 to 1.47.1 ([#832](https://github.com/chanzuckerberg/aws-oidc/issues/832)) ([48b2264](https://github.com/chanzuckerberg/aws-oidc/commit/48b2264c4dbce8537c01cbed360dbcf045f01c55))
* bump github.com/aws/aws-sdk-go from 1.47.1 to 1.47.2 ([#833](https://github.com/chanzuckerberg/aws-oidc/issues/833)) ([e677b80](https://github.com/chanzuckerberg/aws-oidc/commit/e677b803a8036659921e3afe5c75e2899217d35b))
* bump github.com/aws/aws-sdk-go from 1.47.10 to 1.47.11 ([#846](https://github.com/chanzuckerberg/aws-oidc/issues/846)) ([d33bb25](https://github.com/chanzuckerberg/aws-oidc/commit/d33bb2527278e8f3a193769fb4d986fcc13c6f6b))
* bump github.com/aws/aws-sdk-go from 1.47.11 to 1.47.12 ([#847](https://github.com/chanzuckerberg/aws-oidc/issues/847)) ([d6f0e28](https://github.com/chanzuckerberg/aws-oidc/commit/d6f0e28b7d10fe96d825a56c79bef262ed9e0052))
* bump github.com/aws/aws-sdk-go from 1.47.12 to 1.47.13 ([#849](https://github.com/chanzuckerberg/aws-oidc/issues/849)) ([7226ac5](https://github.com/chanzuckerberg/aws-oidc/commit/7226ac52f0a98101ebecfcfaa63dc9d5744f9d9f))
* bump github.com/aws/aws-sdk-go from 1.47.13 to 1.48.0 ([#851](https://github.com/chanzuckerberg/aws-oidc/issues/851)) ([b0b32d1](https://github.com/chanzuckerberg/aws-oidc/commit/b0b32d142da7a436140642b4d3804f09ab3affaa))
* bump github.com/aws/aws-sdk-go from 1.47.2 to 1.47.3 ([#837](https://github.com/chanzuckerberg/aws-oidc/issues/837)) ([4dd0f85](https://github.com/chanzuckerberg/aws-oidc/commit/4dd0f85d555dcc3b309bfdd5c15a9f283e6b0f6a))
* bump github.com/aws/aws-sdk-go from 1.47.3 to 1.47.4 ([#839](https://github.com/chanzuckerberg/aws-oidc/issues/839)) ([cd9df82](https://github.com/chanzuckerberg/aws-oidc/commit/cd9df824063680b595f20d0d8455511c7aa8ec86))
* bump github.com/aws/aws-sdk-go from 1.47.4 to 1.47.5 ([#840](https://github.com/chanzuckerberg/aws-oidc/issues/840)) ([81f6604](https://github.com/chanzuckerberg/aws-oidc/commit/81f6604692014f01306938d9e50ec8e70ac91395))
* bump github.com/aws/aws-sdk-go from 1.47.5 to 1.47.7 ([#841](https://github.com/chanzuckerberg/aws-oidc/issues/841)) ([480c8b5](https://github.com/chanzuckerberg/aws-oidc/commit/480c8b5b9cfd126b41af5d89c7e4ebfcd91ff45e))
* bump github.com/aws/aws-sdk-go from 1.47.7 to 1.47.8 ([#842](https://github.com/chanzuckerberg/aws-oidc/issues/842)) ([93cbe1e](https://github.com/chanzuckerberg/aws-oidc/commit/93cbe1eee29f58ca1baaa058427cf6bf603eab7a))
* bump github.com/aws/aws-sdk-go from 1.47.8 to 1.47.9 ([#843](https://github.com/chanzuckerberg/aws-oidc/issues/843)) ([86c753b](https://github.com/chanzuckerberg/aws-oidc/commit/86c753b2f6c5a384255d4be97d072de906692990))
* bump github.com/aws/aws-sdk-go from 1.47.9 to 1.47.10 ([#844](https://github.com/chanzuckerberg/aws-oidc/issues/844)) ([e45f328](https://github.com/chanzuckerberg/aws-oidc/commit/e45f328a08b609b0b77922181d7e7cad5e50f32b))
* bump github.com/aws/aws-sdk-go from 1.48.0 to 1.48.1 ([#854](https://github.com/chanzuckerberg/aws-oidc/issues/854)) ([cf0964a](https://github.com/chanzuckerberg/aws-oidc/commit/cf0964a644939a67f410dcb6fa892439f48dd866))
* bump github.com/aws/aws-sdk-go from 1.48.1 to 1.48.2 ([#856](https://github.com/chanzuckerberg/aws-oidc/issues/856)) ([ba5511f](https://github.com/chanzuckerberg/aws-oidc/commit/ba5511f64ead2736f06ad267536e2bef98a9fe7d))
* bump github.com/aws/aws-sdk-go from 1.48.10 to 1.48.11 ([#865](https://github.com/chanzuckerberg/aws-oidc/issues/865)) ([e5cc8d2](https://github.com/chanzuckerberg/aws-oidc/commit/e5cc8d241d7102a26d93ae422f97d4d753939f9e))
* bump github.com/aws/aws-sdk-go from 1.48.11 to 1.48.12 ([#866](https://github.com/chanzuckerberg/aws-oidc/issues/866)) ([3c0bd58](https://github.com/chanzuckerberg/aws-oidc/commit/3c0bd58b28a18bc56f577383b0f1ffd67c4655b3))
* bump github.com/aws/aws-sdk-go from 1.48.12 to 1.48.13 ([#868](https://github.com/chanzuckerberg/aws-oidc/issues/868)) ([ca18cd4](https://github.com/chanzuckerberg/aws-oidc/commit/ca18cd478200b733d131e02754f4a524d7f34e28))
* bump github.com/aws/aws-sdk-go from 1.48.13 to 1.48.14 ([#869](https://github.com/chanzuckerberg/aws-oidc/issues/869)) ([639c237](https://github.com/chanzuckerberg/aws-oidc/commit/639c2376b179e792cf79f6bee157bb58108b45ac))
* bump github.com/aws/aws-sdk-go from 1.48.14 to 1.48.15 ([#871](https://github.com/chanzuckerberg/aws-oidc/issues/871)) ([4f3bb02](https://github.com/chanzuckerberg/aws-oidc/commit/4f3bb0224bfd6682a02be68677501f4fdc7c8345))
* bump github.com/aws/aws-sdk-go from 1.48.15 to 1.48.16 ([#873](https://github.com/chanzuckerberg/aws-oidc/issues/873)) ([7e80723](https://github.com/chanzuckerberg/aws-oidc/commit/7e807232ee31852903e69a055a79f58bb2f461e2))
* bump github.com/aws/aws-sdk-go from 1.48.16 to 1.49.0 ([#874](https://github.com/chanzuckerberg/aws-oidc/issues/874)) ([ab97c3e](https://github.com/chanzuckerberg/aws-oidc/commit/ab97c3e45090976e6c7c3e6adc240fd0309728ec))
* bump github.com/aws/aws-sdk-go from 1.48.2 to 1.48.3 ([#857](https://github.com/chanzuckerberg/aws-oidc/issues/857)) ([b67b92e](https://github.com/chanzuckerberg/aws-oidc/commit/b67b92e69cdd631fc33e53faf18f1d2d286a8e60))
* bump github.com/aws/aws-sdk-go from 1.48.3 to 1.48.4 ([#859](https://github.com/chanzuckerberg/aws-oidc/issues/859)) ([eefcae8](https://github.com/chanzuckerberg/aws-oidc/commit/eefcae844f1d9fe3b85912b5aab84c85ba4f1e84))
* bump github.com/aws/aws-sdk-go from 1.48.4 to 1.48.6 ([#861](https://github.com/chanzuckerberg/aws-oidc/issues/861)) ([06e109d](https://github.com/chanzuckerberg/aws-oidc/commit/06e109d7cc2ede975317f17597f4e4952f28f049))
* bump github.com/aws/aws-sdk-go from 1.48.6 to 1.48.7 ([#862](https://github.com/chanzuckerberg/aws-oidc/issues/862)) ([cee2679](https://github.com/chanzuckerberg/aws-oidc/commit/cee26793e3a39b077faa274ebb55d9da929f2230))
* bump github.com/aws/aws-sdk-go from 1.48.7 to 1.48.9 ([#863](https://github.com/chanzuckerberg/aws-oidc/issues/863)) ([43f9266](https://github.com/chanzuckerberg/aws-oidc/commit/43f92668a98c65e1e29f253bc182f39669454185))
* bump github.com/aws/aws-sdk-go from 1.48.9 to 1.48.10 ([#864](https://github.com/chanzuckerberg/aws-oidc/issues/864)) ([0ceff25](https://github.com/chanzuckerberg/aws-oidc/commit/0ceff25a512e959a2cbef5ecb47bd39bb10a4038))
* bump github.com/aws/aws-sdk-go from 1.49.0 to 1.49.1 ([#876](https://github.com/chanzuckerberg/aws-oidc/issues/876)) ([91cb6eb](https://github.com/chanzuckerberg/aws-oidc/commit/91cb6ebcd94161440b5556f20b5bdb162e8d0376))
* bump github.com/aws/aws-sdk-go from 1.49.1 to 1.49.2 ([#878](https://github.com/chanzuckerberg/aws-oidc/issues/878)) ([b0052f6](https://github.com/chanzuckerberg/aws-oidc/commit/b0052f6ee9b66e1ca48055858c1ca23709fe6795))
* bump github.com/aws/aws-sdk-go from 1.49.10 to 1.49.11 ([#890](https://github.com/chanzuckerberg/aws-oidc/issues/890)) ([98c43ff](https://github.com/chanzuckerberg/aws-oidc/commit/98c43ff435ddb54544ba4876f5bdc318808b0d56))
* bump github.com/aws/aws-sdk-go from 1.49.11 to 1.49.12 ([#891](https://github.com/chanzuckerberg/aws-oidc/issues/891)) ([091a8c8](https://github.com/chanzuckerberg/aws-oidc/commit/091a8c8ed5ad7c548f76974863dd990d69872604))
* bump github.com/aws/aws-sdk-go from 1.49.12 to 1.49.13 ([#892](https://github.com/chanzuckerberg/aws-oidc/issues/892)) ([322035b](https://github.com/chanzuckerberg/aws-oidc/commit/322035b565e18c99f0f5a45e0155aca5f520be29))
* bump github.com/aws/aws-sdk-go from 1.49.13 to 1.49.14 ([#893](https://github.com/chanzuckerberg/aws-oidc/issues/893)) ([5985a74](https://github.com/chanzuckerberg/aws-oidc/commit/5985a74c401e864e176eff30488fafc21fadc83b))
* bump github.com/aws/aws-sdk-go from 1.49.14 to 1.49.15 ([#895](https://github.com/chanzuckerberg/aws-oidc/issues/895)) ([689d82a](https://github.com/chanzuckerberg/aws-oidc/commit/689d82a56f68d0e87bb1ed7ebcade2fddd3ec4fd))
* bump github.com/aws/aws-sdk-go from 1.49.15 to 1.49.16 ([#897](https://github.com/chanzuckerberg/aws-oidc/issues/897)) ([1d4646f](https://github.com/chanzuckerberg/aws-oidc/commit/1d4646f2eaedf6c841091233eae057ca50a4ff66))
* bump github.com/aws/aws-sdk-go from 1.49.16 to 1.49.17 ([#899](https://github.com/chanzuckerberg/aws-oidc/issues/899)) ([2a5b7ac](https://github.com/chanzuckerberg/aws-oidc/commit/2a5b7acfe3af39a42eb5c54c16e0404f2064b49b))
* bump github.com/aws/aws-sdk-go from 1.49.17 to 1.49.18 ([#901](https://github.com/chanzuckerberg/aws-oidc/issues/901)) ([5bf9d1b](https://github.com/chanzuckerberg/aws-oidc/commit/5bf9d1b51269422f3c87acd59c787b136aaa3aa9))
* bump github.com/aws/aws-sdk-go from 1.49.18 to 1.49.19 ([#903](https://github.com/chanzuckerberg/aws-oidc/issues/903)) ([2e252ad](https://github.com/chanzuckerberg/aws-oidc/commit/2e252adda10054cdc1c43a411e2a03fbe8275a78))
* bump github.com/aws/aws-sdk-go from 1.49.19 to 1.49.21 ([#905](https://github.com/chanzuckerberg/aws-oidc/issues/905)) ([c56a88d](https://github.com/chanzuckerberg/aws-oidc/commit/c56a88d1e8251d2d79be326d7eb06988541698c3))
* bump github.com/aws/aws-sdk-go from 1.49.2 to 1.49.3 ([#879](https://github.com/chanzuckerberg/aws-oidc/issues/879)) ([2b10b7a](https://github.com/chanzuckerberg/aws-oidc/commit/2b10b7a1f5d9630ce4f1b8c846dcdf95bb581d02))
* bump github.com/aws/aws-sdk-go from 1.49.21 to 1.49.22 ([#907](https://github.com/chanzuckerberg/aws-oidc/issues/907)) ([6b5ac76](https://github.com/chanzuckerberg/aws-oidc/commit/6b5ac7624e12a55cefd96876fbd13f7a1cf804d8))
* bump github.com/aws/aws-sdk-go from 1.49.22 to 1.49.23 ([#909](https://github.com/chanzuckerberg/aws-oidc/issues/909)) ([f94e2e2](https://github.com/chanzuckerberg/aws-oidc/commit/f94e2e24f4e4999f5487de805f198669e159ad21))
* bump github.com/aws/aws-sdk-go from 1.49.23 to 1.49.24 ([#910](https://github.com/chanzuckerberg/aws-oidc/issues/910)) ([7f494b1](https://github.com/chanzuckerberg/aws-oidc/commit/7f494b158ffecff41a2d88e3d81419d3ea420d48))
* bump github.com/aws/aws-sdk-go from 1.49.24 to 1.50.0 ([#912](https://github.com/chanzuckerberg/aws-oidc/issues/912)) ([935c56d](https://github.com/chanzuckerberg/aws-oidc/commit/935c56db61e1782f0c7b22287fc458f42967df0f))
* bump github.com/aws/aws-sdk-go from 1.49.3 to 1.49.4 ([#881](https://github.com/chanzuckerberg/aws-oidc/issues/881)) ([6dbc93f](https://github.com/chanzuckerberg/aws-oidc/commit/6dbc93f558842d7bec7d27257a67a104e98a8318))
* bump github.com/aws/aws-sdk-go from 1.49.4 to 1.49.5 ([#884](https://github.com/chanzuckerberg/aws-oidc/issues/884)) ([4b09c80](https://github.com/chanzuckerberg/aws-oidc/commit/4b09c80cc53dcae1fc379e7693261b2153db080a))
* bump github.com/aws/aws-sdk-go from 1.49.5 to 1.49.6 ([#885](https://github.com/chanzuckerberg/aws-oidc/issues/885)) ([256cfc6](https://github.com/chanzuckerberg/aws-oidc/commit/256cfc61f6c6d339efb27483cb476057b1c4715e))
* bump github.com/aws/aws-sdk-go from 1.49.6 to 1.49.7 ([#886](https://github.com/chanzuckerberg/aws-oidc/issues/886)) ([c8407ea](https://github.com/chanzuckerberg/aws-oidc/commit/c8407ea4c9239cc722d874b94a829554790171c3))
* bump github.com/aws/aws-sdk-go from 1.49.7 to 1.49.8 ([#887](https://github.com/chanzuckerberg/aws-oidc/issues/887)) ([04a5fd9](https://github.com/chanzuckerberg/aws-oidc/commit/04a5fd950a059554f0c1c580849a49b527ac1df5))
* bump github.com/aws/aws-sdk-go from 1.49.8 to 1.49.9 ([#888](https://github.com/chanzuckerberg/aws-oidc/issues/888)) ([89f4603](https://github.com/chanzuckerberg/aws-oidc/commit/89f4603760a0bef0202ae944b9993fb1c77ad8d0))
* bump github.com/aws/aws-sdk-go from 1.49.9 to 1.49.10 ([#889](https://github.com/chanzuckerberg/aws-oidc/issues/889)) ([e22f8ff](https://github.com/chanzuckerberg/aws-oidc/commit/e22f8ffc56ff1250700e069df5c9510355b273db))
* bump github.com/aws/aws-sdk-go from 1.50.0 to 1.50.1 ([#914](https://github.com/chanzuckerberg/aws-oidc/issues/914)) ([6114518](https://github.com/chanzuckerberg/aws-oidc/commit/611451814d75dc000b50c74673bd4a646ef06b64))
* bump github.com/aws/aws-sdk-go from 1.50.1 to 1.50.2 ([#916](https://github.com/chanzuckerberg/aws-oidc/issues/916)) ([921aad7](https://github.com/chanzuckerberg/aws-oidc/commit/921aad71f9f06fec75474d530918dec1a27ce640))
* bump github.com/aws/aws-sdk-go from 1.50.10 to 1.50.11 ([#935](https://github.com/chanzuckerberg/aws-oidc/issues/935)) ([f63d759](https://github.com/chanzuckerberg/aws-oidc/commit/f63d759ca8761bbefcabe7a37a894a90131edce5))
* bump github.com/aws/aws-sdk-go from 1.50.11 to 1.50.12 ([#937](https://github.com/chanzuckerberg/aws-oidc/issues/937)) ([8352cf6](https://github.com/chanzuckerberg/aws-oidc/commit/8352cf63ef3df95af22f27223e2a3783be3e6428))
* bump github.com/aws/aws-sdk-go from 1.50.12 to 1.50.13 ([#938](https://github.com/chanzuckerberg/aws-oidc/issues/938)) ([27a8719](https://github.com/chanzuckerberg/aws-oidc/commit/27a87194090d41d85e2aa90b380a400df06afcca))
* bump github.com/aws/aws-sdk-go from 1.50.13 to 1.50.14 ([#939](https://github.com/chanzuckerberg/aws-oidc/issues/939)) ([21a4300](https://github.com/chanzuckerberg/aws-oidc/commit/21a4300f86b71ebec1dd7f9b09fa9dbc089f9f30))
* bump github.com/aws/aws-sdk-go from 1.50.14 to 1.50.15 ([#941](https://github.com/chanzuckerberg/aws-oidc/issues/941)) ([efe76ec](https://github.com/chanzuckerberg/aws-oidc/commit/efe76ec8406adcd4ea6642c304fca8ce46ad9f9f))
* bump github.com/aws/aws-sdk-go from 1.50.15 to 1.50.16 ([#943](https://github.com/chanzuckerberg/aws-oidc/issues/943)) ([e08f7e4](https://github.com/chanzuckerberg/aws-oidc/commit/e08f7e4f416e8a33c0c1070db5a6891aac03c210))
* bump github.com/aws/aws-sdk-go from 1.50.16 to 1.50.17 ([#945](https://github.com/chanzuckerberg/aws-oidc/issues/945)) ([bb9ef58](https://github.com/chanzuckerberg/aws-oidc/commit/bb9ef58f6951c49217eeee6835ce8b083fb676b3))
* bump github.com/aws/aws-sdk-go from 1.50.17 to 1.50.18 ([#946](https://github.com/chanzuckerberg/aws-oidc/issues/946)) ([7fe9048](https://github.com/chanzuckerberg/aws-oidc/commit/7fe90486dafe24b630c5f6fc17d8592d40ca93ad))
* bump github.com/aws/aws-sdk-go from 1.50.18 to 1.50.19 ([#947](https://github.com/chanzuckerberg/aws-oidc/issues/947)) ([e507540](https://github.com/chanzuckerberg/aws-oidc/commit/e5075405d89bdbf1d82bd0501e5b45ca548ed5b5))
* bump github.com/aws/aws-sdk-go from 1.50.19 to 1.50.20 ([#949](https://github.com/chanzuckerberg/aws-oidc/issues/949)) ([5e2b79a](https://github.com/chanzuckerberg/aws-oidc/commit/5e2b79a4888015b133c50287db7dacb54536c590))
* bump github.com/aws/aws-sdk-go from 1.50.2 to 1.50.3 ([#918](https://github.com/chanzuckerberg/aws-oidc/issues/918)) ([21df5df](https://github.com/chanzuckerberg/aws-oidc/commit/21df5dfdcbc132bea0f78072a2a6a889e2904309))
* bump github.com/aws/aws-sdk-go from 1.50.20 to 1.50.21 ([#950](https://github.com/chanzuckerberg/aws-oidc/issues/950)) ([a2b44ca](https://github.com/chanzuckerberg/aws-oidc/commit/a2b44ca4e2c04d78535f8c280335a12bda699d97))
* bump github.com/aws/aws-sdk-go from 1.50.21 to 1.50.22 ([#951](https://github.com/chanzuckerberg/aws-oidc/issues/951)) ([5144252](https://github.com/chanzuckerberg/aws-oidc/commit/5144252a7d7494bf09a1cca18ba4d1347765ed3e))
* bump github.com/aws/aws-sdk-go from 1.50.22 to 1.50.23 ([#952](https://github.com/chanzuckerberg/aws-oidc/issues/952)) ([29f1ff3](https://github.com/chanzuckerberg/aws-oidc/commit/29f1ff39f4e1db6235e25581ae36d9876a760f97))
* bump github.com/aws/aws-sdk-go from 1.50.23 to 1.50.24 ([#955](https://github.com/chanzuckerberg/aws-oidc/issues/955)) ([48ea93c](https://github.com/chanzuckerberg/aws-oidc/commit/48ea93cae1ce479a8c698c5d7b1e007da8056ec8))
* bump github.com/aws/aws-sdk-go from 1.50.24 to 1.50.25 ([#957](https://github.com/chanzuckerberg/aws-oidc/issues/957)) ([f7df1db](https://github.com/chanzuckerberg/aws-oidc/commit/f7df1db106788369a2496910ad586bf62fd78a7a))
* bump github.com/aws/aws-sdk-go from 1.50.25 to 1.50.26 ([#958](https://github.com/chanzuckerberg/aws-oidc/issues/958)) ([a200c5a](https://github.com/chanzuckerberg/aws-oidc/commit/a200c5a0095fbb9382cbfe33640d3f99d9a3f4df))
* bump github.com/aws/aws-sdk-go from 1.50.26 to 1.50.27 ([#960](https://github.com/chanzuckerberg/aws-oidc/issues/960)) ([e29bd00](https://github.com/chanzuckerberg/aws-oidc/commit/e29bd00d7ad2cbeba757b2f64e96df4aa6a28a64))
* bump github.com/aws/aws-sdk-go from 1.50.27 to 1.50.28 ([#962](https://github.com/chanzuckerberg/aws-oidc/issues/962)) ([19448b3](https://github.com/chanzuckerberg/aws-oidc/commit/19448b34174e5beb1d879ebaaf22b8cf81a31617))
* bump github.com/aws/aws-sdk-go from 1.50.28 to 1.50.29 ([#964](https://github.com/chanzuckerberg/aws-oidc/issues/964)) ([a83c28f](https://github.com/chanzuckerberg/aws-oidc/commit/a83c28fd90bee42bb6b9a09eac68963b146dc17e))
* bump github.com/aws/aws-sdk-go from 1.50.29 to 1.50.30 ([#966](https://github.com/chanzuckerberg/aws-oidc/issues/966)) ([09fb25a](https://github.com/chanzuckerberg/aws-oidc/commit/09fb25aebc16df77ca96dce9e5eedeccc9ba93a2))
* bump github.com/aws/aws-sdk-go from 1.50.3 to 1.50.4 ([#920](https://github.com/chanzuckerberg/aws-oidc/issues/920)) ([ceb7c5a](https://github.com/chanzuckerberg/aws-oidc/commit/ceb7c5af7a249e7e365f71cb5f556e1c880d5356))
* bump github.com/aws/aws-sdk-go from 1.50.30 to 1.50.31 ([#969](https://github.com/chanzuckerberg/aws-oidc/issues/969)) ([cf957f7](https://github.com/chanzuckerberg/aws-oidc/commit/cf957f715c158904595d6f1a0a850bdab56d767a))
* bump github.com/aws/aws-sdk-go from 1.50.31 to 1.50.32 ([#971](https://github.com/chanzuckerberg/aws-oidc/issues/971)) ([2a6272b](https://github.com/chanzuckerberg/aws-oidc/commit/2a6272bafbf18532956f8fff61b2660cefd5295a))
* bump github.com/aws/aws-sdk-go from 1.50.32 to 1.50.33 ([#973](https://github.com/chanzuckerberg/aws-oidc/issues/973)) ([da37c0c](https://github.com/chanzuckerberg/aws-oidc/commit/da37c0c7ed487766625a0770cd636e45a0438809))
* bump github.com/aws/aws-sdk-go from 1.50.33 to 1.50.34 ([#977](https://github.com/chanzuckerberg/aws-oidc/issues/977)) ([68b0365](https://github.com/chanzuckerberg/aws-oidc/commit/68b0365b5843f12ea6edae31065bb886bfa508da))
* bump github.com/aws/aws-sdk-go from 1.50.34 to 1.50.35 ([#978](https://github.com/chanzuckerberg/aws-oidc/issues/978)) ([9914363](https://github.com/chanzuckerberg/aws-oidc/commit/9914363ce14c7c7907776861e0e6675f58bb2da9))
* bump github.com/aws/aws-sdk-go from 1.50.35 to 1.50.36 ([#980](https://github.com/chanzuckerberg/aws-oidc/issues/980)) ([d34b29d](https://github.com/chanzuckerberg/aws-oidc/commit/d34b29dc91c5cee51b636617dc9fcb36d6207322))
* bump github.com/aws/aws-sdk-go from 1.50.36 to 1.50.37 ([#982](https://github.com/chanzuckerberg/aws-oidc/issues/982)) ([34e2007](https://github.com/chanzuckerberg/aws-oidc/commit/34e20073f12f5cc015f1d18d9e430bd1b94af128))
* bump github.com/aws/aws-sdk-go from 1.50.37 to 1.50.38 ([#986](https://github.com/chanzuckerberg/aws-oidc/issues/986)) ([97eba39](https://github.com/chanzuckerberg/aws-oidc/commit/97eba39a637a87aadadf608d5b160f0502b24091))
* bump github.com/aws/aws-sdk-go from 1.50.38 to 1.51.0 ([#987](https://github.com/chanzuckerberg/aws-oidc/issues/987)) ([96e07a3](https://github.com/chanzuckerberg/aws-oidc/commit/96e07a3a1e3108fd31f46861129d9d351df72dd1))
* bump github.com/aws/aws-sdk-go from 1.50.4 to 1.50.5 ([#922](https://github.com/chanzuckerberg/aws-oidc/issues/922)) ([ad80e1c](https://github.com/chanzuckerberg/aws-oidc/commit/ad80e1ce1bb5bf85e1002e63229518f98e269fb7))
* bump github.com/aws/aws-sdk-go from 1.50.5 to 1.50.6 ([#925](https://github.com/chanzuckerberg/aws-oidc/issues/925)) ([2fdca77](https://github.com/chanzuckerberg/aws-oidc/commit/2fdca7790429d22c0ad5cc5cab275866d4a61a47))
* bump github.com/aws/aws-sdk-go from 1.50.6 to 1.50.7 ([#927](https://github.com/chanzuckerberg/aws-oidc/issues/927)) ([0372c74](https://github.com/chanzuckerberg/aws-oidc/commit/0372c74a49c5dbcb096130003296ffcfa0617c8d))
* bump github.com/aws/aws-sdk-go from 1.50.7 to 1.50.8 ([#929](https://github.com/chanzuckerberg/aws-oidc/issues/929)) ([7f9651b](https://github.com/chanzuckerberg/aws-oidc/commit/7f9651b717d551a00791c666b45fe3b0069da918))
* bump github.com/aws/aws-sdk-go from 1.50.8 to 1.50.9 ([#931](https://github.com/chanzuckerberg/aws-oidc/issues/931)) ([fe23b97](https://github.com/chanzuckerberg/aws-oidc/commit/fe23b97e1c24a9a2c0852750d48e5de2b4279e84))
* bump github.com/aws/aws-sdk-go from 1.50.9 to 1.50.10 ([#933](https://github.com/chanzuckerberg/aws-oidc/issues/933)) ([172e9bd](https://github.com/chanzuckerberg/aws-oidc/commit/172e9bde8a4d8308734ff4b7accbf13994d1392d))
* bump github.com/aws/aws-sdk-go from 1.51.0 to 1.51.1 ([#989](https://github.com/chanzuckerberg/aws-oidc/issues/989)) ([93dc8e3](https://github.com/chanzuckerberg/aws-oidc/commit/93dc8e3f733fba8ff00396afd82697c0e0a52e39))
* bump github.com/aws/aws-sdk-go from 1.51.1 to 1.51.2 ([#991](https://github.com/chanzuckerberg/aws-oidc/issues/991)) ([3c1433e](https://github.com/chanzuckerberg/aws-oidc/commit/3c1433ee4fe564e8bfab111974892ac24bc2937b))
* bump github.com/aws/aws-sdk-go from 1.51.10 to 1.51.11 ([#1001](https://github.com/chanzuckerberg/aws-oidc/issues/1001)) ([0f41f3f](https://github.com/chanzuckerberg/aws-oidc/commit/0f41f3ff7f374a121ec3739e83c96f6c19bc56f5))
* bump github.com/aws/aws-sdk-go from 1.51.11 to 1.51.12 ([#1002](https://github.com/chanzuckerberg/aws-oidc/issues/1002)) ([1d96e85](https://github.com/chanzuckerberg/aws-oidc/commit/1d96e85185d98774f949c18dbf8b0a214af917b8))
* bump github.com/aws/aws-sdk-go from 1.51.12 to 1.51.13 ([#1003](https://github.com/chanzuckerberg/aws-oidc/issues/1003)) ([5fd283f](https://github.com/chanzuckerberg/aws-oidc/commit/5fd283f5018c207640f03fbe371d76a5c206e29e))
* bump github.com/aws/aws-sdk-go from 1.51.13 to 1.51.14 ([#1004](https://github.com/chanzuckerberg/aws-oidc/issues/1004)) ([d408e14](https://github.com/chanzuckerberg/aws-oidc/commit/d408e145748c322be229f4fe0c0c7b16bcf44ac1))
* bump github.com/aws/aws-sdk-go from 1.51.14 to 1.51.15 ([#1006](https://github.com/chanzuckerberg/aws-oidc/issues/1006)) ([2e21ae7](https://github.com/chanzuckerberg/aws-oidc/commit/2e21ae7caaa1625c93ddaa5cac5d9d6490d4aaaf))
* bump github.com/aws/aws-sdk-go from 1.51.15 to 1.51.16 ([#1008](https://github.com/chanzuckerberg/aws-oidc/issues/1008)) ([4eee46e](https://github.com/chanzuckerberg/aws-oidc/commit/4eee46e22911b6b13a3fb9e8b7b165e26544a652))
* bump github.com/aws/aws-sdk-go from 1.51.16 to 1.51.17 ([#1009](https://github.com/chanzuckerberg/aws-oidc/issues/1009)) ([5ae0f1f](https://github.com/chanzuckerberg/aws-oidc/commit/5ae0f1f5a4fab52ba53c5bf8a00514ef92f1e75a))
* bump github.com/aws/aws-sdk-go from 1.51.17 to 1.51.18 ([#1010](https://github.com/chanzuckerberg/aws-oidc/issues/1010)) ([db55dbf](https://github.com/chanzuckerberg/aws-oidc/commit/db55dbfd8993bf65b3dc26d8ebd6b8ffb9f2b66f))
* bump github.com/aws/aws-sdk-go from 1.51.18 to 1.51.19 ([#1011](https://github.com/chanzuckerberg/aws-oidc/issues/1011)) ([78cd8df](https://github.com/chanzuckerberg/aws-oidc/commit/78cd8dfbf34848c98ba9706b056545a528404599))
* bump github.com/aws/aws-sdk-go from 1.51.19 to 1.51.20 ([#1013](https://github.com/chanzuckerberg/aws-oidc/issues/1013)) ([3c96519](https://github.com/chanzuckerberg/aws-oidc/commit/3c965190239ec3c5c07707d3b9572150ba2eadcc))
* bump github.com/aws/aws-sdk-go from 1.51.2 to 1.51.3 ([#993](https://github.com/chanzuckerberg/aws-oidc/issues/993)) ([eea9e85](https://github.com/chanzuckerberg/aws-oidc/commit/eea9e851107d0a30cb899a9a89cfef3a4a4b1ed0))
* bump github.com/aws/aws-sdk-go from 1.51.20 to 1.51.21 ([#1014](https://github.com/chanzuckerberg/aws-oidc/issues/1014)) ([653f818](https://github.com/chanzuckerberg/aws-oidc/commit/653f818177d5c52c3dc81725dbb8a6df897d8b12))
* bump github.com/aws/aws-sdk-go from 1.51.21 to 1.51.22 ([#1017](https://github.com/chanzuckerberg/aws-oidc/issues/1017)) ([ed544da](https://github.com/chanzuckerberg/aws-oidc/commit/ed544da2802b7d457e891bf82016f98f02e6e3f2))
* bump github.com/aws/aws-sdk-go from 1.51.22 to 1.51.23 ([#1018](https://github.com/chanzuckerberg/aws-oidc/issues/1018)) ([3d84fe7](https://github.com/chanzuckerberg/aws-oidc/commit/3d84fe7974984fa35e5514355fdd2dd18b1c8eb6))
* bump github.com/aws/aws-sdk-go from 1.51.23 to 1.51.24 ([#1019](https://github.com/chanzuckerberg/aws-oidc/issues/1019)) ([04bde53](https://github.com/chanzuckerberg/aws-oidc/commit/04bde5360f800cb61a5aef0ec9fa090b934eba14))
* bump github.com/aws/aws-sdk-go from 1.51.24 to 1.51.25 ([#1020](https://github.com/chanzuckerberg/aws-oidc/issues/1020)) ([25e7582](https://github.com/chanzuckerberg/aws-oidc/commit/25e758224d10a38f59bcf3117ccdf06a0b1d1a3c))
* bump github.com/aws/aws-sdk-go from 1.51.25 to 1.51.26 ([#1021](https://github.com/chanzuckerberg/aws-oidc/issues/1021)) ([0b45947](https://github.com/chanzuckerberg/aws-oidc/commit/0b45947a37651a6ff33794d449df9cfbc79d90c0))
* bump github.com/aws/aws-sdk-go from 1.51.26 to 1.51.27 ([#1022](https://github.com/chanzuckerberg/aws-oidc/issues/1022)) ([5ac2b1d](https://github.com/chanzuckerberg/aws-oidc/commit/5ac2b1d48c299bfbae6b6956c73e41c224040277))
* bump github.com/aws/aws-sdk-go from 1.51.27 to 1.51.28 ([#1023](https://github.com/chanzuckerberg/aws-oidc/issues/1023)) ([d7f82e3](https://github.com/chanzuckerberg/aws-oidc/commit/d7f82e390e987ee2928a95bf5117f323e2c49774))
* bump github.com/aws/aws-sdk-go from 1.51.28 to 1.51.29 ([#1024](https://github.com/chanzuckerberg/aws-oidc/issues/1024)) ([c5f2f2a](https://github.com/chanzuckerberg/aws-oidc/commit/c5f2f2a016f7822197d0aa67ed1ac46559b68d14))
* bump github.com/aws/aws-sdk-go from 1.51.29 to 1.51.30 ([#1025](https://github.com/chanzuckerberg/aws-oidc/issues/1025)) ([5e53f3a](https://github.com/chanzuckerberg/aws-oidc/commit/5e53f3a1798040b692e664e807062ce8d8542fbb))
* bump github.com/aws/aws-sdk-go from 1.51.3 to 1.51.4 ([#994](https://github.com/chanzuckerberg/aws-oidc/issues/994)) ([85371e1](https://github.com/chanzuckerberg/aws-oidc/commit/85371e185234aae2de5944bdbc5398d6d5b2492e))
* bump github.com/aws/aws-sdk-go from 1.51.30 to 1.51.31 ([#1026](https://github.com/chanzuckerberg/aws-oidc/issues/1026)) ([671b446](https://github.com/chanzuckerberg/aws-oidc/commit/671b446366f83d07d620e23f60977c205bc79565))
* bump github.com/aws/aws-sdk-go from 1.51.31 to 1.51.32 ([#1027](https://github.com/chanzuckerberg/aws-oidc/issues/1027)) ([206753e](https://github.com/chanzuckerberg/aws-oidc/commit/206753eee246c548eb6b9c98ad7c37708f995ea9))
* bump github.com/aws/aws-sdk-go from 1.51.32 to 1.52.0 ([#1028](https://github.com/chanzuckerberg/aws-oidc/issues/1028)) ([cb5fb2e](https://github.com/chanzuckerberg/aws-oidc/commit/cb5fb2ed233d3dd84e624b5dfae14e8920e0980f))
* bump github.com/aws/aws-sdk-go from 1.51.4 to 1.51.5 ([#995](https://github.com/chanzuckerberg/aws-oidc/issues/995)) ([7d5c1e1](https://github.com/chanzuckerberg/aws-oidc/commit/7d5c1e11992a4b090c773a3786c21d2bd563783b))
* bump github.com/aws/aws-sdk-go from 1.51.5 to 1.51.6 ([#996](https://github.com/chanzuckerberg/aws-oidc/issues/996)) ([c0e2f7d](https://github.com/chanzuckerberg/aws-oidc/commit/c0e2f7d14475f65c9bf1c512547351013dcc9266))
* bump github.com/aws/aws-sdk-go from 1.51.6 to 1.51.7 ([#997](https://github.com/chanzuckerberg/aws-oidc/issues/997)) ([98cefab](https://github.com/chanzuckerberg/aws-oidc/commit/98cefab7e4611c5b1bdd72540f891826650f2a05))
* bump github.com/aws/aws-sdk-go from 1.51.7 to 1.51.8 ([#998](https://github.com/chanzuckerberg/aws-oidc/issues/998)) ([e9afcf1](https://github.com/chanzuckerberg/aws-oidc/commit/e9afcf16fdf5de2ed44a7aa46c667873e04d4689))
* bump github.com/aws/aws-sdk-go from 1.51.8 to 1.51.9 ([#999](https://github.com/chanzuckerberg/aws-oidc/issues/999)) ([9d314d2](https://github.com/chanzuckerberg/aws-oidc/commit/9d314d2d71849c189276c48482e03b58cb51729f))
* bump github.com/aws/aws-sdk-go from 1.51.9 to 1.51.10 ([#1000](https://github.com/chanzuckerberg/aws-oidc/issues/1000)) ([1ce9113](https://github.com/chanzuckerberg/aws-oidc/commit/1ce91130e5446b690f2d3373ede537673de00073))
* bump github.com/aws/aws-sdk-go from 1.52.0 to 1.52.1 ([#1029](https://github.com/chanzuckerberg/aws-oidc/issues/1029)) ([808680a](https://github.com/chanzuckerberg/aws-oidc/commit/808680a6fd4752b20ff9411befb6529d94352040))
* bump github.com/aws/aws-sdk-go from 1.52.1 to 1.52.2 ([#1030](https://github.com/chanzuckerberg/aws-oidc/issues/1030)) ([a13dade](https://github.com/chanzuckerberg/aws-oidc/commit/a13dade2800fdbd6def8af5c2e7cdb760cdad732))
* bump github.com/aws/aws-sdk-go from 1.52.2 to 1.52.3 ([#1031](https://github.com/chanzuckerberg/aws-oidc/issues/1031)) ([c0efd3b](https://github.com/chanzuckerberg/aws-oidc/commit/c0efd3bb1c19ed6398f8270a0687c113762650be))
* bump github.com/aws/aws-sdk-go from 1.52.3 to 1.52.4 ([#1032](https://github.com/chanzuckerberg/aws-oidc/issues/1032)) ([968f278](https://github.com/chanzuckerberg/aws-oidc/commit/968f27863d88405e9e222412773cf15c55ebf67d))
* bump github.com/aws/aws-sdk-go from 1.52.4 to 1.52.5 ([#1033](https://github.com/chanzuckerberg/aws-oidc/issues/1033)) ([af18501](https://github.com/chanzuckerberg/aws-oidc/commit/af18501f64a4cbfc8106efc564be72d412622ed1))
* bump github.com/aws/aws-sdk-go from 1.52.5 to 1.52.6 ([#1034](https://github.com/chanzuckerberg/aws-oidc/issues/1034)) ([19e7ebc](https://github.com/chanzuckerberg/aws-oidc/commit/19e7ebcebddb0cf5b3893793b485594a64596f10))
* bump github.com/aws/aws-sdk-go from 1.52.6 to 1.53.0 ([#1035](https://github.com/chanzuckerberg/aws-oidc/issues/1035)) ([e86c914](https://github.com/chanzuckerberg/aws-oidc/commit/e86c914d2fb69afb25ee3bcbdf9d12a9ee71f996))
* bump github.com/aws/aws-sdk-go from 1.53.0 to 1.53.2 ([#1036](https://github.com/chanzuckerberg/aws-oidc/issues/1036)) ([7385238](https://github.com/chanzuckerberg/aws-oidc/commit/7385238cc2cb979ab4c4bab3a5a0c4be7b11fbce))
* bump github.com/aws/aws-sdk-go from 1.53.10 to 1.53.11 ([#1045](https://github.com/chanzuckerberg/aws-oidc/issues/1045)) ([f522f0c](https://github.com/chanzuckerberg/aws-oidc/commit/f522f0c874f065e64ba4ce171b45ddb085f3edc9))
* bump github.com/aws/aws-sdk-go from 1.53.11 to 1.53.12 ([#1046](https://github.com/chanzuckerberg/aws-oidc/issues/1046)) ([e816386](https://github.com/chanzuckerberg/aws-oidc/commit/e816386ad085aa10f269a70ef2bd17d7ca648fc6))
* bump github.com/aws/aws-sdk-go from 1.53.12 to 1.53.13 ([#1047](https://github.com/chanzuckerberg/aws-oidc/issues/1047)) ([d1f72cd](https://github.com/chanzuckerberg/aws-oidc/commit/d1f72cd8af8b43a98259d3dd9cf19529928da690))
* bump github.com/aws/aws-sdk-go from 1.53.13 to 1.53.14 ([#1048](https://github.com/chanzuckerberg/aws-oidc/issues/1048)) ([fadead1](https://github.com/chanzuckerberg/aws-oidc/commit/fadead1e5dba279bb27022f042c2b0ef03102fef))
* bump github.com/aws/aws-sdk-go from 1.53.14 to 1.53.15 ([#1049](https://github.com/chanzuckerberg/aws-oidc/issues/1049)) ([0420556](https://github.com/chanzuckerberg/aws-oidc/commit/04205564f41cdef14bf2dc8799502b4e2762ccf2))
* bump github.com/aws/aws-sdk-go from 1.53.15 to 1.53.16 ([#1050](https://github.com/chanzuckerberg/aws-oidc/issues/1050)) ([118ad6d](https://github.com/chanzuckerberg/aws-oidc/commit/118ad6d70902a25bcfada600e066de91f4f70d03))
* bump github.com/aws/aws-sdk-go from 1.53.16 to 1.53.17 ([#1051](https://github.com/chanzuckerberg/aws-oidc/issues/1051)) ([ae5d5bb](https://github.com/chanzuckerberg/aws-oidc/commit/ae5d5bb568d13597cffcf31249aec7f77bf8860c))
* bump github.com/aws/aws-sdk-go from 1.53.17 to 1.53.18 ([#1052](https://github.com/chanzuckerberg/aws-oidc/issues/1052)) ([b96fba9](https://github.com/chanzuckerberg/aws-oidc/commit/b96fba9f23309ada2a1d9a90c9f333712ece76ee))
* bump github.com/aws/aws-sdk-go from 1.53.18 to 1.53.19 ([#1053](https://github.com/chanzuckerberg/aws-oidc/issues/1053)) ([8a9524b](https://github.com/chanzuckerberg/aws-oidc/commit/8a9524bd05043dd54e0c49d3975be8056bb4a3d9))
* bump github.com/aws/aws-sdk-go from 1.53.19 to 1.53.20 ([#1054](https://github.com/chanzuckerberg/aws-oidc/issues/1054)) ([7e96f11](https://github.com/chanzuckerberg/aws-oidc/commit/7e96f11ed454f5e1c811f3c1fb50465576b4d5a2))
* bump github.com/aws/aws-sdk-go from 1.53.2 to 1.53.3 ([#1037](https://github.com/chanzuckerberg/aws-oidc/issues/1037)) ([91435e0](https://github.com/chanzuckerberg/aws-oidc/commit/91435e04290217184a68ca99cfb1363c39c7e862))
* bump github.com/aws/aws-sdk-go from 1.53.20 to 1.53.21 ([#1055](https://github.com/chanzuckerberg/aws-oidc/issues/1055)) ([d221c1d](https://github.com/chanzuckerberg/aws-oidc/commit/d221c1d5793590f0fed08842c31836afff37a577))
* bump github.com/aws/aws-sdk-go from 1.53.21 to 1.54.0 ([#1056](https://github.com/chanzuckerberg/aws-oidc/issues/1056)) ([4fffa51](https://github.com/chanzuckerberg/aws-oidc/commit/4fffa513cf0266f09b8ba9cf16a98dcc09d7ab69))
* bump github.com/aws/aws-sdk-go from 1.53.3 to 1.53.4 ([#1038](https://github.com/chanzuckerberg/aws-oidc/issues/1038)) ([132dc72](https://github.com/chanzuckerberg/aws-oidc/commit/132dc722c1256d3814b88484589f3765d33e6167))
* bump github.com/aws/aws-sdk-go from 1.53.4 to 1.53.5 ([#1039](https://github.com/chanzuckerberg/aws-oidc/issues/1039)) ([fbb1594](https://github.com/chanzuckerberg/aws-oidc/commit/fbb1594b1d1c62aae5df40cd4e9d58d938ab7eff))
* bump github.com/aws/aws-sdk-go from 1.53.5 to 1.53.6 ([#1040](https://github.com/chanzuckerberg/aws-oidc/issues/1040)) ([b84994f](https://github.com/chanzuckerberg/aws-oidc/commit/b84994f6a0b60999330350d76c4ce08c167f5557))
* bump github.com/aws/aws-sdk-go from 1.53.6 to 1.53.7 ([#1041](https://github.com/chanzuckerberg/aws-oidc/issues/1041)) ([73151f5](https://github.com/chanzuckerberg/aws-oidc/commit/73151f5a0df4ba619e81833e53f54f88c01e22fc))
* bump github.com/aws/aws-sdk-go from 1.53.7 to 1.53.8 ([#1042](https://github.com/chanzuckerberg/aws-oidc/issues/1042)) ([31c6a66](https://github.com/chanzuckerberg/aws-oidc/commit/31c6a66969fe8128d4fd9f8944cd31b3dcf49bee))
* bump github.com/aws/aws-sdk-go from 1.53.8 to 1.53.9 ([#1043](https://github.com/chanzuckerberg/aws-oidc/issues/1043)) ([ce85e7e](https://github.com/chanzuckerberg/aws-oidc/commit/ce85e7ee44831eb3abee2fdafd0b9723d66bed19))
* bump github.com/aws/aws-sdk-go from 1.53.9 to 1.53.10 ([#1044](https://github.com/chanzuckerberg/aws-oidc/issues/1044)) ([0bc1c2b](https://github.com/chanzuckerberg/aws-oidc/commit/0bc1c2bd84263b294c8a86729c9b9b6f499a87ac))
* bump github.com/aws/aws-sdk-go from 1.54.0 to 1.54.1 ([#1058](https://github.com/chanzuckerberg/aws-oidc/issues/1058)) ([c402fa3](https://github.com/chanzuckerberg/aws-oidc/commit/c402fa3e1636deaa11f7b0c72ae26f53d5cd844f))
* bump github.com/aws/aws-sdk-go from 1.54.1 to 1.54.2 ([#1060](https://github.com/chanzuckerberg/aws-oidc/issues/1060)) ([7fe8d69](https://github.com/chanzuckerberg/aws-oidc/commit/7fe8d691552ed6fc09e83490774d9c549a177e05))
* bump github.com/aws/aws-sdk-go from 1.54.10 to 1.54.11 ([#1069](https://github.com/chanzuckerberg/aws-oidc/issues/1069)) ([b537d63](https://github.com/chanzuckerberg/aws-oidc/commit/b537d63f30743a076dadd904b2640643bdfb5a5c))
* bump github.com/aws/aws-sdk-go from 1.54.11 to 1.54.12 ([#1070](https://github.com/chanzuckerberg/aws-oidc/issues/1070)) ([84b23e3](https://github.com/chanzuckerberg/aws-oidc/commit/84b23e3728cc026294e8d15e4d66ea036332021b))
* bump github.com/aws/aws-sdk-go from 1.54.12 to 1.54.13 ([#1071](https://github.com/chanzuckerberg/aws-oidc/issues/1071)) ([731057d](https://github.com/chanzuckerberg/aws-oidc/commit/731057ddcdfafb5edcc90e354f38d5b0e95b1dc5))
* bump github.com/aws/aws-sdk-go from 1.54.13 to 1.54.14 ([#1072](https://github.com/chanzuckerberg/aws-oidc/issues/1072)) ([614b456](https://github.com/chanzuckerberg/aws-oidc/commit/614b456f98345fe72ee6be164f365f84604e15d0))
* bump github.com/aws/aws-sdk-go from 1.54.14 to 1.54.15 ([#1073](https://github.com/chanzuckerberg/aws-oidc/issues/1073)) ([faf019b](https://github.com/chanzuckerberg/aws-oidc/commit/faf019bb9029e9dfb2cf1548e94fee3393d333dd))
* bump github.com/aws/aws-sdk-go from 1.54.15 to 1.54.16 ([#1074](https://github.com/chanzuckerberg/aws-oidc/issues/1074)) ([84508b8](https://github.com/chanzuckerberg/aws-oidc/commit/84508b89f10c448baa395f6d77824b048b6ac05d))
* bump github.com/aws/aws-sdk-go from 1.54.16 to 1.54.17 ([#1075](https://github.com/chanzuckerberg/aws-oidc/issues/1075)) ([fd61abd](https://github.com/chanzuckerberg/aws-oidc/commit/fd61abd866817f992385ebb00b3a7efb4d0c9338))
* bump github.com/aws/aws-sdk-go from 1.54.17 to 1.54.18 ([#1076](https://github.com/chanzuckerberg/aws-oidc/issues/1076)) ([d25603f](https://github.com/chanzuckerberg/aws-oidc/commit/d25603f99e6fa5431aa7e6175db8c8677447442b))
* bump github.com/aws/aws-sdk-go from 1.54.18 to 1.54.19 ([#1077](https://github.com/chanzuckerberg/aws-oidc/issues/1077)) ([f4c23c6](https://github.com/chanzuckerberg/aws-oidc/commit/f4c23c6aa2b630f3a226157fd385c1ac15f55b7e))
* bump github.com/aws/aws-sdk-go from 1.54.19 to 1.54.20 ([#1078](https://github.com/chanzuckerberg/aws-oidc/issues/1078)) ([eedff16](https://github.com/chanzuckerberg/aws-oidc/commit/eedff167bc138c6c453564c8062472b1e428cd4e))
* bump github.com/aws/aws-sdk-go from 1.54.2 to 1.54.3 ([#1061](https://github.com/chanzuckerberg/aws-oidc/issues/1061)) ([845cfad](https://github.com/chanzuckerberg/aws-oidc/commit/845cfad007cdfa9d498d85ebbd6d55a75f1fb78c))
* bump github.com/aws/aws-sdk-go from 1.54.20 to 1.55.0 ([#1079](https://github.com/chanzuckerberg/aws-oidc/issues/1079)) ([c8d38eb](https://github.com/chanzuckerberg/aws-oidc/commit/c8d38ebd2e0339b050a4dc24bea042ada4ae8915))
* bump github.com/aws/aws-sdk-go from 1.54.3 to 1.54.4 ([#1062](https://github.com/chanzuckerberg/aws-oidc/issues/1062)) ([b4a87d3](https://github.com/chanzuckerberg/aws-oidc/commit/b4a87d39ad79cfd36c0a3b81e6855e03d561fc80))
* bump github.com/aws/aws-sdk-go from 1.54.4 to 1.54.5 ([#1063](https://github.com/chanzuckerberg/aws-oidc/issues/1063)) ([73641aa](https://github.com/chanzuckerberg/aws-oidc/commit/73641aab1d95dcd57d16353130c1ec00cbc34173))
* bump github.com/aws/aws-sdk-go from 1.54.5 to 1.54.6 ([#1064](https://github.com/chanzuckerberg/aws-oidc/issues/1064)) ([799ee14](https://github.com/chanzuckerberg/aws-oidc/commit/799ee1498ec23179b916b75d0199049384400abb))
* bump github.com/aws/aws-sdk-go from 1.54.6 to 1.54.7 ([#1065](https://github.com/chanzuckerberg/aws-oidc/issues/1065)) ([461403e](https://github.com/chanzuckerberg/aws-oidc/commit/461403e4c457334767924dd7349efd81aa5a6f4c))
* bump github.com/aws/aws-sdk-go from 1.54.7 to 1.54.8 ([#1066](https://github.com/chanzuckerberg/aws-oidc/issues/1066)) ([0d3aa55](https://github.com/chanzuckerberg/aws-oidc/commit/0d3aa55d88b402d8e22323e26efc267e129cb74d))
* bump github.com/aws/aws-sdk-go from 1.54.8 to 1.54.9 ([#1067](https://github.com/chanzuckerberg/aws-oidc/issues/1067)) ([168af5b](https://github.com/chanzuckerberg/aws-oidc/commit/168af5b81c52681770b7a6e0dc5112f59f92e4d3))
* bump github.com/aws/aws-sdk-go from 1.54.9 to 1.54.10 ([#1068](https://github.com/chanzuckerberg/aws-oidc/issues/1068)) ([956adbe](https://github.com/chanzuckerberg/aws-oidc/commit/956adbe37034e8139ffab95b446b908ddf7c0a6d))
* bump github.com/aws/aws-sdk-go from 1.55.0 to 1.55.1 ([#1080](https://github.com/chanzuckerberg/aws-oidc/issues/1080)) ([d95d52a](https://github.com/chanzuckerberg/aws-oidc/commit/d95d52a5939e166d18a41a374f3620b3d041b9c0))
* bump github.com/aws/aws-sdk-go from 1.55.1 to 1.55.2 ([#1081](https://github.com/chanzuckerberg/aws-oidc/issues/1081)) ([71bb88d](https://github.com/chanzuckerberg/aws-oidc/commit/71bb88dc29bf9e6f98d773dc3f023d0e27c8ad32))
* bump github.com/aws/aws-sdk-go from 1.55.2 to 1.55.3 ([#1082](https://github.com/chanzuckerberg/aws-oidc/issues/1082)) ([a949f79](https://github.com/chanzuckerberg/aws-oidc/commit/a949f7947b16cc00c33853188031a4d9b727d4ce))
* bump github.com/aws/aws-sdk-go from 1.55.3 to 1.55.4 ([#1084](https://github.com/chanzuckerberg/aws-oidc/issues/1084)) ([05bc6b7](https://github.com/chanzuckerberg/aws-oidc/commit/05bc6b7420c3a5acf1c8f89dfad39e1fc3c35bd5))
* bump github.com/aws/aws-sdk-go from 1.55.4 to 1.55.5 ([#1085](https://github.com/chanzuckerberg/aws-oidc/issues/1085)) ([c304a56](https://github.com/chanzuckerberg/aws-oidc/commit/c304a56dbc2de01aee199d9d805232bbeee91529))
* bump github.com/aws/aws-sdk-go from 1.55.5 to 1.55.6 ([#1090](https://github.com/chanzuckerberg/aws-oidc/issues/1090)) ([0d1e46a](https://github.com/chanzuckerberg/aws-oidc/commit/0d1e46a7385315cfae208ab160f282bc66207c90))
* bump github.com/aws/aws-sdk-go from 1.55.6 to 1.55.7 ([#1106](https://github.com/chanzuckerberg/aws-oidc/issues/1106)) ([eebd534](https://github.com/chanzuckerberg/aws-oidc/commit/eebd534f5b3ff0d50c0c76dc65243d03599f90fe))
* bump github.com/chanzuckerberg/go-misc ([fbfdf8b](https://github.com/chanzuckerberg/aws-oidc/commit/fbfdf8b5652a6a0f2ec9fddf6df0408bc98ae4d6))
* bump github.com/chanzuckerberg/go-misc from 1.0.6 to 1.0.7 ([#602](https://github.com/chanzuckerberg/aws-oidc/issues/602)) ([eed9066](https://github.com/chanzuckerberg/aws-oidc/commit/eed906695de2e83b57752d782a193bf272e23151))
* bump github.com/chanzuckerberg/go-misc from 1.0.7 to 1.0.8 ([#617](https://github.com/chanzuckerberg/aws-oidc/issues/617)) ([c3410d6](https://github.com/chanzuckerberg/aws-oidc/commit/c3410d6fd9e495d4a43600ae36246926e2c6844c))
* bump github.com/chanzuckerberg/go-misc from 1.0.8 to 1.0.9 ([#643](https://github.com/chanzuckerberg/aws-oidc/issues/643)) ([9281066](https://github.com/chanzuckerberg/aws-oidc/commit/928106610848feb52afeea225d486ef347f44902))
* bump github.com/chanzuckerberg/go-misc from 1.0.9 to 1.10.0 ([#663](https://github.com/chanzuckerberg/aws-oidc/issues/663)) ([290deb1](https://github.com/chanzuckerberg/aws-oidc/commit/290deb1dabbba15b4d68b65cf9bf68516546b35d))
* bump github.com/chanzuckerberg/go-misc from 1.10.0 to 1.10.1 ([#684](https://github.com/chanzuckerberg/aws-oidc/issues/684)) ([753177e](https://github.com/chanzuckerberg/aws-oidc/commit/753177e6fbd05af18047fc8095e5c2c0c029ee8a))
* bump github.com/chanzuckerberg/go-misc from 1.10.1 to 1.10.2 ([#698](https://github.com/chanzuckerberg/aws-oidc/issues/698)) ([606d3e4](https://github.com/chanzuckerberg/aws-oidc/commit/606d3e4f2d44bc9c129d39abe607e6856aead03c))
* bump github.com/chanzuckerberg/go-misc from 1.10.10 to 1.10.11 ([#831](https://github.com/chanzuckerberg/aws-oidc/issues/831)) ([8172fb2](https://github.com/chanzuckerberg/aws-oidc/commit/8172fb24860d7e6036358732289fefe581e59db7))
* bump github.com/chanzuckerberg/go-misc from 1.10.11 to 1.10.12 ([#845](https://github.com/chanzuckerberg/aws-oidc/issues/845)) ([fd9b8c9](https://github.com/chanzuckerberg/aws-oidc/commit/fd9b8c921170e121c2387f07a7934f3f4fdab9ac))
* bump github.com/chanzuckerberg/go-misc from 1.10.12 to 1.10.13 ([#848](https://github.com/chanzuckerberg/aws-oidc/issues/848)) ([b9a98ab](https://github.com/chanzuckerberg/aws-oidc/commit/b9a98abb187ac3c757cb751227258dd2041826bd))
* bump github.com/chanzuckerberg/go-misc from 1.10.13 to 1.10.14 ([#853](https://github.com/chanzuckerberg/aws-oidc/issues/853)) ([221ebc7](https://github.com/chanzuckerberg/aws-oidc/commit/221ebc7eb8dd48fff66abe1918a5770737628d9f))
* bump github.com/chanzuckerberg/go-misc from 1.10.14 to 1.11.0 ([#870](https://github.com/chanzuckerberg/aws-oidc/issues/870)) ([6440fb4](https://github.com/chanzuckerberg/aws-oidc/commit/6440fb42d8e3764e8db95ba2ead3b270de4ef9ed))
* bump github.com/chanzuckerberg/go-misc from 1.10.2 to 1.10.3 ([#726](https://github.com/chanzuckerberg/aws-oidc/issues/726)) ([3511cbf](https://github.com/chanzuckerberg/aws-oidc/commit/3511cbf799001db1a204cd01b7fe7a026a47385d))
* bump github.com/chanzuckerberg/go-misc from 1.10.3 to 1.10.4 ([#728](https://github.com/chanzuckerberg/aws-oidc/issues/728)) ([f13d2a2](https://github.com/chanzuckerberg/aws-oidc/commit/f13d2a20726e412c076772ce80f24916234651f5))
* bump github.com/chanzuckerberg/go-misc from 1.10.4 to 1.10.5 ([#750](https://github.com/chanzuckerberg/aws-oidc/issues/750)) ([fd27318](https://github.com/chanzuckerberg/aws-oidc/commit/fd2731845a228ebb8c121428201fc28893f18a14))
* bump github.com/chanzuckerberg/go-misc from 1.10.5 to 1.10.6 ([#784](https://github.com/chanzuckerberg/aws-oidc/issues/784)) ([904cf8b](https://github.com/chanzuckerberg/aws-oidc/commit/904cf8b8125ec04b17a0de19ffd54a1ca9b22871))
* bump github.com/chanzuckerberg/go-misc from 1.10.6 to 1.10.7 ([#786](https://github.com/chanzuckerberg/aws-oidc/issues/786)) ([604db98](https://github.com/chanzuckerberg/aws-oidc/commit/604db984b551dcb4cd4f444295a1f54532ee0927))
* bump github.com/chanzuckerberg/go-misc from 1.10.7 to 1.10.8 ([#811](https://github.com/chanzuckerberg/aws-oidc/issues/811)) ([3bebe14](https://github.com/chanzuckerberg/aws-oidc/commit/3bebe141857c2edfb4fecd2fae6bdf92df2aa0fc))
* bump github.com/chanzuckerberg/go-misc from 1.10.8 to 1.10.9 ([#820](https://github.com/chanzuckerberg/aws-oidc/issues/820)) ([2ba886a](https://github.com/chanzuckerberg/aws-oidc/commit/2ba886a9c2fd3aa4197352c7546abbda5b9a8c54))
* bump github.com/chanzuckerberg/go-misc from 1.10.9 to 1.10.10 ([#827](https://github.com/chanzuckerberg/aws-oidc/issues/827)) ([3d2d7f1](https://github.com/chanzuckerberg/aws-oidc/commit/3d2d7f1a2bc3cd579bd79d924adaee2b6a4c4dfb))
* bump github.com/chanzuckerberg/go-misc from 1.11.0 to 1.11.1 ([#923](https://github.com/chanzuckerberg/aws-oidc/issues/923)) ([be53f71](https://github.com/chanzuckerberg/aws-oidc/commit/be53f7184c8b99e01e7f4eea5bd40e19b9f4c871))
* bump github.com/chanzuckerberg/go-misc from 1.11.1 to 1.12.0 ([#953](https://github.com/chanzuckerberg/aws-oidc/issues/953)) ([b6c790d](https://github.com/chanzuckerberg/aws-oidc/commit/b6c790d656dc49263450320146430356e8a3be92))
* bump github.com/chanzuckerberg/go-misc from 1.12.0 to 2.2.0+incompatible ([#1005](https://github.com/chanzuckerberg/aws-oidc/issues/1005)) ([fbfdf8b](https://github.com/chanzuckerberg/aws-oidc/commit/fbfdf8b5652a6a0f2ec9fddf6df0408bc98ae4d6))
* bump github.com/coreos/go-oidc ([00c4603](https://github.com/chanzuckerberg/aws-oidc/commit/00c4603bc78268c9af4949f30d1569c6d126b1dd))
* bump github.com/coreos/go-oidc from 2.2.1+incompatible to 2.3.0+incompatible ([#1093](https://github.com/chanzuckerberg/aws-oidc/issues/1093)) ([00c4603](https://github.com/chanzuckerberg/aws-oidc/commit/00c4603bc78268c9af4949f30d1569c6d126b1dd))
* bump github.com/go-errors/errors from 1.4.2 to 1.5.0 ([#779](https://github.com/chanzuckerberg/aws-oidc/issues/779)) ([e0c0e0e](https://github.com/chanzuckerberg/aws-oidc/commit/e0c0e0e824dbb9c52f04a49ee7aff0437d148c34))
* bump github.com/go-errors/errors from 1.5.0 to 1.5.1 ([#789](https://github.com/chanzuckerberg/aws-oidc/issues/789)) ([adcbe6c](https://github.com/chanzuckerberg/aws-oidc/commit/adcbe6ce2cabb2d35d9a3ce59169fa3c3c8d1f87))
* bump github.com/go-jose/go-jose/v3 from 3.0.0 to 3.0.1 ([#855](https://github.com/chanzuckerberg/aws-oidc/issues/855)) ([ea36f17](https://github.com/chanzuckerberg/aws-oidc/commit/ea36f17ca5ca21bc0cf8d6f06123510d8d98987c))
* bump github.com/go-jose/go-jose/v3 from 3.0.1 to 3.0.3 ([#975](https://github.com/chanzuckerberg/aws-oidc/issues/975)) ([bbdb8c6](https://github.com/chanzuckerberg/aws-oidc/commit/bbdb8c60c03bdb10d70df1f98b65fffa40d83e58))
* bump github.com/go-jose/go-jose/v3 from 3.0.3 to 3.0.4 ([#1099](https://github.com/chanzuckerberg/aws-oidc/issues/1099)) ([123c542](https://github.com/chanzuckerberg/aws-oidc/commit/123c54281275ced928d772db81d5360a3b23c076))
* bump github.com/gorilla/handlers from 1.5.1 to 1.5.2 ([#835](https://github.com/chanzuckerberg/aws-oidc/issues/835)) ([93c3aab](https://github.com/chanzuckerberg/aws-oidc/commit/93c3aab1b8109c6f5976f7aa99973a97de18482e))
* bump github.com/honeycombio/beeline-go from 1.11.1 to 1.12.0 ([#677](https://github.com/chanzuckerberg/aws-oidc/issues/677)) ([b9fc7a2](https://github.com/chanzuckerberg/aws-oidc/commit/b9fc7a258278de99b813bdb5638bea024bdd2d1e))
* bump github.com/honeycombio/beeline-go from 1.12.0 to 1.13.0 ([#729](https://github.com/chanzuckerberg/aws-oidc/issues/729)) ([0170e2f](https://github.com/chanzuckerberg/aws-oidc/commit/0170e2f26e589cff5b6c6550bbf3632280c8bd94))
* bump github.com/honeycombio/beeline-go from 1.13.0 to 1.14.0 ([#867](https://github.com/chanzuckerberg/aws-oidc/issues/867)) ([7e8bcab](https://github.com/chanzuckerberg/aws-oidc/commit/7e8bcabb4ddbe4d22f78efd2cd68a2f3d42736c4))
* bump github.com/honeycombio/beeline-go from 1.14.0 to 1.15.0 ([#972](https://github.com/chanzuckerberg/aws-oidc/issues/972)) ([143998e](https://github.com/chanzuckerberg/aws-oidc/commit/143998e439ac42f94068d6d55403d664680d3583))
* bump github.com/honeycombio/beeline-go from 1.15.0 to 1.16.0 ([#1007](https://github.com/chanzuckerberg/aws-oidc/issues/1007)) ([fcd6511](https://github.com/chanzuckerberg/aws-oidc/commit/fcd65113fadae9e3464ba2c5dc78186c29ff9359))
* bump github.com/honeycombio/beeline-go from 1.16.0 to 1.17.0 ([#1057](https://github.com/chanzuckerberg/aws-oidc/issues/1057)) ([589afa8](https://github.com/chanzuckerberg/aws-oidc/commit/589afa803ab74693850943b47d1698190e763238))
* bump github.com/honeycombio/beeline-go from 1.17.0 to 1.18.0 ([#1087](https://github.com/chanzuckerberg/aws-oidc/issues/1087)) ([14e0c23](https://github.com/chanzuckerberg/aws-oidc/commit/14e0c2335ac479367b7353bb258c59eea76c9874))
* bump github.com/honeycombio/beeline-go from 1.18.0 to 1.19.0 ([#1103](https://github.com/chanzuckerberg/aws-oidc/issues/1103)) ([100be18](https://github.com/chanzuckerberg/aws-oidc/commit/100be18a49f9c474d115a9b1ae03302cc84a510d))
* bump github.com/okta/okta-sdk-golang/v2 from 2.17.0 to 2.18.0 ([#624](https://github.com/chanzuckerberg/aws-oidc/issues/624)) ([f26a56a](https://github.com/chanzuckerberg/aws-oidc/commit/f26a56aa3a59e80b4d2dcca8e3de28eb0fb73f1d))
* bump github.com/okta/okta-sdk-golang/v2 from 2.18.0 to 2.19.0 ([#672](https://github.com/chanzuckerberg/aws-oidc/issues/672)) ([8a684a7](https://github.com/chanzuckerberg/aws-oidc/commit/8a684a7800c2e232075123171c0f77ce7ddce724))
* bump github.com/okta/okta-sdk-golang/v2 from 2.19.0 to 2.20.0 ([#711](https://github.com/chanzuckerberg/aws-oidc/issues/711)) ([c9a8f92](https://github.com/chanzuckerberg/aws-oidc/commit/c9a8f926794790c354f1e30faed3c29e1d46520a))
* bump github.com/sirupsen/logrus from 1.9.0 to 1.9.1 ([#653](https://github.com/chanzuckerberg/aws-oidc/issues/653)) ([fd036db](https://github.com/chanzuckerberg/aws-oidc/commit/fd036db7640b8b1b235dd968ac1d9ea2ba5a7057))
* bump github.com/sirupsen/logrus from 1.9.1 to 1.9.2 ([#656](https://github.com/chanzuckerberg/aws-oidc/issues/656)) ([0dcec7e](https://github.com/chanzuckerberg/aws-oidc/commit/0dcec7e87a8197260507de5a129eec4297749366))
* bump github.com/sirupsen/logrus from 1.9.2 to 1.9.3 ([#675](https://github.com/chanzuckerberg/aws-oidc/issues/675)) ([61c2214](https://github.com/chanzuckerberg/aws-oidc/commit/61c22144301a27d7b8aca04643978bcaae032daa))
* bump github.com/spf13/cobra from 1.6.1 to 1.7.0 ([#609](https://github.com/chanzuckerberg/aws-oidc/issues/609)) ([1b71c91](https://github.com/chanzuckerberg/aws-oidc/commit/1b71c91f2dcf33ae4973b25e232944c53a955ffb))
* bump github.com/spf13/cobra from 1.7.0 to 1.8.0 ([#836](https://github.com/chanzuckerberg/aws-oidc/issues/836)) ([5d70851](https://github.com/chanzuckerberg/aws-oidc/commit/5d70851c54473b0dbbb187d21fb5b9be3caad0fb))
* bump github.com/spf13/cobra from 1.8.0 to 1.8.1 ([#1059](https://github.com/chanzuckerberg/aws-oidc/issues/1059)) ([a2a9269](https://github.com/chanzuckerberg/aws-oidc/commit/a2a926941df23fa58b8dbadd1f411dd106e0e626))
* bump github.com/spf13/cobra from 1.8.1 to 1.9.1 ([#1097](https://github.com/chanzuckerberg/aws-oidc/issues/1097)) ([c792200](https://github.com/chanzuckerberg/aws-oidc/commit/c7922009fe2ae3f854f0103f5f749d04c65a01ec))
* bump github.com/stretchr/testify from 1.8.2 to 1.8.3 ([#658](https://github.com/chanzuckerberg/aws-oidc/issues/658)) ([f46d836](https://github.com/chanzuckerberg/aws-oidc/commit/f46d83665d42e484e44fc3d286ba74556e783221))
* bump github.com/stretchr/testify from 1.8.3 to 1.8.4 ([#666](https://github.com/chanzuckerberg/aws-oidc/issues/666)) ([97c801c](https://github.com/chanzuckerberg/aws-oidc/commit/97c801cacca3af04c782b1dd43f81a4a8a351eaf))
* bump github.com/stretchr/testify from 1.8.4 to 1.9.0 ([#967](https://github.com/chanzuckerberg/aws-oidc/issues/967)) ([e1ef077](https://github.com/chanzuckerberg/aws-oidc/commit/e1ef077a38c15a2a9addba631c8f155b55e3f5b0))
* bump github.com/stretchr/testify from 1.9.0 to 1.10.0 ([#1088](https://github.com/chanzuckerberg/aws-oidc/issues/1088)) ([a6f3e87](https://github.com/chanzuckerberg/aws-oidc/commit/a6f3e8741782e7b2047c82d27080b15511125f47))
* bump golang.org/x/crypto from 0.16.0 to 0.17.0 ([#883](https://github.com/chanzuckerberg/aws-oidc/issues/883)) ([4e905fa](https://github.com/chanzuckerberg/aws-oidc/commit/4e905fade503909ae193d500b5bd9e78b7e42240))
* bump golang.org/x/crypto from 0.25.0 to 0.31.0 ([#1091](https://github.com/chanzuckerberg/aws-oidc/issues/1091)) ([ac0886d](https://github.com/chanzuckerberg/aws-oidc/commit/ac0886d20e1d2a787c7927d99403bac3ac5eca20))
* bump golang.org/x/net from 0.15.0 to 0.17.0 ([#807](https://github.com/chanzuckerberg/aws-oidc/issues/807)) ([818713f](https://github.com/chanzuckerberg/aws-oidc/commit/818713f3b31642d06d8357eb2151ab8dbd2135ec))
* bump google.golang.org/grpc from 1.57.0 to 1.57.1 ([#821](https://github.com/chanzuckerberg/aws-oidc/issues/821)) ([6849631](https://github.com/chanzuckerberg/aws-oidc/commit/6849631bb5ef620a46f7c36be50d6258de8f1593))
* bump google.golang.org/protobuf from 1.32.0 to 1.33.0 ([#984](https://github.com/chanzuckerberg/aws-oidc/issues/984)) ([ed1f7de](https://github.com/chanzuckerberg/aws-oidc/commit/ed1f7deddab36a22ebdfe8446faa1563e9d29e86))
* CCIE-4313 replace tibdex/github-app-token ([#1105](https://github.com/chanzuckerberg/aws-oidc/issues/1105)) ([8d65095](https://github.com/chanzuckerberg/aws-oidc/commit/8d6509572aa610cf62631062ca3cffe2bc6c2938))
* fix goreleaser ([#1100](https://github.com/chanzuckerberg/aws-oidc/issues/1100)) ([e7338b1](https://github.com/chanzuckerberg/aws-oidc/commit/e7338b122c1aa435dfd59d210dc6fe863709711b))
* **main:** release 0.25.47 ([#583](https://github.com/chanzuckerberg/aws-oidc/issues/583)) ([6929ecb](https://github.com/chanzuckerberg/aws-oidc/commit/6929ecbbd063da8a25a6a47b2c6ae8072a42d3fa))
* **main:** release 0.25.48 ([#585](https://github.com/chanzuckerberg/aws-oidc/issues/585)) ([e55ad36](https://github.com/chanzuckerberg/aws-oidc/commit/e55ad365151279d3c5fdfaa26cb2b47ff9415a05))
* **main:** release 0.25.49 ([#587](https://github.com/chanzuckerberg/aws-oidc/issues/587)) ([9d3da26](https://github.com/chanzuckerberg/aws-oidc/commit/9d3da2622de1d880ede1418cd9e9767b1ffa2185))
* **main:** release 0.25.50 ([#589](https://github.com/chanzuckerberg/aws-oidc/issues/589)) ([2255dd9](https://github.com/chanzuckerberg/aws-oidc/commit/2255dd92194a776866bee73b323d10b804a07d1f))
* **main:** release 0.25.51 ([#596](https://github.com/chanzuckerberg/aws-oidc/issues/596)) ([a711302](https://github.com/chanzuckerberg/aws-oidc/commit/a7113024704a897bed593b9c5a9888456aca7751))
* **main:** release 0.25.52 ([#600](https://github.com/chanzuckerberg/aws-oidc/issues/600)) ([f19cf4b](https://github.com/chanzuckerberg/aws-oidc/commit/f19cf4b99ba4e0ab1057fdeda87500eb6d3f01fd))
* **main:** release 0.25.53 ([#604](https://github.com/chanzuckerberg/aws-oidc/issues/604)) ([0577a99](https://github.com/chanzuckerberg/aws-oidc/commit/0577a9915e89806fae5f2c19ec0e382802d55da9))
* **main:** release 0.25.54 ([#605](https://github.com/chanzuckerberg/aws-oidc/issues/605)) ([c9237a3](https://github.com/chanzuckerberg/aws-oidc/commit/c9237a3ab9111acac8083bc9ac800c77b8160d40))
* **main:** release 0.25.55 ([#607](https://github.com/chanzuckerberg/aws-oidc/issues/607)) ([7bec835](https://github.com/chanzuckerberg/aws-oidc/commit/7bec835a0610718899a1f670e4a8f6c624ac716d))
* **main:** release 0.25.56 ([#610](https://github.com/chanzuckerberg/aws-oidc/issues/610)) ([7054f3b](https://github.com/chanzuckerberg/aws-oidc/commit/7054f3b754b3db76205f676c38b30ef81a24b7bd))
* **main:** release 0.25.57 ([#612](https://github.com/chanzuckerberg/aws-oidc/issues/612)) ([2d8df4a](https://github.com/chanzuckerberg/aws-oidc/commit/2d8df4ad3a711c4b99843de2938f90e245293772))
* **main:** release 0.25.58 ([#614](https://github.com/chanzuckerberg/aws-oidc/issues/614)) ([a92223d](https://github.com/chanzuckerberg/aws-oidc/commit/a92223d8d79b020f4099a7b26b774e2a01198ad2))
* **main:** release 0.25.59 ([#619](https://github.com/chanzuckerberg/aws-oidc/issues/619)) ([5e7e787](https://github.com/chanzuckerberg/aws-oidc/commit/5e7e7874284b721550079e7993e9e9123ccfa975))
* **main:** release 0.25.60 ([#622](https://github.com/chanzuckerberg/aws-oidc/issues/622)) ([bfc5bad](https://github.com/chanzuckerberg/aws-oidc/commit/bfc5badd935253d225e883eccb1ef73a5883b1fb))
* **main:** release 0.25.61 ([#625](https://github.com/chanzuckerberg/aws-oidc/issues/625)) ([415eddb](https://github.com/chanzuckerberg/aws-oidc/commit/415eddbd30221eedce7659a8c0100fad61eb6fe9))
* **main:** release 0.25.62 ([#628](https://github.com/chanzuckerberg/aws-oidc/issues/628)) ([08a53ac](https://github.com/chanzuckerberg/aws-oidc/commit/08a53ace434e09cfd81d85ad79775be38865bb8c))
* **main:** release 0.25.63 ([#630](https://github.com/chanzuckerberg/aws-oidc/issues/630)) ([a3d1e1f](https://github.com/chanzuckerberg/aws-oidc/commit/a3d1e1f0d9889e0b89ce8aecd4902cc52f0b0164))
* **main:** release 0.25.64 ([#635](https://github.com/chanzuckerberg/aws-oidc/issues/635)) ([b7e92c0](https://github.com/chanzuckerberg/aws-oidc/commit/b7e92c0a9104d73833acba58ba7080a092534bf0))
* **main:** release 0.25.65 ([#639](https://github.com/chanzuckerberg/aws-oidc/issues/639)) ([949a039](https://github.com/chanzuckerberg/aws-oidc/commit/949a0396664a4917ada2845de69152ba96076cee))
* **main:** release 0.25.66 ([#641](https://github.com/chanzuckerberg/aws-oidc/issues/641)) ([b2c08a0](https://github.com/chanzuckerberg/aws-oidc/commit/b2c08a00617e219eb5e5a05e9406fd0416fbb815))
* **main:** release 0.25.67 ([#645](https://github.com/chanzuckerberg/aws-oidc/issues/645)) ([56c7140](https://github.com/chanzuckerberg/aws-oidc/commit/56c714090599ce88412e3700a55ce2c9be63bac9))
* **main:** release 0.25.68 ([#648](https://github.com/chanzuckerberg/aws-oidc/issues/648)) ([bcc0f0a](https://github.com/chanzuckerberg/aws-oidc/commit/bcc0f0a5087321cb7bc67ace27339f80e8370678))
* **main:** release 0.25.69 ([#655](https://github.com/chanzuckerberg/aws-oidc/issues/655)) ([047231e](https://github.com/chanzuckerberg/aws-oidc/commit/047231ebba50470b8af6dd889d00f668485dc4d2))
* **main:** release 0.25.70 ([#657](https://github.com/chanzuckerberg/aws-oidc/issues/657)) ([8902554](https://github.com/chanzuckerberg/aws-oidc/commit/8902554985e755c703239c1df60b51f3fbc8f737))
* **main:** release 0.25.71 ([#662](https://github.com/chanzuckerberg/aws-oidc/issues/662)) ([3157d23](https://github.com/chanzuckerberg/aws-oidc/commit/3157d232e85b8e5df8f87d778f6f20dc2f20265e))
* **main:** release 0.25.72 ([#664](https://github.com/chanzuckerberg/aws-oidc/issues/664)) ([517a160](https://github.com/chanzuckerberg/aws-oidc/commit/517a1606e9161e7bc4728bafc47d36125b23efde))
* **main:** release 0.25.73 ([#667](https://github.com/chanzuckerberg/aws-oidc/issues/667)) ([6408230](https://github.com/chanzuckerberg/aws-oidc/commit/640823069839637b4873ff6daa70844e8eda08e6))
* **main:** release 0.25.74 ([#669](https://github.com/chanzuckerberg/aws-oidc/issues/669)) ([d753c05](https://github.com/chanzuckerberg/aws-oidc/commit/d753c05a534e31b949cf4e7448686a66ecdd5da1))
* **main:** release 0.25.75 ([#673](https://github.com/chanzuckerberg/aws-oidc/issues/673)) ([23b7611](https://github.com/chanzuckerberg/aws-oidc/commit/23b76117bf3a32fdfef39b557a627d89e64254c3))
* **main:** release 0.25.76 ([#678](https://github.com/chanzuckerberg/aws-oidc/issues/678)) ([fed4bc6](https://github.com/chanzuckerberg/aws-oidc/commit/fed4bc6a96419698e3a4e9b76ba3292cf28703cc))
* **main:** release 0.25.77 ([#680](https://github.com/chanzuckerberg/aws-oidc/issues/680)) ([ab37cd1](https://github.com/chanzuckerberg/aws-oidc/commit/ab37cd14c24ede8e935ab159b1b0bd922dd7732e))
* **main:** release 0.25.78 ([#685](https://github.com/chanzuckerberg/aws-oidc/issues/685)) ([55722e4](https://github.com/chanzuckerberg/aws-oidc/commit/55722e41ef9d86d6ece66c3f8cabed999e00c491))
* **main:** release 0.25.79 ([#687](https://github.com/chanzuckerberg/aws-oidc/issues/687)) ([3994972](https://github.com/chanzuckerberg/aws-oidc/commit/39949723479d547b62fe7eeb5a5c044c7ff3f003))
* **main:** release 0.25.80 ([#689](https://github.com/chanzuckerberg/aws-oidc/issues/689)) ([2b9d6af](https://github.com/chanzuckerberg/aws-oidc/commit/2b9d6af1269e3920be79e240f9b6e648898313d9))
* **main:** release 0.25.81 ([#692](https://github.com/chanzuckerberg/aws-oidc/issues/692)) ([6628754](https://github.com/chanzuckerberg/aws-oidc/commit/6628754ab40d95c66422f28a1ce50f4254cd8137))
* **main:** release 0.25.82 ([#696](https://github.com/chanzuckerberg/aws-oidc/issues/696)) ([3f70aef](https://github.com/chanzuckerberg/aws-oidc/commit/3f70aef998a71c6a7dc772e803cdbd9168d28c6f))
* **main:** release 0.25.83 ([#700](https://github.com/chanzuckerberg/aws-oidc/issues/700)) ([7019568](https://github.com/chanzuckerberg/aws-oidc/commit/701956849daa5832e9af4acf0b9b486379e99641))
* **main:** release 0.25.84 ([#704](https://github.com/chanzuckerberg/aws-oidc/issues/704)) ([8856ca0](https://github.com/chanzuckerberg/aws-oidc/commit/8856ca0d8261267d235abd826a46608e9efc06e5))
* **main:** release 0.25.85 ([#706](https://github.com/chanzuckerberg/aws-oidc/issues/706)) ([d36731a](https://github.com/chanzuckerberg/aws-oidc/commit/d36731a51020eae79f90df3da3a04b2c504a3f6f))
* **main:** release 0.25.86 ([#710](https://github.com/chanzuckerberg/aws-oidc/issues/710)) ([e8f6fc5](https://github.com/chanzuckerberg/aws-oidc/commit/e8f6fc58bbd7aa1aa71532e2a4e752ea0a9b4b1a))
* **main:** release 0.25.87 ([#717](https://github.com/chanzuckerberg/aws-oidc/issues/717)) ([d0d21f5](https://github.com/chanzuckerberg/aws-oidc/commit/d0d21f5cdf8a8ccd2a2a1b485acdf52f55d03330))
* **main:** release 0.25.88 ([#721](https://github.com/chanzuckerberg/aws-oidc/issues/721)) ([ba2dfd7](https://github.com/chanzuckerberg/aws-oidc/commit/ba2dfd73a1f620c9f493e8e052d392fc814b13ad))
* **main:** release 0.25.89 ([#723](https://github.com/chanzuckerberg/aws-oidc/issues/723)) ([35709d6](https://github.com/chanzuckerberg/aws-oidc/commit/35709d61352edf45dacf1ebb241655857db0e1f3))
* **main:** release 0.25.90 ([#727](https://github.com/chanzuckerberg/aws-oidc/issues/727)) ([0203ee1](https://github.com/chanzuckerberg/aws-oidc/commit/0203ee19445d49d501526ebb7489d8d08b162da8))
* **main:** release 0.25.91 ([#733](https://github.com/chanzuckerberg/aws-oidc/issues/733)) ([81aaf2a](https://github.com/chanzuckerberg/aws-oidc/commit/81aaf2a3aeec3162036075a608835e98556056d7))
* **main:** release 0.25.92 ([#735](https://github.com/chanzuckerberg/aws-oidc/issues/735)) ([3e37b6f](https://github.com/chanzuckerberg/aws-oidc/commit/3e37b6fa32a85dfec326bd8acf072664c7351535))
* **main:** release 0.26.0 ([#738](https://github.com/chanzuckerberg/aws-oidc/issues/738)) ([5743984](https://github.com/chanzuckerberg/aws-oidc/commit/5743984966033e8492f623d1b4c28ce242b964ca))
* **main:** release 0.26.1 ([#741](https://github.com/chanzuckerberg/aws-oidc/issues/741)) ([7a692dc](https://github.com/chanzuckerberg/aws-oidc/commit/7a692dc9771abc265d8eb42a3505513796078a6a))
* **main:** release 0.26.10 ([#761](https://github.com/chanzuckerberg/aws-oidc/issues/761)) ([0a4cdef](https://github.com/chanzuckerberg/aws-oidc/commit/0a4cdef324deab653117285f566ef512a8aea5fc))
* **main:** release 0.26.11 ([#763](https://github.com/chanzuckerberg/aws-oidc/issues/763)) ([cda5be0](https://github.com/chanzuckerberg/aws-oidc/commit/cda5be06cabaf7befc40c811cd040c3eec5a336e))
* **main:** release 0.26.12 ([#765](https://github.com/chanzuckerberg/aws-oidc/issues/765)) ([e11138a](https://github.com/chanzuckerberg/aws-oidc/commit/e11138a1e28b95fc498acf3986358569ec6895f9))
* **main:** release 0.26.13 ([#768](https://github.com/chanzuckerberg/aws-oidc/issues/768)) ([2cfc987](https://github.com/chanzuckerberg/aws-oidc/commit/2cfc987b14e1beefcd30f83030cfb032a6e1cc48))
* **main:** release 0.26.14 ([#770](https://github.com/chanzuckerberg/aws-oidc/issues/770)) ([6b84936](https://github.com/chanzuckerberg/aws-oidc/commit/6b84936128b6a41fd578695984d84722049baec6))
* **main:** release 0.26.15 ([#772](https://github.com/chanzuckerberg/aws-oidc/issues/772)) ([e8d020f](https://github.com/chanzuckerberg/aws-oidc/commit/e8d020fa6dc2e3d70221bf3b2beece22fd11031a))
* **main:** release 0.26.16 ([#775](https://github.com/chanzuckerberg/aws-oidc/issues/775)) ([6f9d834](https://github.com/chanzuckerberg/aws-oidc/commit/6f9d8346a587d76e23dfa53a09dd59f8ee2e66d0))
* **main:** release 0.26.17 ([#777](https://github.com/chanzuckerberg/aws-oidc/issues/777)) ([1993687](https://github.com/chanzuckerberg/aws-oidc/commit/1993687de351c0f8bd65038e50862c367dd89f66))
* **main:** release 0.26.2 ([#743](https://github.com/chanzuckerberg/aws-oidc/issues/743)) ([672520b](https://github.com/chanzuckerberg/aws-oidc/commit/672520ba037d26769c73b0f1d94818dea2700762))
* **main:** release 0.26.3 ([#745](https://github.com/chanzuckerberg/aws-oidc/issues/745)) ([df19d35](https://github.com/chanzuckerberg/aws-oidc/commit/df19d359deace74629f53d13638777c0dd32f544))
* **main:** release 0.26.4 ([#747](https://github.com/chanzuckerberg/aws-oidc/issues/747)) ([5661bb2](https://github.com/chanzuckerberg/aws-oidc/commit/5661bb204281ba4ef5c6e1ac8a1d7038ba5c218e))
* **main:** release 0.26.5 ([#749](https://github.com/chanzuckerberg/aws-oidc/issues/749)) ([4e40c48](https://github.com/chanzuckerberg/aws-oidc/commit/4e40c481b47e54ae331b3ba423e8331f83c2f39f))
* **main:** release 0.26.6 ([#752](https://github.com/chanzuckerberg/aws-oidc/issues/752)) ([153401b](https://github.com/chanzuckerberg/aws-oidc/commit/153401b7582cd7229254f4087d780f102b9f1127))
* **main:** release 0.26.7 ([#754](https://github.com/chanzuckerberg/aws-oidc/issues/754)) ([c4f6b39](https://github.com/chanzuckerberg/aws-oidc/commit/c4f6b392403774c699fed91a29027af57fdcf8cb))
* **main:** release 0.26.8 ([#756](https://github.com/chanzuckerberg/aws-oidc/issues/756)) ([bfa18d9](https://github.com/chanzuckerberg/aws-oidc/commit/bfa18d9cc43cce94a4514af805ab579565b41154))
* **main:** release 0.26.9 ([#759](https://github.com/chanzuckerberg/aws-oidc/issues/759)) ([009b809](https://github.com/chanzuckerberg/aws-oidc/commit/009b809747afaf13d42ef86612c4824f49ce288c))
* **main:** release 0.27.0 ([#780](https://github.com/chanzuckerberg/aws-oidc/issues/780)) ([0246e5b](https://github.com/chanzuckerberg/aws-oidc/commit/0246e5b95ba7868f9729439533677e6b76cfea0a))
* **main:** release 0.28.0 ([#785](https://github.com/chanzuckerberg/aws-oidc/issues/785)) ([7a57714](https://github.com/chanzuckerberg/aws-oidc/commit/7a577144063a8596cd1bcb5d230987fbcfd2e72f))
* **main:** release 0.28.1 ([#795](https://github.com/chanzuckerberg/aws-oidc/issues/795)) ([c472789](https://github.com/chanzuckerberg/aws-oidc/commit/c472789045a89604c8b5581c9c828251dd63449c))
* **main:** release 0.28.10 ([#818](https://github.com/chanzuckerberg/aws-oidc/issues/818)) ([be30173](https://github.com/chanzuckerberg/aws-oidc/commit/be301732dc9e986b1dc9c26ecf46d5830ed16bec))
* **main:** release 0.28.11 ([#822](https://github.com/chanzuckerberg/aws-oidc/issues/822)) ([328eb49](https://github.com/chanzuckerberg/aws-oidc/commit/328eb49ce06255e7cdd4743945a83f5f240499b2))
* **main:** release 0.28.12 ([#825](https://github.com/chanzuckerberg/aws-oidc/issues/825)) ([e6b7c9b](https://github.com/chanzuckerberg/aws-oidc/commit/e6b7c9be38ee573c7a04e4756fed564fade8fcfd))
* **main:** release 0.28.13 ([#830](https://github.com/chanzuckerberg/aws-oidc/issues/830)) ([fceca09](https://github.com/chanzuckerberg/aws-oidc/commit/fceca09c123d9e2b9b3965ca7708d64bd258d4e6))
* **main:** release 0.28.14 ([#834](https://github.com/chanzuckerberg/aws-oidc/issues/834)) ([db0a05c](https://github.com/chanzuckerberg/aws-oidc/commit/db0a05c36f01c57b9dc5d73c30f2ea5d4a11ceca))
* **main:** release 0.28.15 ([#838](https://github.com/chanzuckerberg/aws-oidc/issues/838)) ([9a69738](https://github.com/chanzuckerberg/aws-oidc/commit/9a6973832873426b6835297c0619fc07a0afd366))
* **main:** release 0.28.16 ([#850](https://github.com/chanzuckerberg/aws-oidc/issues/850)) ([d5603a0](https://github.com/chanzuckerberg/aws-oidc/commit/d5603a093f1ca33cf422ea71741d0efa5ee8b010))
* **main:** release 0.28.17 ([#852](https://github.com/chanzuckerberg/aws-oidc/issues/852)) ([2a74cd8](https://github.com/chanzuckerberg/aws-oidc/commit/2a74cd8040a9e54910983c255c0219211be06814))
* **main:** release 0.28.18 ([#858](https://github.com/chanzuckerberg/aws-oidc/issues/858)) ([6b8de9c](https://github.com/chanzuckerberg/aws-oidc/commit/6b8de9ca270acf86adb754115abf3d156cac266e))
* **main:** release 0.28.19 ([#860](https://github.com/chanzuckerberg/aws-oidc/issues/860)) ([05620e5](https://github.com/chanzuckerberg/aws-oidc/commit/05620e54a15334414a9e82b8192ad9df1c1bbce7))
* **main:** release 0.28.2 ([#797](https://github.com/chanzuckerberg/aws-oidc/issues/797)) ([390c66e](https://github.com/chanzuckerberg/aws-oidc/commit/390c66e81a58b15f3b0da8a0309b15e4e5861eeb))
* **main:** release 0.28.20 ([#872](https://github.com/chanzuckerberg/aws-oidc/issues/872)) ([d7e3005](https://github.com/chanzuckerberg/aws-oidc/commit/d7e300540f1bbc000c94c76d8b4303013482edb9))
* **main:** release 0.28.21 ([#875](https://github.com/chanzuckerberg/aws-oidc/issues/875)) ([9dbaf44](https://github.com/chanzuckerberg/aws-oidc/commit/9dbaf446a364a46165918a10984ad6bbc91ce334))
* **main:** release 0.28.22 ([#877](https://github.com/chanzuckerberg/aws-oidc/issues/877)) ([811df84](https://github.com/chanzuckerberg/aws-oidc/commit/811df84698904d4f5c6e3a05e9f6602ee7ee2276))
* **main:** release 0.28.23 ([#880](https://github.com/chanzuckerberg/aws-oidc/issues/880)) ([a15c98b](https://github.com/chanzuckerberg/aws-oidc/commit/a15c98b41d7fb963f0ea748b81dc758f61753cc1))
* **main:** release 0.28.24 ([#882](https://github.com/chanzuckerberg/aws-oidc/issues/882)) ([1fb768b](https://github.com/chanzuckerberg/aws-oidc/commit/1fb768b03fdacc1015d8223184a63edf6abfb607))
* **main:** release 0.28.25 ([#894](https://github.com/chanzuckerberg/aws-oidc/issues/894)) ([3a7c26c](https://github.com/chanzuckerberg/aws-oidc/commit/3a7c26c513886e2e208c6a50b114453934ca0bfa))
* **main:** release 0.28.26 ([#896](https://github.com/chanzuckerberg/aws-oidc/issues/896)) ([37d99a9](https://github.com/chanzuckerberg/aws-oidc/commit/37d99a99dd49d99b7cff0dffb76d3bb5ced424ec))
* **main:** release 0.28.27 ([#898](https://github.com/chanzuckerberg/aws-oidc/issues/898)) ([f8c4b0e](https://github.com/chanzuckerberg/aws-oidc/commit/f8c4b0ecf9b62f3a777857bd2733045461fd6cd0))
* **main:** release 0.28.28 ([#900](https://github.com/chanzuckerberg/aws-oidc/issues/900)) ([435c1e7](https://github.com/chanzuckerberg/aws-oidc/commit/435c1e7f2b83dfd39252157e12dc3fe7e693cdd1))
* **main:** release 0.28.29 ([#902](https://github.com/chanzuckerberg/aws-oidc/issues/902)) ([1fcd095](https://github.com/chanzuckerberg/aws-oidc/commit/1fcd095eaf1a146b8808a4350ba46f63483c3fb4))
* **main:** release 0.28.3 ([#799](https://github.com/chanzuckerberg/aws-oidc/issues/799)) ([a9403c8](https://github.com/chanzuckerberg/aws-oidc/commit/a9403c8699b2cab3b36ddeccfa9cf743bc1d18ca))
* **main:** release 0.28.30 ([#904](https://github.com/chanzuckerberg/aws-oidc/issues/904)) ([bdfa333](https://github.com/chanzuckerberg/aws-oidc/commit/bdfa333691e1c5fc2e9f2ec5e4150a7092765464))
* **main:** release 0.28.31 ([#906](https://github.com/chanzuckerberg/aws-oidc/issues/906)) ([4956398](https://github.com/chanzuckerberg/aws-oidc/commit/49563981c9cc036deee05dd6037e4f05b6f9c217))
* **main:** release 0.28.32 ([#908](https://github.com/chanzuckerberg/aws-oidc/issues/908)) ([d759ed9](https://github.com/chanzuckerberg/aws-oidc/commit/d759ed932ec00080a6d3b271ea39809e707ed3bc))
* **main:** release 0.28.33 ([#911](https://github.com/chanzuckerberg/aws-oidc/issues/911)) ([ad701d9](https://github.com/chanzuckerberg/aws-oidc/commit/ad701d9f246f51f7f03cdb1248826050d524b654))
* **main:** release 0.28.34 ([#913](https://github.com/chanzuckerberg/aws-oidc/issues/913)) ([31b6b03](https://github.com/chanzuckerberg/aws-oidc/commit/31b6b03dbdea38633174b2383efa3e51ae420072))
* **main:** release 0.28.35 ([#915](https://github.com/chanzuckerberg/aws-oidc/issues/915)) ([b507d96](https://github.com/chanzuckerberg/aws-oidc/commit/b507d96f326803dec45ac03e1e373ae272ab8b13))
* **main:** release 0.28.36 ([#917](https://github.com/chanzuckerberg/aws-oidc/issues/917)) ([c3c8430](https://github.com/chanzuckerberg/aws-oidc/commit/c3c8430a27ee73fcfb2a64b8ac9df739526b4dd0))
* **main:** release 0.28.37 ([#919](https://github.com/chanzuckerberg/aws-oidc/issues/919)) ([8c001c5](https://github.com/chanzuckerberg/aws-oidc/commit/8c001c508927a68a799c3ca7c859d48ae80d8338))
* **main:** release 0.28.38 ([#921](https://github.com/chanzuckerberg/aws-oidc/issues/921)) ([5984bc9](https://github.com/chanzuckerberg/aws-oidc/commit/5984bc9c8e184b5b34f6d97f0fe48247e3b57b92))
* **main:** release 0.28.39 ([#924](https://github.com/chanzuckerberg/aws-oidc/issues/924)) ([6439b54](https://github.com/chanzuckerberg/aws-oidc/commit/6439b54d29a87cd7b7b6ae99beea87e4f1fcf497))
* **main:** release 0.28.4 ([#801](https://github.com/chanzuckerberg/aws-oidc/issues/801)) ([8255af5](https://github.com/chanzuckerberg/aws-oidc/commit/8255af5d80e2b44fb519e3c217293ff59fed33bd))
* **main:** release 0.28.40 ([#926](https://github.com/chanzuckerberg/aws-oidc/issues/926)) ([ac9e883](https://github.com/chanzuckerberg/aws-oidc/commit/ac9e88302fecb817cf2a125c8278c735cc2cb4f1))
* **main:** release 0.28.41 ([#928](https://github.com/chanzuckerberg/aws-oidc/issues/928)) ([92a1cdc](https://github.com/chanzuckerberg/aws-oidc/commit/92a1cdc812865a9d8dd89f2293b18a08ce143bcb))
* **main:** release 0.28.42 ([#930](https://github.com/chanzuckerberg/aws-oidc/issues/930)) ([b71a934](https://github.com/chanzuckerberg/aws-oidc/commit/b71a9345360f5ce347fc9bebde46804ea2d8c651))
* **main:** release 0.28.43 ([#932](https://github.com/chanzuckerberg/aws-oidc/issues/932)) ([29c3fee](https://github.com/chanzuckerberg/aws-oidc/commit/29c3fee3ef5b81e45e97450aa5db747d5845703f))
* **main:** release 0.28.44 ([#934](https://github.com/chanzuckerberg/aws-oidc/issues/934)) ([762daa4](https://github.com/chanzuckerberg/aws-oidc/commit/762daa41601e54a4ea47db96f6f0dee67064e541))
* **main:** release 0.28.45 ([#936](https://github.com/chanzuckerberg/aws-oidc/issues/936)) ([4b409a5](https://github.com/chanzuckerberg/aws-oidc/commit/4b409a5894f776bce1247da808a67a926d292734))
* **main:** release 0.28.46 ([#940](https://github.com/chanzuckerberg/aws-oidc/issues/940)) ([2632272](https://github.com/chanzuckerberg/aws-oidc/commit/2632272d401d94702294d8b5753ae7d30aff5ea3))
* **main:** release 0.28.47 ([#942](https://github.com/chanzuckerberg/aws-oidc/issues/942)) ([48dd8cb](https://github.com/chanzuckerberg/aws-oidc/commit/48dd8cb118e13c353fd06222d3a16311a2f2eb09))
* **main:** release 0.28.48 ([#944](https://github.com/chanzuckerberg/aws-oidc/issues/944)) ([28037ad](https://github.com/chanzuckerberg/aws-oidc/commit/28037ad6af77395775357b5e057f782748b1af64))
* **main:** release 0.28.49 ([#948](https://github.com/chanzuckerberg/aws-oidc/issues/948)) ([cb5b5ab](https://github.com/chanzuckerberg/aws-oidc/commit/cb5b5ab07c8a645d8abd2da7faba700e4e42ca48))
* **main:** release 0.28.5 ([#803](https://github.com/chanzuckerberg/aws-oidc/issues/803)) ([b3e1deb](https://github.com/chanzuckerberg/aws-oidc/commit/b3e1debd1b298180e5bbeeb9b0458226851fa31a))
* **main:** release 0.28.50 ([#954](https://github.com/chanzuckerberg/aws-oidc/issues/954)) ([6965c37](https://github.com/chanzuckerberg/aws-oidc/commit/6965c37e77fe80d7c0e1284192af536a7baebdcf))
* **main:** release 0.28.51 ([#956](https://github.com/chanzuckerberg/aws-oidc/issues/956)) ([255c636](https://github.com/chanzuckerberg/aws-oidc/commit/255c636f701e6b58241aa873a6a85aa17bba6b3f))
* **main:** release 0.28.52 ([#959](https://github.com/chanzuckerberg/aws-oidc/issues/959)) ([0bfe5b9](https://github.com/chanzuckerberg/aws-oidc/commit/0bfe5b97cdf834d05987cdb3c9be19a2e647c261))
* **main:** release 0.28.53 ([#961](https://github.com/chanzuckerberg/aws-oidc/issues/961)) ([30a70d0](https://github.com/chanzuckerberg/aws-oidc/commit/30a70d02d1c63ebe86b0015e265823781083feb7))
* **main:** release 0.28.54 ([#963](https://github.com/chanzuckerberg/aws-oidc/issues/963)) ([249be14](https://github.com/chanzuckerberg/aws-oidc/commit/249be14adf002c19903e283c6302f7fbdecfe2b4))
* **main:** release 0.28.55 ([#965](https://github.com/chanzuckerberg/aws-oidc/issues/965)) ([c060973](https://github.com/chanzuckerberg/aws-oidc/commit/c06097354976991fb9d4639f0e993a7274a7d8d8))
* **main:** release 0.28.56 ([#968](https://github.com/chanzuckerberg/aws-oidc/issues/968)) ([e9a69d5](https://github.com/chanzuckerberg/aws-oidc/commit/e9a69d51d0f364beda564b6fcded2ac1478cc252))
* **main:** release 0.28.57 ([#970](https://github.com/chanzuckerberg/aws-oidc/issues/970)) ([b1e5dbe](https://github.com/chanzuckerberg/aws-oidc/commit/b1e5dbe596c7d4153065a0d88b7e3a10da37ec66))
* **main:** release 0.28.58 ([#974](https://github.com/chanzuckerberg/aws-oidc/issues/974)) ([24f9a36](https://github.com/chanzuckerberg/aws-oidc/commit/24f9a364bc0b99cfb583abdfadbe18dedb164e70))
* **main:** release 0.28.59 ([#976](https://github.com/chanzuckerberg/aws-oidc/issues/976)) ([4de8705](https://github.com/chanzuckerberg/aws-oidc/commit/4de8705a0fec5f8738dae21b806a74d143cb1427))
* **main:** release 0.28.6 ([#806](https://github.com/chanzuckerberg/aws-oidc/issues/806)) ([567b6d1](https://github.com/chanzuckerberg/aws-oidc/commit/567b6d1bab3f76de49371e0be377c5e6ba8ac7e6))
* **main:** release 0.28.60 ([#979](https://github.com/chanzuckerberg/aws-oidc/issues/979)) ([3e9ab6f](https://github.com/chanzuckerberg/aws-oidc/commit/3e9ab6f8ad9635d86ffd6627770d9cc204598585))
* **main:** release 0.28.61 ([#981](https://github.com/chanzuckerberg/aws-oidc/issues/981)) ([f90a3db](https://github.com/chanzuckerberg/aws-oidc/commit/f90a3dbb85d1b0ce9078b625f15e19e25e075d32))
* **main:** release 0.28.62 ([#983](https://github.com/chanzuckerberg/aws-oidc/issues/983)) ([aec0d35](https://github.com/chanzuckerberg/aws-oidc/commit/aec0d35f192e8f6de1de320afc578499c0c87dd9))
* **main:** release 0.28.63 ([#985](https://github.com/chanzuckerberg/aws-oidc/issues/985)) ([32781fc](https://github.com/chanzuckerberg/aws-oidc/commit/32781fcef9516a1ce9b9a0b5163b3c1c4b34032f))
* **main:** release 0.28.64 ([#988](https://github.com/chanzuckerberg/aws-oidc/issues/988)) ([dc6e781](https://github.com/chanzuckerberg/aws-oidc/commit/dc6e7815a82a765cf692702b718ae00084ac3333))
* **main:** release 0.28.65 ([#990](https://github.com/chanzuckerberg/aws-oidc/issues/990)) ([df579fa](https://github.com/chanzuckerberg/aws-oidc/commit/df579fab2f3a0af5ae314367ce81208a1fdb9905))
* **main:** release 0.28.66 ([#992](https://github.com/chanzuckerberg/aws-oidc/issues/992)) ([3e4d66a](https://github.com/chanzuckerberg/aws-oidc/commit/3e4d66aaa969a4c24980e1b0ecfc4083725e0883))
* **main:** release 0.28.67 ([#1012](https://github.com/chanzuckerberg/aws-oidc/issues/1012)) ([06d9a8f](https://github.com/chanzuckerberg/aws-oidc/commit/06d9a8fb90b1b338c62c5b1d1cfc4ebe00c1d8e6))
* **main:** release 0.28.68 ([#1015](https://github.com/chanzuckerberg/aws-oidc/issues/1015)) ([7192ae5](https://github.com/chanzuckerberg/aws-oidc/commit/7192ae5ddfef044aaac83a45cc7806937d6f1afb))
* **main:** release 0.28.69 ([#1089](https://github.com/chanzuckerberg/aws-oidc/issues/1089)) ([90fd37b](https://github.com/chanzuckerberg/aws-oidc/commit/90fd37bf6d1d6ed19c2276aa147e0bb732ca51d6))
* **main:** release 0.28.7 ([#809](https://github.com/chanzuckerberg/aws-oidc/issues/809)) ([01a6244](https://github.com/chanzuckerberg/aws-oidc/commit/01a62440af406ec7dddd1fb42229b7f8a91d16e1))
* **main:** release 0.28.70 ([#1092](https://github.com/chanzuckerberg/aws-oidc/issues/1092)) ([5d2551e](https://github.com/chanzuckerberg/aws-oidc/commit/5d2551eb8d36e2f7595175c07fafbe0facd9377c))
* **main:** release 0.28.71 ([#1094](https://github.com/chanzuckerberg/aws-oidc/issues/1094)) ([cabd4bd](https://github.com/chanzuckerberg/aws-oidc/commit/cabd4bd7a641a2d19f2508e06b1fff735fbcb034))
* **main:** release 0.28.72 ([#1096](https://github.com/chanzuckerberg/aws-oidc/issues/1096)) ([3702c0c](https://github.com/chanzuckerberg/aws-oidc/commit/3702c0cf3fb030d811ebe13edd2b881542760c1d))
* **main:** release 0.28.73 ([#1098](https://github.com/chanzuckerberg/aws-oidc/issues/1098)) ([9e33b71](https://github.com/chanzuckerberg/aws-oidc/commit/9e33b7191d9989049d2fc0ea30dfbc08bb89acce))
* **main:** release 0.28.74 ([#1102](https://github.com/chanzuckerberg/aws-oidc/issues/1102)) ([f5ea625](https://github.com/chanzuckerberg/aws-oidc/commit/f5ea62593deb355bfaca9f5b2d45a7082a15f446))
* **main:** release 0.28.8 ([#814](https://github.com/chanzuckerberg/aws-oidc/issues/814)) ([749ce25](https://github.com/chanzuckerberg/aws-oidc/commit/749ce25e97e177a62ee740719d49f1d7ae5a9b3b))
* **main:** release 0.28.9 ([#816](https://github.com/chanzuckerberg/aws-oidc/issues/816)) ([e6e112c](https://github.com/chanzuckerberg/aws-oidc/commit/e6e112cbee19d1c5f223e5cc84ae884834cfae44))
* replace deprecated fork usage ([#1083](https://github.com/chanzuckerberg/aws-oidc/issues/1083)) ([867094e](https://github.com/chanzuckerberg/aws-oidc/commit/867094efdfa6a4a04792b1cc32bea549b0e6e7a5))

## [0.28.74](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.73...v0.28.74) (2025-03-10)


### BugFixes

* remove toolchain directive ([#1101](https://github.com/chanzuckerberg/aws-oidc/issues/1101)) ([107923e](https://github.com/chanzuckerberg/aws-oidc/commit/107923e8b457f957a5248d490b07ccc0495ca237))

## [0.28.73](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.72...v0.28.73) (2025-03-07)


### Misc

* bump github.com/go-jose/go-jose/v3 from 3.0.3 to 3.0.4 ([#1099](https://github.com/chanzuckerberg/aws-oidc/issues/1099)) ([123c542](https://github.com/chanzuckerberg/aws-oidc/commit/123c54281275ced928d772db81d5360a3b23c076))
* bump github.com/spf13/cobra from 1.8.1 to 1.9.1 ([#1097](https://github.com/chanzuckerberg/aws-oidc/issues/1097)) ([c792200](https://github.com/chanzuckerberg/aws-oidc/commit/c7922009fe2ae3f854f0103f5f749d04c65a01ec))
* fix goreleaser ([#1100](https://github.com/chanzuckerberg/aws-oidc/issues/1100)) ([e7338b1](https://github.com/chanzuckerberg/aws-oidc/commit/e7338b122c1aa435dfd59d210dc6fe863709711b))

## [0.28.72](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.71...v0.28.72) (2025-02-07)


### BugFixes

* Update runs-on labels in GitHub Actions workflows ([#1095](https://github.com/chanzuckerberg/aws-oidc/issues/1095)) ([fe0228b](https://github.com/chanzuckerberg/aws-oidc/commit/fe0228b88d9b3c92fee5b6d7a59c9f185e8cd5d9))
* Update runs-on to use ARM64 or X64 ([fe0228b](https://github.com/chanzuckerberg/aws-oidc/commit/fe0228b88d9b3c92fee5b6d7a59c9f185e8cd5d9))

## [0.28.71](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.70...v0.28.71) (2025-01-23)


### Misc

* bump github.com/coreos/go-oidc from 2.2.1+incompatible to 2.3.0+incompatible ([#1093](https://github.com/chanzuckerberg/aws-oidc/issues/1093)) ([00c4603](https://github.com/chanzuckerberg/aws-oidc/commit/00c4603bc78268c9af4949f30d1569c6d126b1dd))

## [0.28.70](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.69...v0.28.70) (2025-01-16)


### Misc

* bump github.com/aws/aws-sdk-go from 1.55.5 to 1.55.6 ([#1090](https://github.com/chanzuckerberg/aws-oidc/issues/1090)) ([0d1e46a](https://github.com/chanzuckerberg/aws-oidc/commit/0d1e46a7385315cfae208ab160f282bc66207c90))
* bump golang.org/x/crypto from 0.25.0 to 0.31.0 ([#1091](https://github.com/chanzuckerberg/aws-oidc/issues/1091)) ([ac0886d](https://github.com/chanzuckerberg/aws-oidc/commit/ac0886d20e1d2a787c7927d99403bac3ac5eca20))

## [0.28.69](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.68...v0.28.69) (2024-11-25)


### Misc

* bump github.com/honeycombio/beeline-go from 1.17.0 to 1.18.0 ([#1087](https://github.com/chanzuckerberg/aws-oidc/issues/1087)) ([14e0c23](https://github.com/chanzuckerberg/aws-oidc/commit/14e0c2335ac479367b7353bb258c59eea76c9874))
* bump github.com/stretchr/testify from 1.9.0 to 1.10.0 ([#1088](https://github.com/chanzuckerberg/aws-oidc/issues/1088)) ([a6f3e87](https://github.com/chanzuckerberg/aws-oidc/commit/a6f3e8741782e7b2047c82d27080b15511125f47))

## [0.28.68](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.67...v0.28.68) (2024-08-01)


### Misc

* bump github.com/aws/aws-sdk-go from 1.51.20 to 1.51.21 ([#1014](https://github.com/chanzuckerberg/aws-oidc/issues/1014)) ([653f818](https://github.com/chanzuckerberg/aws-oidc/commit/653f818177d5c52c3dc81725dbb8a6df897d8b12))
* bump github.com/aws/aws-sdk-go from 1.51.21 to 1.51.22 ([#1017](https://github.com/chanzuckerberg/aws-oidc/issues/1017)) ([ed544da](https://github.com/chanzuckerberg/aws-oidc/commit/ed544da2802b7d457e891bf82016f98f02e6e3f2))
* bump github.com/aws/aws-sdk-go from 1.51.22 to 1.51.23 ([#1018](https://github.com/chanzuckerberg/aws-oidc/issues/1018)) ([3d84fe7](https://github.com/chanzuckerberg/aws-oidc/commit/3d84fe7974984fa35e5514355fdd2dd18b1c8eb6))
* bump github.com/aws/aws-sdk-go from 1.51.23 to 1.51.24 ([#1019](https://github.com/chanzuckerberg/aws-oidc/issues/1019)) ([04bde53](https://github.com/chanzuckerberg/aws-oidc/commit/04bde5360f800cb61a5aef0ec9fa090b934eba14))
* bump github.com/aws/aws-sdk-go from 1.51.24 to 1.51.25 ([#1020](https://github.com/chanzuckerberg/aws-oidc/issues/1020)) ([25e7582](https://github.com/chanzuckerberg/aws-oidc/commit/25e758224d10a38f59bcf3117ccdf06a0b1d1a3c))
* bump github.com/aws/aws-sdk-go from 1.51.25 to 1.51.26 ([#1021](https://github.com/chanzuckerberg/aws-oidc/issues/1021)) ([0b45947](https://github.com/chanzuckerberg/aws-oidc/commit/0b45947a37651a6ff33794d449df9cfbc79d90c0))
* bump github.com/aws/aws-sdk-go from 1.51.26 to 1.51.27 ([#1022](https://github.com/chanzuckerberg/aws-oidc/issues/1022)) ([5ac2b1d](https://github.com/chanzuckerberg/aws-oidc/commit/5ac2b1d48c299bfbae6b6956c73e41c224040277))
* bump github.com/aws/aws-sdk-go from 1.51.27 to 1.51.28 ([#1023](https://github.com/chanzuckerberg/aws-oidc/issues/1023)) ([d7f82e3](https://github.com/chanzuckerberg/aws-oidc/commit/d7f82e390e987ee2928a95bf5117f323e2c49774))
* bump github.com/aws/aws-sdk-go from 1.51.28 to 1.51.29 ([#1024](https://github.com/chanzuckerberg/aws-oidc/issues/1024)) ([c5f2f2a](https://github.com/chanzuckerberg/aws-oidc/commit/c5f2f2a016f7822197d0aa67ed1ac46559b68d14))
* bump github.com/aws/aws-sdk-go from 1.51.29 to 1.51.30 ([#1025](https://github.com/chanzuckerberg/aws-oidc/issues/1025)) ([5e53f3a](https://github.com/chanzuckerberg/aws-oidc/commit/5e53f3a1798040b692e664e807062ce8d8542fbb))
* bump github.com/aws/aws-sdk-go from 1.51.30 to 1.51.31 ([#1026](https://github.com/chanzuckerberg/aws-oidc/issues/1026)) ([671b446](https://github.com/chanzuckerberg/aws-oidc/commit/671b446366f83d07d620e23f60977c205bc79565))
* bump github.com/aws/aws-sdk-go from 1.51.31 to 1.51.32 ([#1027](https://github.com/chanzuckerberg/aws-oidc/issues/1027)) ([206753e](https://github.com/chanzuckerberg/aws-oidc/commit/206753eee246c548eb6b9c98ad7c37708f995ea9))
* bump github.com/aws/aws-sdk-go from 1.51.32 to 1.52.0 ([#1028](https://github.com/chanzuckerberg/aws-oidc/issues/1028)) ([cb5fb2e](https://github.com/chanzuckerberg/aws-oidc/commit/cb5fb2ed233d3dd84e624b5dfae14e8920e0980f))
* bump github.com/aws/aws-sdk-go from 1.52.0 to 1.52.1 ([#1029](https://github.com/chanzuckerberg/aws-oidc/issues/1029)) ([808680a](https://github.com/chanzuckerberg/aws-oidc/commit/808680a6fd4752b20ff9411befb6529d94352040))
* bump github.com/aws/aws-sdk-go from 1.52.1 to 1.52.2 ([#1030](https://github.com/chanzuckerberg/aws-oidc/issues/1030)) ([a13dade](https://github.com/chanzuckerberg/aws-oidc/commit/a13dade2800fdbd6def8af5c2e7cdb760cdad732))
* bump github.com/aws/aws-sdk-go from 1.52.2 to 1.52.3 ([#1031](https://github.com/chanzuckerberg/aws-oidc/issues/1031)) ([c0efd3b](https://github.com/chanzuckerberg/aws-oidc/commit/c0efd3bb1c19ed6398f8270a0687c113762650be))
* bump github.com/aws/aws-sdk-go from 1.52.3 to 1.52.4 ([#1032](https://github.com/chanzuckerberg/aws-oidc/issues/1032)) ([968f278](https://github.com/chanzuckerberg/aws-oidc/commit/968f27863d88405e9e222412773cf15c55ebf67d))
* bump github.com/aws/aws-sdk-go from 1.52.4 to 1.52.5 ([#1033](https://github.com/chanzuckerberg/aws-oidc/issues/1033)) ([af18501](https://github.com/chanzuckerberg/aws-oidc/commit/af18501f64a4cbfc8106efc564be72d412622ed1))
* bump github.com/aws/aws-sdk-go from 1.52.5 to 1.52.6 ([#1034](https://github.com/chanzuckerberg/aws-oidc/issues/1034)) ([19e7ebc](https://github.com/chanzuckerberg/aws-oidc/commit/19e7ebcebddb0cf5b3893793b485594a64596f10))
* bump github.com/aws/aws-sdk-go from 1.52.6 to 1.53.0 ([#1035](https://github.com/chanzuckerberg/aws-oidc/issues/1035)) ([e86c914](https://github.com/chanzuckerberg/aws-oidc/commit/e86c914d2fb69afb25ee3bcbdf9d12a9ee71f996))
* bump github.com/aws/aws-sdk-go from 1.53.0 to 1.53.2 ([#1036](https://github.com/chanzuckerberg/aws-oidc/issues/1036)) ([7385238](https://github.com/chanzuckerberg/aws-oidc/commit/7385238cc2cb979ab4c4bab3a5a0c4be7b11fbce))
* bump github.com/aws/aws-sdk-go from 1.53.10 to 1.53.11 ([#1045](https://github.com/chanzuckerberg/aws-oidc/issues/1045)) ([f522f0c](https://github.com/chanzuckerberg/aws-oidc/commit/f522f0c874f065e64ba4ce171b45ddb085f3edc9))
* bump github.com/aws/aws-sdk-go from 1.53.11 to 1.53.12 ([#1046](https://github.com/chanzuckerberg/aws-oidc/issues/1046)) ([e816386](https://github.com/chanzuckerberg/aws-oidc/commit/e816386ad085aa10f269a70ef2bd17d7ca648fc6))
* bump github.com/aws/aws-sdk-go from 1.53.12 to 1.53.13 ([#1047](https://github.com/chanzuckerberg/aws-oidc/issues/1047)) ([d1f72cd](https://github.com/chanzuckerberg/aws-oidc/commit/d1f72cd8af8b43a98259d3dd9cf19529928da690))
* bump github.com/aws/aws-sdk-go from 1.53.13 to 1.53.14 ([#1048](https://github.com/chanzuckerberg/aws-oidc/issues/1048)) ([fadead1](https://github.com/chanzuckerberg/aws-oidc/commit/fadead1e5dba279bb27022f042c2b0ef03102fef))
* bump github.com/aws/aws-sdk-go from 1.53.14 to 1.53.15 ([#1049](https://github.com/chanzuckerberg/aws-oidc/issues/1049)) ([0420556](https://github.com/chanzuckerberg/aws-oidc/commit/04205564f41cdef14bf2dc8799502b4e2762ccf2))
* bump github.com/aws/aws-sdk-go from 1.53.15 to 1.53.16 ([#1050](https://github.com/chanzuckerberg/aws-oidc/issues/1050)) ([118ad6d](https://github.com/chanzuckerberg/aws-oidc/commit/118ad6d70902a25bcfada600e066de91f4f70d03))
* bump github.com/aws/aws-sdk-go from 1.53.16 to 1.53.17 ([#1051](https://github.com/chanzuckerberg/aws-oidc/issues/1051)) ([ae5d5bb](https://github.com/chanzuckerberg/aws-oidc/commit/ae5d5bb568d13597cffcf31249aec7f77bf8860c))
* bump github.com/aws/aws-sdk-go from 1.53.17 to 1.53.18 ([#1052](https://github.com/chanzuckerberg/aws-oidc/issues/1052)) ([b96fba9](https://github.com/chanzuckerberg/aws-oidc/commit/b96fba9f23309ada2a1d9a90c9f333712ece76ee))
* bump github.com/aws/aws-sdk-go from 1.53.18 to 1.53.19 ([#1053](https://github.com/chanzuckerberg/aws-oidc/issues/1053)) ([8a9524b](https://github.com/chanzuckerberg/aws-oidc/commit/8a9524bd05043dd54e0c49d3975be8056bb4a3d9))
* bump github.com/aws/aws-sdk-go from 1.53.19 to 1.53.20 ([#1054](https://github.com/chanzuckerberg/aws-oidc/issues/1054)) ([7e96f11](https://github.com/chanzuckerberg/aws-oidc/commit/7e96f11ed454f5e1c811f3c1fb50465576b4d5a2))
* bump github.com/aws/aws-sdk-go from 1.53.2 to 1.53.3 ([#1037](https://github.com/chanzuckerberg/aws-oidc/issues/1037)) ([91435e0](https://github.com/chanzuckerberg/aws-oidc/commit/91435e04290217184a68ca99cfb1363c39c7e862))
* bump github.com/aws/aws-sdk-go from 1.53.20 to 1.53.21 ([#1055](https://github.com/chanzuckerberg/aws-oidc/issues/1055)) ([d221c1d](https://github.com/chanzuckerberg/aws-oidc/commit/d221c1d5793590f0fed08842c31836afff37a577))
* bump github.com/aws/aws-sdk-go from 1.53.21 to 1.54.0 ([#1056](https://github.com/chanzuckerberg/aws-oidc/issues/1056)) ([4fffa51](https://github.com/chanzuckerberg/aws-oidc/commit/4fffa513cf0266f09b8ba9cf16a98dcc09d7ab69))
* bump github.com/aws/aws-sdk-go from 1.53.3 to 1.53.4 ([#1038](https://github.com/chanzuckerberg/aws-oidc/issues/1038)) ([132dc72](https://github.com/chanzuckerberg/aws-oidc/commit/132dc722c1256d3814b88484589f3765d33e6167))
* bump github.com/aws/aws-sdk-go from 1.53.4 to 1.53.5 ([#1039](https://github.com/chanzuckerberg/aws-oidc/issues/1039)) ([fbb1594](https://github.com/chanzuckerberg/aws-oidc/commit/fbb1594b1d1c62aae5df40cd4e9d58d938ab7eff))
* bump github.com/aws/aws-sdk-go from 1.53.5 to 1.53.6 ([#1040](https://github.com/chanzuckerberg/aws-oidc/issues/1040)) ([b84994f](https://github.com/chanzuckerberg/aws-oidc/commit/b84994f6a0b60999330350d76c4ce08c167f5557))
* bump github.com/aws/aws-sdk-go from 1.53.6 to 1.53.7 ([#1041](https://github.com/chanzuckerberg/aws-oidc/issues/1041)) ([73151f5](https://github.com/chanzuckerberg/aws-oidc/commit/73151f5a0df4ba619e81833e53f54f88c01e22fc))
* bump github.com/aws/aws-sdk-go from 1.53.7 to 1.53.8 ([#1042](https://github.com/chanzuckerberg/aws-oidc/issues/1042)) ([31c6a66](https://github.com/chanzuckerberg/aws-oidc/commit/31c6a66969fe8128d4fd9f8944cd31b3dcf49bee))
* bump github.com/aws/aws-sdk-go from 1.53.8 to 1.53.9 ([#1043](https://github.com/chanzuckerberg/aws-oidc/issues/1043)) ([ce85e7e](https://github.com/chanzuckerberg/aws-oidc/commit/ce85e7ee44831eb3abee2fdafd0b9723d66bed19))
* bump github.com/aws/aws-sdk-go from 1.53.9 to 1.53.10 ([#1044](https://github.com/chanzuckerberg/aws-oidc/issues/1044)) ([0bc1c2b](https://github.com/chanzuckerberg/aws-oidc/commit/0bc1c2bd84263b294c8a86729c9b9b6f499a87ac))
* bump github.com/aws/aws-sdk-go from 1.54.0 to 1.54.1 ([#1058](https://github.com/chanzuckerberg/aws-oidc/issues/1058)) ([c402fa3](https://github.com/chanzuckerberg/aws-oidc/commit/c402fa3e1636deaa11f7b0c72ae26f53d5cd844f))
* bump github.com/aws/aws-sdk-go from 1.54.1 to 1.54.2 ([#1060](https://github.com/chanzuckerberg/aws-oidc/issues/1060)) ([7fe8d69](https://github.com/chanzuckerberg/aws-oidc/commit/7fe8d691552ed6fc09e83490774d9c549a177e05))
* bump github.com/aws/aws-sdk-go from 1.54.10 to 1.54.11 ([#1069](https://github.com/chanzuckerberg/aws-oidc/issues/1069)) ([b537d63](https://github.com/chanzuckerberg/aws-oidc/commit/b537d63f30743a076dadd904b2640643bdfb5a5c))
* bump github.com/aws/aws-sdk-go from 1.54.11 to 1.54.12 ([#1070](https://github.com/chanzuckerberg/aws-oidc/issues/1070)) ([84b23e3](https://github.com/chanzuckerberg/aws-oidc/commit/84b23e3728cc026294e8d15e4d66ea036332021b))
* bump github.com/aws/aws-sdk-go from 1.54.12 to 1.54.13 ([#1071](https://github.com/chanzuckerberg/aws-oidc/issues/1071)) ([731057d](https://github.com/chanzuckerberg/aws-oidc/commit/731057ddcdfafb5edcc90e354f38d5b0e95b1dc5))
* bump github.com/aws/aws-sdk-go from 1.54.13 to 1.54.14 ([#1072](https://github.com/chanzuckerberg/aws-oidc/issues/1072)) ([614b456](https://github.com/chanzuckerberg/aws-oidc/commit/614b456f98345fe72ee6be164f365f84604e15d0))
* bump github.com/aws/aws-sdk-go from 1.54.14 to 1.54.15 ([#1073](https://github.com/chanzuckerberg/aws-oidc/issues/1073)) ([faf019b](https://github.com/chanzuckerberg/aws-oidc/commit/faf019bb9029e9dfb2cf1548e94fee3393d333dd))
* bump github.com/aws/aws-sdk-go from 1.54.15 to 1.54.16 ([#1074](https://github.com/chanzuckerberg/aws-oidc/issues/1074)) ([84508b8](https://github.com/chanzuckerberg/aws-oidc/commit/84508b89f10c448baa395f6d77824b048b6ac05d))
* bump github.com/aws/aws-sdk-go from 1.54.16 to 1.54.17 ([#1075](https://github.com/chanzuckerberg/aws-oidc/issues/1075)) ([fd61abd](https://github.com/chanzuckerberg/aws-oidc/commit/fd61abd866817f992385ebb00b3a7efb4d0c9338))
* bump github.com/aws/aws-sdk-go from 1.54.17 to 1.54.18 ([#1076](https://github.com/chanzuckerberg/aws-oidc/issues/1076)) ([d25603f](https://github.com/chanzuckerberg/aws-oidc/commit/d25603f99e6fa5431aa7e6175db8c8677447442b))
* bump github.com/aws/aws-sdk-go from 1.54.18 to 1.54.19 ([#1077](https://github.com/chanzuckerberg/aws-oidc/issues/1077)) ([f4c23c6](https://github.com/chanzuckerberg/aws-oidc/commit/f4c23c6aa2b630f3a226157fd385c1ac15f55b7e))
* bump github.com/aws/aws-sdk-go from 1.54.19 to 1.54.20 ([#1078](https://github.com/chanzuckerberg/aws-oidc/issues/1078)) ([eedff16](https://github.com/chanzuckerberg/aws-oidc/commit/eedff167bc138c6c453564c8062472b1e428cd4e))
* bump github.com/aws/aws-sdk-go from 1.54.2 to 1.54.3 ([#1061](https://github.com/chanzuckerberg/aws-oidc/issues/1061)) ([845cfad](https://github.com/chanzuckerberg/aws-oidc/commit/845cfad007cdfa9d498d85ebbd6d55a75f1fb78c))
* bump github.com/aws/aws-sdk-go from 1.54.20 to 1.55.0 ([#1079](https://github.com/chanzuckerberg/aws-oidc/issues/1079)) ([c8d38eb](https://github.com/chanzuckerberg/aws-oidc/commit/c8d38ebd2e0339b050a4dc24bea042ada4ae8915))
* bump github.com/aws/aws-sdk-go from 1.54.3 to 1.54.4 ([#1062](https://github.com/chanzuckerberg/aws-oidc/issues/1062)) ([b4a87d3](https://github.com/chanzuckerberg/aws-oidc/commit/b4a87d39ad79cfd36c0a3b81e6855e03d561fc80))
* bump github.com/aws/aws-sdk-go from 1.54.4 to 1.54.5 ([#1063](https://github.com/chanzuckerberg/aws-oidc/issues/1063)) ([73641aa](https://github.com/chanzuckerberg/aws-oidc/commit/73641aab1d95dcd57d16353130c1ec00cbc34173))
* bump github.com/aws/aws-sdk-go from 1.54.5 to 1.54.6 ([#1064](https://github.com/chanzuckerberg/aws-oidc/issues/1064)) ([799ee14](https://github.com/chanzuckerberg/aws-oidc/commit/799ee1498ec23179b916b75d0199049384400abb))
* bump github.com/aws/aws-sdk-go from 1.54.6 to 1.54.7 ([#1065](https://github.com/chanzuckerberg/aws-oidc/issues/1065)) ([461403e](https://github.com/chanzuckerberg/aws-oidc/commit/461403e4c457334767924dd7349efd81aa5a6f4c))
* bump github.com/aws/aws-sdk-go from 1.54.7 to 1.54.8 ([#1066](https://github.com/chanzuckerberg/aws-oidc/issues/1066)) ([0d3aa55](https://github.com/chanzuckerberg/aws-oidc/commit/0d3aa55d88b402d8e22323e26efc267e129cb74d))
* bump github.com/aws/aws-sdk-go from 1.54.8 to 1.54.9 ([#1067](https://github.com/chanzuckerberg/aws-oidc/issues/1067)) ([168af5b](https://github.com/chanzuckerberg/aws-oidc/commit/168af5b81c52681770b7a6e0dc5112f59f92e4d3))
* bump github.com/aws/aws-sdk-go from 1.54.9 to 1.54.10 ([#1068](https://github.com/chanzuckerberg/aws-oidc/issues/1068)) ([956adbe](https://github.com/chanzuckerberg/aws-oidc/commit/956adbe37034e8139ffab95b446b908ddf7c0a6d))
* bump github.com/aws/aws-sdk-go from 1.55.0 to 1.55.1 ([#1080](https://github.com/chanzuckerberg/aws-oidc/issues/1080)) ([d95d52a](https://github.com/chanzuckerberg/aws-oidc/commit/d95d52a5939e166d18a41a374f3620b3d041b9c0))
* bump github.com/aws/aws-sdk-go from 1.55.1 to 1.55.2 ([#1081](https://github.com/chanzuckerberg/aws-oidc/issues/1081)) ([71bb88d](https://github.com/chanzuckerberg/aws-oidc/commit/71bb88dc29bf9e6f98d773dc3f023d0e27c8ad32))
* bump github.com/aws/aws-sdk-go from 1.55.2 to 1.55.3 ([#1082](https://github.com/chanzuckerberg/aws-oidc/issues/1082)) ([a949f79](https://github.com/chanzuckerberg/aws-oidc/commit/a949f7947b16cc00c33853188031a4d9b727d4ce))
* bump github.com/aws/aws-sdk-go from 1.55.3 to 1.55.4 ([#1084](https://github.com/chanzuckerberg/aws-oidc/issues/1084)) ([05bc6b7](https://github.com/chanzuckerberg/aws-oidc/commit/05bc6b7420c3a5acf1c8f89dfad39e1fc3c35bd5))
* bump github.com/aws/aws-sdk-go from 1.55.4 to 1.55.5 ([#1085](https://github.com/chanzuckerberg/aws-oidc/issues/1085)) ([c304a56](https://github.com/chanzuckerberg/aws-oidc/commit/c304a56dbc2de01aee199d9d805232bbeee91529))
* bump github.com/honeycombio/beeline-go from 1.16.0 to 1.17.0 ([#1057](https://github.com/chanzuckerberg/aws-oidc/issues/1057)) ([589afa8](https://github.com/chanzuckerberg/aws-oidc/commit/589afa803ab74693850943b47d1698190e763238))
* bump github.com/spf13/cobra from 1.8.0 to 1.8.1 ([#1059](https://github.com/chanzuckerberg/aws-oidc/issues/1059)) ([a2a9269](https://github.com/chanzuckerberg/aws-oidc/commit/a2a926941df23fa58b8dbadd1f411dd106e0e626))
* replace deprecated fork usage ([#1083](https://github.com/chanzuckerberg/aws-oidc/issues/1083)) ([867094e](https://github.com/chanzuckerberg/aws-oidc/commit/867094efdfa6a4a04792b1cc32bea549b0e6e7a5))


### BugFixes

* helper scripts to update deps ([#1086](https://github.com/chanzuckerberg/aws-oidc/issues/1086)) ([6cba7e1](https://github.com/chanzuckerberg/aws-oidc/commit/6cba7e1c11c9cebcd4c4d2698471c5c216e89096))

## [0.28.67](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.66...v0.28.67) (2024-04-12)


### Misc

* bump github.com/aws/aws-sdk-go from 1.51.18 to 1.51.19 ([#1011](https://github.com/chanzuckerberg/aws-oidc/issues/1011)) ([78cd8df](https://github.com/chanzuckerberg/aws-oidc/commit/78cd8dfbf34848c98ba9706b056545a528404599))
* bump github.com/aws/aws-sdk-go from 1.51.19 to 1.51.20 ([#1013](https://github.com/chanzuckerberg/aws-oidc/issues/1013)) ([3c96519](https://github.com/chanzuckerberg/aws-oidc/commit/3c965190239ec3c5c07707d3b9572150ba2eadcc))

## [0.28.66](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.65...v0.28.66) (2024-04-10)


### Misc

* bump github.com/aws/aws-sdk-go from 1.51.1 to 1.51.2 ([#991](https://github.com/chanzuckerberg/aws-oidc/issues/991)) ([3c1433e](https://github.com/chanzuckerberg/aws-oidc/commit/3c1433ee4fe564e8bfab111974892ac24bc2937b))
* bump github.com/aws/aws-sdk-go from 1.51.10 to 1.51.11 ([#1001](https://github.com/chanzuckerberg/aws-oidc/issues/1001)) ([0f41f3f](https://github.com/chanzuckerberg/aws-oidc/commit/0f41f3ff7f374a121ec3739e83c96f6c19bc56f5))
* bump github.com/aws/aws-sdk-go from 1.51.11 to 1.51.12 ([#1002](https://github.com/chanzuckerberg/aws-oidc/issues/1002)) ([1d96e85](https://github.com/chanzuckerberg/aws-oidc/commit/1d96e85185d98774f949c18dbf8b0a214af917b8))
* bump github.com/aws/aws-sdk-go from 1.51.12 to 1.51.13 ([#1003](https://github.com/chanzuckerberg/aws-oidc/issues/1003)) ([5fd283f](https://github.com/chanzuckerberg/aws-oidc/commit/5fd283f5018c207640f03fbe371d76a5c206e29e))
* bump github.com/aws/aws-sdk-go from 1.51.13 to 1.51.14 ([#1004](https://github.com/chanzuckerberg/aws-oidc/issues/1004)) ([d408e14](https://github.com/chanzuckerberg/aws-oidc/commit/d408e145748c322be229f4fe0c0c7b16bcf44ac1))
* bump github.com/aws/aws-sdk-go from 1.51.14 to 1.51.15 ([#1006](https://github.com/chanzuckerberg/aws-oidc/issues/1006)) ([2e21ae7](https://github.com/chanzuckerberg/aws-oidc/commit/2e21ae7caaa1625c93ddaa5cac5d9d6490d4aaaf))
* bump github.com/aws/aws-sdk-go from 1.51.15 to 1.51.16 ([#1008](https://github.com/chanzuckerberg/aws-oidc/issues/1008)) ([4eee46e](https://github.com/chanzuckerberg/aws-oidc/commit/4eee46e22911b6b13a3fb9e8b7b165e26544a652))
* bump github.com/aws/aws-sdk-go from 1.51.16 to 1.51.17 ([#1009](https://github.com/chanzuckerberg/aws-oidc/issues/1009)) ([5ae0f1f](https://github.com/chanzuckerberg/aws-oidc/commit/5ae0f1f5a4fab52ba53c5bf8a00514ef92f1e75a))
* bump github.com/aws/aws-sdk-go from 1.51.17 to 1.51.18 ([#1010](https://github.com/chanzuckerberg/aws-oidc/issues/1010)) ([db55dbf](https://github.com/chanzuckerberg/aws-oidc/commit/db55dbfd8993bf65b3dc26d8ebd6b8ffb9f2b66f))
* bump github.com/aws/aws-sdk-go from 1.51.2 to 1.51.3 ([#993](https://github.com/chanzuckerberg/aws-oidc/issues/993)) ([eea9e85](https://github.com/chanzuckerberg/aws-oidc/commit/eea9e851107d0a30cb899a9a89cfef3a4a4b1ed0))
* bump github.com/aws/aws-sdk-go from 1.51.3 to 1.51.4 ([#994](https://github.com/chanzuckerberg/aws-oidc/issues/994)) ([85371e1](https://github.com/chanzuckerberg/aws-oidc/commit/85371e185234aae2de5944bdbc5398d6d5b2492e))
* bump github.com/aws/aws-sdk-go from 1.51.4 to 1.51.5 ([#995](https://github.com/chanzuckerberg/aws-oidc/issues/995)) ([7d5c1e1](https://github.com/chanzuckerberg/aws-oidc/commit/7d5c1e11992a4b090c773a3786c21d2bd563783b))
* bump github.com/aws/aws-sdk-go from 1.51.5 to 1.51.6 ([#996](https://github.com/chanzuckerberg/aws-oidc/issues/996)) ([c0e2f7d](https://github.com/chanzuckerberg/aws-oidc/commit/c0e2f7d14475f65c9bf1c512547351013dcc9266))
* bump github.com/aws/aws-sdk-go from 1.51.6 to 1.51.7 ([#997](https://github.com/chanzuckerberg/aws-oidc/issues/997)) ([98cefab](https://github.com/chanzuckerberg/aws-oidc/commit/98cefab7e4611c5b1bdd72540f891826650f2a05))
* bump github.com/aws/aws-sdk-go from 1.51.7 to 1.51.8 ([#998](https://github.com/chanzuckerberg/aws-oidc/issues/998)) ([e9afcf1](https://github.com/chanzuckerberg/aws-oidc/commit/e9afcf16fdf5de2ed44a7aa46c667873e04d4689))
* bump github.com/aws/aws-sdk-go from 1.51.8 to 1.51.9 ([#999](https://github.com/chanzuckerberg/aws-oidc/issues/999)) ([9d314d2](https://github.com/chanzuckerberg/aws-oidc/commit/9d314d2d71849c189276c48482e03b58cb51729f))
* bump github.com/aws/aws-sdk-go from 1.51.9 to 1.51.10 ([#1000](https://github.com/chanzuckerberg/aws-oidc/issues/1000)) ([1ce9113](https://github.com/chanzuckerberg/aws-oidc/commit/1ce91130e5446b690f2d3373ede537673de00073))
* bump github.com/chanzuckerberg/go-misc from 1.12.0 to 2.2.0+incompatible ([#1005](https://github.com/chanzuckerberg/aws-oidc/issues/1005)) ([fbfdf8b](https://github.com/chanzuckerberg/aws-oidc/commit/fbfdf8b5652a6a0f2ec9fddf6df0408bc98ae4d6))
* bump github.com/honeycombio/beeline-go from 1.15.0 to 1.16.0 ([#1007](https://github.com/chanzuckerberg/aws-oidc/issues/1007)) ([fcd6511](https://github.com/chanzuckerberg/aws-oidc/commit/fcd65113fadae9e3464ba2c5dc78186c29ff9359))

## [0.28.65](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.64...v0.28.65) (2024-03-18)


### Misc

* bump github.com/aws/aws-sdk-go from 1.51.0 to 1.51.1 ([#989](https://github.com/chanzuckerberg/aws-oidc/issues/989)) ([93dc8e3](https://github.com/chanzuckerberg/aws-oidc/commit/93dc8e3f733fba8ff00396afd82697c0e0a52e39))

## [0.28.64](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.63...v0.28.64) (2024-03-15)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.38 to 1.51.0 ([#987](https://github.com/chanzuckerberg/aws-oidc/issues/987)) ([96e07a3](https://github.com/chanzuckerberg/aws-oidc/commit/96e07a3a1e3108fd31f46861129d9d351df72dd1))

## [0.28.63](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.62...v0.28.63) (2024-03-14)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.37 to 1.50.38 ([#986](https://github.com/chanzuckerberg/aws-oidc/issues/986)) ([97eba39](https://github.com/chanzuckerberg/aws-oidc/commit/97eba39a637a87aadadf608d5b160f0502b24091))
* bump google.golang.org/protobuf from 1.32.0 to 1.33.0 ([#984](https://github.com/chanzuckerberg/aws-oidc/issues/984)) ([ed1f7de](https://github.com/chanzuckerberg/aws-oidc/commit/ed1f7deddab36a22ebdfe8446faa1563e9d29e86))

## [0.28.62](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.61...v0.28.62) (2024-03-13)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.36 to 1.50.37 ([#982](https://github.com/chanzuckerberg/aws-oidc/issues/982)) ([34e2007](https://github.com/chanzuckerberg/aws-oidc/commit/34e20073f12f5cc015f1d18d9e430bd1b94af128))

## [0.28.61](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.60...v0.28.61) (2024-03-12)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.35 to 1.50.36 ([#980](https://github.com/chanzuckerberg/aws-oidc/issues/980)) ([d34b29d](https://github.com/chanzuckerberg/aws-oidc/commit/d34b29dc91c5cee51b636617dc9fcb36d6207322))

## [0.28.60](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.59...v0.28.60) (2024-03-11)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.34 to 1.50.35 ([#978](https://github.com/chanzuckerberg/aws-oidc/issues/978)) ([9914363](https://github.com/chanzuckerberg/aws-oidc/commit/9914363ce14c7c7907776861e0e6675f58bb2da9))

## [0.28.59](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.58...v0.28.59) (2024-03-08)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.33 to 1.50.34 ([#977](https://github.com/chanzuckerberg/aws-oidc/issues/977)) ([68b0365](https://github.com/chanzuckerberg/aws-oidc/commit/68b0365b5843f12ea6edae31065bb886bfa508da))
* bump github.com/go-jose/go-jose/v3 from 3.0.1 to 3.0.3 ([#975](https://github.com/chanzuckerberg/aws-oidc/issues/975)) ([bbdb8c6](https://github.com/chanzuckerberg/aws-oidc/commit/bbdb8c60c03bdb10d70df1f98b65fffa40d83e58))

## [0.28.58](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.57...v0.28.58) (2024-03-07)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.32 to 1.50.33 ([#973](https://github.com/chanzuckerberg/aws-oidc/issues/973)) ([da37c0c](https://github.com/chanzuckerberg/aws-oidc/commit/da37c0c7ed487766625a0770cd636e45a0438809))

## [0.28.57](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.56...v0.28.57) (2024-03-06)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.30 to 1.50.31 ([#969](https://github.com/chanzuckerberg/aws-oidc/issues/969)) ([cf957f7](https://github.com/chanzuckerberg/aws-oidc/commit/cf957f715c158904595d6f1a0a850bdab56d767a))
* bump github.com/aws/aws-sdk-go from 1.50.31 to 1.50.32 ([#971](https://github.com/chanzuckerberg/aws-oidc/issues/971)) ([2a6272b](https://github.com/chanzuckerberg/aws-oidc/commit/2a6272bafbf18532956f8fff61b2660cefd5295a))
* bump github.com/honeycombio/beeline-go from 1.14.0 to 1.15.0 ([#972](https://github.com/chanzuckerberg/aws-oidc/issues/972)) ([143998e](https://github.com/chanzuckerberg/aws-oidc/commit/143998e439ac42f94068d6d55403d664680d3583))

## [0.28.56](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.55...v0.28.56) (2024-03-04)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.29 to 1.50.30 ([#966](https://github.com/chanzuckerberg/aws-oidc/issues/966)) ([09fb25a](https://github.com/chanzuckerberg/aws-oidc/commit/09fb25aebc16df77ca96dce9e5eedeccc9ba93a2))
* bump github.com/stretchr/testify from 1.8.4 to 1.9.0 ([#967](https://github.com/chanzuckerberg/aws-oidc/issues/967)) ([e1ef077](https://github.com/chanzuckerberg/aws-oidc/commit/e1ef077a38c15a2a9addba631c8f155b55e3f5b0))

## [0.28.55](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.54...v0.28.55) (2024-03-01)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.28 to 1.50.29 ([#964](https://github.com/chanzuckerberg/aws-oidc/issues/964)) ([a83c28f](https://github.com/chanzuckerberg/aws-oidc/commit/a83c28fd90bee42bb6b9a09eac68963b146dc17e))

## [0.28.54](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.53...v0.28.54) (2024-02-29)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.27 to 1.50.28 ([#962](https://github.com/chanzuckerberg/aws-oidc/issues/962)) ([19448b3](https://github.com/chanzuckerberg/aws-oidc/commit/19448b34174e5beb1d879ebaaf22b8cf81a31617))

## [0.28.53](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.52...v0.28.53) (2024-02-28)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.26 to 1.50.27 ([#960](https://github.com/chanzuckerberg/aws-oidc/issues/960)) ([e29bd00](https://github.com/chanzuckerberg/aws-oidc/commit/e29bd00d7ad2cbeba757b2f64e96df4aa6a28a64))

## [0.28.52](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.51...v0.28.52) (2024-02-27)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.25 to 1.50.26 ([#958](https://github.com/chanzuckerberg/aws-oidc/issues/958)) ([a200c5a](https://github.com/chanzuckerberg/aws-oidc/commit/a200c5a0095fbb9382cbfe33640d3f99d9a3f4df))

## [0.28.51](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.50...v0.28.51) (2024-02-26)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.23 to 1.50.24 ([#955](https://github.com/chanzuckerberg/aws-oidc/issues/955)) ([48ea93c](https://github.com/chanzuckerberg/aws-oidc/commit/48ea93cae1ce479a8c698c5d7b1e007da8056ec8))
* bump github.com/aws/aws-sdk-go from 1.50.24 to 1.50.25 ([#957](https://github.com/chanzuckerberg/aws-oidc/issues/957)) ([f7df1db](https://github.com/chanzuckerberg/aws-oidc/commit/f7df1db106788369a2496910ad586bf62fd78a7a))

## [0.28.50](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.49...v0.28.50) (2024-02-22)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.22 to 1.50.23 ([#952](https://github.com/chanzuckerberg/aws-oidc/issues/952)) ([29f1ff3](https://github.com/chanzuckerberg/aws-oidc/commit/29f1ff39f4e1db6235e25581ae36d9876a760f97))
* bump github.com/chanzuckerberg/go-misc from 1.11.1 to 1.12.0 ([#953](https://github.com/chanzuckerberg/aws-oidc/issues/953)) ([b6c790d](https://github.com/chanzuckerberg/aws-oidc/commit/b6c790d656dc49263450320146430356e8a3be92))

## [0.28.49](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.48...v0.28.49) (2024-02-21)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.18 to 1.50.19 ([#947](https://github.com/chanzuckerberg/aws-oidc/issues/947)) ([e507540](https://github.com/chanzuckerberg/aws-oidc/commit/e5075405d89bdbf1d82bd0501e5b45ca548ed5b5))
* bump github.com/aws/aws-sdk-go from 1.50.19 to 1.50.20 ([#949](https://github.com/chanzuckerberg/aws-oidc/issues/949)) ([5e2b79a](https://github.com/chanzuckerberg/aws-oidc/commit/5e2b79a4888015b133c50287db7dacb54536c590))
* bump github.com/aws/aws-sdk-go from 1.50.20 to 1.50.21 ([#950](https://github.com/chanzuckerberg/aws-oidc/issues/950)) ([a2b44ca](https://github.com/chanzuckerberg/aws-oidc/commit/a2b44ca4e2c04d78535f8c280335a12bda699d97))
* bump github.com/aws/aws-sdk-go from 1.50.21 to 1.50.22 ([#951](https://github.com/chanzuckerberg/aws-oidc/issues/951)) ([5144252](https://github.com/chanzuckerberg/aws-oidc/commit/5144252a7d7494bf09a1cca18ba4d1347765ed3e))

## [0.28.48](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.47...v0.28.48) (2024-02-15)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.15 to 1.50.16 ([#943](https://github.com/chanzuckerberg/aws-oidc/issues/943)) ([e08f7e4](https://github.com/chanzuckerberg/aws-oidc/commit/e08f7e4f416e8a33c0c1070db5a6891aac03c210))
* bump github.com/aws/aws-sdk-go from 1.50.16 to 1.50.17 ([#945](https://github.com/chanzuckerberg/aws-oidc/issues/945)) ([bb9ef58](https://github.com/chanzuckerberg/aws-oidc/commit/bb9ef58f6951c49217eeee6835ce8b083fb676b3))
* bump github.com/aws/aws-sdk-go from 1.50.17 to 1.50.18 ([#946](https://github.com/chanzuckerberg/aws-oidc/issues/946)) ([7fe9048](https://github.com/chanzuckerberg/aws-oidc/commit/7fe90486dafe24b630c5f6fc17d8592d40ca93ad))

## [0.28.47](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.46...v0.28.47) (2024-02-12)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.14 to 1.50.15 ([#941](https://github.com/chanzuckerberg/aws-oidc/issues/941)) ([efe76ec](https://github.com/chanzuckerberg/aws-oidc/commit/efe76ec8406adcd4ea6642c304fca8ce46ad9f9f))

## [0.28.46](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.45...v0.28.46) (2024-02-09)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.13 to 1.50.14 ([#939](https://github.com/chanzuckerberg/aws-oidc/issues/939)) ([21a4300](https://github.com/chanzuckerberg/aws-oidc/commit/21a4300f86b71ebec1dd7f9b09fa9dbc089f9f30))

## [0.28.45](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.44...v0.28.45) (2024-02-08)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.10 to 1.50.11 ([#935](https://github.com/chanzuckerberg/aws-oidc/issues/935)) ([f63d759](https://github.com/chanzuckerberg/aws-oidc/commit/f63d759ca8761bbefcabe7a37a894a90131edce5))
* bump github.com/aws/aws-sdk-go from 1.50.11 to 1.50.12 ([#937](https://github.com/chanzuckerberg/aws-oidc/issues/937)) ([8352cf6](https://github.com/chanzuckerberg/aws-oidc/commit/8352cf63ef3df95af22f27223e2a3783be3e6428))
* bump github.com/aws/aws-sdk-go from 1.50.12 to 1.50.13 ([#938](https://github.com/chanzuckerberg/aws-oidc/issues/938)) ([27a8719](https://github.com/chanzuckerberg/aws-oidc/commit/27a87194090d41d85e2aa90b380a400df06afcca))

## [0.28.44](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.43...v0.28.44) (2024-02-05)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.9 to 1.50.10 ([#933](https://github.com/chanzuckerberg/aws-oidc/issues/933)) ([172e9bd](https://github.com/chanzuckerberg/aws-oidc/commit/172e9bde8a4d8308734ff4b7accbf13994d1392d))

## [0.28.43](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.42...v0.28.43) (2024-02-02)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.8 to 1.50.9 ([#931](https://github.com/chanzuckerberg/aws-oidc/issues/931)) ([fe23b97](https://github.com/chanzuckerberg/aws-oidc/commit/fe23b97e1c24a9a2c0852750d48e5de2b4279e84))

## [0.28.42](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.41...v0.28.42) (2024-02-01)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.7 to 1.50.8 ([#929](https://github.com/chanzuckerberg/aws-oidc/issues/929)) ([7f9651b](https://github.com/chanzuckerberg/aws-oidc/commit/7f9651b717d551a00791c666b45fe3b0069da918))

## [0.28.41](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.40...v0.28.41) (2024-01-31)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.6 to 1.50.7 ([#927](https://github.com/chanzuckerberg/aws-oidc/issues/927)) ([0372c74](https://github.com/chanzuckerberg/aws-oidc/commit/0372c74a49c5dbcb096130003296ffcfa0617c8d))

## [0.28.40](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.39...v0.28.40) (2024-01-30)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.5 to 1.50.6 ([#925](https://github.com/chanzuckerberg/aws-oidc/issues/925)) ([2fdca77](https://github.com/chanzuckerberg/aws-oidc/commit/2fdca7790429d22c0ad5cc5cab275866d4a61a47))

## [0.28.39](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.38...v0.28.39) (2024-01-29)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.4 to 1.50.5 ([#922](https://github.com/chanzuckerberg/aws-oidc/issues/922)) ([ad80e1c](https://github.com/chanzuckerberg/aws-oidc/commit/ad80e1ce1bb5bf85e1002e63229518f98e269fb7))
* bump github.com/chanzuckerberg/go-misc from 1.11.0 to 1.11.1 ([#923](https://github.com/chanzuckerberg/aws-oidc/issues/923)) ([be53f71](https://github.com/chanzuckerberg/aws-oidc/commit/be53f7184c8b99e01e7f4eea5bd40e19b9f4c871))

## [0.28.38](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.37...v0.28.38) (2024-01-26)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.3 to 1.50.4 ([#920](https://github.com/chanzuckerberg/aws-oidc/issues/920)) ([ceb7c5a](https://github.com/chanzuckerberg/aws-oidc/commit/ceb7c5af7a249e7e365f71cb5f556e1c880d5356))

## [0.28.37](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.36...v0.28.37) (2024-01-25)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.2 to 1.50.3 ([#918](https://github.com/chanzuckerberg/aws-oidc/issues/918)) ([21df5df](https://github.com/chanzuckerberg/aws-oidc/commit/21df5dfdcbc132bea0f78072a2a6a889e2904309))

## [0.28.36](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.35...v0.28.36) (2024-01-24)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.1 to 1.50.2 ([#916](https://github.com/chanzuckerberg/aws-oidc/issues/916)) ([921aad7](https://github.com/chanzuckerberg/aws-oidc/commit/921aad71f9f06fec75474d530918dec1a27ce640))

## [0.28.35](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.34...v0.28.35) (2024-01-23)


### Misc

* bump github.com/aws/aws-sdk-go from 1.50.0 to 1.50.1 ([#914](https://github.com/chanzuckerberg/aws-oidc/issues/914)) ([6114518](https://github.com/chanzuckerberg/aws-oidc/commit/611451814d75dc000b50c74673bd4a646ef06b64))

## [0.28.34](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.33...v0.28.34) (2024-01-22)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.24 to 1.50.0 ([#912](https://github.com/chanzuckerberg/aws-oidc/issues/912)) ([935c56d](https://github.com/chanzuckerberg/aws-oidc/commit/935c56db61e1782f0c7b22287fc458f42967df0f))

## [0.28.33](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.32...v0.28.33) (2024-01-19)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.23 to 1.49.24 ([#910](https://github.com/chanzuckerberg/aws-oidc/issues/910)) ([7f494b1](https://github.com/chanzuckerberg/aws-oidc/commit/7f494b158ffecff41a2d88e3d81419d3ea420d48))

## [0.28.32](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.31...v0.28.32) (2024-01-18)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.21 to 1.49.22 ([#907](https://github.com/chanzuckerberg/aws-oidc/issues/907)) ([6b5ac76](https://github.com/chanzuckerberg/aws-oidc/commit/6b5ac7624e12a55cefd96876fbd13f7a1cf804d8))
* bump github.com/aws/aws-sdk-go from 1.49.22 to 1.49.23 ([#909](https://github.com/chanzuckerberg/aws-oidc/issues/909)) ([f94e2e2](https://github.com/chanzuckerberg/aws-oidc/commit/f94e2e24f4e4999f5487de805f198669e159ad21))

## [0.28.31](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.30...v0.28.31) (2024-01-15)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.19 to 1.49.21 ([#905](https://github.com/chanzuckerberg/aws-oidc/issues/905)) ([c56a88d](https://github.com/chanzuckerberg/aws-oidc/commit/c56a88d1e8251d2d79be326d7eb06988541698c3))

## [0.28.30](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.29...v0.28.30) (2024-01-12)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.18 to 1.49.19 ([#903](https://github.com/chanzuckerberg/aws-oidc/issues/903)) ([2e252ad](https://github.com/chanzuckerberg/aws-oidc/commit/2e252adda10054cdc1c43a411e2a03fbe8275a78))

## [0.28.29](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.28...v0.28.29) (2024-01-11)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.17 to 1.49.18 ([#901](https://github.com/chanzuckerberg/aws-oidc/issues/901)) ([5bf9d1b](https://github.com/chanzuckerberg/aws-oidc/commit/5bf9d1b51269422f3c87acd59c787b136aaa3aa9))

## [0.28.28](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.27...v0.28.28) (2024-01-09)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.16 to 1.49.17 ([#899](https://github.com/chanzuckerberg/aws-oidc/issues/899)) ([2a5b7ac](https://github.com/chanzuckerberg/aws-oidc/commit/2a5b7acfe3af39a42eb5c54c16e0404f2064b49b))

## [0.28.27](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.26...v0.28.27) (2024-01-08)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.15 to 1.49.16 ([#897](https://github.com/chanzuckerberg/aws-oidc/issues/897)) ([1d4646f](https://github.com/chanzuckerberg/aws-oidc/commit/1d4646f2eaedf6c841091233eae057ca50a4ff66))

## [0.28.26](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.25...v0.28.26) (2024-01-05)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.14 to 1.49.15 ([#895](https://github.com/chanzuckerberg/aws-oidc/issues/895)) ([689d82a](https://github.com/chanzuckerberg/aws-oidc/commit/689d82a56f68d0e87bb1ed7ebcade2fddd3ec4fd))

## [0.28.25](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.24...v0.28.25) (2024-01-04)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.13 to 1.49.14 ([#893](https://github.com/chanzuckerberg/aws-oidc/issues/893)) ([5985a74](https://github.com/chanzuckerberg/aws-oidc/commit/5985a74c401e864e176eff30488fafc21fadc83b))

## [0.28.24](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.23...v0.28.24) (2024-01-01)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.10 to 1.49.11 ([#890](https://github.com/chanzuckerberg/aws-oidc/issues/890)) ([98c43ff](https://github.com/chanzuckerberg/aws-oidc/commit/98c43ff435ddb54544ba4876f5bdc318808b0d56))
* bump github.com/aws/aws-sdk-go from 1.49.11 to 1.49.12 ([#891](https://github.com/chanzuckerberg/aws-oidc/issues/891)) ([091a8c8](https://github.com/chanzuckerberg/aws-oidc/commit/091a8c8ed5ad7c548f76974863dd990d69872604))
* bump github.com/aws/aws-sdk-go from 1.49.12 to 1.49.13 ([#892](https://github.com/chanzuckerberg/aws-oidc/issues/892)) ([322035b](https://github.com/chanzuckerberg/aws-oidc/commit/322035b565e18c99f0f5a45e0155aca5f520be29))
* bump github.com/aws/aws-sdk-go from 1.49.3 to 1.49.4 ([#881](https://github.com/chanzuckerberg/aws-oidc/issues/881)) ([6dbc93f](https://github.com/chanzuckerberg/aws-oidc/commit/6dbc93f558842d7bec7d27257a67a104e98a8318))
* bump github.com/aws/aws-sdk-go from 1.49.4 to 1.49.5 ([#884](https://github.com/chanzuckerberg/aws-oidc/issues/884)) ([4b09c80](https://github.com/chanzuckerberg/aws-oidc/commit/4b09c80cc53dcae1fc379e7693261b2153db080a))
* bump github.com/aws/aws-sdk-go from 1.49.5 to 1.49.6 ([#885](https://github.com/chanzuckerberg/aws-oidc/issues/885)) ([256cfc6](https://github.com/chanzuckerberg/aws-oidc/commit/256cfc61f6c6d339efb27483cb476057b1c4715e))
* bump github.com/aws/aws-sdk-go from 1.49.6 to 1.49.7 ([#886](https://github.com/chanzuckerberg/aws-oidc/issues/886)) ([c8407ea](https://github.com/chanzuckerberg/aws-oidc/commit/c8407ea4c9239cc722d874b94a829554790171c3))
* bump github.com/aws/aws-sdk-go from 1.49.7 to 1.49.8 ([#887](https://github.com/chanzuckerberg/aws-oidc/issues/887)) ([04a5fd9](https://github.com/chanzuckerberg/aws-oidc/commit/04a5fd950a059554f0c1c580849a49b527ac1df5))
* bump github.com/aws/aws-sdk-go from 1.49.8 to 1.49.9 ([#888](https://github.com/chanzuckerberg/aws-oidc/issues/888)) ([89f4603](https://github.com/chanzuckerberg/aws-oidc/commit/89f4603760a0bef0202ae944b9993fb1c77ad8d0))
* bump github.com/aws/aws-sdk-go from 1.49.9 to 1.49.10 ([#889](https://github.com/chanzuckerberg/aws-oidc/issues/889)) ([e22f8ff](https://github.com/chanzuckerberg/aws-oidc/commit/e22f8ffc56ff1250700e069df5c9510355b273db))
* bump golang.org/x/crypto from 0.16.0 to 0.17.0 ([#883](https://github.com/chanzuckerberg/aws-oidc/issues/883)) ([4e905fa](https://github.com/chanzuckerberg/aws-oidc/commit/4e905fade503909ae193d500b5bd9e78b7e42240))

## [0.28.23](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.22...v0.28.23) (2023-12-15)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.2 to 1.49.3 ([#879](https://github.com/chanzuckerberg/aws-oidc/issues/879)) ([2b10b7a](https://github.com/chanzuckerberg/aws-oidc/commit/2b10b7a1f5d9630ce4f1b8c846dcdf95bb581d02))

## [0.28.22](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.21...v0.28.22) (2023-12-14)


### Misc

* bump github.com/aws/aws-sdk-go from 1.49.0 to 1.49.1 ([#876](https://github.com/chanzuckerberg/aws-oidc/issues/876)) ([91cb6eb](https://github.com/chanzuckerberg/aws-oidc/commit/91cb6ebcd94161440b5556f20b5bdb162e8d0376))
* bump github.com/aws/aws-sdk-go from 1.49.1 to 1.49.2 ([#878](https://github.com/chanzuckerberg/aws-oidc/issues/878)) ([b0052f6](https://github.com/chanzuckerberg/aws-oidc/commit/b0052f6ee9b66e1ca48055858c1ca23709fe6795))

## [0.28.21](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.20...v0.28.21) (2023-12-12)


### Misc

* bump github.com/aws/aws-sdk-go from 1.48.16 to 1.49.0 ([#874](https://github.com/chanzuckerberg/aws-oidc/issues/874)) ([ab97c3e](https://github.com/chanzuckerberg/aws-oidc/commit/ab97c3e45090976e6c7c3e6adc240fd0309728ec))

## [0.28.20](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.19...v0.28.20) (2023-12-11)


### Misc

* bump github.com/aws/aws-sdk-go from 1.48.14 to 1.48.15 ([#871](https://github.com/chanzuckerberg/aws-oidc/issues/871)) ([4f3bb02](https://github.com/chanzuckerberg/aws-oidc/commit/4f3bb0224bfd6682a02be68677501f4fdc7c8345))
* bump github.com/aws/aws-sdk-go from 1.48.15 to 1.48.16 ([#873](https://github.com/chanzuckerberg/aws-oidc/issues/873)) ([7e80723](https://github.com/chanzuckerberg/aws-oidc/commit/7e807232ee31852903e69a055a79f58bb2f461e2))

## [0.28.19](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.18...v0.28.19) (2023-12-07)


### Misc

* bump github.com/aws/aws-sdk-go from 1.48.10 to 1.48.11 ([#865](https://github.com/chanzuckerberg/aws-oidc/issues/865)) ([e5cc8d2](https://github.com/chanzuckerberg/aws-oidc/commit/e5cc8d241d7102a26d93ae422f97d4d753939f9e))
* bump github.com/aws/aws-sdk-go from 1.48.11 to 1.48.12 ([#866](https://github.com/chanzuckerberg/aws-oidc/issues/866)) ([3c0bd58](https://github.com/chanzuckerberg/aws-oidc/commit/3c0bd58b28a18bc56f577383b0f1ffd67c4655b3))
* bump github.com/aws/aws-sdk-go from 1.48.12 to 1.48.13 ([#868](https://github.com/chanzuckerberg/aws-oidc/issues/868)) ([ca18cd4](https://github.com/chanzuckerberg/aws-oidc/commit/ca18cd478200b733d131e02754f4a524d7f34e28))
* bump github.com/aws/aws-sdk-go from 1.48.13 to 1.48.14 ([#869](https://github.com/chanzuckerberg/aws-oidc/issues/869)) ([639c237](https://github.com/chanzuckerberg/aws-oidc/commit/639c2376b179e792cf79f6bee157bb58108b45ac))
* bump github.com/aws/aws-sdk-go from 1.48.3 to 1.48.4 ([#859](https://github.com/chanzuckerberg/aws-oidc/issues/859)) ([eefcae8](https://github.com/chanzuckerberg/aws-oidc/commit/eefcae844f1d9fe3b85912b5aab84c85ba4f1e84))
* bump github.com/aws/aws-sdk-go from 1.48.4 to 1.48.6 ([#861](https://github.com/chanzuckerberg/aws-oidc/issues/861)) ([06e109d](https://github.com/chanzuckerberg/aws-oidc/commit/06e109d7cc2ede975317f17597f4e4952f28f049))
* bump github.com/aws/aws-sdk-go from 1.48.6 to 1.48.7 ([#862](https://github.com/chanzuckerberg/aws-oidc/issues/862)) ([cee2679](https://github.com/chanzuckerberg/aws-oidc/commit/cee26793e3a39b077faa274ebb55d9da929f2230))
* bump github.com/aws/aws-sdk-go from 1.48.7 to 1.48.9 ([#863](https://github.com/chanzuckerberg/aws-oidc/issues/863)) ([43f9266](https://github.com/chanzuckerberg/aws-oidc/commit/43f92668a98c65e1e29f253bc182f39669454185))
* bump github.com/aws/aws-sdk-go from 1.48.9 to 1.48.10 ([#864](https://github.com/chanzuckerberg/aws-oidc/issues/864)) ([0ceff25](https://github.com/chanzuckerberg/aws-oidc/commit/0ceff25a512e959a2cbef5ecb47bd39bb10a4038))
* bump github.com/chanzuckerberg/go-misc from 1.10.14 to 1.11.0 ([#870](https://github.com/chanzuckerberg/aws-oidc/issues/870)) ([6440fb4](https://github.com/chanzuckerberg/aws-oidc/commit/6440fb42d8e3764e8db95ba2ead3b270de4ef9ed))
* bump github.com/honeycombio/beeline-go from 1.13.0 to 1.14.0 ([#867](https://github.com/chanzuckerberg/aws-oidc/issues/867)) ([7e8bcab](https://github.com/chanzuckerberg/aws-oidc/commit/7e8bcabb4ddbe4d22f78efd2cd68a2f3d42736c4))

## [0.28.18](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.17...v0.28.18) (2023-11-23)


### Misc

* bump github.com/aws/aws-sdk-go from 1.48.2 to 1.48.3 ([#857](https://github.com/chanzuckerberg/aws-oidc/issues/857)) ([b67b92e](https://github.com/chanzuckerberg/aws-oidc/commit/b67b92e69cdd631fc33e53faf18f1d2d286a8e60))

## [0.28.17](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.16...v0.28.17) (2023-11-22)


### Misc

* bump github.com/aws/aws-sdk-go from 1.47.13 to 1.48.0 ([#851](https://github.com/chanzuckerberg/aws-oidc/issues/851)) ([b0b32d1](https://github.com/chanzuckerberg/aws-oidc/commit/b0b32d142da7a436140642b4d3804f09ab3affaa))
* bump github.com/aws/aws-sdk-go from 1.48.0 to 1.48.1 ([#854](https://github.com/chanzuckerberg/aws-oidc/issues/854)) ([cf0964a](https://github.com/chanzuckerberg/aws-oidc/commit/cf0964a644939a67f410dcb6fa892439f48dd866))
* bump github.com/aws/aws-sdk-go from 1.48.1 to 1.48.2 ([#856](https://github.com/chanzuckerberg/aws-oidc/issues/856)) ([ba5511f](https://github.com/chanzuckerberg/aws-oidc/commit/ba5511f64ead2736f06ad267536e2bef98a9fe7d))
* bump github.com/chanzuckerberg/go-misc from 1.10.13 to 1.10.14 ([#853](https://github.com/chanzuckerberg/aws-oidc/issues/853)) ([221ebc7](https://github.com/chanzuckerberg/aws-oidc/commit/221ebc7eb8dd48fff66abe1918a5770737628d9f))
* bump github.com/go-jose/go-jose/v3 from 3.0.0 to 3.0.1 ([#855](https://github.com/chanzuckerberg/aws-oidc/issues/855)) ([ea36f17](https://github.com/chanzuckerberg/aws-oidc/commit/ea36f17ca5ca21bc0cf8d6f06123510d8d98987c))

## [0.28.16](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.15...v0.28.16) (2023-11-17)


### Misc

* bump github.com/aws/aws-sdk-go from 1.47.12 to 1.47.13 ([#849](https://github.com/chanzuckerberg/aws-oidc/issues/849)) ([7226ac5](https://github.com/chanzuckerberg/aws-oidc/commit/7226ac52f0a98101ebecfcfaa63dc9d5744f9d9f))

## [0.28.15](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.14...v0.28.15) (2023-11-16)


### Misc

* bump github.com/aws/aws-sdk-go from 1.47.10 to 1.47.11 ([#846](https://github.com/chanzuckerberg/aws-oidc/issues/846)) ([d33bb25](https://github.com/chanzuckerberg/aws-oidc/commit/d33bb2527278e8f3a193769fb4d986fcc13c6f6b))
* bump github.com/aws/aws-sdk-go from 1.47.11 to 1.47.12 ([#847](https://github.com/chanzuckerberg/aws-oidc/issues/847)) ([d6f0e28](https://github.com/chanzuckerberg/aws-oidc/commit/d6f0e28b7d10fe96d825a56c79bef262ed9e0052))
* bump github.com/aws/aws-sdk-go from 1.47.2 to 1.47.3 ([#837](https://github.com/chanzuckerberg/aws-oidc/issues/837)) ([4dd0f85](https://github.com/chanzuckerberg/aws-oidc/commit/4dd0f85d555dcc3b309bfdd5c15a9f283e6b0f6a))
* bump github.com/aws/aws-sdk-go from 1.47.3 to 1.47.4 ([#839](https://github.com/chanzuckerberg/aws-oidc/issues/839)) ([cd9df82](https://github.com/chanzuckerberg/aws-oidc/commit/cd9df824063680b595f20d0d8455511c7aa8ec86))
* bump github.com/aws/aws-sdk-go from 1.47.4 to 1.47.5 ([#840](https://github.com/chanzuckerberg/aws-oidc/issues/840)) ([81f6604](https://github.com/chanzuckerberg/aws-oidc/commit/81f6604692014f01306938d9e50ec8e70ac91395))
* bump github.com/aws/aws-sdk-go from 1.47.5 to 1.47.7 ([#841](https://github.com/chanzuckerberg/aws-oidc/issues/841)) ([480c8b5](https://github.com/chanzuckerberg/aws-oidc/commit/480c8b5b9cfd126b41af5d89c7e4ebfcd91ff45e))
* bump github.com/aws/aws-sdk-go from 1.47.7 to 1.47.8 ([#842](https://github.com/chanzuckerberg/aws-oidc/issues/842)) ([93cbe1e](https://github.com/chanzuckerberg/aws-oidc/commit/93cbe1eee29f58ca1baaa058427cf6bf603eab7a))
* bump github.com/aws/aws-sdk-go from 1.47.8 to 1.47.9 ([#843](https://github.com/chanzuckerberg/aws-oidc/issues/843)) ([86c753b](https://github.com/chanzuckerberg/aws-oidc/commit/86c753b2f6c5a384255d4be97d072de906692990))
* bump github.com/aws/aws-sdk-go from 1.47.9 to 1.47.10 ([#844](https://github.com/chanzuckerberg/aws-oidc/issues/844)) ([e45f328](https://github.com/chanzuckerberg/aws-oidc/commit/e45f328a08b609b0b77922181d7e7cad5e50f32b))
* bump github.com/chanzuckerberg/go-misc from 1.10.11 to 1.10.12 ([#845](https://github.com/chanzuckerberg/aws-oidc/issues/845)) ([fd9b8c9](https://github.com/chanzuckerberg/aws-oidc/commit/fd9b8c921170e121c2387f07a7934f3f4fdab9ac))
* bump github.com/chanzuckerberg/go-misc from 1.10.12 to 1.10.13 ([#848](https://github.com/chanzuckerberg/aws-oidc/issues/848)) ([b9a98ab](https://github.com/chanzuckerberg/aws-oidc/commit/b9a98abb187ac3c757cb751227258dd2041826bd))
* bump github.com/gorilla/handlers from 1.5.1 to 1.5.2 ([#835](https://github.com/chanzuckerberg/aws-oidc/issues/835)) ([93c3aab](https://github.com/chanzuckerberg/aws-oidc/commit/93c3aab1b8109c6f5976f7aa99973a97de18482e))
* bump github.com/spf13/cobra from 1.7.0 to 1.8.0 ([#836](https://github.com/chanzuckerberg/aws-oidc/issues/836)) ([5d70851](https://github.com/chanzuckerberg/aws-oidc/commit/5d70851c54473b0dbbb187d21fb5b9be3caad0fb))

## [0.28.14](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.13...v0.28.14) (2023-11-03)


### Misc

* bump github.com/aws/aws-sdk-go from 1.47.1 to 1.47.2 ([#833](https://github.com/chanzuckerberg/aws-oidc/issues/833)) ([e677b80](https://github.com/chanzuckerberg/aws-oidc/commit/e677b803a8036659921e3afe5c75e2899217d35b))

## [0.28.13](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.12...v0.28.13) (2023-11-02)


### Misc

* bump github.com/aws/aws-sdk-go from 1.46.7 to 1.47.0 ([#829](https://github.com/chanzuckerberg/aws-oidc/issues/829)) ([b65fc40](https://github.com/chanzuckerberg/aws-oidc/commit/b65fc407737a46ae979cbb574208b340233b6687))
* bump github.com/aws/aws-sdk-go from 1.47.0 to 1.47.1 ([#832](https://github.com/chanzuckerberg/aws-oidc/issues/832)) ([48b2264](https://github.com/chanzuckerberg/aws-oidc/commit/48b2264c4dbce8537c01cbed360dbcf045f01c55))
* bump github.com/chanzuckerberg/go-misc from 1.10.10 to 1.10.11 ([#831](https://github.com/chanzuckerberg/aws-oidc/issues/831)) ([8172fb2](https://github.com/chanzuckerberg/aws-oidc/commit/8172fb24860d7e6036358732289fefe581e59db7))

## [0.28.12](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.11...v0.28.12) (2023-10-31)


### Misc

* bump github.com/aws/aws-sdk-go from 1.46.4 to 1.46.5 ([#824](https://github.com/chanzuckerberg/aws-oidc/issues/824)) ([2d9613c](https://github.com/chanzuckerberg/aws-oidc/commit/2d9613c1738b65a31ec791a5107d8a09b53e5f76))
* bump github.com/aws/aws-sdk-go from 1.46.5 to 1.46.6 ([#826](https://github.com/chanzuckerberg/aws-oidc/issues/826)) ([5dd5a70](https://github.com/chanzuckerberg/aws-oidc/commit/5dd5a703dfdbfc8cd6ff0e0cfa445849d9596a74))
* bump github.com/aws/aws-sdk-go from 1.46.6 to 1.46.7 ([#828](https://github.com/chanzuckerberg/aws-oidc/issues/828)) ([bd12246](https://github.com/chanzuckerberg/aws-oidc/commit/bd122469f373b68ee78bb024eeb84a533a186916))
* bump github.com/chanzuckerberg/go-misc from 1.10.9 to 1.10.10 ([#827](https://github.com/chanzuckerberg/aws-oidc/issues/827)) ([3d2d7f1](https://github.com/chanzuckerberg/aws-oidc/commit/3d2d7f1a2bc3cd579bd79d924adaee2b6a4c4dfb))

## [0.28.11](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.10...v0.28.11) (2023-10-26)


### Misc

* bump github.com/aws/aws-sdk-go from 1.46.3 to 1.46.4 ([#823](https://github.com/chanzuckerberg/aws-oidc/issues/823)) ([afbe61f](https://github.com/chanzuckerberg/aws-oidc/commit/afbe61f0fa9d1a72e1a52b09ac7d01163d767f2b))
* bump google.golang.org/grpc from 1.57.0 to 1.57.1 ([#821](https://github.com/chanzuckerberg/aws-oidc/issues/821)) ([6849631](https://github.com/chanzuckerberg/aws-oidc/commit/6849631bb5ef620a46f7c36be50d6258de8f1593))

## [0.28.10](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.9...v0.28.10) (2023-10-25)


### Misc

* bump github.com/aws/aws-sdk-go from 1.46.0 to 1.46.2 ([#817](https://github.com/chanzuckerberg/aws-oidc/issues/817)) ([20c5914](https://github.com/chanzuckerberg/aws-oidc/commit/20c5914bf335039875f6a2a9aa06b3f3b0a3f748))
* bump github.com/aws/aws-sdk-go from 1.46.2 to 1.46.3 ([#819](https://github.com/chanzuckerberg/aws-oidc/issues/819)) ([2a7cb24](https://github.com/chanzuckerberg/aws-oidc/commit/2a7cb24a93a233e0a4063b655d7ae41ad78a8c70))
* bump github.com/chanzuckerberg/go-misc from 1.10.8 to 1.10.9 ([#820](https://github.com/chanzuckerberg/aws-oidc/issues/820)) ([2ba886a](https://github.com/chanzuckerberg/aws-oidc/commit/2ba886a9c2fd3aa4197352c7546abbda5b9a8c54))

## [0.28.9](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.8...v0.28.9) (2023-10-20)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.28 to 1.46.0 ([#815](https://github.com/chanzuckerberg/aws-oidc/issues/815)) ([077a83b](https://github.com/chanzuckerberg/aws-oidc/commit/077a83b71f2a9dd711c0976253907015271183eb))

## [0.28.8](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.7...v0.28.8) (2023-10-19)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.27 to 1.45.28 ([#813](https://github.com/chanzuckerberg/aws-oidc/issues/813)) ([15a39e7](https://github.com/chanzuckerberg/aws-oidc/commit/15a39e7f6ea759346a65a98119018d9de65e73e5))

## [0.28.7](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.6...v0.28.7) (2023-10-18)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.24 to 1.45.25 ([#808](https://github.com/chanzuckerberg/aws-oidc/issues/808)) ([ec54b0e](https://github.com/chanzuckerberg/aws-oidc/commit/ec54b0edce908c4704b969d5d23f193ace669d8f))
* bump github.com/aws/aws-sdk-go from 1.45.25 to 1.45.26 ([#810](https://github.com/chanzuckerberg/aws-oidc/issues/810)) ([30f2291](https://github.com/chanzuckerberg/aws-oidc/commit/30f2291a6f7ea973d94ba6d5ee290314cb6871fd))
* bump github.com/aws/aws-sdk-go from 1.45.26 to 1.45.27 ([#812](https://github.com/chanzuckerberg/aws-oidc/issues/812)) ([5068557](https://github.com/chanzuckerberg/aws-oidc/commit/5068557313ccec30e3c2cf46ab985e8828b53370))
* bump github.com/chanzuckerberg/go-misc from 1.10.7 to 1.10.8 ([#811](https://github.com/chanzuckerberg/aws-oidc/issues/811)) ([3bebe14](https://github.com/chanzuckerberg/aws-oidc/commit/3bebe141857c2edfb4fecd2fae6bdf92df2aa0fc))

## [0.28.6](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.5...v0.28.6) (2023-10-11)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.23 to 1.45.24 ([#805](https://github.com/chanzuckerberg/aws-oidc/issues/805)) ([2e9572a](https://github.com/chanzuckerberg/aws-oidc/commit/2e9572acc3aea10fcaca5a42cbcdb2dde9f1c370))
* bump golang.org/x/net from 0.15.0 to 0.17.0 ([#807](https://github.com/chanzuckerberg/aws-oidc/issues/807)) ([818713f](https://github.com/chanzuckerberg/aws-oidc/commit/818713f3b31642d06d8357eb2151ab8dbd2135ec))

## [0.28.5](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.4...v0.28.5) (2023-10-06)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.21 to 1.45.22 ([#802](https://github.com/chanzuckerberg/aws-oidc/issues/802)) ([4e28dad](https://github.com/chanzuckerberg/aws-oidc/commit/4e28dad6e1d3987f8de7963c0d59a6bc0769bf62))
* bump github.com/aws/aws-sdk-go from 1.45.22 to 1.45.23 ([#804](https://github.com/chanzuckerberg/aws-oidc/issues/804)) ([f519943](https://github.com/chanzuckerberg/aws-oidc/commit/f519943eabf72da894c7b9142a9cfe9580b9f133))

## [0.28.4](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.3...v0.28.4) (2023-10-04)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.20 to 1.45.21 ([#800](https://github.com/chanzuckerberg/aws-oidc/issues/800)) ([9824bf1](https://github.com/chanzuckerberg/aws-oidc/commit/9824bf17159fe6caf44bedb373ca0cdd79493dba))

## [0.28.3](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.2...v0.28.3) (2023-10-03)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.19 to 1.45.20 ([#798](https://github.com/chanzuckerberg/aws-oidc/issues/798)) ([7477b22](https://github.com/chanzuckerberg/aws-oidc/commit/7477b22607715ccadc5f961a1299d3737ce14c6b))

## [0.28.2](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.1...v0.28.2) (2023-09-29)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.18 to 1.45.19 ([#796](https://github.com/chanzuckerberg/aws-oidc/issues/796)) ([f617607](https://github.com/chanzuckerberg/aws-oidc/commit/f61760799ded9849e4f936206174551ed72d7919))

## [0.28.1](https://github.com/chanzuckerberg/aws-oidc/compare/v0.28.0...v0.28.1) (2023-09-28)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.17 to 1.45.18 ([#794](https://github.com/chanzuckerberg/aws-oidc/issues/794)) ([d0b2889](https://github.com/chanzuckerberg/aws-oidc/commit/d0b2889272e69f0f9c6a76a1fbbad044f54a61a9))

## [0.28.0](https://github.com/chanzuckerberg/aws-oidc/compare/v0.27.0...v0.28.0) (2023-09-27)


### Features

* Downgrade go to 1.20 ([#792](https://github.com/chanzuckerberg/aws-oidc/issues/792)) ([d57d4e1](https://github.com/chanzuckerberg/aws-oidc/commit/d57d4e184366e0a340218605cf68a370b218d615))


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.12 to 1.45.13 ([#787](https://github.com/chanzuckerberg/aws-oidc/issues/787)) ([9d42752](https://github.com/chanzuckerberg/aws-oidc/commit/9d42752bf1be7b3bac580d99d981be8622a644a1))
* bump github.com/aws/aws-sdk-go from 1.45.13 to 1.45.14 ([#788](https://github.com/chanzuckerberg/aws-oidc/issues/788)) ([77c2d5e](https://github.com/chanzuckerberg/aws-oidc/commit/77c2d5e39bb4ad0f0f32bbabea4bba8df9eb300d))
* bump github.com/aws/aws-sdk-go from 1.45.14 to 1.45.15 ([#790](https://github.com/chanzuckerberg/aws-oidc/issues/790)) ([0ab7df4](https://github.com/chanzuckerberg/aws-oidc/commit/0ab7df4b92566a042856b61658d3a7b0b86239de))
* bump github.com/aws/aws-sdk-go from 1.45.15 to 1.45.16 ([#791](https://github.com/chanzuckerberg/aws-oidc/issues/791)) ([c255f29](https://github.com/chanzuckerberg/aws-oidc/commit/c255f29848da0e80b932a43fd8ef13cce9407b3f))
* bump github.com/aws/aws-sdk-go from 1.45.16 to 1.45.17 ([#793](https://github.com/chanzuckerberg/aws-oidc/issues/793)) ([b098e18](https://github.com/chanzuckerberg/aws-oidc/commit/b098e18803a9cfeb036b39df080be8ae3471b146))
* bump github.com/aws/aws-sdk-go from 1.45.5 to 1.45.12 ([#783](https://github.com/chanzuckerberg/aws-oidc/issues/783)) ([adfb562](https://github.com/chanzuckerberg/aws-oidc/commit/adfb56234bf139fa043a61879df8a21bb3284b2c))
* bump github.com/chanzuckerberg/go-misc from 1.10.5 to 1.10.6 ([#784](https://github.com/chanzuckerberg/aws-oidc/issues/784)) ([904cf8b](https://github.com/chanzuckerberg/aws-oidc/commit/904cf8b8125ec04b17a0de19ffd54a1ca9b22871))
* bump github.com/chanzuckerberg/go-misc from 1.10.6 to 1.10.7 ([#786](https://github.com/chanzuckerberg/aws-oidc/issues/786)) ([604db98](https://github.com/chanzuckerberg/aws-oidc/commit/604db984b551dcb4cd4f444295a1f54532ee0927))
* bump github.com/go-errors/errors from 1.5.0 to 1.5.1 ([#789](https://github.com/chanzuckerberg/aws-oidc/issues/789)) ([adcbe6c](https://github.com/chanzuckerberg/aws-oidc/commit/adcbe6ce2cabb2d35d9a3ce59169fa3c3c8d1f87))

## [0.27.0](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.17...v0.27.0) (2023-09-08)


### Features

* Upgrade to go 1.21 ([#782](https://github.com/chanzuckerberg/aws-oidc/issues/782)) ([5783aff](https://github.com/chanzuckerberg/aws-oidc/commit/5783aff5ffd80b3e0b9ebb43e8405ac5e01ec710))


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.3 to 1.45.4 ([#778](https://github.com/chanzuckerberg/aws-oidc/issues/778)) ([b1f9d1d](https://github.com/chanzuckerberg/aws-oidc/commit/b1f9d1d3e6ee1601fbfd1ac544409fc538a359d0))
* bump github.com/aws/aws-sdk-go from 1.45.4 to 1.45.5 ([#781](https://github.com/chanzuckerberg/aws-oidc/issues/781)) ([b48f389](https://github.com/chanzuckerberg/aws-oidc/commit/b48f3891608a7db5a7b2e27b81d089dbabb586f5))
* bump github.com/go-errors/errors from 1.4.2 to 1.5.0 ([#779](https://github.com/chanzuckerberg/aws-oidc/issues/779)) ([e0c0e0e](https://github.com/chanzuckerberg/aws-oidc/commit/e0c0e0e824dbb9c52f04a49ee7aff0437d148c34))

## [0.26.17](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.16...v0.26.17) (2023-09-06)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.2 to 1.45.3 ([#776](https://github.com/chanzuckerberg/aws-oidc/issues/776)) ([ff8ab5b](https://github.com/chanzuckerberg/aws-oidc/commit/ff8ab5b72030a624d2a2b0072dcfa50401e4f550))

## [0.26.16](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.15...v0.26.16) (2023-09-04)


### Misc

* bump github.com/aws/aws-sdk-go from 1.45.1 to 1.45.2 ([#774](https://github.com/chanzuckerberg/aws-oidc/issues/774)) ([4b2ff31](https://github.com/chanzuckerberg/aws-oidc/commit/4b2ff31660cbe0dfd1cd008ea4c0b2680c477470))

## [0.26.15](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.14...v0.26.15) (2023-09-01)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.334 to 1.45.0 ([#771](https://github.com/chanzuckerberg/aws-oidc/issues/771)) ([0dea440](https://github.com/chanzuckerberg/aws-oidc/commit/0dea4406df86ab77863786886a207492a5187a6d))
* bump github.com/aws/aws-sdk-go from 1.45.0 to 1.45.1 ([#773](https://github.com/chanzuckerberg/aws-oidc/issues/773)) ([e37ab65](https://github.com/chanzuckerberg/aws-oidc/commit/e37ab657fc6076384e4ebbb4bb4e29fea82fd68d))

## [0.26.14](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.13...v0.26.14) (2023-08-30)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.333 to 1.44.334 ([#769](https://github.com/chanzuckerberg/aws-oidc/issues/769)) ([4e110f3](https://github.com/chanzuckerberg/aws-oidc/commit/4e110f31d653a1ec6889b3bcd309c098c50b2bef))

## [0.26.13](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.12...v0.26.13) (2023-08-29)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.332 to 1.44.333 ([#767](https://github.com/chanzuckerberg/aws-oidc/issues/767)) ([5f06eff](https://github.com/chanzuckerberg/aws-oidc/commit/5f06eff05666c4efaf7a76f50835322298ef0fbe))

## [0.26.12](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.11...v0.26.12) (2023-08-28)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.330 to 1.44.331 ([#764](https://github.com/chanzuckerberg/aws-oidc/issues/764)) ([44edfd2](https://github.com/chanzuckerberg/aws-oidc/commit/44edfd2d313851dbc2543bf9b996183bf9455382))
* bump github.com/aws/aws-sdk-go from 1.44.331 to 1.44.332 ([#766](https://github.com/chanzuckerberg/aws-oidc/issues/766)) ([da235c0](https://github.com/chanzuckerberg/aws-oidc/commit/da235c0f10a3a4df3d9a75675507715f860e2e99))

## [0.26.11](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.10...v0.26.11) (2023-08-24)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.329 to 1.44.330 ([#762](https://github.com/chanzuckerberg/aws-oidc/issues/762)) ([6d0d5aa](https://github.com/chanzuckerberg/aws-oidc/commit/6d0d5aa57762ba577ddf9994e0a071c420bb9e25))

## [0.26.10](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.9...v0.26.10) (2023-08-23)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.328 to 1.44.329 ([#760](https://github.com/chanzuckerberg/aws-oidc/issues/760)) ([6f0adcc](https://github.com/chanzuckerberg/aws-oidc/commit/6f0adcc828398818a1a37fe54fa34312a2b35d8f))

## [0.26.9](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.8...v0.26.9) (2023-08-22)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.327 to 1.44.328 ([#758](https://github.com/chanzuckerberg/aws-oidc/issues/758)) ([0d28f89](https://github.com/chanzuckerberg/aws-oidc/commit/0d28f89cea711d21b5b72fc8abe17ad682301daa))

## [0.26.8](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.7...v0.26.8) (2023-08-21)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.325 to 1.44.326 ([#755](https://github.com/chanzuckerberg/aws-oidc/issues/755)) ([816e61d](https://github.com/chanzuckerberg/aws-oidc/commit/816e61d1a0b0e46547146c9f150051f04672077a))
* bump github.com/aws/aws-sdk-go from 1.44.326 to 1.44.327 ([#757](https://github.com/chanzuckerberg/aws-oidc/issues/757)) ([5a194e9](https://github.com/chanzuckerberg/aws-oidc/commit/5a194e973c044be0f2ed7c561abda90e39444a72))

## [0.26.7](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.6...v0.26.7) (2023-08-17)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.324 to 1.44.325 ([#753](https://github.com/chanzuckerberg/aws-oidc/issues/753)) ([70a7c03](https://github.com/chanzuckerberg/aws-oidc/commit/70a7c03431339b2646e8866339d7c515a2b3aa8c))

## [0.26.6](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.5...v0.26.6) (2023-08-16)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.323 to 1.44.324 ([#751](https://github.com/chanzuckerberg/aws-oidc/issues/751)) ([5396e68](https://github.com/chanzuckerberg/aws-oidc/commit/5396e68f6d22f7c626a3c92121c5cff17622790b))
* bump github.com/chanzuckerberg/go-misc from 1.10.4 to 1.10.5 ([#750](https://github.com/chanzuckerberg/aws-oidc/issues/750)) ([fd27318](https://github.com/chanzuckerberg/aws-oidc/commit/fd2731845a228ebb8c121428201fc28893f18a14))

## [0.26.5](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.4...v0.26.5) (2023-08-15)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.322 to 1.44.323 ([#748](https://github.com/chanzuckerberg/aws-oidc/issues/748)) ([922380c](https://github.com/chanzuckerberg/aws-oidc/commit/922380cf5f4261169940723b34f8c369cb85f1ae))

## [0.26.4](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.3...v0.26.4) (2023-08-14)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.321 to 1.44.322 ([#746](https://github.com/chanzuckerberg/aws-oidc/issues/746)) ([042fbd9](https://github.com/chanzuckerberg/aws-oidc/commit/042fbd9cf0035836e87fba36507ba7b10f0f2f03))

## [0.26.3](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.2...v0.26.3) (2023-08-11)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.320 to 1.44.321 ([#744](https://github.com/chanzuckerberg/aws-oidc/issues/744)) ([a4e27b2](https://github.com/chanzuckerberg/aws-oidc/commit/a4e27b2b4c8d8660f0cdaa9b237ef00159a3967a))

## [0.26.2](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.1...v0.26.2) (2023-08-10)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.319 to 1.44.320 ([#742](https://github.com/chanzuckerberg/aws-oidc/issues/742)) ([3cffbcc](https://github.com/chanzuckerberg/aws-oidc/commit/3cffbcc1fc179ce6626253f1d98567dc84891173))

## [0.26.1](https://github.com/chanzuckerberg/aws-oidc/compare/v0.26.0...v0.26.1) (2023-08-09)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.318 to 1.44.319 ([#740](https://github.com/chanzuckerberg/aws-oidc/issues/740)) ([dc62272](https://github.com/chanzuckerberg/aws-oidc/commit/dc622720478d1c181e2b85287e51e19f68e7cd92))

## [0.26.0](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.92...v0.26.0) (2023-08-08)


### Features

* allow dependabot to automerge ([#737](https://github.com/chanzuckerberg/aws-oidc/issues/737)) ([a20bce2](https://github.com/chanzuckerberg/aws-oidc/commit/a20bce2169ae6b93673e78341e7ff1b9570d2dbc))


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.317 to 1.44.318 ([#739](https://github.com/chanzuckerberg/aws-oidc/issues/739)) ([42eb492](https://github.com/chanzuckerberg/aws-oidc/commit/42eb49221ed81ae24b805e225f9777bc951feb66))

## [0.25.92](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.91...v0.25.92) (2023-08-04)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.315 to 1.44.316 ([#734](https://github.com/chanzuckerberg/aws-oidc/issues/734)) ([49f5c3e](https://github.com/chanzuckerberg/aws-oidc/commit/49f5c3edeafddbc923abdb5eee025d190f54f8b0))

## [0.25.91](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.90...v0.25.91) (2023-08-03)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.314 to 1.44.315 ([#732](https://github.com/chanzuckerberg/aws-oidc/issues/732)) ([9384956](https://github.com/chanzuckerberg/aws-oidc/commit/9384956644239515bf53a0d7f18fcb2308f847fa))

## [0.25.90](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.89...v0.25.90) (2023-08-02)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.309 to 1.44.312 ([#725](https://github.com/chanzuckerberg/aws-oidc/issues/725)) ([a2dc901](https://github.com/chanzuckerberg/aws-oidc/commit/a2dc901e0050dde66d1403806786efbc338dd664))
* bump github.com/aws/aws-sdk-go from 1.44.312 to 1.44.313 ([#730](https://github.com/chanzuckerberg/aws-oidc/issues/730)) ([360fcb0](https://github.com/chanzuckerberg/aws-oidc/commit/360fcb0368b8d464ec8ef6f4d02fe9a431454504))
* bump github.com/aws/aws-sdk-go from 1.44.313 to 1.44.314 ([#731](https://github.com/chanzuckerberg/aws-oidc/issues/731)) ([2d2602b](https://github.com/chanzuckerberg/aws-oidc/commit/2d2602b3d858ce48f0ff93afcad09290468d3264))
* bump github.com/chanzuckerberg/go-misc from 1.10.2 to 1.10.3 ([#726](https://github.com/chanzuckerberg/aws-oidc/issues/726)) ([3511cbf](https://github.com/chanzuckerberg/aws-oidc/commit/3511cbf799001db1a204cd01b7fe7a026a47385d))
* bump github.com/chanzuckerberg/go-misc from 1.10.3 to 1.10.4 ([#728](https://github.com/chanzuckerberg/aws-oidc/issues/728)) ([f13d2a2](https://github.com/chanzuckerberg/aws-oidc/commit/f13d2a20726e412c076772ce80f24916234651f5))
* bump github.com/honeycombio/beeline-go from 1.12.0 to 1.13.0 ([#729](https://github.com/chanzuckerberg/aws-oidc/issues/729)) ([0170e2f](https://github.com/chanzuckerberg/aws-oidc/commit/0170e2f26e589cff5b6c6550bbf3632280c8bd94))

## [0.25.89](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.88...v0.25.89) (2023-07-27)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.308 to 1.44.309 ([#722](https://github.com/chanzuckerberg/aws-oidc/issues/722)) ([bc82803](https://github.com/chanzuckerberg/aws-oidc/commit/bc82803409c9c98f0ab03a356b72b043be9d0151))

## [0.25.88](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.87...v0.25.88) (2023-07-26)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.306 to 1.44.307 ([#718](https://github.com/chanzuckerberg/aws-oidc/issues/718)) ([ddc9508](https://github.com/chanzuckerberg/aws-oidc/commit/ddc95082ce6d97486a7cc096f59e30feee3a384b))
* bump github.com/aws/aws-sdk-go from 1.44.307 to 1.44.308 ([#720](https://github.com/chanzuckerberg/aws-oidc/issues/720)) ([e968ddb](https://github.com/chanzuckerberg/aws-oidc/commit/e968ddb8bc120bc7ecb84c9eab924aaa8fafdda0))

## [0.25.87](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.86...v0.25.87) (2023-07-24)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.300 to 1.44.306 ([#716](https://github.com/chanzuckerberg/aws-oidc/issues/716)) ([2b548b8](https://github.com/chanzuckerberg/aws-oidc/commit/2b548b821ea04458aa9fbf0ffc9c52cfe1d647b5))
* bump github.com/okta/okta-sdk-golang/v2 from 2.19.0 to 2.20.0 ([#711](https://github.com/chanzuckerberg/aws-oidc/issues/711)) ([c9a8f92](https://github.com/chanzuckerberg/aws-oidc/commit/c9a8f926794790c354f1e30faed3c29e1d46520a))

## [0.25.86](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.85...v0.25.86) (2023-07-14)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.299 to 1.44.300 ([#709](https://github.com/chanzuckerberg/aws-oidc/issues/709)) ([fdc0ca4](https://github.com/chanzuckerberg/aws-oidc/commit/fdc0ca47c051abf9c21eefd5b5dfe42b9522ccad))

## [0.25.85](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.84...v0.25.85) (2023-07-11)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.294 to 1.44.295 ([#705](https://github.com/chanzuckerberg/aws-oidc/issues/705)) ([ee1e4ea](https://github.com/chanzuckerberg/aws-oidc/commit/ee1e4ea5beb5d4552179f1caa4306df02d5789d2))
* bump github.com/aws/aws-sdk-go from 1.44.295 to 1.44.298 ([#707](https://github.com/chanzuckerberg/aws-oidc/issues/707)) ([a6e5079](https://github.com/chanzuckerberg/aws-oidc/commit/a6e5079f1491a17e12aa8061f5eb8e08c555c370))
* bump github.com/aws/aws-sdk-go from 1.44.298 to 1.44.299 ([#708](https://github.com/chanzuckerberg/aws-oidc/issues/708)) ([d3d0b02](https://github.com/chanzuckerberg/aws-oidc/commit/d3d0b024df7967222124adf690dc9ea4b389c29f))

## [0.25.84](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.83...v0.25.84) (2023-07-03)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.291 to 1.44.294 ([#703](https://github.com/chanzuckerberg/aws-oidc/issues/703)) ([fc7eba2](https://github.com/chanzuckerberg/aws-oidc/commit/fc7eba276514bc7cf8c1d8dbb7776d5593016c39))

## [0.25.83](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.82...v0.25.83) (2023-06-28)


### Misc

* bump github.com/chanzuckerberg/go-misc from 1.10.1 to 1.10.2 ([#698](https://github.com/chanzuckerberg/aws-oidc/issues/698)) ([606d3e4](https://github.com/chanzuckerberg/aws-oidc/commit/606d3e4f2d44bc9c129d39abe607e6856aead03c))

## [0.25.82](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.81...v0.25.82) (2023-06-27)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.286 to 1.44.289 ([#695](https://github.com/chanzuckerberg/aws-oidc/issues/695)) ([05b6872](https://github.com/chanzuckerberg/aws-oidc/commit/05b6872731301cdecc72513c60c577cfa3f5b791))

## [0.25.81](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.80...v0.25.81) (2023-06-21)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.285 to 1.44.286 ([#691](https://github.com/chanzuckerberg/aws-oidc/issues/691)) ([da5e2f1](https://github.com/chanzuckerberg/aws-oidc/commit/da5e2f119ed330df66c88d36f24a9962fecd239d))

## [0.25.80](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.79...v0.25.80) (2023-06-20)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.283 to 1.44.284 ([#688](https://github.com/chanzuckerberg/aws-oidc/issues/688)) ([5c3aae4](https://github.com/chanzuckerberg/aws-oidc/commit/5c3aae498947383963e52fea14a98d03cb480be6))
* bump github.com/aws/aws-sdk-go from 1.44.284 to 1.44.285 ([#690](https://github.com/chanzuckerberg/aws-oidc/issues/690)) ([e424ecb](https://github.com/chanzuckerberg/aws-oidc/commit/e424ecb9d9d732b88e91fce98e4706c3ca75f803))

## [0.25.79](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.78...v0.25.79) (2023-06-16)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.282 to 1.44.283 ([#686](https://github.com/chanzuckerberg/aws-oidc/issues/686)) ([035dd6e](https://github.com/chanzuckerberg/aws-oidc/commit/035dd6e61bb76492b5953aa63e8a4acfec561e5c))

## [0.25.78](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.77...v0.25.78) (2023-06-15)


### Misc

* bump github.com/chanzuckerberg/go-misc from 1.10.0 to 1.10.1 ([#684](https://github.com/chanzuckerberg/aws-oidc/issues/684)) ([753177e](https://github.com/chanzuckerberg/aws-oidc/commit/753177e6fbd05af18047fc8095e5c2c0c029ee8a))

## [0.25.77](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.76...v0.25.77) (2023-06-14)


### Misc

* bump github.com/AlecAivazis/survey/v2 from 2.3.6 to 2.3.7 ([#682](https://github.com/chanzuckerberg/aws-oidc/issues/682)) ([eb639d0](https://github.com/chanzuckerberg/aws-oidc/commit/eb639d098f87499b8cd9b350ea3c88dcc4873c3e))
* bump github.com/aws/aws-sdk-go from 1.44.277 to 1.44.280 ([#679](https://github.com/chanzuckerberg/aws-oidc/issues/679)) ([5fb236f](https://github.com/chanzuckerberg/aws-oidc/commit/5fb236fcf38923cba34bf9c7f2abddcb7a09fa79))

## [0.25.76](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.75...v0.25.76) (2023-06-09)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.275 to 1.44.277 ([#676](https://github.com/chanzuckerberg/aws-oidc/issues/676)) ([1e4ac15](https://github.com/chanzuckerberg/aws-oidc/commit/1e4ac15c561adf23357817819d7ba1935fdb5464))
* bump github.com/honeycombio/beeline-go from 1.11.1 to 1.12.0 ([#677](https://github.com/chanzuckerberg/aws-oidc/issues/677)) ([b9fc7a2](https://github.com/chanzuckerberg/aws-oidc/commit/b9fc7a258278de99b813bdb5638bea024bdd2d1e))

## [0.25.75](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.74...v0.25.75) (2023-06-06)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.273 to 1.44.274 ([#671](https://github.com/chanzuckerberg/aws-oidc/issues/671)) ([410ac0c](https://github.com/chanzuckerberg/aws-oidc/commit/410ac0cc4fa6e80fe39731ac87162e125e732a27))
* bump github.com/aws/aws-sdk-go from 1.44.274 to 1.44.275 ([#674](https://github.com/chanzuckerberg/aws-oidc/issues/674)) ([c12f1e0](https://github.com/chanzuckerberg/aws-oidc/commit/c12f1e0a13e2b32de7f6bc2b9a9e8096a923ca44))
* bump github.com/okta/okta-sdk-golang/v2 from 2.18.0 to 2.19.0 ([#672](https://github.com/chanzuckerberg/aws-oidc/issues/672)) ([8a684a7](https://github.com/chanzuckerberg/aws-oidc/commit/8a684a7800c2e232075123171c0f77ce7ddce724))
* bump github.com/sirupsen/logrus from 1.9.2 to 1.9.3 ([#675](https://github.com/chanzuckerberg/aws-oidc/issues/675)) ([61c2214](https://github.com/chanzuckerberg/aws-oidc/commit/61c22144301a27d7b8aca04643978bcaae032daa))

## [0.25.74](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.73...v0.25.74) (2023-06-01)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.271 to 1.44.272 ([#668](https://github.com/chanzuckerberg/aws-oidc/issues/668)) ([48e51fd](https://github.com/chanzuckerberg/aws-oidc/commit/48e51fd170485007cadef3fbd4d88d8fe5bf4fb6))
* bump github.com/aws/aws-sdk-go from 1.44.272 to 1.44.273 ([#670](https://github.com/chanzuckerberg/aws-oidc/issues/670)) ([a068c94](https://github.com/chanzuckerberg/aws-oidc/commit/a068c949f0f80223cbf0d7f9096e0df69b5021d6))

## [0.25.73](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.72...v0.25.73) (2023-05-30)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.269 to 1.44.271 ([#665](https://github.com/chanzuckerberg/aws-oidc/issues/665)) ([1ee6eb5](https://github.com/chanzuckerberg/aws-oidc/commit/1ee6eb5025ab8bec585031aea7400d6725767194))
* bump github.com/stretchr/testify from 1.8.3 to 1.8.4 ([#666](https://github.com/chanzuckerberg/aws-oidc/issues/666)) ([97c801c](https://github.com/chanzuckerberg/aws-oidc/commit/97c801cacca3af04c782b1dd43f81a4a8a351eaf))

## [0.25.72](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.71...v0.25.72) (2023-05-26)


### Misc

* bump github.com/chanzuckerberg/go-misc from 1.0.9 to 1.10.0 ([#663](https://github.com/chanzuckerberg/aws-oidc/issues/663)) ([290deb1](https://github.com/chanzuckerberg/aws-oidc/commit/290deb1dabbba15b4d68b65cf9bf68516546b35d))

## [0.25.71](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.70...v0.25.71) (2023-05-23)


### Misc

* bump github.com/stretchr/testify from 1.8.2 to 1.8.3 ([#658](https://github.com/chanzuckerberg/aws-oidc/issues/658)) ([f46d836](https://github.com/chanzuckerberg/aws-oidc/commit/f46d83665d42e484e44fc3d286ba74556e783221))

## [0.25.70](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.69...v0.25.70) (2023-05-18)


### Misc

* bump github.com/sirupsen/logrus from 1.9.1 to 1.9.2 ([#656](https://github.com/chanzuckerberg/aws-oidc/issues/656)) ([0dcec7e](https://github.com/chanzuckerberg/aws-oidc/commit/0dcec7e87a8197260507de5a129eec4297749366))

## [0.25.69](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.68...v0.25.69) (2023-05-17)


### Misc

* bump github.com/sirupsen/logrus from 1.9.0 to 1.9.1 ([#653](https://github.com/chanzuckerberg/aws-oidc/issues/653)) ([fd036db](https://github.com/chanzuckerberg/aws-oidc/commit/fd036db7640b8b1b235dd968ac1d9ea2ba5a7057))

## [0.25.68](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.67...v0.25.68) (2023-05-11)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.258 to 1.44.259 ([#647](https://github.com/chanzuckerberg/aws-oidc/issues/647)) ([3120d7f](https://github.com/chanzuckerberg/aws-oidc/commit/3120d7f598425d3175a151929b02c025c00b173b))
* bump github.com/aws/aws-sdk-go from 1.44.259 to 1.44.260 ([#649](https://github.com/chanzuckerberg/aws-oidc/issues/649)) ([a69ed11](https://github.com/chanzuckerberg/aws-oidc/commit/a69ed11ed5b3a1d34e427e0724d2ec7333c7adef))
* bump github.com/aws/aws-sdk-go from 1.44.260 to 1.44.261 ([#650](https://github.com/chanzuckerberg/aws-oidc/issues/650)) ([cb28a4f](https://github.com/chanzuckerberg/aws-oidc/commit/cb28a4f377a881ad58386604a2600857b5c3a3e5))

## [0.25.67](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.66...v0.25.67) (2023-05-05)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.255 to 1.44.257 ([#644](https://github.com/chanzuckerberg/aws-oidc/issues/644)) ([0f7462c](https://github.com/chanzuckerberg/aws-oidc/commit/0f7462c809b792fb0f8efd6f699e2c5ca90afb8c))
* bump github.com/chanzuckerberg/go-misc from 1.0.8 to 1.0.9 ([#643](https://github.com/chanzuckerberg/aws-oidc/issues/643)) ([9281066](https://github.com/chanzuckerberg/aws-oidc/commit/928106610848feb52afeea225d486ef347f44902))

## [0.25.66](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.65...v0.25.66) (2023-05-03)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.254 to 1.44.255 ([#640](https://github.com/chanzuckerberg/aws-oidc/issues/640)) ([70c45e7](https://github.com/chanzuckerberg/aws-oidc/commit/70c45e726befbb45d7991e6dc5d5b3dcdd544822))

## [0.25.65](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.64...v0.25.65) (2023-05-02)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.253 to 1.44.254 ([#638](https://github.com/chanzuckerberg/aws-oidc/issues/638)) ([9e5dd6d](https://github.com/chanzuckerberg/aws-oidc/commit/9e5dd6df77409107cedcfa1af5d3eab84edc7692))

## [0.25.64](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.63...v0.25.64) (2023-05-01)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.250 to 1.44.251 ([#634](https://github.com/chanzuckerberg/aws-oidc/issues/634)) ([f5496e5](https://github.com/chanzuckerberg/aws-oidc/commit/f5496e54a7c1461dc5eefd83e630201311ee5cf5))
* bump github.com/aws/aws-sdk-go from 1.44.251 to 1.44.253 ([#637](https://github.com/chanzuckerberg/aws-oidc/issues/637)) ([3ba839f](https://github.com/chanzuckerberg/aws-oidc/commit/3ba839fbcf07db479caf6b000ae79cca7ddaa325))

## [0.25.63](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.62...v0.25.63) (2023-04-25)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.246 to 1.44.247 ([#629](https://github.com/chanzuckerberg/aws-oidc/issues/629)) ([5c61e60](https://github.com/chanzuckerberg/aws-oidc/commit/5c61e6053aa778a9834bac9e45df12fe0e33afb4))
* bump github.com/aws/aws-sdk-go from 1.44.247 to 1.44.248 ([#631](https://github.com/chanzuckerberg/aws-oidc/issues/631)) ([51e2f37](https://github.com/chanzuckerberg/aws-oidc/commit/51e2f37ce4ddba41ce71b6d57379ab44614f0aa0))
* bump github.com/aws/aws-sdk-go from 1.44.248 to 1.44.249 ([#632](https://github.com/chanzuckerberg/aws-oidc/issues/632)) ([fe948c0](https://github.com/chanzuckerberg/aws-oidc/commit/fe948c024542a70f0e44694865c75e22ef8827e6))

## [0.25.62](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.61...v0.25.62) (2023-04-20)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.245 to 1.44.246 ([#627](https://github.com/chanzuckerberg/aws-oidc/issues/627)) ([36eeb5a](https://github.com/chanzuckerberg/aws-oidc/commit/36eeb5af391e3bd3dfadd4d20206b0aee5693985))

## [0.25.61](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.60...v0.25.61) (2023-04-18)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.243 to 1.44.244 ([#623](https://github.com/chanzuckerberg/aws-oidc/issues/623)) ([9bdd071](https://github.com/chanzuckerberg/aws-oidc/commit/9bdd071d5543e7a15a9f107d5b2ffb1e1d3df8a9))
* bump github.com/aws/aws-sdk-go from 1.44.244 to 1.44.245 ([#626](https://github.com/chanzuckerberg/aws-oidc/issues/626)) ([215b869](https://github.com/chanzuckerberg/aws-oidc/commit/215b8692a85a32f4eb41bef9ca0c91aea958a5e0))
* bump github.com/okta/okta-sdk-golang/v2 from 2.17.0 to 2.18.0 ([#624](https://github.com/chanzuckerberg/aws-oidc/issues/624)) ([f26a56a](https://github.com/chanzuckerberg/aws-oidc/commit/f26a56aa3a59e80b4d2dcca8e3de28eb0fb73f1d))

## [0.25.60](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.59...v0.25.60) (2023-04-14)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.242 to 1.44.243 ([#621](https://github.com/chanzuckerberg/aws-oidc/issues/621)) ([e7e58f6](https://github.com/chanzuckerberg/aws-oidc/commit/e7e58f6be724e3153a6c7dffaf325625f8d81a9e))

## [0.25.59](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.58...v0.25.59) (2023-04-13)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.238 to 1.44.241 ([#618](https://github.com/chanzuckerberg/aws-oidc/issues/618)) ([0d6fafd](https://github.com/chanzuckerberg/aws-oidc/commit/0d6fafd1e8497b99a5edda2f0a5b10bbec8b19e2))
* bump github.com/aws/aws-sdk-go from 1.44.241 to 1.44.242 ([#620](https://github.com/chanzuckerberg/aws-oidc/issues/620)) ([9f41fc1](https://github.com/chanzuckerberg/aws-oidc/commit/9f41fc15a9d100a1ac55ad21693b3ebb65c930e1))
* bump github.com/chanzuckerberg/go-misc from 1.0.7 to 1.0.8 ([#617](https://github.com/chanzuckerberg/aws-oidc/issues/617)) ([c3410d6](https://github.com/chanzuckerberg/aws-oidc/commit/c3410d6fd9e495d4a43600ae36246926e2c6844c))

## [0.25.58](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.57...v0.25.58) (2023-04-07)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.237 to 1.44.238 ([#613](https://github.com/chanzuckerberg/aws-oidc/issues/613)) ([dc1ef13](https://github.com/chanzuckerberg/aws-oidc/commit/dc1ef13221e3a73f4331fdd2d2855c06126f93ed))

## [0.25.57](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.56...v0.25.57) (2023-04-06)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.236 to 1.44.237 ([#611](https://github.com/chanzuckerberg/aws-oidc/issues/611)) ([f8f55fd](https://github.com/chanzuckerberg/aws-oidc/commit/f8f55fdb596a1296b7ca140da7d7c63428c9119e))

## [0.25.56](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.55...v0.25.56) (2023-04-05)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.235 to 1.44.236 ([#608](https://github.com/chanzuckerberg/aws-oidc/issues/608)) ([44891bf](https://github.com/chanzuckerberg/aws-oidc/commit/44891bf88d5022c9fde19570dee25bdabaf297f7))
* bump github.com/spf13/cobra from 1.6.1 to 1.7.0 ([#609](https://github.com/chanzuckerberg/aws-oidc/issues/609)) ([1b71c91](https://github.com/chanzuckerberg/aws-oidc/commit/1b71c91f2dcf33ae4973b25e232944c53a955ffb))

## [0.25.55](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.54...v0.25.55) (2023-04-04)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.234 to 1.44.235 ([#606](https://github.com/chanzuckerberg/aws-oidc/issues/606)) ([a48a02d](https://github.com/chanzuckerberg/aws-oidc/commit/a48a02df0e45e9d9d7d50233f2080e4cfef57a8f))

## [0.25.54](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.53...v0.25.54) (2023-04-03)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.232 to 1.44.234 ([#603](https://github.com/chanzuckerberg/aws-oidc/issues/603)) ([5b10b25](https://github.com/chanzuckerberg/aws-oidc/commit/5b10b254e10122aa871df9d26090559a2203b5d6))

## [0.25.53](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.52...v0.25.53) (2023-04-03)


### Misc

* bump github.com/chanzuckerberg/go-misc from 1.0.6 to 1.0.7 ([#602](https://github.com/chanzuckerberg/aws-oidc/issues/602)) ([eed9066](https://github.com/chanzuckerberg/aws-oidc/commit/eed906695de2e83b57752d782a193bf272e23151))

## [0.25.52](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.51...v0.25.52) (2023-03-30)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.230 to 1.44.231 ([#598](https://github.com/chanzuckerberg/aws-oidc/issues/598)) ([34cf4ea](https://github.com/chanzuckerberg/aws-oidc/commit/34cf4eaaa938a14668b74b4531677bdc9e1d0a93))
* bump github.com/aws/aws-sdk-go from 1.44.231 to 1.44.232 ([#599](https://github.com/chanzuckerberg/aws-oidc/issues/599)) ([d48a89f](https://github.com/chanzuckerberg/aws-oidc/commit/d48a89f8a9cd22106c0bfa29df2f7229a1e5cb7a))

## [0.25.51](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.50...v0.25.51) (2023-03-28)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.223 to 1.44.229 ([#595](https://github.com/chanzuckerberg/aws-oidc/issues/595)) ([78416dd](https://github.com/chanzuckerberg/aws-oidc/commit/78416dd47123b5e612ef63d6f8dfc4f630ee2a7b))
* bump github.com/aws/aws-sdk-go from 1.44.229 to 1.44.230 ([#597](https://github.com/chanzuckerberg/aws-oidc/issues/597)) ([5473ad0](https://github.com/chanzuckerberg/aws-oidc/commit/5473ad0d37fd2d7ba977f23fbc1f9f722467ff03))

## [0.25.50](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.49...v0.25.50) (2023-03-17)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.222 to 1.44.223 ([#588](https://github.com/chanzuckerberg/aws-oidc/issues/588)) ([64a8cd2](https://github.com/chanzuckerberg/aws-oidc/commit/64a8cd2ec6cad4b30b20f4a1235068399a856377))

## [0.25.49](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.48...v0.25.49) (2023-03-16)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.221 to 1.44.222 ([#586](https://github.com/chanzuckerberg/aws-oidc/issues/586)) ([d89221e](https://github.com/chanzuckerberg/aws-oidc/commit/d89221ee1f073ea8918765bc9f468f77f05676bd))

## [0.25.48](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.47...v0.25.48) (2023-03-15)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.220 to 1.44.221 ([#584](https://github.com/chanzuckerberg/aws-oidc/issues/584)) ([3998a45](https://github.com/chanzuckerberg/aws-oidc/commit/3998a45ef7230fe18bd81dbb072c264fdcef70e4))

## [0.25.47](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.46...v0.25.47) (2023-03-14)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.219 to 1.44.220 ([#582](https://github.com/chanzuckerberg/aws-oidc/issues/582)) ([871a4cf](https://github.com/chanzuckerberg/aws-oidc/commit/871a4cfd548c3722a0ad7e72ea6b330d9ba340a2))

## [0.25.46](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.45...v0.25.46) (2023-03-13)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.218 to 1.44.219 ([#580](https://github.com/chanzuckerberg/aws-oidc/issues/580)) ([3572637](https://github.com/chanzuckerberg/aws-oidc/commit/3572637b8d6ff3bb41fc082dde08e31b7fbb3382))

## [0.25.45](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.44...v0.25.45) (2023-03-09)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.214 to 1.44.215 ([#574](https://github.com/chanzuckerberg/aws-oidc/issues/574)) ([fdeb778](https://github.com/chanzuckerberg/aws-oidc/commit/fdeb778f2e12bd9f36c94ddf4561d5ccdc13c90b))
* bump github.com/aws/aws-sdk-go from 1.44.215 to 1.44.216 ([#577](https://github.com/chanzuckerberg/aws-oidc/issues/577)) ([5cb733e](https://github.com/chanzuckerberg/aws-oidc/commit/5cb733eb74aa2adedf73d64cef34eb7f8a411675))
* bump github.com/aws/aws-sdk-go from 1.44.216 to 1.44.217 ([#578](https://github.com/chanzuckerberg/aws-oidc/issues/578)) ([7b65a0f](https://github.com/chanzuckerberg/aws-oidc/commit/7b65a0ff9c7e9299fa6b850cd68a3d4c73abb5b1))
* bump github.com/okta/okta-sdk-golang/v2 from 2.16.0 to 2.17.0 ([#575](https://github.com/chanzuckerberg/aws-oidc/issues/575)) ([90e1c4a](https://github.com/chanzuckerberg/aws-oidc/commit/90e1c4a8d491adb134456da1f1b03b063d43c0d8))

## [0.25.44](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.43...v0.25.44) (2023-03-06)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.210 to 1.44.211 ([#569](https://github.com/chanzuckerberg/aws-oidc/issues/569)) ([4925981](https://github.com/chanzuckerberg/aws-oidc/commit/4925981378a520268005deffaec4d56ed90a5137))
* bump github.com/aws/aws-sdk-go from 1.44.211 to 1.44.212 ([#570](https://github.com/chanzuckerberg/aws-oidc/issues/570)) ([cc0d8e0](https://github.com/chanzuckerberg/aws-oidc/commit/cc0d8e05b9f99a81a056f228d5d597b0d8549be8))
* bump github.com/aws/aws-sdk-go from 1.44.212 to 1.44.213 ([#572](https://github.com/chanzuckerberg/aws-oidc/issues/572)) ([0aa9e83](https://github.com/chanzuckerberg/aws-oidc/commit/0aa9e83d2cd1164a05346f5b4e907a7d2b7af2ed))
* bump github.com/aws/aws-sdk-go from 1.44.213 to 1.44.214 ([#573](https://github.com/chanzuckerberg/aws-oidc/issues/573)) ([a43da5a](https://github.com/chanzuckerberg/aws-oidc/commit/a43da5ae772989e92fe1f94878472d38a3b5b550))

## [0.25.43](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.42...v0.25.43) (2023-02-28)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.208 to 1.44.210 ([#567](https://github.com/chanzuckerberg/aws-oidc/issues/567)) ([ae60f5b](https://github.com/chanzuckerberg/aws-oidc/commit/ae60f5bd4d286f24dbb32d009d339812824c5df8))
* bump github.com/stretchr/testify from 1.8.1 to 1.8.2 ([#565](https://github.com/chanzuckerberg/aws-oidc/issues/565)) ([45f2de9](https://github.com/chanzuckerberg/aws-oidc/commit/45f2de947edae6f709b6e2d8a3be6b3eac37a7bd))

## [0.25.42](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.41...v0.25.42) (2023-02-24)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.207 to 1.44.208 ([#563](https://github.com/chanzuckerberg/aws-oidc/issues/563)) ([763f19d](https://github.com/chanzuckerberg/aws-oidc/commit/763f19d347a728ca3780a062aa6bfb55a4237ddb))

## [0.25.41](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.40...v0.25.41) (2023-02-23)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.205 to 1.44.206 ([#560](https://github.com/chanzuckerberg/aws-oidc/issues/560)) ([bf90843](https://github.com/chanzuckerberg/aws-oidc/commit/bf90843ecf4270c1ca44dcf631f11c2682554449))
* bump github.com/aws/aws-sdk-go from 1.44.206 to 1.44.207 ([#562](https://github.com/chanzuckerberg/aws-oidc/issues/562)) ([1df4f93](https://github.com/chanzuckerberg/aws-oidc/commit/1df4f937b38f4f2bb5a35bb23ef35c55c0f1bdca))

## [0.25.40](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.39...v0.25.40) (2023-02-21)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.203 to 1.44.205 ([#557](https://github.com/chanzuckerberg/aws-oidc/issues/557)) ([137c3eb](https://github.com/chanzuckerberg/aws-oidc/commit/137c3ebab69bd97556d59333af31b4e2a52f0378))
* bump golang.org/x/net from 0.6.0 to 0.7.0 ([#559](https://github.com/chanzuckerberg/aws-oidc/issues/559)) ([27686f8](https://github.com/chanzuckerberg/aws-oidc/commit/27686f853b311b52debeb7d65ad955085dcc3e26))

## [0.25.39](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.38...v0.25.39) (2023-02-17)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.202 to 1.44.203 ([#554](https://github.com/chanzuckerberg/aws-oidc/issues/554)) ([035e337](https://github.com/chanzuckerberg/aws-oidc/commit/035e3378a335639fe472c22e26fb14123a5e5dea))

## [0.25.38](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.37...v0.25.38) (2023-02-16)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.201 to 1.44.202 ([#552](https://github.com/chanzuckerberg/aws-oidc/issues/552)) ([e6fb39c](https://github.com/chanzuckerberg/aws-oidc/commit/e6fb39c71bbac48d3c3abb69a92b2fd6a2748af0))

## [0.25.37](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.36...v0.25.37) (2023-02-15)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.200 to 1.44.201 ([#550](https://github.com/chanzuckerberg/aws-oidc/issues/550)) ([59666a7](https://github.com/chanzuckerberg/aws-oidc/commit/59666a79fdce1094644ec692212d5cf2dd956b67))

## [0.25.36](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.35...v0.25.36) (2023-02-14)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.199 to 1.44.200 ([#548](https://github.com/chanzuckerberg/aws-oidc/issues/548)) ([fff623b](https://github.com/chanzuckerberg/aws-oidc/commit/fff623b26b325d44d71011c722b6fc9ba3f96899))

## [0.25.35](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.34...v0.25.35) (2023-02-13)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.198 to 1.44.199 ([#546](https://github.com/chanzuckerberg/aws-oidc/issues/546)) ([567213d](https://github.com/chanzuckerberg/aws-oidc/commit/567213d8dd95420d15b0abd15a72d2d000b2ccb5))

## [0.25.34](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.33...v0.25.34) (2023-02-10)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.197 to 1.44.198 ([#544](https://github.com/chanzuckerberg/aws-oidc/issues/544)) ([cdcde2d](https://github.com/chanzuckerberg/aws-oidc/commit/cdcde2dfa018b03648bae3063ea69b5fdb01fece))

## [0.25.33](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.32...v0.25.33) (2023-02-09)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.196 to 1.44.197 ([#542](https://github.com/chanzuckerberg/aws-oidc/issues/542)) ([c1aba0b](https://github.com/chanzuckerberg/aws-oidc/commit/c1aba0b35eab6a957d3f7113efe775282ce03aca))

## [0.25.32](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.31...v0.25.32) (2023-02-08)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.195 to 1.44.196 ([#540](https://github.com/chanzuckerberg/aws-oidc/issues/540)) ([2407f66](https://github.com/chanzuckerberg/aws-oidc/commit/2407f66352d3ccb4bd37192d28adb0f3e61d2f3c))

## [0.25.31](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.30...v0.25.31) (2023-02-07)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.194 to 1.44.195 ([#538](https://github.com/chanzuckerberg/aws-oidc/issues/538)) ([d15378c](https://github.com/chanzuckerberg/aws-oidc/commit/d15378ccc3bfbd128a485e53c3d60a3044d3bdea))

## [0.25.30](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.29...v0.25.30) (2023-02-06)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.193 to 1.44.194 ([#536](https://github.com/chanzuckerberg/aws-oidc/issues/536)) ([3d81f50](https://github.com/chanzuckerberg/aws-oidc/commit/3d81f509e7ba6f35b62b9e9d0faa12cddace04c8))

## [0.25.29](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.28...v0.25.29) (2023-02-03)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.192 to 1.44.193 ([#534](https://github.com/chanzuckerberg/aws-oidc/issues/534)) ([9d84c51](https://github.com/chanzuckerberg/aws-oidc/commit/9d84c51c0d6c0f0153b9f7db46d730b35dfc8123))

## [0.25.28](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.27...v0.25.28) (2023-02-02)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.191 to 1.44.192 ([#532](https://github.com/chanzuckerberg/aws-oidc/issues/532)) ([fdf957e](https://github.com/chanzuckerberg/aws-oidc/commit/fdf957e3f9a73354e8ca7e7f29ddec7a41cc65dd))

## [0.25.27](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.26...v0.25.27) (2023-02-01)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.190 to 1.44.191 ([#530](https://github.com/chanzuckerberg/aws-oidc/issues/530)) ([7d658e5](https://github.com/chanzuckerberg/aws-oidc/commit/7d658e55fffd610e5c57f472dafd795d485e7d8a))

## [0.25.26](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.25...v0.25.26) (2023-01-31)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.186 to 1.44.188 ([#527](https://github.com/chanzuckerberg/aws-oidc/issues/527)) ([6340eb7](https://github.com/chanzuckerberg/aws-oidc/commit/6340eb71a8b0b960ca6c8fa9fb312aab50464e74))
* bump github.com/aws/aws-sdk-go from 1.44.188 to 1.44.190 ([#529](https://github.com/chanzuckerberg/aws-oidc/issues/529)) ([067318a](https://github.com/chanzuckerberg/aws-oidc/commit/067318afdfaeb1349ffae159c684480cc99dc7ff))
* bump github.com/chanzuckerberg/go-misc from 1.0.2 to 1.0.3 ([#526](https://github.com/chanzuckerberg/aws-oidc/issues/526)) ([8668366](https://github.com/chanzuckerberg/aws-oidc/commit/86683666bc3602aaa331f82b7f70e036939fe010))

## [0.25.25](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.24...v0.25.25) (2023-01-25)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.185 to 1.44.186 ([#522](https://github.com/chanzuckerberg/aws-oidc/issues/522)) ([07bd3a3](https://github.com/chanzuckerberg/aws-oidc/commit/07bd3a35852f06026fe6639f0c532cc4ced9fb0b))

## [0.25.24](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.23...v0.25.24) (2023-01-25)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.184 to 1.44.185 ([#520](https://github.com/chanzuckerberg/aws-oidc/issues/520)) ([0e346cb](https://github.com/chanzuckerberg/aws-oidc/commit/0e346cbcd5a049a96744daffd9b00154ddde9511))
* bump github.com/chanzuckerberg/go-misc from 1.0.1 to 1.0.2 ([#521](https://github.com/chanzuckerberg/aws-oidc/issues/521)) ([b96d3c7](https://github.com/chanzuckerberg/aws-oidc/commit/b96d3c78e5527f54981087036d3ac7697bc410dd))

## [0.25.23](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.22...v0.25.23) (2023-01-23)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.183 to 1.44.184 ([#518](https://github.com/chanzuckerberg/aws-oidc/issues/518)) ([2557cd6](https://github.com/chanzuckerberg/aws-oidc/commit/2557cd63b501bc49851e7a3e68b2d8539740de56))

## [0.25.22](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.21...v0.25.22) (2023-01-20)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.182 to 1.44.183 ([#516](https://github.com/chanzuckerberg/aws-oidc/issues/516)) ([74853bd](https://github.com/chanzuckerberg/aws-oidc/commit/74853bd242b232fbe1c49b8de50787021130571f))

## [0.25.21](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.20...v0.25.21) (2023-01-19)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.181 to 1.44.182 ([#514](https://github.com/chanzuckerberg/aws-oidc/issues/514)) ([f6ce861](https://github.com/chanzuckerberg/aws-oidc/commit/f6ce861bd48f84e4824f99168df2e92ad623db54))

## [0.25.20](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.19...v0.25.20) (2023-01-18)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.180 to 1.44.181 ([#512](https://github.com/chanzuckerberg/aws-oidc/issues/512)) ([d2c8e86](https://github.com/chanzuckerberg/aws-oidc/commit/d2c8e8653a150c1d816c055846e6e3c4f063ea30))

## [0.25.19](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.18...v0.25.19) (2023-01-16)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.179 to 1.44.180 ([#510](https://github.com/chanzuckerberg/aws-oidc/issues/510)) ([55afcd5](https://github.com/chanzuckerberg/aws-oidc/commit/55afcd5bb7847834cc3ad44a474bdaa587758e27))

## [0.25.18](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.17...v0.25.18) (2023-01-13)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.178 to 1.44.179 ([#508](https://github.com/chanzuckerberg/aws-oidc/issues/508)) ([29b4232](https://github.com/chanzuckerberg/aws-oidc/commit/29b423297b00bf229f153ba2c291017887278316))

## [0.25.17](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.16...v0.25.17) (2023-01-12)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.177 to 1.44.178 ([#506](https://github.com/chanzuckerberg/aws-oidc/issues/506)) ([1b4fad0](https://github.com/chanzuckerberg/aws-oidc/commit/1b4fad00f62c1b65e9fbe775354137aeb6bf5d7f))

## [0.25.16](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.15...v0.25.16) (2023-01-11)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.176 to 1.44.177 ([#504](https://github.com/chanzuckerberg/aws-oidc/issues/504)) ([057d613](https://github.com/chanzuckerberg/aws-oidc/commit/057d613b9835baf64d68146fb01ff71947cb805c))

## [0.25.15](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.14...v0.25.15) (2023-01-10)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.175 to 1.44.176 ([#502](https://github.com/chanzuckerberg/aws-oidc/issues/502)) ([ae58e44](https://github.com/chanzuckerberg/aws-oidc/commit/ae58e441df664cac8bef0161f8b7bb65997d97f6))

## [0.25.14](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.13...v0.25.14) (2023-01-09)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.171 to 1.44.172 ([#497](https://github.com/chanzuckerberg/aws-oidc/issues/497)) ([f6f6d98](https://github.com/chanzuckerberg/aws-oidc/commit/f6f6d98235655c643707f1d09fc2252ee5a7c42a))
* bump github.com/aws/aws-sdk-go from 1.44.172 to 1.44.173 ([#499](https://github.com/chanzuckerberg/aws-oidc/issues/499)) ([b1f60bd](https://github.com/chanzuckerberg/aws-oidc/commit/b1f60bdfa6c9d7a878f615fbc2ec6cf7b2b751a3))
* bump github.com/aws/aws-sdk-go from 1.44.173 to 1.44.174 ([#500](https://github.com/chanzuckerberg/aws-oidc/issues/500)) ([e8d386a](https://github.com/chanzuckerberg/aws-oidc/commit/e8d386ad14670781a993ade8dcb11a756fe95b64))
* bump github.com/aws/aws-sdk-go from 1.44.174 to 1.44.175 ([#501](https://github.com/chanzuckerberg/aws-oidc/issues/501)) ([83526fb](https://github.com/chanzuckerberg/aws-oidc/commit/83526fbfc980cfbcd7b2da564a3e7e04508b7abf))

## [0.25.13](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.12...v0.25.13) (2023-01-03)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.170 to 1.44.171 ([#495](https://github.com/chanzuckerberg/aws-oidc/issues/495)) ([abdc369](https://github.com/chanzuckerberg/aws-oidc/commit/abdc369dde87d748baa12ea99a4be3ec66d3bc59))

## [0.25.12](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.11...v0.25.12) (2023-01-01)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.166 to 1.44.167 ([#490](https://github.com/chanzuckerberg/aws-oidc/issues/490)) ([98c9ac5](https://github.com/chanzuckerberg/aws-oidc/commit/98c9ac5081f1167a4852bd25b10dcbfd02da421f))
* bump github.com/aws/aws-sdk-go from 1.44.167 to 1.44.170 ([#494](https://github.com/chanzuckerberg/aws-oidc/issues/494)) ([2a63f30](https://github.com/chanzuckerberg/aws-oidc/commit/2a63f30c541d9db0cf877612c58d9527db853505))

## [0.25.11](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.10...v0.25.11) (2022-12-23)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.163 to 1.44.166 ([#488](https://github.com/chanzuckerberg/aws-oidc/issues/488)) ([270626a](https://github.com/chanzuckerberg/aws-oidc/commit/270626a9df28b9029cdadc7b5411c538cdef758c))

## [0.25.10](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.9...v0.25.10) (2022-12-20)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.162 to 1.44.163 ([#484](https://github.com/chanzuckerberg/aws-oidc/issues/484)) ([ea7e62c](https://github.com/chanzuckerberg/aws-oidc/commit/ea7e62c780ea7ddff2d41fe8c816c7b7e756b9ac))

## [0.25.9](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.8...v0.25.9) (2022-12-19)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.159 to 1.44.160 ([#480](https://github.com/chanzuckerberg/aws-oidc/issues/480)) ([950eaa0](https://github.com/chanzuckerberg/aws-oidc/commit/950eaa04240a9fb56c55c62f0fbf87297e0339e3))
* bump github.com/aws/aws-sdk-go from 1.44.160 to 1.44.161 ([#482](https://github.com/chanzuckerberg/aws-oidc/issues/482)) ([fefd19c](https://github.com/chanzuckerberg/aws-oidc/commit/fefd19c0862a251c33c5061d637502c2c0227548))
* bump github.com/aws/aws-sdk-go from 1.44.161 to 1.44.162 ([#483](https://github.com/chanzuckerberg/aws-oidc/issues/483)) ([0f3ddd3](https://github.com/chanzuckerberg/aws-oidc/commit/0f3ddd3cf4e6dbd8a4f315dfde9a3b61db7ed941))

## [0.25.8](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.7...v0.25.8) (2022-12-14)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.154 to 1.44.158 ([#477](https://github.com/chanzuckerberg/aws-oidc/issues/477)) ([5317f62](https://github.com/chanzuckerberg/aws-oidc/commit/5317f62b9d6131c80d70c0c550092b6ae73308a1))
* bump github.com/aws/aws-sdk-go from 1.44.158 to 1.44.159 ([#479](https://github.com/chanzuckerberg/aws-oidc/issues/479)) ([96eae33](https://github.com/chanzuckerberg/aws-oidc/commit/96eae33f220a1cf39eaf20181dfaef997baf0d92))

## [0.25.7](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.6...v0.25.7) (2022-12-07)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.147 to 1.44.151 ([#468](https://github.com/chanzuckerberg/aws-oidc/issues/468)) ([adb6b47](https://github.com/chanzuckerberg/aws-oidc/commit/adb6b47f046a70720887d412a4765dcdbf870178))
* bump github.com/aws/aws-sdk-go from 1.44.151 to 1.44.154 ([#471](https://github.com/chanzuckerberg/aws-oidc/issues/471)) ([8cfe181](https://github.com/chanzuckerberg/aws-oidc/commit/8cfe181de333bc56c41502d378f48a6ace980a00))
* bump github.com/okta/okta-sdk-golang/v2 from 2.15.0 to 2.16.0 ([#466](https://github.com/chanzuckerberg/aws-oidc/issues/466)) ([9fb7d8c](https://github.com/chanzuckerberg/aws-oidc/commit/9fb7d8cbbc54147762230d01a5517e510610b4cb))
* Upgrade Docker builder image to go 1.19 ([#473](https://github.com/chanzuckerberg/aws-oidc/issues/473)) ([0f4165f](https://github.com/chanzuckerberg/aws-oidc/commit/0f4165fcdf2736fe4667a2a09e88b24fcaf6b4dd))

## [0.25.6](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.5...v0.25.6) (2022-10-14)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.114 to 1.44.115 ([#429](https://github.com/chanzuckerberg/aws-oidc/issues/429)) ([e756658](https://github.com/chanzuckerberg/aws-oidc/commit/e756658b0517171c3073e31df7c147aaa742248e))

## [0.25.5](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.4...v0.25.5) (2022-10-13)


### Misc

* bump github.com/honeycombio/beeline-go from 1.10.0 to 1.11.0 ([#427](https://github.com/chanzuckerberg/aws-oidc/issues/427)) ([3052a70](https://github.com/chanzuckerberg/aws-oidc/commit/3052a70a6e7737a5f5dab11e485e13dcab9a1945))

## [0.25.4](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.3...v0.25.4) (2022-10-12)


### Misc

* bump github.com/spf13/cobra from 1.5.0 to 1.6.0 ([#424](https://github.com/chanzuckerberg/aws-oidc/issues/424)) ([eecf02a](https://github.com/chanzuckerberg/aws-oidc/commit/eecf02ae0536db737ca73e6a1596b9d486849d94))
* Prevent dependabot from upgrading broken dep ([#426](https://github.com/chanzuckerberg/aws-oidc/issues/426)) ([72a83cc](https://github.com/chanzuckerberg/aws-oidc/commit/72a83cc2def989058cb963de19c2f3306122a20c))

## [0.25.3](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.2...v0.25.3) (2022-10-11)


### BugFixes

* go-keyring downgrade ([#422](https://github.com/chanzuckerberg/aws-oidc/issues/422)) ([ee588a6](https://github.com/chanzuckerberg/aws-oidc/commit/ee588a6a19973fae682c00ad80d873ff6245716e))

## [0.25.2](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.1...v0.25.2) (2022-10-10)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.108 to 1.44.109 ([#415](https://github.com/chanzuckerberg/aws-oidc/issues/415)) ([12e5d44](https://github.com/chanzuckerberg/aws-oidc/commit/12e5d44e2abcbf7dfe30c4f388f1defc8ac65d63))
* bump github.com/aws/aws-sdk-go from 1.44.109 to 1.44.110 ([#417](https://github.com/chanzuckerberg/aws-oidc/issues/417)) ([1050ecf](https://github.com/chanzuckerberg/aws-oidc/commit/1050ecfc90774a2983054f511a0636f3789ad59d))
* bump github.com/aws/aws-sdk-go from 1.44.110 to 1.44.111 ([#418](https://github.com/chanzuckerberg/aws-oidc/issues/418)) ([0b605d0](https://github.com/chanzuckerberg/aws-oidc/commit/0b605d01c1226ab5229cfe21da94d4760f122fca))
* bump github.com/aws/aws-sdk-go from 1.44.111 to 1.44.112 ([#419](https://github.com/chanzuckerberg/aws-oidc/issues/419)) ([da89119](https://github.com/chanzuckerberg/aws-oidc/commit/da891192adc07249f368961acb00fbf307fbb1e1))
* bump github.com/aws/aws-sdk-go from 1.44.112 to 1.44.113 ([#420](https://github.com/chanzuckerberg/aws-oidc/issues/420)) ([c49773a](https://github.com/chanzuckerberg/aws-oidc/commit/c49773aab4dfd071dc3fb1b3263a0bcfdbcfa4e8))
* bump github.com/aws/aws-sdk-go from 1.44.113 to 1.44.114 ([#421](https://github.com/chanzuckerberg/aws-oidc/issues/421)) ([a0cf0b6](https://github.com/chanzuckerberg/aws-oidc/commit/a0cf0b6d2eb455e7a75e3d235c7e328d59f4dffc))

## [0.25.1](https://github.com/chanzuckerberg/aws-oidc/compare/v0.25.0...v0.25.1) (2022-09-28)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.101 to 1.44.103 ([#409](https://github.com/chanzuckerberg/aws-oidc/issues/409)) ([1169b50](https://github.com/chanzuckerberg/aws-oidc/commit/1169b5020c96f63932161779aa1802cec5e1324d))
* bump github.com/aws/aws-sdk-go from 1.44.103 to 1.44.104 ([#410](https://github.com/chanzuckerberg/aws-oidc/issues/410)) ([235bc62](https://github.com/chanzuckerberg/aws-oidc/commit/235bc6298d239973a88099bd7bce5155972b20b6))
* bump github.com/aws/aws-sdk-go from 1.44.104 to 1.44.105 ([#411](https://github.com/chanzuckerberg/aws-oidc/issues/411)) ([942da46](https://github.com/chanzuckerberg/aws-oidc/commit/942da460262b17051c75ac586a64dc39cb9a063c))
* bump github.com/aws/aws-sdk-go from 1.44.105 to 1.44.106 ([#412](https://github.com/chanzuckerberg/aws-oidc/issues/412)) ([2896342](https://github.com/chanzuckerberg/aws-oidc/commit/289634293e3539ad3e23a6cf1e73205b50e030d0))
* bump github.com/aws/aws-sdk-go from 1.44.106 to 1.44.107 ([#413](https://github.com/chanzuckerberg/aws-oidc/issues/413)) ([6356657](https://github.com/chanzuckerberg/aws-oidc/commit/6356657ae88d1ce8515f9ea04e8956576ac5eca9))
* bump github.com/aws/aws-sdk-go from 1.44.96 to 1.44.101 ([#406](https://github.com/chanzuckerberg/aws-oidc/issues/406)) ([7ba9fd2](https://github.com/chanzuckerberg/aws-oidc/commit/7ba9fd2349e07f85f59f8475bc4c3d2e48479e0b))

## [0.25.0](https://github.com/chanzuckerberg/aws-oidc/compare/v0.24.11...v0.25.0) (2022-09-13)


### Features

* **docs:** Add some instructions to configure the aws config generation service ([#164](https://github.com/chanzuckerberg/aws-oidc/issues/164)) ([aca8b01](https://github.com/chanzuckerberg/aws-oidc/commit/aca8b01d2cc210b62aa8cec236c76b77f272eeb1))
* Enable CodeQL ([#168](https://github.com/chanzuckerberg/aws-oidc/issues/168)) ([173720d](https://github.com/chanzuckerberg/aws-oidc/commit/173720d09eeccbee62aac7211670290bfe0b8454))
* fix release ([#364](https://github.com/chanzuckerberg/aws-oidc/issues/364)) ([76e1b4c](https://github.com/chanzuckerberg/aws-oidc/commit/76e1b4ca86be757528a861e14a7e487e3fdc3259))
* fix release ([#365](https://github.com/chanzuckerberg/aws-oidc/issues/365)) ([122ab97](https://github.com/chanzuckerberg/aws-oidc/commit/122ab974823e66764fa5f46bbc35f7031ffa3394))
* Trigger Release ([#363](https://github.com/chanzuckerberg/aws-oidc/issues/363)) ([b4c83be](https://github.com/chanzuckerberg/aws-oidc/commit/b4c83bef5be3e88659257902aea8669208aa85ba))
* Upgrade Go to 1.17 and build darwin arm64 ([#173](https://github.com/chanzuckerberg/aws-oidc/issues/173)) ([641140a](https://github.com/chanzuckerberg/aws-oidc/commit/641140ab54df15bca1274016cd06b8400f920b08))


### BugFixes

* add docker login step to release actions ([#374](https://github.com/chanzuckerberg/aws-oidc/issues/374)) ([5fe6c12](https://github.com/chanzuckerberg/aws-oidc/commit/5fe6c1214700db17c54dd92e37c11a8a8904676f))
* broken setup scripts ([#181](https://github.com/chanzuckerberg/aws-oidc/issues/181)) ([9c3eae6](https://github.com/chanzuckerberg/aws-oidc/commit/9c3eae6e8999f75b9941433234e1b2ec08863d55))
* change version fileformat to help prerelease process ([#371](https://github.com/chanzuckerberg/aws-oidc/issues/371)) ([db5e252](https://github.com/chanzuckerberg/aws-oidc/commit/db5e25220ac02281241a9db8d8bbdd5d602c2b6c))
* changes ([#368](https://github.com/chanzuckerberg/aws-oidc/issues/368)) ([629a195](https://github.com/chanzuckerberg/aws-oidc/commit/629a19516fd1a85c4912ab26b768416cc715cdbe))
* configure: create .aws dir if missing ([#174](https://github.com/chanzuckerberg/aws-oidc/issues/174)) ([fc890ca](https://github.com/chanzuckerberg/aws-oidc/commit/fc890ca478881ecb19357b8c73b8d0bb4bedb664))
* release process automation ([#369](https://github.com/chanzuckerberg/aws-oidc/issues/369)) ([f56d1e6](https://github.com/chanzuckerberg/aws-oidc/commit/f56d1e6546b00e2d7dc70d3bfccd28e0201337a8))
* restore version back to 1.17 for now ([#366](https://github.com/chanzuckerberg/aws-oidc/issues/366)) ([d78a662](https://github.com/chanzuckerberg/aws-oidc/commit/d78a662cc5b4f4ed94ea5e2ad73dfb48ba2ddd7b))
* update release github action ([16baa23](https://github.com/chanzuckerberg/aws-oidc/commit/16baa2380eef401f4967823619c3f1433977e29e))
* updating go-misc for latest AWS SDK version ([#184](https://github.com/chanzuckerberg/aws-oidc/issues/184)) ([412378a](https://github.com/chanzuckerberg/aws-oidc/commit/412378a6611950581bd207442c7de7c45b7b9812))
* upgrade specific dependencies ([#353](https://github.com/chanzuckerberg/aws-oidc/issues/353)) ([1dff77d](https://github.com/chanzuckerberg/aws-oidc/commit/1dff77d36c810fe3d299679e3443711bd232d18d))
* yaml formatting in goreleaser file ([#367](https://github.com/chanzuckerberg/aws-oidc/issues/367)) ([aa66f2f](https://github.com/chanzuckerberg/aws-oidc/commit/aa66f2f118ca85672c191f1d67fe4fcac9f54229))


### Misc

* bump github.com/AlecAivazis/survey/v2 from 2.3.5 to 2.3.6 ([#399](https://github.com/chanzuckerberg/aws-oidc/issues/399)) ([a44b056](https://github.com/chanzuckerberg/aws-oidc/commit/a44b056e8fc650fbb92339c55a7bbfc60dec888e))
* bump github.com/aws/aws-sdk-go from 1.44.76 to 1.44.77 ([#377](https://github.com/chanzuckerberg/aws-oidc/issues/377)) ([5622d90](https://github.com/chanzuckerberg/aws-oidc/commit/5622d905cf21ce5509a81a0d2e87c0c0ef9c163e))
* bump github.com/aws/aws-sdk-go from 1.44.77 to 1.44.78 ([#379](https://github.com/chanzuckerberg/aws-oidc/issues/379)) ([fb66973](https://github.com/chanzuckerberg/aws-oidc/commit/fb66973e70eb257c3acf8c38f705baa2a28c20be))
* bump github.com/aws/aws-sdk-go from 1.44.78 to 1.44.80 ([#380](https://github.com/chanzuckerberg/aws-oidc/issues/380)) ([33192ff](https://github.com/chanzuckerberg/aws-oidc/commit/33192ffb63690ec7244cc6d0f7d196b553fba754))
* bump github.com/aws/aws-sdk-go from 1.44.80 to 1.44.81 ([#381](https://github.com/chanzuckerberg/aws-oidc/issues/381)) ([66b4980](https://github.com/chanzuckerberg/aws-oidc/commit/66b4980759f9a62e59ed6a27bac729dece0b083c))
* bump github.com/aws/aws-sdk-go from 1.44.81 to 1.44.82 ([#382](https://github.com/chanzuckerberg/aws-oidc/issues/382)) ([cefed4c](https://github.com/chanzuckerberg/aws-oidc/commit/cefed4ccae0a8a04ed86e575f337e1d18bb6a5b9))
* bump github.com/aws/aws-sdk-go from 1.44.82 to 1.44.85 ([#387](https://github.com/chanzuckerberg/aws-oidc/issues/387)) ([dc1167f](https://github.com/chanzuckerberg/aws-oidc/commit/dc1167f5908d4b2948412e06ddc89fc087964d50))
* bump github.com/aws/aws-sdk-go from 1.44.85 to 1.44.86 ([#388](https://github.com/chanzuckerberg/aws-oidc/issues/388)) ([fc564d5](https://github.com/chanzuckerberg/aws-oidc/commit/fc564d58706a03f736aaf690dcf3e9823f297d2b))
* bump github.com/aws/aws-sdk-go from 1.44.86 to 1.44.87 ([#389](https://github.com/chanzuckerberg/aws-oidc/issues/389)) ([7f222ff](https://github.com/chanzuckerberg/aws-oidc/commit/7f222ffae1e7f875c6c4205f40eb44e156caa293))
* bump github.com/aws/aws-sdk-go from 1.44.87 to 1.44.88 ([#390](https://github.com/chanzuckerberg/aws-oidc/issues/390)) ([95427bf](https://github.com/chanzuckerberg/aws-oidc/commit/95427bfcb217303455a4a25560d168cfd7f92117))
* bump github.com/aws/aws-sdk-go from 1.44.88 to 1.44.89 ([#391](https://github.com/chanzuckerberg/aws-oidc/issues/391)) ([b9a5e7d](https://github.com/chanzuckerberg/aws-oidc/commit/b9a5e7d14a6b6339122fa8533714dc1ff5b71d41))
* bump github.com/aws/aws-sdk-go from 1.44.89 to 1.44.90 ([#392](https://github.com/chanzuckerberg/aws-oidc/issues/392)) ([9ab78e3](https://github.com/chanzuckerberg/aws-oidc/commit/9ab78e385e17a10a9b74da7e44f7fda85aef858f))
* bump github.com/aws/aws-sdk-go from 1.44.90 to 1.44.91 ([#394](https://github.com/chanzuckerberg/aws-oidc/issues/394)) ([3483ea5](https://github.com/chanzuckerberg/aws-oidc/commit/3483ea5ab7ea3b165fbdd8e18dff4a70070caa7b))
* bump github.com/aws/aws-sdk-go from 1.44.91 to 1.44.92 ([#395](https://github.com/chanzuckerberg/aws-oidc/issues/395)) ([fe42968](https://github.com/chanzuckerberg/aws-oidc/commit/fe429683372c958ebf6818f9aa801358d1a53744))
* bump github.com/aws/aws-sdk-go from 1.44.92 to 1.44.93 ([#396](https://github.com/chanzuckerberg/aws-oidc/issues/396)) ([c950181](https://github.com/chanzuckerberg/aws-oidc/commit/c950181b1f8478345ab27e9b676238d39e5d95ce))
* bump github.com/aws/aws-sdk-go from 1.44.93 to 1.44.94 ([#397](https://github.com/chanzuckerberg/aws-oidc/issues/397)) ([933b259](https://github.com/chanzuckerberg/aws-oidc/commit/933b259087c88eec97dc59f6fc405a858f929fb9))
* bump github.com/aws/aws-sdk-go from 1.44.94 to 1.44.95 ([#398](https://github.com/chanzuckerberg/aws-oidc/issues/398)) ([fcfc6a6](https://github.com/chanzuckerberg/aws-oidc/commit/fcfc6a67fc3cf0778c5b9821922369ce1263e8ab))
* bump github.com/aws/aws-sdk-go from 1.44.95 to 1.44.96 ([#400](https://github.com/chanzuckerberg/aws-oidc/issues/400)) ([ca260f0](https://github.com/chanzuckerberg/aws-oidc/commit/ca260f0779eda1ea534546a74578673e7a6294f2))
* bump github.com/honeycombio/beeline-go from 1.9.0 to 1.10.0 ([#386](https://github.com/chanzuckerberg/aws-oidc/issues/386)) ([a9bc4f8](https://github.com/chanzuckerberg/aws-oidc/commit/a9bc4f80d6342eb6bff986dddd131e593d8af859))
* bump github.com/okta/okta-sdk-golang/v2 from 2.13.0 to 2.14.0 ([#384](https://github.com/chanzuckerberg/aws-oidc/issues/384)) ([179ddea](https://github.com/chanzuckerberg/aws-oidc/commit/179ddeab6ee5ab76139b829ff5a1489fb43223af))
* **CCIE-0:** release 0.24.7 ([#370](https://github.com/chanzuckerberg/aws-oidc/issues/370)) ([2e521d3](https://github.com/chanzuckerberg/aws-oidc/commit/2e521d33251214bad2951f74f9f94e771f7c6e63))
* dependency update ([#281](https://github.com/chanzuckerberg/aws-oidc/issues/281)) ([5765c82](https://github.com/chanzuckerberg/aws-oidc/commit/5765c82a76c9d455b3a07717e9bb1b05473193a2))
* **deps:** bump github.com/aws/aws-sdk-go from 1.44.73 to 1.44.76 ([#376](https://github.com/chanzuckerberg/aws-oidc/issues/376)) ([2a0d35c](https://github.com/chanzuckerberg/aws-oidc/commit/2a0d35cc1b982fd0e2f144f07fa2d0f6dd173d30))
* **main:** release 0.24.10 ([#378](https://github.com/chanzuckerberg/aws-oidc/issues/378)) ([a765710](https://github.com/chanzuckerberg/aws-oidc/commit/a7657109741d1d7017d97785e07876714d534e44))
* **main:** release 0.24.11 ([#393](https://github.com/chanzuckerberg/aws-oidc/issues/393)) ([6a301eb](https://github.com/chanzuckerberg/aws-oidc/commit/6a301eb2e08c9fd5a4891bb7bbfdf412e940b6e9))
* **main:** release 0.24.8 ([22998ca](https://github.com/chanzuckerberg/aws-oidc/commit/22998cac6519dc02b2b7fb1ddb88e4c097cfc1d8))
* release 0.24.8 ([#372](https://github.com/chanzuckerberg/aws-oidc/issues/372)) ([22998ca](https://github.com/chanzuckerberg/aws-oidc/commit/22998cac6519dc02b2b7fb1ddb88e4c097cfc1d8))
* release 0.24.9 ([#375](https://github.com/chanzuckerberg/aws-oidc/issues/375)) ([8bc1d79](https://github.com/chanzuckerberg/aws-oidc/commit/8bc1d794cd22fd11e5184102235724c7514d4461))
* upgrade go-misc and aws mocks ([#285](https://github.com/chanzuckerberg/aws-oidc/issues/285)) ([b19d33e](https://github.com/chanzuckerberg/aws-oidc/commit/b19d33e8d649b5830b654be0ba80a90ad391884d))

## [0.24.11](https://github.com/chanzuckerberg/aws-oidc/compare/v0.24.10...v0.24.11) (2022-09-12)


### Misc

* bump github.com/AlecAivazis/survey/v2 from 2.3.5 to 2.3.6 ([#399](https://github.com/chanzuckerberg/aws-oidc/issues/399)) ([a44b056](https://github.com/chanzuckerberg/aws-oidc/commit/a44b056e8fc650fbb92339c55a7bbfc60dec888e))
* bump github.com/aws/aws-sdk-go from 1.44.89 to 1.44.90 ([#392](https://github.com/chanzuckerberg/aws-oidc/issues/392)) ([9ab78e3](https://github.com/chanzuckerberg/aws-oidc/commit/9ab78e385e17a10a9b74da7e44f7fda85aef858f))
* bump github.com/aws/aws-sdk-go from 1.44.90 to 1.44.91 ([#394](https://github.com/chanzuckerberg/aws-oidc/issues/394)) ([3483ea5](https://github.com/chanzuckerberg/aws-oidc/commit/3483ea5ab7ea3b165fbdd8e18dff4a70070caa7b))
* bump github.com/aws/aws-sdk-go from 1.44.91 to 1.44.92 ([#395](https://github.com/chanzuckerberg/aws-oidc/issues/395)) ([fe42968](https://github.com/chanzuckerberg/aws-oidc/commit/fe429683372c958ebf6818f9aa801358d1a53744))
* bump github.com/aws/aws-sdk-go from 1.44.92 to 1.44.93 ([#396](https://github.com/chanzuckerberg/aws-oidc/issues/396)) ([c950181](https://github.com/chanzuckerberg/aws-oidc/commit/c950181b1f8478345ab27e9b676238d39e5d95ce))
* bump github.com/aws/aws-sdk-go from 1.44.93 to 1.44.94 ([#397](https://github.com/chanzuckerberg/aws-oidc/issues/397)) ([933b259](https://github.com/chanzuckerberg/aws-oidc/commit/933b259087c88eec97dc59f6fc405a858f929fb9))
* bump github.com/aws/aws-sdk-go from 1.44.94 to 1.44.95 ([#398](https://github.com/chanzuckerberg/aws-oidc/issues/398)) ([fcfc6a6](https://github.com/chanzuckerberg/aws-oidc/commit/fcfc6a67fc3cf0778c5b9821922369ce1263e8ab))

## [0.24.10](https://github.com/chanzuckerberg/aws-oidc/compare/v0.24.9...v0.24.10) (2022-09-01)


### Misc

* bump github.com/aws/aws-sdk-go from 1.44.76 to 1.44.77 ([#377](https://github.com/chanzuckerberg/aws-oidc/issues/377)) ([5622d90](https://github.com/chanzuckerberg/aws-oidc/commit/5622d905cf21ce5509a81a0d2e87c0c0ef9c163e))
* bump github.com/aws/aws-sdk-go from 1.44.77 to 1.44.78 ([#379](https://github.com/chanzuckerberg/aws-oidc/issues/379)) ([fb66973](https://github.com/chanzuckerberg/aws-oidc/commit/fb66973e70eb257c3acf8c38f705baa2a28c20be))
* bump github.com/aws/aws-sdk-go from 1.44.78 to 1.44.80 ([#380](https://github.com/chanzuckerberg/aws-oidc/issues/380)) ([33192ff](https://github.com/chanzuckerberg/aws-oidc/commit/33192ffb63690ec7244cc6d0f7d196b553fba754))
* bump github.com/aws/aws-sdk-go from 1.44.80 to 1.44.81 ([#381](https://github.com/chanzuckerberg/aws-oidc/issues/381)) ([66b4980](https://github.com/chanzuckerberg/aws-oidc/commit/66b4980759f9a62e59ed6a27bac729dece0b083c))
* bump github.com/aws/aws-sdk-go from 1.44.81 to 1.44.82 ([#382](https://github.com/chanzuckerberg/aws-oidc/issues/382)) ([cefed4c](https://github.com/chanzuckerberg/aws-oidc/commit/cefed4ccae0a8a04ed86e575f337e1d18bb6a5b9))
* bump github.com/aws/aws-sdk-go from 1.44.82 to 1.44.85 ([#387](https://github.com/chanzuckerberg/aws-oidc/issues/387)) ([dc1167f](https://github.com/chanzuckerberg/aws-oidc/commit/dc1167f5908d4b2948412e06ddc89fc087964d50))
* bump github.com/aws/aws-sdk-go from 1.44.85 to 1.44.86 ([#388](https://github.com/chanzuckerberg/aws-oidc/issues/388)) ([fc564d5](https://github.com/chanzuckerberg/aws-oidc/commit/fc564d58706a03f736aaf690dcf3e9823f297d2b))
* bump github.com/aws/aws-sdk-go from 1.44.86 to 1.44.87 ([#389](https://github.com/chanzuckerberg/aws-oidc/issues/389)) ([7f222ff](https://github.com/chanzuckerberg/aws-oidc/commit/7f222ffae1e7f875c6c4205f40eb44e156caa293))
* bump github.com/aws/aws-sdk-go from 1.44.87 to 1.44.88 ([#390](https://github.com/chanzuckerberg/aws-oidc/issues/390)) ([95427bf](https://github.com/chanzuckerberg/aws-oidc/commit/95427bfcb217303455a4a25560d168cfd7f92117))
* bump github.com/aws/aws-sdk-go from 1.44.88 to 1.44.89 ([#391](https://github.com/chanzuckerberg/aws-oidc/issues/391)) ([b9a5e7d](https://github.com/chanzuckerberg/aws-oidc/commit/b9a5e7d14a6b6339122fa8533714dc1ff5b71d41))
* bump github.com/honeycombio/beeline-go from 1.9.0 to 1.10.0 ([#386](https://github.com/chanzuckerberg/aws-oidc/issues/386)) ([a9bc4f8](https://github.com/chanzuckerberg/aws-oidc/commit/a9bc4f80d6342eb6bff986dddd131e593d8af859))
* bump github.com/okta/okta-sdk-golang/v2 from 2.13.0 to 2.14.0 ([#384](https://github.com/chanzuckerberg/aws-oidc/issues/384)) ([179ddea](https://github.com/chanzuckerberg/aws-oidc/commit/179ddeab6ee5ab76139b829ff5a1489fb43223af))

## [0.24.9](https://github.com/chanzuckerberg/aws-oidc/compare/v0.24.8...v0.24.9) (2022-08-15)


### BugFixes

* add docker login step to release actions ([#374](https://github.com/chanzuckerberg/aws-oidc/issues/374)) ([5fe6c12](https://github.com/chanzuckerberg/aws-oidc/commit/5fe6c1214700db17c54dd92e37c11a8a8904676f))


### Misc

* **deps:** bump github.com/aws/aws-sdk-go from 1.44.73 to 1.44.76 ([#376](https://github.com/chanzuckerberg/aws-oidc/issues/376)) ([2a0d35c](https://github.com/chanzuckerberg/aws-oidc/commit/2a0d35cc1b982fd0e2f144f07fa2d0f6dd173d30))

## [0.24.8](https://github.com/chanzuckerberg/aws-oidc/compare/v0.24.7...v0.24.8) (2022-08-11)


### BugFixes

* change version fileformat to help prerelease process ([#371](https://github.com/chanzuckerberg/aws-oidc/issues/371)) ([db5e252](https://github.com/chanzuckerberg/aws-oidc/commit/db5e25220ac02281241a9db8d8bbdd5d602c2b6c))

## [0.24.7](https://github.com/chanzuckerberg/aws-oidc/compare/v0.24.6...v0.24.7) (2022-08-11)


### BugFixes

* release process automation ([#369](https://github.com/chanzuckerberg/aws-oidc/issues/369)) ([f56d1e6](https://github.com/chanzuckerberg/aws-oidc/commit/f56d1e6546b00e2d7dc70d3bfccd28e0201337a8))
