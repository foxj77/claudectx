# Fish completion for claudectx
# To install: copy to ~/.config/fish/completions/claudectx.fish

# Disable file completion by default
complete -c claudectx -f

# Helper function to get profiles
function __fish_claudectx_profiles
    claudectx 2>/dev/null | awk '{print $1}'
end

# Main commands
complete -c claudectx -s h -l help -d 'Show help'
complete -c claudectx -s v -l version -d 'Show version'
complete -c claudectx -s c -l current -d 'Show current profile'
complete -c claudectx -s n -d 'Create new profile' -x
complete -c claudectx -s d -d 'Delete profile' -a '(__fish_claudectx_profiles)'
complete -c claudectx -a '-' -d 'Switch to previous profile'

# Export command
complete -c claudectx -n "__fish_use_subcommand" -a 'export' -d 'Export profile to JSON'
complete -c claudectx -n "__fish_seen_subcommand_from export" -a '(__fish_claudectx_profiles)' -d 'Profile to export'
complete -c claudectx -n "__fish_seen_subcommand_from export" -F -d 'Output file'

# Import command
complete -c claudectx -n "__fish_use_subcommand" -a 'import' -d 'Import profile from JSON'
complete -c claudectx -n "__fish_seen_subcommand_from import" -F -d 'Input file'
complete -c claudectx -n "__fish_seen_subcommand_from import" -x -d 'New profile name'

# Health command
complete -c claudectx -n "__fish_use_subcommand" -a 'health' -d 'Check profile health'
complete -c claudectx -n "__fish_seen_subcommand_from health" -a '(__fish_claudectx_profiles)' -d 'Profile to check'

# Profile names for switching
complete -c claudectx -n "__fish_use_subcommand" -a '(__fish_claudectx_profiles)' -d 'Switch to profile'
