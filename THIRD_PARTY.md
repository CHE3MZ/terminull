# third-party forks

terminull embeds the following upstream sources as internal engines. Forks are literal copies at the pinned version unless the modifications column notes adaptation. The upstream license is reproduced next to each fork.

| fork | upstream | version | commit | license | local path | forked | modifications |
| --- | --- | --- | --- | --- | --- | --- | --- |
| go-isatty | github.com/mattn/go-isatty | v0.0.24 | c44dc0b9c702c76577fdb7898032969e0611efc2 | MIT | termcore/internal/isatty | 2026-10-22 | vendored as-is; tests omitted |
| go-colorable | github.com/mattn/go-colorable | v0.1.16 | 56c5db567780daeef4a4ffbb2c14d7d4b6b94b9b | MIT | termcore/internal/colorable | 2026-10-22 | isatty import repointed to termcore/internal/isatty; tests, cmd, examples omitted |
| go-runewidth | github.com/mattn/go-runewidth | v0.0.31 | 466251be9763946c75d1d9491e190cc5e80cf2f8 | MIT | termcore/internal/width | 2026-10-22 | uax29 import repointed to termcore/internal/uax29/graphemes; regen script kept under script/; tests omitted |
| uax29 | github.com/clipperhouse/uax29/v2 | v2.7.0 | b03477d1fbba89df95a6b55da0a222b2e2228610 | MIT | termcore/internal/uax29/graphemes | 2026-10-22 | graphemes package only; tests omitted |
| shlex | github.com/google/shlex | e7afc7f (2019-12-02) | e7afc7fbc51079733e9468cdfd1efcd7d196cd1d | Apache-2.0 | termcore/internal/shlex | 2026-10-22 | vendored as-is; tests omitted |
| echelon | github.com/cirruslabs/echelon | v1.9.0 | 87c7b423b345e0900da54f9243316589bb5b4de4 | MIT | termrender | 2026-10-14 | import paths and formatting changed; x/text usage removed; CI fold decorators added; tests reworked |
| cirrus-cli foldable logs | github.com/cirruslabs/cirrus-cli | v1.0.0 | 1769788a20dc72f6aef85bcdc92abaa1bf061cdc | MIT | termrender/renderers/ci.go | 2026-10-14 | foldable log decorators only, adapted to termrender renderer interfaces |

Fork dates are per row; upstream licenses sit beside each tree (`termrender/LICENSE`, `termrender/renderers/LICENSE.cirrus-cli`, and the `LICENSE`/`COPYING` files under `termcore/internal`).

Forked trees are excluded from repo lint, staticcheck, errcheck, and gosec; upstream CI and the advisory watch are their quality signal.

Maintenance: upstream fixes are cherry-picked only for a bug we hit or a capability we document. Because the forks are renamed, osv and govulncheck cannot see them; `task upstream` is the advisory sweep. Update this table whenever a fork version changes.
