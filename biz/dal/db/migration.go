package db

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// MigrateCredentials 将现有认证数据迁移到凭证池
// 幂等操作：重复执行不会重复创建
func MigrateCredentials() {
	credDAO := NewCredentialDAO()

	// 1. 为每个 SSH 密钥创建对应的凭证
	migrateSSHKeysToCredentials(credDAO)

	// 2. 从 Repo 的认证配置中提取凭证
	migrateRepoAuthToCredentials(credDAO)

	log.Println("Credential migration completed.")
}

func migrateSSHKeysToCredentials(credDAO *CredentialDAO) {
	sshKeyDAO := NewSSHKeyDAO()
	keys, err := sshKeyDAO.FindAll()
	if err != nil {
		log.Printf("Warning: failed to load SSH keys for migration: %v", err)
		return
	}

	for _, key := range keys {
		credName := fmt.Sprintf("SSH: %s", key.Name)
		exists, _ := credDAO.ExistsByName(credName)
		if exists {
			continue
		}

		// 检查是否已有引用此 ssh_key_id 的凭证
		existing, _ := credDAO.FindBySSHKeyID(key.ID)
		if len(existing) > 0 {
			continue
		}

		cred := &po.Credential{
			Name:        credName,
			Type:        "ssh_key",
			Description: key.Description,
			SSHKeyID:    key.ID,
		}
		if err := credDAO.Create(cred); err != nil {
			log.Printf("Warning: failed to create credential for SSH key %s: %v", key.Name, err)
		}
	}
}

func migrateRepoAuthToCredentials(credDAO *CredentialDAO) {
	repoDAO := NewRepoDAO()
	repos, err := repoDAO.FindAll()
	if err != nil {
		log.Printf("Warning: failed to load repos for migration: %v", err)
		return
	}

	for i := range repos {
		repo := &repos[i]
		changed := false

		// 处理主认证
		if repo.AuthType != "" && repo.AuthType != "none" && repo.DefaultCredentialID == 0 {
			credID := findOrCreateCredentialFromAuth(credDAO, repo.AuthType, repo.AuthKey, repo.AuthSecret, "", 0)
			if credID > 0 {
				repo.DefaultCredentialID = credID
				changed = true
			}
		}

		// 处理远程认证
		if repo.RemoteAuthsJSON != "" && len(repo.RemoteCredentials) == 0 {
			var remoteAuths map[string]domain.AuthInfo
			if err := json.Unmarshal([]byte(repo.RemoteAuthsJSON), &remoteAuths); err == nil {
				remoteCreds := make(map[string]uint)
				for remoteName, authInfo := range remoteAuths {
					if authInfo.Type == "" || authInfo.Type == "none" {
						continue
					}
					credID := findOrCreateCredentialFromAuth(credDAO, authInfo.Type, authInfo.Key, authInfo.Secret, authInfo.Source, authInfo.SSHKeyID)
					if credID > 0 {
						remoteCreds[remoteName] = credID
					}
				}
				if len(remoteCreds) > 0 {
					repo.RemoteCredentials = remoteCreds
					changed = true
				}
			}
		}

		if changed {
			if err := repoDAO.Save(repo); err != nil {
				log.Printf("Warning: failed to update repo %s with credential refs: %v", repo.Key, err)
			}
		}
	}
}

func findOrCreateCredentialFromAuth(credDAO *CredentialDAO, authType, authKey, authSecret, source string, sshKeyID uint) uint {
	switch authType {
	case "ssh":
		if source == "database" && sshKeyID > 0 {
			// 查找已有的凭证
			existing, _ := credDAO.FindBySSHKeyID(sshKeyID)
			if len(existing) > 0 {
				return existing[0].ID
			}
			// 创建新凭证
			cred := &po.Credential{
				Name:     fmt.Sprintf("SSH Key #%d (migrated)", sshKeyID),
				Type:     "ssh_key",
				SSHKeyID: sshKeyID,
			}
			if err := credDAO.Create(cred); err == nil {
				return cred.ID
			}
		} else if authKey != "" {
			// 本地密钥 - 按路径去重
			name := fmt.Sprintf("SSH Local: %s", authKey)
			if existing, err := credDAO.FindByName(name); err == nil {
				return existing.ID
			}
			cred := &po.Credential{
				Name:       name,
				Type:       "ssh_key",
				SSHKeyPath: authKey,
				Secret:     authSecret,
			}
			if err := credDAO.Create(cred); err == nil {
				return cred.ID
			}
		}
	case "http":
		if authKey != "" {
			name := fmt.Sprintf("HTTP: %s (migrated)", authKey)
			if existing, err := credDAO.FindByName(name); err == nil {
				return existing.ID
			}
			cred := &po.Credential{
				Name:     name,
				Type:     "http_basic",
				Username: authKey,
				Secret:   authSecret,
			}
			if err := credDAO.Create(cred); err == nil {
				return cred.ID
			}
		}
	}
	return 0
}
