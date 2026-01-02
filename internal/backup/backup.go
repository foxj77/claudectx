package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/mcpconfig"
	"github.com/johnfox/claudectx/internal/paths"
)

// Backup represents a single backup snapshot
type Backup struct {
	ID        string
	CreatedAt time.Time
}

// Manager handles backup operations
type Manager struct {
	backupDir string
}

// NewManager creates a new backup manager
func NewManager() (*Manager, error) {
	claudeDir, err := paths.ClaudeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get claude directory: %w", err)
	}

	backupDir := filepath.Join(claudeDir, "backups")

	// Ensure backup directory exists
	err = os.MkdirAll(backupDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &Manager{
		backupDir: backupDir,
	}, nil
}

// Create creates a new backup of current configuration
func (m *Manager) Create() (string, error) {
	// Generate backup ID (timestamp-based)
	backupID := fmt.Sprintf("backup-%d", time.Now().UnixNano())
	backupPath := filepath.Join(m.backupDir, backupID)

	// Create backup directory
	err := os.MkdirAll(backupPath, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup settings.json
	settingsPath, err := paths.SettingsFile()
	if err != nil {
		return "", err
	}

	if config.FileExists(settingsPath) {
		destSettings := filepath.Join(backupPath, "settings.json")
		err = config.CopyFile(settingsPath, destSettings)
		if err != nil {
			os.RemoveAll(backupPath) // Clean up on failure
			return "", fmt.Errorf("failed to backup settings.json: %w", err)
		}
	}

	// Backup CLAUDE.md if it exists
	claudeMDPath, err := paths.ClaudeMDFile()
	if err != nil {
		return "", err
	}

	if config.FileExists(claudeMDPath) {
		destClaudeMD := filepath.Join(backupPath, "CLAUDE.md")
		err = config.CopyFile(claudeMDPath, destClaudeMD)
		if err != nil {
			os.RemoveAll(backupPath) // Clean up on failure
			return "", fmt.Errorf("failed to backup CLAUDE.md: %w", err)
		}
	}

	// Backup MCP servers from ~/.claude.json
	claudeJSONPath, err := paths.ClaudeJSONFile()
	if err != nil {
		return "", err
	}

	if config.FileExists(claudeJSONPath) {
		mcpServers, err := mcpconfig.LoadMCPServers(claudeJSONPath)
		if err == nil && len(mcpServers) > 0 {
			destMCP := filepath.Join(backupPath, "mcp.json")
			err = mcpconfig.SaveToFile(destMCP, mcpServers)
			if err != nil {
				os.RemoveAll(backupPath) // Clean up on failure
				return "", fmt.Errorf("failed to backup MCP servers: %w", err)
			}
		}
	}

	return backupID, nil
}

// Restore restores configuration from a backup
func (m *Manager) Restore(backupID string) error {
	backupPath := filepath.Join(m.backupDir, backupID)

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup %q does not exist", backupID)
	}

	// Restore settings.json
	backupSettings := filepath.Join(backupPath, "settings.json")
	if config.FileExists(backupSettings) {
		settingsPath, err := paths.SettingsFile()
		if err != nil {
			return err
		}

		err = config.CopyFile(backupSettings, settingsPath)
		if err != nil {
			return fmt.Errorf("failed to restore settings.json: %w", err)
		}
	}

	// Restore CLAUDE.md if it was backed up
	backupClaudeMD := filepath.Join(backupPath, "CLAUDE.md")
	claudeMDPath, err := paths.ClaudeMDFile()
	if err != nil {
		return err
	}

	if config.FileExists(backupClaudeMD) {
		err = config.CopyFile(backupClaudeMD, claudeMDPath)
		if err != nil {
			return fmt.Errorf("failed to restore CLAUDE.md: %w", err)
		}
	} else {
		// If CLAUDE.md wasn't in backup, remove it if it exists
		if config.FileExists(claudeMDPath) {
			os.Remove(claudeMDPath)
		}
	}

	// Restore MCP servers if they were backed up
	backupMCP := filepath.Join(backupPath, "mcp.json")
	claudeJSONPath, err := paths.ClaudeJSONFile()
	if err != nil {
		return err
	}

	if config.FileExists(backupMCP) {
		mcpServers, err := mcpconfig.LoadFromFile(backupMCP)
		if err != nil {
			return fmt.Errorf("failed to load backup MCP servers: %w", err)
		}
		err = mcpconfig.SaveMCPServers(claudeJSONPath, mcpServers)
		if err != nil {
			return fmt.Errorf("failed to restore MCP servers: %w", err)
		}
	} else {
		// If MCP servers weren't in backup, remove them from ~/.claude.json
		if config.FileExists(claudeJSONPath) {
			err = mcpconfig.SaveMCPServers(claudeJSONPath, make(mcpconfig.MCPServers))
			if err != nil {
				return fmt.Errorf("failed to clear MCP servers: %w", err)
			}
		}
	}

	return nil
}

// RestoreLatest restores the most recent backup
func (m *Manager) RestoreLatest() error {
	latest := m.GetLatest()
	if latest == "" {
		return fmt.Errorf("no backups available to restore")
	}

	return m.Restore(latest)
}

// List returns all backups, sorted by creation time (newest first)
func (m *Manager) List() ([]Backup, error) {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Backup{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []Backup
	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			backups = append(backups, Backup{
				ID:        entry.Name(),
				CreatedAt: info.ModTime(),
			})
		}
	}

	// Sort by creation time, newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// GetLatest returns the ID of the most recent backup
func (m *Manager) GetLatest() string {
	backups, err := m.List()
	if err != nil || len(backups) == 0 {
		return ""
	}

	return backups[0].ID
}

// Delete removes a backup
func (m *Manager) Delete(backupID string) error {
	backupPath := filepath.Join(m.backupDir, backupID)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup %q does not exist", backupID)
	}

	err := os.RemoveAll(backupPath)
	if err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	return nil
}

// Prune removes old backups, keeping only the specified number
func (m *Manager) Prune(keep int) error {
	backups, err := m.List()
	if err != nil {
		return err
	}

	if len(backups) <= keep {
		return nil // Nothing to prune
	}

	// Delete oldest backups
	toDelete := backups[keep:]
	for _, backup := range toDelete {
		err := m.Delete(backup.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
