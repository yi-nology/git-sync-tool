package domain

type AuthInfo struct {
	Type     string `json:"type"`       // ssh, http, none
	Key      string `json:"key"`        // SSH Key Path or Username
	Secret   string `json:"secret"`     // Passphrase or Password (Encrypted in DB)
	Source   string `json:"source"`     // "local"(file path) or "database"(db key)
	SSHKeyID uint   `json:"ssh_key_id"` // Database SSH Key ID (when Source="database")
}
