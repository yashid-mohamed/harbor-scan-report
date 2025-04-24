# Harbor Scan Report Action

One of the fantastic features of [Harbor](https://goharbor.io/) is its integration with various vulnerability scanners. These scanners provide a list of vulnerabilities found in container images.

This action retrieves the scan report and optionally posts a comment with the scan results for a given Docker image.

Currently, it can comment on Pull Requests (PRs) and issues.

This action consists of two parts:
1. Retrieving the scan report.
2. Generating a GitHub comment (optional).

## Why is it useful?

* Your image is stored in [Harbor](https://goharbor.io/).
* You want to prevent insecure images from being deployed.

## Examples

### Clean Image
![CleanImage](clean-image.png?raw=true)

### Vulnerable Image
![VulnerableImage](vulnerable-image.png)

## Configuration

### Minimal Valid Example

```yaml
- name: Run Report
  uses: kyberorg/harbor-scan-report@v0.1
  with:
    harbor-host: my_harbor.tld
    image: my_harbor.tld/hub/redhat/ubi8:latest
```

### Full Example

```yaml
- name: Run Report
  uses: kyberorg/harbor-scan-report@v0.1
  with:
    harbor-host: my_harbor.tld
    harbor-proto: http
    harbor-port: 8080
    harbor-robot: ${{ secrets.HARBOR_ROBOT }}
    harbor-token: ${{ secrets.HARBOR_TOKEN }}
    image: my_harbor.tld:8080/hub/redhat/ubi8:latest
    digest: sha256:01814f4b10f321f09244a919d34b0d5706d95624b4c69d75866bb9935a89582d
    timeout: 150
    check-interval: 10
    max-allowed-severity: high
    report-sort-by: score
    report-only-fixable: true
    github-url: ${{ github.event.pull_request.comments_url }}
    github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Inputs

### `harbor-host`

The hostname of the Harbor instance, without protocol or port.

Required: `yes`

### `image`

Image to scan. Format: `registry.tld/project/repo:tag[@sha256:digest]` or `project/repo:tag[@sha256:digest]`.

Tag is optional, if tag missing the default tag `latest` will be used.  

Digest is also optional. You can set it using the [digest input](#digest)

Required: `yes`

### `digest`

Specifies the expected image digest. This is useful for rolling tags

Action will wait until image has given digest.

Format: `sha256:01814f4b10f321f09244a919d34b0d5706d95624b4c69d75866bb9935a89582d`

Required: `no`

### `harbor-robot`

The robot account or username to access Harbor. A robot account with limited privileges is preferred.

Without credentials, the action can only access public repositories.

Required: `no`

### `harbor-token`

The token for the robot account or the password for the user defined in `harbor-robot`.

This parameter is optional, but without credentials action can access public repositories only.

Required: `no`

### `max-allowed-severity`

Maximum Vulnerability severity after which action considered as failed.

Valid values: `none`, `low`, `medium`, `high`, `critical`

Default value: `critical`

* `none` means zero-tolerance to any vulnerabilities i.e. action succeeds only if image hasn't any vulnerabilities.

* `critical` means that action never fails, even if image has critical vulnerabilities.

### `report-sort-by`

The sorting criteria for the vulnerability report.

Valid Values: `severity`, `score`

Default value: `severity`

* `severity` means that report will be sorted by Harbor's Severity
* `score` means that report will be sorted by CVSSv3 Score

### `report-only-fixable`

If enabled (set to true), the vulnerability report will only include items with a fix version available.

Valid values: `true` and `false`

Default value: `false`

Required: `no`

### `github-url`

The GitHub API endpoint to use. Typically, you would use built-in variables:
- `github.event.issue.comments_url` for commenting on issues.
- `github.event.pull_request.comments_url` for commenting pull requests.

If not defined, commenting mode is disabled.

Required: `no`

```yaml
github-url: ${{ github.event.pull_request.comments_url }}
```

### `github-token`

A GitHub personal access token used to comment on your behalf. Normally, this is added to the action's secrets as `${{ secrets.GITHUB_TOKEN }}`

```yaml
github-token: ${{ secrets.GITHUB_TOKEN }}
```
Note: This token must have the following permissions to comment on the issue or pull request.

```yaml
permissions:
  contents: write
  pull-requests: write
```

### `comment-title`

String that is used as comment title.

Default: `Docker Image Vulnerability Report`

### `comment-mode`

Specifies the behavior for subsequent runs. The action can create a new comment or update the last one.

Valid values: `create_new` and `update_last`

Default value: `create_new`

### `timeout`

The time in seconds after which the action fails. This includes waiting for the image to appear in Harbor and for the scan report to be ready.

Must be a positive integer.

Default: `120` (2 minutes)

### `check-interval`

The time in seconds between requests to Harbor. This applies to waiting for the image to appear and for the scan report to be ready.

Must be a positive integer.

Default: `5` (5 seconds)

### `harbor-proto`

The protocol of the Harbor instance. Use this if your Harbor instance is accessible only via `http`.

Valid values: `http` and `https`.

Default value is `https`. 

### `harbor-port`

The custom port of the Harbor instance. Use this if the Harbor instance uses a non-default port.

Any port within `0-65535` range.
