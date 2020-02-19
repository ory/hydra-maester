<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Changelog](#changelog)
  - [Unreleased](#unreleased)
  - [v0.0.14 (2020-02-11)](#v0014-2020-02-11)
  - [v0.0.13 (2020-02-11)](#v0013-2020-02-11)
  - [v0.0.12 (2020-02-05)](#v0012-2020-02-05)
  - [v0.0.11 (2020-02-03)](#v0011-2020-02-03)
  - [v0.0.10 (2020-02-01)](#v0010-2020-02-01)
  - [v0.0.9 (2019-12-26)](#v009-2019-12-26)
  - [v0.0.8 (2019-12-16)](#v008-2019-12-16)
  - [v0.0.7 (2019-12-16)](#v007-2019-12-16)
  - [v0.0.6 (2019-11-26)](#v006-2019-11-26)
  - [v0.0.5 (2019-11-14)](#v005-2019-11-14)
  - [v0.0.4 (2019-09-19)](#v004-2019-09-19)
  - [v0.0.3 (2019-08-30)](#v003-2019-08-30)
  - [v0.0.2-test2 (2019-08-30)](#v002-test2-2019-08-30)
  - [v0.0.2-test1 (2019-08-30)](#v002-test1-2019-08-30)
  - [v0.0.2 (2019-08-30)](#v002-2019-08-30)
  - [v0.0.1 (2019-08-29)](#v001-2019-08-29)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Changelog

## [Unreleased](https://github.com/ory/hydra-maester/tree/HEAD)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.14...HEAD)

**Closed issues:**

- Restarting hydra-maester should trigger a reconcile check [\#43](https://github.com/ory/hydra-maester/issues/43)

## [v0.0.14](https://github.com/ory/hydra-maester/tree/v0.0.14) (2020-02-11)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.13...v0.0.14)

## [v0.0.13](https://github.com/ory/hydra-maester/tree/v0.0.13) (2020-02-11)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.12...v0.0.13)

**Merged pull requests:**

- feat: Sync mode - least work approach [\#46](https://github.com/ory/hydra-maester/pull/46) ([jakkab](https://github.com/jakkab))

## [v0.0.12](https://github.com/ory/hydra-maester/tree/v0.0.12) (2020-02-05)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.11...v0.0.12)

**Merged pull requests:**

- fix\(resolver\): Fix client name detection for updates [\#45](https://github.com/ory/hydra-maester/pull/45) ([paulbdavis](https://github.com/paulbdavis))

## [v0.0.11](https://github.com/ory/hydra-maester/tree/v0.0.11) (2020-02-03)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.10...v0.0.11)

**Merged pull requests:**

- adding support for hydra client metadata property [\#38](https://github.com/ory/hydra-maester/pull/38) ([amihalj](https://github.com/amihalj))

## [v0.0.10](https://github.com/ory/hydra-maester/tree/v0.0.10) (2020-02-01)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.9...v0.0.10)

**Closed issues:**

- hydra\_v1alpha1\_oauth2client\_user\_credentials.yaml is using invalid apiVersion for OAuth2Client [\#41](https://github.com/ory/hydra-maester/issues/41)

**Merged pull requests:**

- adding audience into API calls [\#44](https://github.com/ory/hydra-maester/pull/44) ([amihalj](https://github.com/amihalj))
- Fix sample apiVersion [\#42](https://github.com/ory/hydra-maester/pull/42) ([PoulpiFr](https://github.com/PoulpiFr))

## [v0.0.9](https://github.com/ory/hydra-maester/tree/v0.0.9) (2019-12-26)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.8...v0.0.9)

**Closed issues:**

- Set owner reference on secrets created by the controller. [\#20](https://github.com/ory/hydra-maester/issues/20)

**Merged pull requests:**

- Use binary kustomize release for CI [\#40](https://github.com/ory/hydra-maester/pull/40) ([aeneasr](https://github.com/aeneasr))

## [v0.0.8](https://github.com/ory/hydra-maester/tree/v0.0.8) (2019-12-16)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.7...v0.0.8)

## [v0.0.7](https://github.com/ory/hydra-maester/tree/v0.0.7) (2019-12-16)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.6...v0.0.7)

**Merged pull requests:**

- Set OwnerReference on Secrets created by controller [\#39](https://github.com/ory/hydra-maester/pull/39) ([kubadz](https://github.com/kubadz))

## [v0.0.6](https://github.com/ory/hydra-maester/tree/v0.0.6) (2019-11-26)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.5...v0.0.6)

**Closed issues:**

- Improvement: Allow multiple hydra with single maester [\#34](https://github.com/ory/hydra-maester/issues/34)
- cannot use hydra's allow\_termination\_from support without X-Forwarded-Proto header [\#32](https://github.com/ory/hydra-maester/issues/32)
- How do we set the redirect\_uri [\#26](https://github.com/ory/hydra-maester/issues/26)

**Merged pull requests:**

- adding support for token\_endpoint\_auth\_method [\#37](https://github.com/ory/hydra-maester/pull/37) ([amihalj](https://github.com/amihalj))

## [v0.0.5](https://github.com/ory/hydra-maester/tree/v0.0.5) (2019-11-14)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.4...v0.0.5)

**Fixed bugs:**

- BUG: "do" func unmarshalls body of unsuccessful requests to oauth2clientjson  [\#21](https://github.com/ory/hydra-maester/issues/21)

**Closed issues:**

- Created secret name should match `secretName` in spec [\#29](https://github.com/ory/hydra-maester/issues/29)
- Testing issues [\#27](https://github.com/ory/hydra-maester/issues/27)

**Merged pull requests:**

- Feature: multi hydra [\#35](https://github.com/ory/hydra-maester/pull/35) ([paulbdavis](https://github.com/paulbdavis))
- README: add kubebuilder as a prerequisite [\#31](https://github.com/ory/hydra-maester/pull/31) ([paulbdavis](https://github.com/paulbdavis))

## [v0.0.4](https://github.com/ory/hydra-maester/tree/v0.0.4) (2019-09-19)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.3...v0.0.4)

**Fixed bugs:**

- Controller can override clients that don't correspond with current CR. [\#22](https://github.com/ory/hydra-maester/issues/22)

**Closed issues:**

- Create CI/CD & release [\#6](https://github.com/ory/hydra-maester/issues/6)
- Plug it in in the hydra chart [\#5](https://github.com/ory/hydra-maester/issues/5)
- Define CRD [\#2](https://github.com/ory/hydra-maester/issues/2)

**Merged pull requests:**

- fix JSON decode bug [\#25](https://github.com/ory/hydra-maester/pull/25) ([jakkab](https://github.com/jakkab))
- Fix override clients bug [\#23](https://github.com/ory/hydra-maester/pull/23) ([jakkab](https://github.com/jakkab))
- Full upgrade [\#19](https://github.com/ory/hydra-maester/pull/19) ([jakkab](https://github.com/jakkab))
- CR status [\#17](https://github.com/ory/hydra-maester/pull/17) ([jakkab](https://github.com/jakkab))

## [v0.0.3](https://github.com/ory/hydra-maester/tree/v0.0.3) (2019-08-30)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.2-test2...v0.0.3)

**Merged pull requests:**

- Extend readme: command-line flags [\#15](https://github.com/ory/hydra-maester/pull/15) ([jakkab](https://github.com/jakkab))

## [v0.0.2-test2](https://github.com/ory/hydra-maester/tree/v0.0.2-test2) (2019-08-30)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.2-test1...v0.0.2-test2)

**Merged pull requests:**

- Update release-changelog pipeline [\#16](https://github.com/ory/hydra-maester/pull/16) ([Demonsthere](https://github.com/Demonsthere))

## [v0.0.2-test1](https://github.com/ory/hydra-maester/tree/v0.0.2-test1) (2019-08-30)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.2...v0.0.2-test1)

**Merged pull requests:**

- Update Dockerfile [\#14](https://github.com/ory/hydra-maester/pull/14) ([Demonsthere](https://github.com/Demonsthere))

## [v0.0.2](https://github.com/ory/hydra-maester/tree/v0.0.2) (2019-08-30)

[Full Changelog](https://github.com/ory/hydra-maester/compare/v0.0.1...v0.0.2)

**Merged pull requests:**

- Add valid sample [\#13](https://github.com/ory/hydra-maester/pull/13) ([jakkab](https://github.com/jakkab))

## [v0.0.1](https://github.com/ory/hydra-maester/tree/v0.0.1) (2019-08-29)

[Full Changelog](https://github.com/ory/hydra-maester/compare/7cd60343d02b63ae0ab438f9e8ee26724f28a87a...v0.0.1)

**Merged pull requests:**

- Scaffold controller [\#12](https://github.com/ory/hydra-maester/pull/12) ([jakkab](https://github.com/jakkab))
- Add goreleaser file [\#11](https://github.com/ory/hydra-maester/pull/11) ([piotrmsc](https://github.com/piotrmsc))
- Small info [\#10](https://github.com/ory/hydra-maester/pull/10) ([piotrmsc](https://github.com/piotrmsc))
- Design documents [\#9](https://github.com/ory/hydra-maester/pull/9) ([piotrmsc](https://github.com/piotrmsc))
- Fix circleci config [\#8](https://github.com/ory/hydra-maester/pull/8) ([piotrmsc](https://github.com/piotrmsc))
- Initial ci/cd + release config [\#7](https://github.com/ory/hydra-maester/pull/7) ([piotrmsc](https://github.com/piotrmsc))
- Initial readme [\#1](https://github.com/ory/hydra-maester/pull/1) ([piotrmsc](https://github.com/piotrmsc))



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
