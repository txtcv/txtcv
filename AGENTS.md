# AGENTS.md

txtcv is a CLI-first plaintext CV builder for developers: a JSON Resume
`cv.json`/`cv.yaml` on disk, rendered to themed HTML and printed to unwatermarked PDF
locally; paid AI features (tailoring, cover letters) come from a separate hosted server.

## What txtcv is about

- **Plaintext all the way down.** The CV is a file you own — versioned in git, edited in
  any editor. Variants, job applications, notes, and cover letters are files too:
  greppable, git-friendly, no database, no lock-in.
- **The CLI is the moat.** Free, offline, useful year-round even when not job hunting.
  The free tier is a complete product, not a trial.
- **Paid features are power-ups during a job search** (AI tailoring, AI cover letters),
  served by a lightweight hosted server. Monetization sells the sprint, not the year: a
  time-bounded pass, never an auto-renewing subscription.

## Product features

- **Commands:** `init`, `validate`, `lint` (offline style heuristics), `serve` (live
  preview + hot reload, default port 9090), `export` (PDF/HTML/MD/TXT), `theme`
  (list/set/path), `variant` (create/list/rm), a job tracker (`job`, `jobs`, `note`,
  `interview`, `board`, `stats`), `match` (local keyword-match report vs a job
  description), and Pro commands `login`/`logout`/`whoami`, `tailor`, `letter`.
- **Export** is unlimited and unwatermarked, produced locally — the founding credibility
  claim.
- **Themes:** CSS custom properties over one shared semantic HTML template. Embedded
  defaults plus runtime-loadable themes from `./themes/<name>/` — no recompile needed.
- **Variants:** an overlay merged onto the base CV with RFC 7396 (JSON Merge Patch)
  semantics — objects deep-merge, arrays replace wholesale, `null` deletes a key. Escape
  hatch: `meta.txtcv.variant.mode: full` (standalone CV, no merge). `serve`, `export`,
  and `tailor --apply` all operate on base ⊕ overlay.
- **Job tracker:** one markdown file per application (`jobs/<slug>.md`) — YAML
  frontmatter for structure (company, role, status, dates, variant link), body for
  free-form notes and interview logs.
- **Pro (server-backed):** AI tailoring against a JD — suggestions rendered as a
  reviewable diff; `--apply` writes only the variant overlay, never the base CV. AI
  cover letters land as markdown drafts in `letters/`. Interview prep, funnel analytics,
  and premium theme packs round out the paid tier.

## Technology stack

- Go, shipped as a **single static binary**. Module path: `github.com/txtcv/txtcv`.
- urfave/cli/v3 for commands; goccy/go-yaml for YAML; draft-07 JSON Resume schema
  embedded via `go:embed` and validated with google/jsonschema-go; partial dates
  (`2006`, `2006-01`, `2006-01-01`) accepted.
- **PDF export via chromedp** (pure-Go CDP client) with installed-browser discovery
  (Chrome/Edge/Brave/Chromium); `TXTCV_BROWSER=/path` overrides; fallbacks down to the
  manual print path from `serve`. No Node, no Puppeteer — the single-binary ethos also
  rules out npm `jsonresume-theme-*` packages.
- `meta.txtcv` is the vendor-extension namespace on the JSON Resume schema (`theme`,
  `page`, `job`, `variant.mode`).
- Distribution: goreleaser + Homebrew tap; `go install github.com/txtcv/txtcv@latest` as
  the fallback path.
- **Repo boundary:** this repo is the OSS CLI. The paid server (device-flow auth,
  stateless LLM endpoints that never store CV content) is a separate private repository.

## Product principles that are engineering rules

- **Free-tier CLI commands make zero network calls.** Only `login`/`tailor`/`letter`
  (Pro) touch the network. Enforce in CI: fail if any non-Pro package imports a network
  client; the tracker package must not import `net/http` at all.
- **No dark patterns, ever:** no watermarks, no PDF paywalls, no auto-renewing trials,
  no data hostage-taking.
- **Free/pro rule:** useful while maintaining a CV → free; useful while using a CV to
  get a job → paid. Local code is free; server compute is paid.
- **The LLM suggests; the human decides.** AI output is suggestions reviewed as diffs,
  never direct writes to CV files.
- **Privacy is architectural:** the server is stateless and never stores CVs — "your CV
  is never stored on our servers" must remain literally true.

## Testing conventions

- Golden-file tests: every theme renders a fixture CV to HTML; snapshots are committed;
  drift fails CI.
- Property tests for merge-patch composition (base ⊕ overlay must validate against the
  schema); tracker file mutations must be idempotent and slug-stable.
- PDF smoke tests skip gracefully when no browser is available (e.g., in CI).
