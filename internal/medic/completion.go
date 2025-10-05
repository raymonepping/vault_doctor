package medic

import (
	"fmt"
	"os"
)

func PrintCompletion(shell string) int {
	switch shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported shell: %s (use bash|zsh|fish)\n", shell)
		return 1
	}
	return 0
}

const bashCompletion = `# bash completion for vault_doctor
_vault_doctor()
{
    local cur prev words cword
    _init_completion || return

    local subcmds="medic completion -h --help -V --version"
    local global_flags="-h --help -V --version"
    local medic_flags="--json --quiet --no-color"

    if [[ ${#COMP_WORDS[@]} -le 2 ]]; then
        COMPREPLY=( $(compgen -W "${subcmds}" -- "$cur") )
        return
    fi

    case "${COMP_WORDS[1]}" in
        medic)
            COMPREPLY=( $(compgen -W "${medic_flags}" -- "$cur") )
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
            ;;
        *)
            COMPREPLY=( $(compgen -W "${global_flags}" -- "$cur") )
            ;;
    esac
}
complete -F _vault_doctor vault_doctor
`

const zshCompletion = `#compdef vault_doctor

_arguments -C \
  '1: :((medic\:Run\ diagnostics completion\:Generate\ shell\ completions -h\:\:Help --help\:\:Help -V\:\:Version --version\:\:Version))' \
  '*::arg:->args'

case $words[2] in
  medic)
    _values 'flags' --json --quiet --no-color
    ;;
  completion)
    _values 'shell' bash zsh fish
    ;;
  *)
    _values 'global' -h --help -V --version
    ;;
esac
`

const fishCompletion = `# fish completion for vault_doctor
complete -c vault_doctor -f -n "__fish_use_subcommand" -a "medic" -d "Run diagnostics"
complete -c vault_doctor -f -n "__fish_use_subcommand" -a "completion" -d "Generate shell completions"

# medic flags
complete -c vault_doctor -n "__fish_seen_subcommand_from medic" -l json -d "Output JSON"
complete -c vault_doctor -n "__fish_seen_subcommand_from medic" -l quiet -d "Quiet mode"
complete -c vault_doctor -n "__fish_seen_subcommand_from medic" -l no-color -d "Disable colors"

# completion args
complete -c vault_doctor -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
`
