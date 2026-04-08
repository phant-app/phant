package license

type KeyResult struct {
	LicenseKey string `json:"licenseKey"`
	Error      string `json:"error"`
}

type SaveResult struct {
	Success    bool   `json:"success"`
	LicenseKey string `json:"licenseKey"`
	Message    string `json:"message"`
	Error      string `json:"error"`
}
