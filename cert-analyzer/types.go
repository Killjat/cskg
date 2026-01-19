package main

import (
	"crypto/x509"
	"time"
)

// CertificateInfo 证书基本信息
type CertificateInfo struct {
	Subject           SubjectInfo     `json:"subject"`
	Issuer            IssuerInfo      `json:"issuer"`
	Validity          ValidityInfo    `json:"validity"`
	SANDomains        []string        `json:"san_domains"`
	SignatureAlgorithm string         `json:"signature_algorithm"`
	PublicKey         PublicKeyInfo   `json:"public_key"`
	SerialNumber      string          `json:"serial_number"`
	FingerprintSHA1   string          `json:"fingerprint_sha1"`
	FingerprintSHA256 string          `json:"fingerprint_sha256"`
	Version           int             `json:"version"`
	Extensions        []ExtensionInfo `json:"extensions,omitempty"`
	RelatedSites      *RelatedSites   `json:"related_sites,omitempty"`
}

// SubjectInfo 证书主题信息
type SubjectInfo struct {
	CommonName         string `json:"common_name"`
	Organization       string `json:"organization,omitempty"`
	OrganizationalUnit string `json:"organizational_unit,omitempty"`
	Country            string `json:"country,omitempty"`
	Province           string `json:"province,omitempty"`
	Locality           string `json:"locality,omitempty"`
}

// IssuerInfo 证书颁发者信息
type IssuerInfo struct {
	CommonName         string `json:"common_name"`
	Organization       string `json:"organization,omitempty"`
	OrganizationalUnit string `json:"organizational_unit,omitempty"`
	Country            string `json:"country,omitempty"`
}

// ValidityInfo 证书有效期信息
type ValidityInfo struct {
	NotBefore      time.Time `json:"not_before"`
	NotAfter       time.Time `json:"not_after"`
	DaysRemaining  int       `json:"days_remaining"`
	IsExpired      bool      `json:"is_expired"`
	ExpiresSoon    bool      `json:"expires_soon"` // 30天内过期
}

// PublicKeyInfo 公钥信息
type PublicKeyInfo struct {
	Algorithm string `json:"algorithm"`
	Size      int    `json:"size"`
}

// ExtensionInfo 证书扩展信息
type ExtensionInfo struct {
	OID      string `json:"oid"`
	Critical bool   `json:"critical"`
	Value    string `json:"value,omitempty"`
}

// SecurityAnalysis 安全分析结果
type SecurityAnalysis struct {
	IsExpired            bool     `json:"is_expired"`
	ExpiresSoon          bool     `json:"expires_soon"`
	IsSelfSigned         bool     `json:"is_self_signed"`
	WeakSignature        bool     `json:"weak_signature"`
	CertificateChainValid bool    `json:"certificate_chain_valid"`
	SecurityScore        int      `json:"security_score"`
	Warnings             []string `json:"warnings,omitempty"`
	Recommendations      []string `json:"recommendations,omitempty"`
}

// CertificateResult 完整的证书分析结果
type CertificateResult struct {
	URL               string                   `json:"url"`
	Timestamp         time.Time                `json:"timestamp"`
	Status            string                   `json:"status"`
	Error             string                   `json:"error,omitempty"`
	Certificate       *CertificateInfo         `json:"certificate,omitempty"`
	CertificateChain  []CertificateInfo        `json:"certificate_chain,omitempty"`
	SecurityAnalysis  *SecurityAnalysis        `json:"security_analysis,omitempty"`
	AdvancedAnalysis  *AdvancedAnalysisResult  `json:"advanced_analysis,omitempty"`
	ConnectionInfo    *ConnectionInfo          `json:"connection_info,omitempty"`
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	TLSVersion    string   `json:"tls_version"`
	CipherSuite   string   `json:"cipher_suite"`
	ServerName    string   `json:"server_name"`
	PeerCerts     int      `json:"peer_certificates_count"`
	Protocols     []string `json:"supported_protocols,omitempty"`
	ConnectTime   int64    `json:"connect_time_ms"`
}

// BatchResult 批量检测结果
type BatchResult struct {
	TotalURLs    int                 `json:"total_urls"`
	SuccessCount int                 `json:"success_count"`
	FailureCount int                 `json:"failure_count"`
	Results      []CertificateResult `json:"results"`
	Summary      *BatchSummary       `json:"summary,omitempty"`
}

// BatchSummary 批量检测摘要
type BatchSummary struct {
	ExpiredCerts     int      `json:"expired_certificates"`
	ExpiringSoon     int      `json:"expiring_soon"`
	SelfSignedCerts  int      `json:"self_signed_certificates"`
	WeakSignatures   int      `json:"weak_signatures"`
	CommonIssuers    []string `json:"common_issuers"`
	AverageScore     float64  `json:"average_security_score"`
}

// Config 配置信息
type Config struct {
	Timeout         time.Duration            `json:"timeout"`
	SkipVerify      bool                     `json:"skip_verify"`
	FollowRedirects bool                     `json:"follow_redirects"`
	MaxRedirects    int                      `json:"max_redirects"`
	UserAgent       string                   `json:"user_agent"`
	Verbose         bool                     `json:"verbose"`
	SearchConfig    *CertificateSearchConfig `json:"search_config,omitempty"`
	AdvancedConfig  *AdvancedAnalysisConfig  `json:"advanced_config,omitempty"`
}

// CertificateChain 证书链
type CertificateChain struct {
	Certificates []*x509.Certificate
	Verified     bool
	Error        error
}

// RelatedSites 使用相同证书的相关网站
type RelatedSites struct {
	SearchMethod    string          `json:"search_method"`
	TotalFound      int             `json:"total_found"`
	SearchTime      int64           `json:"search_time_ms"`
	Sites           []RelatedSite   `json:"sites"`
	SearchError     string          `json:"search_error,omitempty"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// RelatedSite 相关网站信息
type RelatedSite struct {
	URL             string    `json:"url"`
	Domain          string    `json:"domain"`
	Port            int       `json:"port"`
	Title           string    `json:"title,omitempty"`
	Server          string    `json:"server,omitempty"`
	Country         string    `json:"country,omitempty"`
	Organization    string    `json:"organization,omitempty"`
	LastSeen        time.Time `json:"last_seen,omitempty"`
	Confidence      float64   `json:"confidence"`
	Source          string    `json:"source"`
}

// CertificateSearchConfig 证书搜索配置
type CertificateSearchConfig struct {
	EnableSearch     bool          `json:"enable_search"`
	SearchMethods    []string      `json:"search_methods"`
	MaxResults       int           `json:"max_results"`
	Timeout          time.Duration `json:"timeout"`
	FOFAConfig       *FOFAConfig   `json:"fofa_config,omitempty"`
	ShodanConfig     *ShodanConfig `json:"shodan_config,omitempty"`
	CensysConfig     *CensysConfig `json:"censys_config,omitempty"`
}

// FOFAConfig FOFA搜索配置
type FOFAConfig struct {
	Email    string `json:"email"`
	Key      string `json:"key"`
	Enabled  bool   `json:"enabled"`
}

// ShodanConfig Shodan搜索配置
type ShodanConfig struct {
	APIKey   string `json:"api_key"`
	Enabled  bool   `json:"enabled"`
}

// CensysConfig Censys搜索配置
type CensysConfig struct {
	AppID    string `json:"app_id"`
	Secret   string `json:"secret"`
	Enabled  bool   `json:"enabled"`
}