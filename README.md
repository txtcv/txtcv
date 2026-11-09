# txtcv

txtcv is a CLI-first plaintext CV builder for software developers.

## Preview and save as PDF

Create a CV and preview it in your browser:

```sh
txtcv init
txtcv validate
txtcv serve
```

`txtcv validate` checks `cv.yaml` (preferred) or `cv.json` in the current
directory against the bundled JSON Resume schema, entirely offline. To check
another file, run `txtcv validate path/to/resume.yaml` (also supports `.yml` and
`.json`). Validation prints a success message, or reports an error and exits
with status 1 for an invalid, malformed, or unreadable CV. Partial dates such
as `2024`, `2024-06`, and `2024-06-15` are supported.

Open http://localhost:9090, then use your browser’s Print command
(⌘P on macOS, Ctrl+P on Windows/Linux) and choose **Save as PDF**.

Choose a theme in your CV (`default`, `minimal`, `code`, or `creative`):

```yaml
meta:
  txtcv:
    theme: code
```

`txtcv serve --theme creative` overrides the CV's theme. Without either setting,
the theme is `default`. Refreshing the preview picks up changes to the CV's theme.

The preview includes print CSS with A4 paper, margins, and page-break rules.
Check the print preview, use 100% scale, and disable browser-added headers and
footers. Enable background graphics to preserve the page’s colors.

PDF creation happens locally in your browser, with no account, watermark, or
export limit. Built-in PDF export is deferred; txtcv does not launch or manage
a browser.
