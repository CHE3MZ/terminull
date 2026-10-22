# third-party forks

terminull embeds the following upstream sources as internal engines. Forks are literal copies at the pinned version unless the modifications column notes adaptation. The upstream license is reproduced next to each fork.

| fork | upstream | version | commit | license | local path | forked | modifications |
| --- | --- | --- | --- | --- | --- | --- | --- |
| go-isatty | github.com/mattn/go-isatty | v0.0.24 | c44dc0b9c702c76577fdb7898032969e0611efc2 | MIT | termcore/internal/isatty | 2026-10-22 | vendored as-is; tests omitted |
| go-colorable | github.com/mattn/go-colorable | v0.1.16 | 56c5db567780daeef4a4ffbb2c14d7d4b6b94b9b | MIT | termcore/internal/colorable | 2026-10-22 | isatty import repointed to termcore/internal/isatty; tests, cmd, examples omitted |
| go-runewidth | github.com/mattn/go-runewidth | v0.0.31 | 466251be9763946c75d1d9491e190cc5e80cf2f8 | MIT | termcore/internal/width | 2026-10-22 | uax29 import repointed to termcore/internal/uax29/graphemes; regen script kept under script/; tests omitted |
| uax29 | github.com/clipperhouse/uax29/v2 | v2.7.0 | b03477d1fbba89df95a6b55da0a222b2e2228610 | MIT | termcore/internal/uax29/graphemes | 2026-10-22 | graphemes package only; tests omitted |
| shlex | github.com/google/shlex | e7afc7f (2019-12-02) | e7afc7fbc51079733e9468cdfd1efcd7d196cd1d | Apache-2.0 | termcore/internal/shlex | 2026-10-22 | vendored as-is; tests omitted |
| color | github.com/fatih/color | v1.19.0 | ca25f6e17f118a5a259f3c2c0d395949d1103a5a | MIT | termstyle/internal/color | 2026-10-22 | imports repointed to termcore for tty detection and console writers; Windows init delegates to termcore.EnsureConsole; tests omitted |
| retry-go | github.com/avast/retry-go/v4 | v4.7.0 | 375037b55dceec39815fd3d0c5bfbccf30255a38 | MIT | termretry/internal/retry | 2026-10-22 | production retry.go and options.go only; tests and examples omitted |
| go-humanize | github.com/dustin/go-humanize | v1.0.1 | 9ec74ab2f7a7161664182fd4e5e292fccffbc75f | MIT | termoutput/internal/humanize | 2026-10-22 | bytes.go and package documentation only; remaining formats and tests omitted |
| liner | github.com/peterh/liner | v1.2.3 | 67e095d9354414a42bb18709a0f6df73e328f6b4 | MIT | termprompt/internal/lineedit | 2026-10-22 | width uses termcore.RuneWidth; Solaris keeps the plain fallback instead of the x/sys editing path; tests omitted |
| koanf | github.com/knadh/koanf/v2 | v2.3.8 | 182a04e8fdcaf5cec7cfcfd62254ea97efd8ac4b | MIT | termconfig/internal/koanf | 2026-10-22 | production getters/interfaces/koanf/options only; imports repointed to internal maps, mapstructure, and copystructure; tests omitted |
| koanf maps | github.com/knadh/koanf/maps | v0.1.2 | d7a5fdf75d575ab06cec33ddc54dd28c3b140589 | MIT | termconfig/internal/koanf/maps | 2026-10-22 | copystructure import repointed; tests omitted |
| koanf yaml parser | github.com/knadh/koanf/parsers/yaml | v1.1.1 | 75e13d9bb6c95f62c3d8f7d871f8e0217dc38185 | MIT | termconfig/internal/koanf/yaml | 2026-10-22 | parser switched from go.yaml.in/yaml/v3 to gopkg.in/yaml.v3; tests omitted |
| koanf confmap provider | github.com/knadh/koanf/providers/confmap | v1.0.1 | 75e13d9bb6c95f62c3d8f7d871f8e0217dc38185 | MIT | termconfig/internal/koanf/confmap | 2026-10-22 | maps import repointed to the internal fork; tests omitted |
| koanf env provider | github.com/knadh/koanf/providers/env/v2 | v2.0.2 | 3e00cf1b07537ffe51b3d9b82a50e4cd00a5b746 | MIT | termconfig/internal/koanf/env | 2026-10-22 | maps import repointed to the internal fork; tests omitted |
| koanf file provider | github.com/knadh/koanf/providers/file | v1.2.1 | 9c90b76c2bb5106019000c2b60f42f807d6e4fa8 | MIT | termconfig/internal/koanf/file | 2026-10-22 | Watch/Unwatch and fsnotify dependency removed; load-only provider |
| mapstructure | github.com/go-viper/mapstructure/v2 | v2.4.0 | b9794a5f0e73d425210d6614ed833067029155f5 | MIT | termconfig/internal/mapstructure | 2026-10-22 | internal errors import repointed to the local fork; tests omitted |
| copystructure | github.com/mitchellh/copystructure | v1.2.0 | d4ce1f938f7a7ea2a40bff4544b56be9c00b5e84 | MIT | termconfig/internal/copystructure | 2026-10-22 | reflectwalk import repointed to the local fork; tests omitted |
| reflectwalk | github.com/mitchellh/reflectwalk | v1.0.2 | e0c24fdb021963cd2c4013097a0b993a7c4e344f | MIT | termconfig/internal/reflectwalk | 2026-10-22 | vendored as-is; tests omitted |
| cobra | github.com/spf13/cobra | v1.10.2 | 88b30ab89da2d0d0abb153818746c5a2d30eccec | Apache-2.0 | termapp/internal/cli | 2026-10-22 | package renamed cli; pflag import repointed to the internal fork; mousetrap win hook and variables removed; tests, docs, and generators omitted |
| pflag | github.com/spf13/pflag | v1.0.9 | 10438578954bba2527fe5cae3684d4532b064bbe | BSD-3-Clause | termapp/internal/cli/pflag | 2026-10-22 | non-constant Fprintf calls replaced with Fprint to satisfy go vet; tests omitted |
| echelon | github.com/cirruslabs/echelon | v1.9.0 | 87c7b423b345e0900da54f9243316589bb5b4de4 | MIT | termrender | 2026-10-14 | import paths and formatting changed; x/text usage removed; CI fold decorators added; tests reworked |
| cirrus-cli foldable logs | github.com/cirruslabs/cirrus-cli | v1.0.0 | 1769788a20dc72f6aef85bcdc92abaa1bf061cdc | MIT | termrender/renderers/ci.go | 2026-10-14 | foldable log decorators only, adapted to termrender renderer interfaces |

Fork dates are per row; upstream licenses sit beside each tree (`termrender/LICENSE`, `termrender/renderers/LICENSE.cirrus-cli`, `termstyle/internal/color/LICENSE.md`, `termretry/internal/retry/LICENSE`, `termoutput/internal/humanize/LICENSE`, `termprompt/internal/lineedit/COPYING`, the `LICENSE`/`COPYING` files under `termcore/internal`, the `LICENSE` files under `termconfig/internal`, and `termapp/internal/cli/LICENSE.txt` plus `termapp/internal/cli/pflag/LICENSE`).

Forked trees are excluded from repo lint, staticcheck, errcheck, and gosec; upstream CI and the advisory watch are their quality signal.

Maintenance: upstream fixes are cherry-picked only for a bug we hit or a capability we document. Because the forks are renamed, osv and govulncheck cannot see them; `task upstream` is the advisory sweep. Update this table whenever a fork version changes.
