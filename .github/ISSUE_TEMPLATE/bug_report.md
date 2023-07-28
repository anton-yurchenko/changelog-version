---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: anton-yurchenko

---

**Describe the bug**
A clear and concise description of what the bug is.

**GitHub Actions workflow**
```yaml
name: release
on:
  push:
    tags:
      - v[0-9]+.[0-9]+.[0-9]+

jobs:
  release:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
```

**Input values (please complete the following information):**
 - Version: `v1.2.3`
 - Update Tags: `<true/false>`

**Changelog content**
Content of the changelog file (feel free to _mask_ it).

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Additional context**
Add any other context about the problem here.
