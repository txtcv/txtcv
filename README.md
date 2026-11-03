# txtcv

txtcv is a CLI-first plaintext CV builder for software developers.

## Preview and save as PDF

Create a CV and preview it in your browser:

```sh
txtcv init
txtcv serve
```

Open http://localhost:9090, then use your browser’s Print command
(⌘P on macOS, Ctrl+P on Windows/Linux) and choose **Save as PDF**.

The preview includes print CSS with A4 paper, margins, and page-break rules.
Check the print preview, use 100% scale, and disable browser-added headers and
footers. Enable background graphics to preserve the page’s colors.

PDF creation happens locally in your browser, with no account, watermark, or
export limit. Built-in PDF export is deferred; txtcv does not launch or manage
a browser.
