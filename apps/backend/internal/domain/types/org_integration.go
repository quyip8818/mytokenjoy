package types

type OrgIntegration struct {
	Platform                  *Platform
	Connected                 bool
	LastImport                *string
	LastImportOK              *int
	LastImportFail            *int
	Enabled                   bool
	StartTime                 string
	FrequencyHours            int
	DeleteMemberThreshold     int
	DeleteDepartmentThreshold int
	NotifyPhone               bool
	NotifyEmail               bool
	NotifyIm                  bool
	EncryptedCredential       []byte
}

func (i OrgIntegration) ToDataSourceStatus() DataSourceStatus {
	status := DataSourceStatus{Connected: i.Connected, Platform: i.Platform}
	if i.LastImport != nil {
		s := *i.LastImport
		status.LastImport = &s
	}
	if i.LastImportOK != nil || i.LastImportFail != nil {
		result := ImportResult{}
		if i.LastImportOK != nil {
			result.SuccessMembers = *i.LastImportOK
		}
		if i.LastImportFail != nil {
			result.Failures = make([]ImportFailure, *i.LastImportFail)
		}
		status.LastImportResult = &result
	}
	return status
}

func (i *OrgIntegration) ApplyDataSourceStatus(status DataSourceStatus) {
	i.Connected = status.Connected
	i.Platform = status.Platform
	i.LastImport = status.LastImport
	i.LastImportOK = nil
	i.LastImportFail = nil
	if status.LastImportResult != nil {
		ok := status.LastImportResult.SuccessMembers
		i.LastImportOK = &ok
		fail := len(status.LastImportResult.Failures)
		i.LastImportFail = &fail
	}
}

func (i OrgIntegration) ToSyncConfig() SyncConfig {
	return SyncConfig{
		Enabled:                   i.Enabled,
		StartTime:                 i.StartTime,
		FrequencyHours:            i.FrequencyHours,
		DeleteMemberThreshold:     i.DeleteMemberThreshold,
		DeleteDepartmentThreshold: i.DeleteDepartmentThreshold,
		NotifyPhone:               i.NotifyPhone,
		NotifyEmail:               i.NotifyEmail,
		NotifyIm:                  i.NotifyIm,
	}
}

func (i *OrgIntegration) ApplySyncConfig(cfg SyncConfig) {
	i.Enabled = cfg.Enabled
	i.StartTime = cfg.StartTime
	i.FrequencyHours = cfg.FrequencyHours
	i.DeleteMemberThreshold = cfg.DeleteMemberThreshold
	i.DeleteDepartmentThreshold = cfg.DeleteDepartmentThreshold
	i.NotifyPhone = cfg.NotifyPhone
	i.NotifyEmail = cfg.NotifyEmail
	i.NotifyIm = cfg.NotifyIm
}

func (i OrgIntegration) HasCredential() bool {
	return i.Platform != nil && len(i.EncryptedCredential) > 0
}

func (i OrgIntegration) ToStoredCredential() *StoredCredential {
	if !i.HasCredential() {
		return nil
	}
	encrypted := make([]byte, len(i.EncryptedCredential))
	copy(encrypted, i.EncryptedCredential)
	return &StoredCredential{Platform: *i.Platform, Encrypted: encrypted}
}

func OrgIntegrationFromStatusAndConfig(status DataSourceStatus, cfg SyncConfig) OrgIntegration {
	integration := OrgIntegration{}
	integration.ApplyDataSourceStatus(status)
	integration.ApplySyncConfig(cfg)
	return integration
}
