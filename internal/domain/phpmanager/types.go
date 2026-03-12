package phpmanager

type Version struct {
	Version   string `json:"version"`
	Installed bool   `json:"installed"`
	Active    bool   `json:"active"`
}

type IniSettings struct {
	UploadMaxFilesize string `json:"uploadMaxFilesize"`
	PostMaxSize       string `json:"postMaxSize"`
	MemoryLimit       string `json:"memoryLimit"`
	MaxExecutionTime  string `json:"maxExecutionTime"`
}

type Extension struct {
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Scope     string `json:"scope"`
	INIPath   string `json:"iniPath"`
	INIExists bool   `json:"iniExists"`
}

type Snapshot struct {
	GeneratedAt   string      `json:"generatedAt"`
	Supported     bool        `json:"supported"`
	Platform      string      `json:"platform"`
	ActiveVersion string      `json:"activeVersion"`
	Versions      []Version   `json:"versions"`
	Settings      IniSettings `json:"settings"`
	Extensions    []Extension `json:"extensions"`
	Warnings      []string    `json:"warnings"`
	LastError     string      `json:"lastError"`
}

type IniSettingsUpdateRequest struct {
	UploadMaxFilesize string `json:"uploadMaxFilesize"`
	PostMaxSize       string `json:"postMaxSize"`
	MemoryLimit       string `json:"memoryLimit"`
	MaxExecutionTime  string `json:"maxExecutionTime"`
}

type ExtensionToggleRequest struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type ActionResult struct {
	Success           bool     `json:"success"`
	Supported         bool     `json:"supported"`
	Version           string   `json:"version"`
	Command           string   `json:"command"`
	RequiresPrivilege bool     `json:"requiresPrivilege"`
	SuggestedCommands []string `json:"suggestedCommands"`
	Message           string   `json:"message"`
	Error             string   `json:"error"`
}
