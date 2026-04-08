package update

type Manifest struct {
	Version  string `json:"version"`
	LinuxURL string `json:"linux_url"`
	Notes    string `json:"notes"`
}

type CheckResult struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	DownloadURL     string `json:"downloadURL"`
	Notes           string `json:"notes"`
	Error           string `json:"error"`
}

type DownloadResult struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	Downloaded      bool   `json:"downloaded"`
	FilePath        string `json:"filePath"`
	FinalURL        string `json:"finalURL"`
	StatusCode      int    `json:"statusCode"`
	BytesWritten    int64  `json:"bytesWritten"`
	Notes           string `json:"notes"`
	Error           string `json:"error"`
}
