# Contributing

<!-- toc -->
-   [Getting Started](#getting-started)
-   [Workflow](#workflow)
    -   [Creating Pull Requests](#creating-pull-requests)
    -   [Code Review](#code-review)
    -   [Testing](#testing)
<!-- /toc -->

# Getting Started

- Fork the repository on GitHub
- Read the [Workflow](#workflow) for contribution instructions.

# Workflow

Feel free to ask questions, report bugs, or send pull requests. Here's a breakdown of the workflow:

Asking a Question or Reporting a Bug:

- Create an issue with a clear and descriptive title.
- Select the appropriate label from the options on the right.
- Provide a detailed description of the situation.
- As the discussion progresses, the issue's label may change or more details might be requested. Long-term unresponsive or unreproducible issues will be closed automatically by a [stable bot](https://github.com/actions/stale).


Sending a Pull Request

- Create a corresponding issue as an admission ticket and associate it with your pull request.
- Make commits in logical units, and be prepared for potential requests for unit tests.
- Submit clear commit messages.
- The pull request must receive approval from at least one maintainer and one developer.

## Creating Pull Requests

1Block.AI follows the standard [GitHub pull request](https://help.github.com/articles/about-pull-requests/) process. To submit a proposed change, please develop the code/fix and add new testing cases if possible. Before submitting a pull request, run local verifications to predict the pass or fail of continuous integration:


- Developing the Command Line Tool:
    - Run and pass `make test`
    - Run and pass `make build`
    - Run and pass the CI test locally using [act](https://github.com/nektos/act):
      ```sh
      $ act -j validate
      ```
- Developing the API-server:
    - Run and pass `make test`
    - Run and pass `make build`
    - Run and pass the CI test locally using [act](https://github.com/nektos/act):
      ```sh
      $ act -j validate
      ```

## Code Review

To facilitate effective reviews of your PR, please consider the following guidelines that reviewers will appreciate:

- Follow [good coding guidelines](https://github.com/golang/go/wiki/CodeReviewComments).
- Write [well-structured commit messages](https://chris.beams.io/posts/git-commit/).
- Break down large changes into a series of smaller, logically ordered patches, each contributing to a more comprehensive solution.
- [Label](https://github.com/oneblock-ai/oneblock/labels) your PR with the required label(e.g., `priority/*`, `area/*`, `milestone version`).
- [Request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/requesting-a-pull-request-review) appropriate reviewers to review.


### Commit Message Format

1Block.AI follows the [Conventional Commits](https://www.conventionalcommits.org/) format for commit messages. The structure should be as follows:

```text
<type>[optional scope]: <description>
<BLANK LINE>
[optional body]
<BLANK LINE>
[optional footer]
```

This convention enables automatic changelog generation and facilitates easy navigation of the git history. An example commit message:

```text
feat(api-server): add `--enable-metrics` option to enable metrics

Add `--enable-metrics` option to enable metrics, default is false.

Address #7
```

For more examples, refer to [here](https://www.conventionalcommits.org/en/v1.0.0/#examples).

#### Rules for Commit Messages

1. Separate the subject from the body with a blank line and the footer from the body with another blank line.
2. The subject line should not end with a period and should be limited to 70 characters.
3. Use the body to describe what, why, or how, wrapping it at 80 characters.
4. Keep all content lowercase except for `BREAKING CHANGE`.
5. Use the imperative mood to describe actions: "change," not "changed" or "changes."

#### Details of Message Subject

The subject line must include `type` and `scope`:

- Allowed `type` values:
  + **fix**: a bug fix.
  + **feat**: a new feature.
  + **chore**: updating code dependencies.
  + **docs**: changes to documentation.
  + **test**: changes to add or refactor tests.
  + **ci**: changes to CI and scripts.
  + **style**: changes that do not affect the code's meaning (e.g., white-space, formatting, missing semicolons).
  + **perf**: changes to improve performance.
  + **refactor**: changes that neither fix a bug nor add a feature (e.g., renaming a variable or extracting a method).
  + **revert**: changes that revert previous commits.
- Example `scope` values:
  + **api-server**/**cmd**/**cluster**: the related scope of the project.

`Scope` can be empty for changes that are global or difficult to assign to a single component.

#### Details of Message Body

The body should include the motivation for the change and contrasts with the previous behavior. For more information about the body message, please view: [Writing Good Commit Messages: A Practical Guide](https://www.freecodecamp.org/news/writing-good-commit-messages-a-practical-guide/)

#### Details of Message Footer

The footer should be used for:

- Referencing issues:
  Closed issues should be listed on a separate line prefixed with the `address` keyword like this:

    ```text
    address #7
    ``` 

  or in the case of multiple issues:

    ```text
    address #8, #9, #10
    ```

- Recording breaking changes:
  Breaking changes should be prefixed with the `BREAKING CHANGE` keyword like this:

    ```text
    BREAKING CHANGE: bump Go version to 1.13
    ```

  or use multiple lines to mention the description of the change, justification, and migration notes:

    ```text
    BREAKING CHANGE:
    
    `enable-metrics` option has dropped; therefore, monitoring metrics are enabled by default.
    
    To migrate your project, change all commands where you use `--enable-metrics`.
    ```

## Testing

There are multiple types of tests. The location of the test code varies with type, as do the specifics of the environment needed to successfully run the test:

- Unit: These confirm that a particular function behaves as intended. Unit test source code can be found adjacent to the corresponding source code within a given package.
- Integration: These tests cover interactions of package components or interactions between 1block.ai components like `api-server` and Kubernetes cluster. Integration test source code can be found in the `tests` directory.
- End-to-end ("e2e"): These are broad tests of overall system behavior and coherence. E2E test source code can be found in the `oneblock-ai/test` repo.

Continuous integration will run unit and integration tests on PRs.

### Unit Testing

Unit tests are usually easily run locally by any developer. Developed code can be PR'd after passing unit tests. You can run unit tests with:

```sh
$ make test
```

### Integration Testing

Integration testing is based on [the envtest of sigs.k8s.io/controller-runtime](https://book.kubebuilder.io/reference/testing/envtest.html), using [Ginkgo](http://onsi.github.io/ginkgo/), a testing framework that supports [Behavior-Driven Development(BDD)](https://en.wikipedia.org/wiki/Behavior-driven_development) style. You can run integration tests with:

```sh
$ make test && make envtest
```

### E2E Testing

TODO

### Documentation

Currently, 1Block.AI hosts its documentation on the [oneblock-ai/doc](https://github.com/oneblock-ai/docs) repo. Please submit your PR against this repo for documentation changes.
