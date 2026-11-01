# txtcv

txtcv is a CLI-first plaintext CV builder for software developers.

## Export

`txtcv export` renders the CV — the exact same render `serve` previews — to an
unwatermarked, ATS-friendly PDF, entirely locally. There are no limits and no
account needed.

```sh
txtcv export                        # cv.yaml → alice-cv.pdf (a4)
txtcv export --paper letter         # US letter paper
txtcv export --out 2026-acme.pdf    # pick the output path
txtcv export cv.json                # explicit file instead of cv.yaml/cv.json
```

The PDF is **text-selectable, tagged, and carries a document outline** from the
headings, so ATS parsers and screen readers read the real structure. Section
headings are the ones parsers look for (`EXPERIENCE`, `EDUCATION`, `SKILLS`,
`CERTIFICATIONS`), keyword lists extract as plain text, and there are no
headers, footers, page numbers, or watermarks. Unlimited exports, forever.

### The browser

PDF printing needs a locally installed Chromium-family browser (Google Chrome,
Chromium, Edge, Brave); txtcv finds it automatically. To choose one yourself:

```sh
TXTCV_BROWSER=/path/to/my/chrome txtcv export
```

Discovery never touches the network: the CV is handed to the browser over a
`file://` URL and the browser talks only to your own machine.

No browser installed? The manual fallback still produces the same PDF:

```sh
txtcv serve     # open http://localhost:9090, then ⌘P → Save as PDF
```

The page's print CSS already sets the right paper, so the browser's print
dialog needs no geometry fiddling.
