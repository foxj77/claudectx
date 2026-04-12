package mcpconfig

import (
	"encoding/json"
	"fmt"
	"os"
)

// MCPServer represents a single MCP server configuration
type MCPServer struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// MCPServers is a map of server name to server configuration
type MCPServers map[string]MCPServer

// claudeJSON represents the structure of ~/.claude.json
// We only define the fields we care about; json.RawMessage preserves the rest
type claudeJSON struct {
	MCPServers MCPServers `json:"mcpServers,omitempty"`
}

// LoadMCPServers reads the MCP servers from ~/.claude.json
func LoadMCPServers(path string) (MCPServers, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty servers
			return make(MCPServers), nil
		}
		return nil, fmt.Errorf("failed to read claude.json: %w", err)
	}

	var config claudeJSON
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse claude.json: %w", err)
	}

	if config.MCPServers == nil {
		return make(MCPServers), nil
	}

	return config.MCPServers, nil
}

// SaveMCPServers updates only the mcpServers field in ~/.claude.json
// preserving all other fields in the file
func SaveMCPServers(path string, servers MCPServers) error {
	// Read existing file content
	var existingData map[string]json.RawMessage
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, create new one with just mcpServers
			existingData = make(map[string]json.RawMessage)
		} else {
			return fmt.Errorf("failed to read claude.json: %w", err)
		}
	} else {
		err = json.Unmarshal(data, &existingData)
		if err != nil {
			return fmt.Errorf("failed to parse claude.json: %w", err)
		}
	}

	// Update or remove mcpServers field
	if len(servers) == 0 {
		// Remove the field if empty
		delete(existingData, "mcpServers")
	} else {
		// Marshal the new servers
		serversJSON, err := json.Marshal(servers)
		if err != nil {
			return fmt.Errorf("failed to marshal MCP servers: %w", err)
		}
		existingData["mcpServers"] = serversJSON
	}

	// Marshal the complete file
	output, err := json.MarshalIndent(existingData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal claude.json: %w", err)
	}

	// Write back to file
	err = os.WriteFile(path, output, 0644)
	if err != nil {
		return fmt.Errorf("failed to write claude.json: %w", err)
	}

	return nil
}

// SaveToFile saves MCP servers to a standalone JSON file (for profile storage)
func SaveToFile(path string, servers MCPServers) error {
	data, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal MCP servers: %w", err)
	}

	data = append(data, '\n')

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write MCP servers file: %w", err)
	}

	return nil
}

// LoadFromFile loads MCP servers from a standalone JSON file (for profile storage)
func LoadFromFile(path string) (MCPServers, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(MCPServers), nil
		}
		return nil, fmt.Errorf("failed to read MCP servers file: %w", err)
	}

	var servers MCPServers
	err = json.Unmarshal(data, &servers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MCP servers file: %w", err)
	}

	if servers == nil {
		return make(MCPServers), nil
	}

	return servers, nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
