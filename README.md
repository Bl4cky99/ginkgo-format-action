<a id="readme-top"></a>

<br />
<div align="center">
  <h1 align="center">ginkgo-format-action</h1>
  <p align="center">
    A GitHub Action that turns <b>Ginkgo JSON test reports</b> into a rich GitHub Actions step summary —<br />
    with per-suite breakdown, expandable failure details and structured outputs.
    <br /><br />
    <a href="https://github.com/Bl4cky99/ginkgo-format-action/issues/new?template=bug_report.yaml">Report Bug</a>
    &middot;
    <a href="https://github.com/Bl4cky99/ginkgo-format-action/issues/new?template=feature_request.yaml">Request Feature</a>
    <br /><br />
  </p>
</div>

<details>
<summary>Table of Contents</summary>
<ol>
  <li><a href="#features">Features</a></li>
  <li><a href="#usage">Usage</a>
    <ul>
      <li><a href="#quickstart">Quickstart</a></li>
      <li><a href="#inputs">Inputs</a></li>
      <li><a href="#outputs">Outputs</a></li>
    </ul>
  </li>
  <li><a href="#examples">Examples</a></li>
  <li><a href="#step-summary-preview">Step Summary Preview</a></li>
  <li><a href="#license">License</a></li>
</ol>
</details>

---

## <span id="features">Features</span>

- **Rich step summary**: renders a formatted Markdown summary directly to the GitHub Actions step summary panel.
- **Per-suite breakdown**: a table showing spec counts, pass / fail / skip / pending per suite and wall-clock duration.
- **Expandable failure details**: collapsible `<details>` blocks per failing spec with error message, source location and full spec path.
- **Structured outputs**: `total`, `passed`, `failed`, `skipped`, `pending`, `succeeded` — usable in downstream steps.
- **Fail on failures**: optionally exit with code `1` after rendering when any spec failed.
- **Configurable display**: toggle breakdown / failure details, cap the number of failure blocks rendered, set a custom heading title.
- **Graceful missing report**: emits a `::warning::` annotation and a helpful summary callout if no report file is found.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="usage">Usage</span>

### <span id="quickstart">Quickstart</span>

Run Ginkgo with `--json-report` to produce the report, then pass it to this action:

```yaml
- name: Run tests
  run: ginkgo --json-report=report.json ./...

- name: Format Ginkgo report
  uses: Bl4cky99/ginkgo-format-action@v1.0.2
  with:
    report-path: report.json
```

> Relative report paths are resolved against `GITHUB_WORKSPACE`.

---

### <span id="inputs">Inputs</span>

| Input | Required | Default | Description |
| :--- | :---: | :---: | :--- |
| `report-path` | no | `report.json` | Path to the Ginkgo JSON report file. Relative paths are resolved against `GITHUB_WORKSPACE`. |
| `title` | no | `Ginkgo Test Results` | Heading rendered above the summary. |
| `fail-on-failures` | no | `false` | Exit with code `1` after rendering if any spec failed. |
| `show-breakdown` | no | `true` | Render the per-suite breakdown table. |
| `show-failure-details` | no | `true` | Render expandable failure detail blocks. |
| `max-failure-details` | no | `0` | Maximum number of failure blocks to render. `0` means unlimited. |

---

### <span id="outputs">Outputs</span>

| Output | Type | Description |
| :--- | :---: | :--- |
| `total` | `int` | Total number of specs (`It` nodes). |
| `passed` | `int` | Number of passed specs. |
| `failed` | `int` | Number of failed specs (includes panicked, timed-out, aborted, interrupted). |
| `skipped` | `int` | Number of skipped specs. |
| `pending` | `int` | Number of pending specs. |
| `succeeded` | `bool` | `true` if no specs failed, `false` otherwise. |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="examples">Examples</span>

**Fail the workflow when tests fail**

```yaml
- name: Run tests
  id: tests
  run: ginkgo --json-report=report.json ./...
  continue-on-error: true   # let the formatter always run

- name: Format Ginkgo report
  uses: Bl4cky99/ginkgo-format-action@v1.0.2
  with:
    report-path: report.json
    fail-on-failures: "true"
```

> `continue-on-error: true` on the Ginkgo step ensures the formatter always runs even when tests fail, so the summary is always written before the workflow exits.

---

**Limit failure detail blocks**

When a suite has many failures, rendering all of them can produce a very long summary. Cap it with `max-failure-details`:

```yaml
- uses: Bl4cky99/ginkgo-format-action@v1.0.2
  with:
    report-path: report.json
    max-failure-details: "10"
```

A `> [!NOTE]` callout is appended when the limit is reached, telling the reader how many failures were omitted.

---

**Use outputs in downstream steps**

```yaml
- name: Format Ginkgo report
  id: ginkgo
  uses: Bl4cky99/ginkgo-format-action@v1.0.2
  with:
    report-path: report.json

- name: Post notification on failure
  if: steps.ginkgo.outputs.succeeded == 'false'
  run: |
    echo "Tests failed: ${{ steps.ginkgo.outputs.failed }}/${{ steps.ginkgo.outputs.total }}"
```

---

**Multiple suites — one report**

Run Ginkgo recursively and collect all suite results into a single JSON report:

```yaml
- name: Run all tests
  run: ginkgo --json-report=report.json -r ./...

- uses: Bl4cky99/ginkgo-format-action@v1.0.2
  with:
    report-path: report.json
    title: "Full Test Suite"
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="step-summary-preview">Step Summary Preview</span>

**Passed suite** — a `[!NOTE]` callout with total counts and duration:

> **[!NOTE]**  
> **Ginkgo Test Suite Passed**  
> **7** specs total │ Passed: **6** │ Failed: **0** │ Skipped: 1 │ Pending: 0  
> *Total Duration: 2.034s*

**Failed suite** — a `[!CAUTION]` callout followed by expandable failure blocks:

<details>
<summary><code>FAIL</code> <b>Token › validates expiry</b> — <em>auth/token_test.go:42</em></summary>
<br>

| Property | Details |
| :--- | :--- |
| **Spec** | ` Token › validates expiry ` |
| **State** | ` failed ` |
| **Error** | ` Expected <bool>: false to be true ` |

</details>

<br>

**Breakdown table** — one row per suite:

| Result | Suite | Specs | Passed | Failed | Skipped | Pending | Duration |
| :---: | :--- | ---: | ---: | ---: | ---: | ---: | ---: |
| ❌ | Auth Suite | 4 | 3 | 1 | 0 | 0 | 1.234s |
| ✔ | Storage Suite | 3 | 2 | 0 | 1 | 0 | 800ms |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="license">License</span>

This project is licensed under the **MIT License**.

- Copyright © 2026 [Jason Giese (Bl4cky99)](https://github.com/Bl4cky99)
- See the full text in [LICENSE](./LICENSE).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

**Happy testing!**
