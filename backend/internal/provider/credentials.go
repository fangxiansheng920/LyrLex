package provider

// CredentialsProvider 提供外部服务所需的密钥（由设置中心维护，可动态变更）。
type CredentialsProvider interface {
	YoudaoKeys() (appKey, appSecret string)
}
