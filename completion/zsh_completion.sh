#compdef claudectx

# Zsh completion for claudectx
# To install: copy to a directory in your $fpath (e.g., /usr/local/share/zsh/site-functions/_claudectx)

_claudectx_profiles() {
    local profiles
    profiles=(${(f)"$(claudectx 2>/dev/null | awk '{print $1}')"})
    _describe 'profiles' profiles
}

_claudectx() {
    local context state state_descr line
    typeset -A opt_args

    _arguments -C \
        '1: :->command' \
        '*:: :->args'

    case $state in
        command)
            local -a commands profiles
            commands=(
                '-h:Show help'
                '--help:Show help'
                '-v:Show version'
                '--version:Show version'
                '-c:Show current profile'
                '--current:Show current profile'
                '-n:Create new profile'
                '-d:Delete profile'
                '-:Switch to previous profile'
                'export:Export profile to JSON'
                'import:Import profile from JSON'
                'health:Check profile health'
            )

            # Get profiles
            profiles=(${(f)"$(claudectx 2>/dev/null | awk '{print $1}')"})

            _describe 'command' commands
            _describe 'profile' profiles
            ;;
        args)
            case $line[1] in
                -n)
                    _message 'profile name'
                    ;;
                -d)
                    _claudectx_profiles
                    ;;
                export)
                    if (( CURRENT == 2 )); then
                        _claudectx_profiles
                    else
                        _files -g '*.json'
                    fi
                    ;;
                import)
                    if (( CURRENT == 2 )); then
                        _files -g '*.json'
                    else
                        _message 'new profile name (optional)'
                    fi
                    ;;
                health)
                    _claudectx_profiles
                    ;;
            esac
            ;;
    esac
}

_claudectx "$@"
