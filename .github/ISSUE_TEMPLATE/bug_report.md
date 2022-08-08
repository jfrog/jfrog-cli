---
name: "🐞 Bug Report"
description: Create a report to help us improve
title: "(short issue description)"
labels: [bug]
assignees: []
body:
- type: textarea
  id: description
  attributes:
  label: Describe the bug
  description: What is the problem? A clear and concise description of the bug.
  validations:
  required: true

- type: textarea
  id: current
  attributes:
  label: Current Behavior
  description: |
  What actually happened?

      Please include full errors, uncaught exceptions, stack traces, and relevant logs.
      Using environment variable 'JFROG_CLI_LOG_LEVEL="DEBUG"' upon running the command will increase the log level.
  validations:
  required: false

- type: textarea
  id: reproduction
  attributes:
  label: Reproduction Steps
  description: |
  Provide a self-contained, concise snippet of code that can be used to reproduce the issue.
  For more complex issues provide a repo with the smallest sample that reproduces the bug.

      Avoid including business logic or credentials.
  validations:
  required: false
- type: textarea
  id: solution
  attributes:
  label: Possible Solution
  description: |
  Suggest a fix/reason for the bug
  validations:
  required: false

- type: textarea
  id: expected
  attributes:
  label: Expected Behavior
  description: |
  What did you expect to happen?
  validations:
  required: false

- type: input
  id: environment
  attributes:
  label: JFrog CLI version
  validations:
  required: false

- type: input
  id: environment
  attributes:
  label: OS
  validations:
  required: false

- type: input
  id: environment
  attributes:
  label: JFrog Artifactory version
  validations:
  required: false

- type: input
  id: environment
  attributes:
  label: JFrog Xray version
  validations:
  required: false