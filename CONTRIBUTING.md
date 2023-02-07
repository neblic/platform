# Contributing

By participating in this project, you agree to abide our [code of conduct](./CODE_OF_CONDUCT.md).

## Support

Do not open issues for general support questions. `Issues` usage is limited to bug reports and feature requests. Instead, please ask a question in Stack Overflow using the tag `neblic`. Contributors are subscribed to the tag and will be notified of new questions.

Stack Overflow platform is designed specifically to provide support and it excels at that, so it is much better to keep a knowledge base in there than to use the `Issues` section.

## Reporting bugs and requesting features

Bugs and new features are tracked in the repository `Issues` section. Please, search for existing bugs and features (open and closed issues) before opening a new one. If it seems to be an unreported bug or an unrequested feature, please create a new issue with as much detail as possible.

When an issue is opened, it will get the `triage` tag set automatically. A contributor will then take a look, remove the `triage` tag, relabel it if necessary, and add a comment with their findings.

Planning is publicly available in the [`Projects`](https://github.com/neblic/platform/projects) section. In there, you can see if someone is working on an issue or when it is planned to be addressed. If an issue is not assigned to any project, it doesn't mean that it won't be fixed or implemented soon, it just means that, for now, it is not known when will it be fixed or implemented.

## Getting started

The best way to start contributing is to look for issues with the label `good first issue`. These are simple tasks that do not require a lot of background knowledge. If you would like to take one, please add a comment or submit a PR. You will receive all the support you need from other contributors in the same issue discussion, feel free to ask anything!

For larger contributions, the short and long term road-map is available in the [`Projects`](https://github.com/neblic/platform/projects) section. Before starting a big feature or contribution, take a look to make sure no one is actually working or plans to work on something similar. Even if it is planned, you are welcome to share your thoughts and ask how you could contribute so the solution includes everyone's requirements.

## Creating PRs

### PR structure

Unless they include trivial changes, it is encouraged to keep PRs under 500 lines  (excluding go mod/sum changes). If not possible, split them into meaningful commits that can be reviewed independently.

### Commit message
Commit messages need to follow the conventional commits style (see Appendix C), it is important to follow the proper style because releases changelogs are generated automatically from the commits. Please, keep that in mind when writing the commit messages.

In addition to the format, the Conventional Commits specification contains recommendations about what and how to write in the commit message, but for a more detailed explanation check this [link](https://cbea.ms/git-commit/) (ignore the format recommendations if they are in conflict with the Conventional Commits specification), especially the sections:
* Limit the subject line to (approximately) 50 characters
* Use the imperative form in the subject line
* Use the body to explain what and why vs. how

### Tests

The project implements two types of tests. Behavioural tests, written following a BDD style with [Ginkgo](https://github.com/onsi/ginkgo), and standard Go unit tests.

The rationale behind implementing behavioural tests with a potentially large scope is to cover and highlight major user functionality so it is straightforward to validate and understand how they work. These tests should only use public methods and fields, and generally, they will cover multiple structures or packages. If required, they can use doubles to replace parts that are out of scope, but it is encouraged to use as many real objects as possible.

It is not necessary to strive for 100% code coverage. PRs should include enough tests so reviewers are confident that the code works as expected. If it is a critical part of the code or if its complexity is high, it will require thorough testing. And if it introduces a new major feature, it is likely it will require a behaviour test.

## Appendix A: Project planning

The [`Projects`](https://github.com/neblic/platform/projects) section contains one or more projects associated with the repository. Contributors use GiHub Projects to coordinate development and tasks, and to plan and share long-term goals.

## Appendix B: Conventional commits types and scopes

Commit messages follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#summary) specification so it's possible to automatically generate release changelogs and track version numbers.

Type must be one of the following (based on [Angular type definitions](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#type)):
* build: Changes that affect the build system or external dependencies
* ci: Changes to the CI configuration files
* docs: Documentation updates
* feat: New features
* fix: Bugfixes
* perf: Performance improvements
* refactor: Code changes that do not modify its functionality
* style: Formatting
* test: New or updated tests

Accepted scopes are:
* controlplane: Internal control plane definition and implementation
* neblictl: CLI client
* sampler: Go sampler library
* kafkasampler: Kafka sampler service
