#!/usr/bin/env bash
#
# Bash completion for deploy.sh (delivery-pulse)
#
# Source this in your .bashrc or .bash_profile:
#   source /path/to/delivery-pulse/deploy-completion.bash
#
# Or if you've aliased deploy.sh:
#   alias dpulse='/path/to/delivery-pulse/deploy.sh'
#   complete -F _deploy_sh_completions dpulse

_deploy_sh_completions() {
  local cur prev commands envs flags
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"

  commands="install uninstall help"
  envs="local dev"
  flags="--test -t"

  # First argument: command
  if [ "$COMP_CWORD" -eq 1 ]; then
    COMPREPLY=($(compgen -W "$commands" -- "$cur"))
    return 0
  fi

  # Second argument: environment (for install/uninstall)
  if [ "$COMP_CWORD" -eq 2 ]; then
    case "${COMP_WORDS[1]}" in
      install|uninstall)
        COMPREPLY=($(compgen -W "$envs" -- "$cur"))
        ;;
    esac
    return 0
  fi

  # Third+ argument: flags (only for install)
  if [ "$COMP_CWORD" -ge 3 ] && [ "${COMP_WORDS[1]}" == "install" ]; then
    COMPREPLY=($(compgen -W "$flags" -- "$cur"))
    return 0
  fi
}

# Register completion for both the script name and common aliases
complete -F _deploy_sh_completions deploy.sh
complete -F _deploy_sh_completions dpulse
