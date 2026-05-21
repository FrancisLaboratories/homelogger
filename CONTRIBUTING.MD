# Contributing to HomeLogger

Thank you for your interest in contributing! To ensure a productive environment and to save you from doing work that might not be merged, please review these guidelines before starting.

---

## The "Before You Code" Rule

To keep the project focused and maintainable, we have specific rules regarding Pull Requests:

### Bug Fixes
* **Unsolicited PRs:** Generally accepted after consideration! If you find a bug and have a fix, feel free to submit a PR.
* **Process:** We strongly recommend discussing the bug in a **GitHub Issue** first to ensure the fix aligns with our architecture.

### New Features
* **Unsolicited PRs:** **Please do not submit a PR for a new feature without prior discussion + approval.**
* **Process:** If you have an idea for a feature, please open a [Feature Request Issue/Discussion] first. I will only accept PRs for features that I have explicitly indicated I am willing to merge. 
* **Why?** This prevents you from wasting time on a feature that may not fit the project's long-term roadmap and design.

### General Feedback & Questions
* All general feedback, "How-to" questions, or non-technical ideas should be kept in **GitHub Discussions**.


### Developer Certificate of Origin (DCO)
All contributions must comply with the [Developer Certificate of Origin (DCO)](DCO.md). By contributing, you certify that you have the right to submit the work under the project's open source license and that you agree to the terms outlined in the DCO.

---

## AI-Generated Code Policy

We value human oversight. If you use AI tools (GitHub Copilot, ChatGPT, Claude Code, etc.), you must follow the guidelines in the [AI Policy](AI_POLICY.md) file.

---

## Pull Request Process

1.  **Check for an Issue:** Ensure there is an existing Issue or Discussion for your work. Ensure you have notified the maintainer that you plan to work on this issue.
2.  **Fork & Branch:** Create a branch from `main`.
3.  **Tests:** Ensure all existing tests pass and add new ones if applicable.
4.  **Describe:** Clearly explain what your PR does and link to the relevant Issue.

---

## Style & Standards
* Follow the existing code style and naming conventions.
* Keep commits clean and descriptive. If I can't figure out the clear purpose of your PR, it's not getting merged. 
* Code in the `client` folder is formatted by [Prettier](https://prettier.io/docs/)
* Code is scanned by SonarQube for quality. We are shooting for 80% quality rating. Please fix as many issues on your branch as possible before submitting a PR.

---

## Tests

Please make sure your code includes or updates any tests. We are shooting for 80% test coverage.

Note: Building tests into the project is currently a work in progress. There is a lot of code that doesn't have tests at this time. With that said, I will be implementing tests soon and will require updated tests for new code and any code that is being modified.

---

**Note:** Quality is prioritized over quantity. I would much rather see one well-documented bug fix than five unrequested features!
