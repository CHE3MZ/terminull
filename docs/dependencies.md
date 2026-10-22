# external (v2 target):

gopkg.in/yaml.v3 v3.0.1 ( file codec , config and output )
golang.org/x/sys v0.49.0 ( windows console + platform syscalls )
golang.org/x/term v0.44.0 ( terminal size + raw mode )
log/slog , fmt , and the rest of the standard library ( comes with go )

nothing else ships. after phase 14 the module graph holds these three plus stdlib.

# forked ( owned , literal , internal ):

mattn/go-isatty v0.0.24 → termcore/internal/isatty ( MIT )
mattn/go-colorable v0.1.16 → termcore/internal/colorable ( MIT )
mattn/go-runewidth v0.0.31 → termcore/internal/width + unicode tables ( MIT )
google/shlex e7afc7fbc510 → termcore/internal/shlex ( Apache-2.0 )
fatih/color v1.19.0 → termstyle/internal/color ( MIT )
avast/retry-go/v4 v4.7.0 → termretry/internal/retry ( MIT )
dustin/go-humanize v1.0.1 → termoutput/internal/humanize ( MIT )
peterh/liner v1.2.3 → termprompt/internal/lineedit ( MIT )
knadh/koanf/v2 v2.3.8 → termconfig/internal/koanf ( MIT )
knadh/koanf/parsers/yaml v1.1.1 → termconfig/internal/koanf/yaml ( MIT )
knadh/koanf/providers/{confmap,env/v2,file} → termconfig/internal/koanf ( MIT )
go-viper/mapstructure/v2 v2.4.0 → termconfig/internal/mapstructure ( MIT )
spf13/cobra v1.10.2 → termapp/internal/cli ( Apache-2.0 )
spf13/pflag v1.0.9 → termapp/internal/cli/pflag ( BSD-3-Clause )

# removed by v2:

cobra , pflag , koanf ( all ) , mapstructure , go.yaml.in/yaml/v3 , liner , retry-go , humanize , shlex , fatih/color , isatty , colorable , runewidth , x/text.

stale indirects cleared by tidy: fsnotify , clipperhouse/uax29 , mousetrap , testify.

# policy:

forks are literal copies pinned by version, minimally modified at integration seams, and never surfaced in public apis.
provenance lives in THIRD_PARTY.md ( repo , version , commit , license , fork date ).
task upstream sweeps the recorded origins for releases and advisories, because renamed modules fall out of osv and govulncheck visibility.
gopkg.in/yaml.v3 and golang.org/x/* stay external by design; everything else is owned.
see docs/v2-plan.md for the phase order and acceptance gates.
