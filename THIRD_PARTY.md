# third-party forks

terminull embeds the following upstream sources as internal engines. Each copy is literal at the pinned version: only import paths and formatting changed, plus the omissions noted below. The upstream license is reproduced next to each fork.

| fork | upstream | version | license | local path | modifications |
| --- | --- | --- | --- | --- | --- |
| go-isatty | github.com/mattn/go-isatty | v0.0.24 | MIT | termcore/internal/isatty | vendored as-is; tests omitted |
| go-colorable | github.com/mattn/go-colorable | v0.1.16 | MIT | termcore/internal/colorable | isatty import repointed to termcore/internal/isatty; tests, cmd, examples omitted |
| go-runewidth | github.com/mattn/go-runewidth | v0.0.31 | MIT | termcore/internal/width | uax29 import repointed to termcore/internal/uax29/graphemes; regen script kept under script/; tests omitted |
| uax29 | github.com/clipperhouse/uax29/v2 | v2.7.0 | MIT | termcore/internal/uax29/graphemes | graphemes package only; tests omitted |
| shlex | github.com/google/shlex | e7afc7fbc510 (2019-12-02) | Apache-2.0 | termcore/internal/shlex | vendored as-is; tests omitted |

Forked: 2026-10-22.

Forked trees are excluded from repo lint, staticcheck, errcheck, and gosec; upstream CI and the advisory watch are their quality signal.

Maintenance: upstream fixes are cherry-picked only for a bug we hit or a capability we document. Because the forks are renamed, osv and govulncheck cannot see them; `task upstream` is the advisory sweep. Update this table whenever a fork version changes.
