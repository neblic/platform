# Contributing

By participating in this project, you agree to abide our [code of conduct](./CODE_OF_CONDUCT.md).

## Support

Do not open issues for general support questions. The use of `Issues` is limited to bug reports and feature requests. Instead, please post a question on Stack Overflow using the `neblic` tag. Contributors are subscribed to the tag and will monitor for any new questions.

Stack Overflow platform is designed specifically to provide support and it excels at it, so it is much better to maintain a knowledge base in there than to use the `Issues` section.

## Reporting bugs and requesting features

Bugs and new features are tracked in the `Issues` section of the repository. Please, search for existing bugs and features (open and closed issues) before opening a new issue. If it seems to be an unreported bug or an unrequested feature, please create a new issue with as much detail as possible.

When an issue is opened, it will get the `triage` tag automatically. A contributor will take a look at it, remove the `triage` tag, relabel it if necessary, and add a comment with their findings.

Planning is publicly available in the [`Projects`](https://github.com/neblic/platform/projects) section. In there, you can see if someone is working on an issue or when it is planned to be addressed. If an issue is not assigned to a project, it doesn't mean that it won't be fixed or implemented soon, it just means there's no current estimate of when it will be fixed or implemented.

## Getting started

The best way to start contributing is to look for issues labeled `good first issue`. These are simple tasks that do not require much background knowledge. If you would like to take one, please add a comment or submit a PR. You will get all the support you need from other contributors in the same issue discussion, so feel free to ask anything!

For larger contributions, the short and long term road-map is available in the [`Projects`](https://github.com/neblic/platform/projects) section. Before starting a major feature or contribution, take a look at the road-map to make sure that no one else is working on or planning to work on something similar. Even if it is planned, feel free to share your thoughts and ask how you can contribute so that the solution meets everyone's needs. 

## Creating PRs

### PR structure

Unless they contain trivial changes, PRs are encouraged to be under 500 lines (excluding go mod/sum changes). If this is not possible, break them into meaningful commits that can be reviewed independently.

### Commit message

Commit messages must follow the conventional commit style (see Appendix B); it is important to follow the correct style because release changelogs are automatically generated from the commits. Please keep this in mind when writing commit messages.

In addition to the format, the Conventional Commits specification makes recommendations about what and how to write in the commit message, but for a more detailed explanation, see this [link](https://cbea.ms/git-commit/) (ignore the format recommendations if they conflict with the Conventional Commits specification), especially the sections:
* Limit the subject line to (approximately) 50 characters
* Use the imperative form in the subject line
* Use the body to explain what and why vs. how

### Tests

The project implements two kinds of tests. Behavioral tests written in a BDD style using [Ginkgo](https://github.com/onsi/ginkgo), and standard Go unit tests.

The rationale behind implementing behavioral tests with a potentially large scope is to cover and highlight important user functionality so it is easy to validate and understand how it works. These tests should only use public methods and fields and will generally cover multiple structures or packages. If necessary, they can use doubles to replace parts that are out of scope, but it is recommended to use as many real objects as possible.

It is not necessary to aim for 100% code coverage. PRs should include enough tests to give reviewers confidence that the code will work as expected. If it is a critical part of the code, or if its complexity is high, it will require thorough testing. And if it introduces a major new feature, it will probably require behavioral testing.

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
* dataplane: Data collection pipeline and transformations
* collector: Collector setup and builds
* neblictl: CLI client
* sampler: Go sampler library
* kafkasampler: Kafka sampler service
