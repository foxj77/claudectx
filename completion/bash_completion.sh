#!/usr/bin/env bash

# Bash completion for claudectx
# To install: source this file or copy to /etc/bash_completion.d/claudectx

_claudectx() {
    local cur prev opts profiles
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Main commands and flags
    opts="-h --help -v --version -c --current -n -d - export import"

    # Get list of profiles (if claudectx is in PATH)
    if command -v claudectx &> /dev/null; then
        profiles=$(claudectx 2>/dev/null | awk '{print $1}')
    fi

    case "${prev}" in
        -n)
            # After -n, suggest a name (no completion)
            return 0
            ;;
        -d)
            # After -d, suggest profiles to delete
            COMPREPLY=( $(compgen -W "${profiles}" -- ${cur}) )
            return 0
            ;;
        export)
            # After export, suggest profiles
            COMPREPLY=( $(compgen -W "${profiles}" -- ${cur}) )
            return 0
            ;;
        import)
            # After import, suggest files
            COMPREPLY=( $(compgen -f -X '!*.json' -- ${cur}) )
            return 0
            ;;
        claudectx)
            # First argument: suggest commands and profiles
            COMPREPLY=( $(compgen -W "${opts} ${profiles}" -- ${cur}) )
            return 0
            ;;
        *)
            # Check if we're completing after 'export <profile>'
            if [ "${COMP_WORDS[COMP_CWORD-2]}" = "export" ]; then
                # Suggest filename for export
                COMPREPLY=( $(compgen -f -X '!*.json' -- ${cur}) )
                return 0
            fi

            # Check if we're completing after 'import <file>'
            if [ "${COMP_WORDS[COMP_CWORD-2]}" = "import" ]; then
                # Suggest profile name for import
                return 0
            fi

            # Default: suggest profiles and options
            COMPREPLY=( $(compgen -W "${opts} ${profiles}" -- ${cur}) )
            return 0
            ;;
    esac
}

# Register the completion function
complete -F _claudectx claudectx
