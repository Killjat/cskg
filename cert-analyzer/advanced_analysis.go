package main

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// AdvancedAnalyzer 高级安全分析器
type AdvancedAnalyzer struct {
	config *AdvancedAnalysisConfig
}

// AdvancedAnalysisConfig 高级分析配置
type AdvancedAnalysisConfig struct {
	EnableThreatIntel     bool     `json:"enable_threat_intel"`
	EnablePhishingDetect  bool     `json:"enable_phishing_detect"`
	EnableDGADetection    bool     `json:"enable_dga_detection"`
	EnableTimelineAnalysis bool    `json:"enable_timeline_analysis"`
	LegitimateOrgs        []string `json:"legitimate_orgs"`
	SuspiciousKeywords    []string `json:"suspicious_keywords"`
	MinValidityDays       int      `json:"min_validity_days"`
	MaxValidityDays       int      `json:"max_validity_days"`
}

// AdvancedAnalysisResult 高级分析结果
type AdvancedAnalysisResult struct {
	ThreatIntelligence  *ThreatIntelResult  `json:"threat_intelligence,omitempty"`
	PhishingAnalysis    *PhishingAnalysis   `json:"phishing_analysis,omitempty"`
	DGAAnalysis         *DGAAnalysis        `json:"dga_analysis,omitempty"`
	TimelineAnalysis    *TimelineAnalysis   `json:"timeline_analysis,omitempty"`
	AnomalyDetection    *AnomalyDetection   `json:"anomaly_detection,omitempty"`
	RiskScore           int                 `json:"risk_score"`
	RiskFactors         []string            `json:"risk_factors"`
	Recommendations     []string            `json:"recommendations"`
}

// ThreatIntelResult 威胁情报分析结果
type ThreatIntelResult struct {
	IsMalicious         bool                `json:"is_malicious"`
	ThreatType          string              `json:"threat_type,omitempty"`
	APTAttribution      string              `json:"apt_attribution,omitempty"`
	IOCMatches          []IOCMatch          `json:"ioc_matches,omitempty"`
	SimilarThreats      []SimilarThreat     `json:"similar_threats,omitempty"`
	InfrastructureLinks []InfrastructureLink `json:"infrastructure_links,omitempty"`
}

// IOCMatch IOC匹配结果
type IOCMatch struct {
	Type        string  `json:"type"`        // fingerprint, serial, domain
	Value       string  `json:"value"`
	Source      string  `json:"source"`      // threat_feed, blacklist
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// SimilarThreat 相似威胁
type SimilarThreat struct {
	ThreatName   string  `json:"threat_name"`
	Similarity   float64 `json:"similarity"`
	SharedIOCs   []string `json:"shared_iocs"`
	TimeOverlap  bool    `json:"time_overlap"`
}

// InfrastructureLink 基础设施关联
type InfrastructureLink struct {
	LinkedDomain    string    `json:"linked_domain"`
	LinkType        string    `json:"link_type"`        // same_cert, same_ca, same_org
	Confidence      float64   `json:"confidence"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
}

// PhishingAnalysis 钓鱼分析
type PhishingAnalysis struct {
	IsPhishing          bool                `json:"is_phishing"`
	TargetBrand         string              `json:"target_brand,omitempty"`
	SimilarityScore     float64             `json:"similarity_score"`
	PhishingIndicators  []PhishingIndicator `json:"phishing_indicators,omitempty"`
	LegitimateAlternatives []string         `json:"legitimate_alternatives,omitempty"`
}

// PhishingIndicator 钓鱼指标
type PhishingIndicator struct {
	Type        string  `json:"type"`        // typosquatting, homograph, subdomain
	Description string  `json:"description"`
	Severity    string  `json:"severity"`    // low, medium, high, critical
	Confidence  float64 `json:"confidence"`
}

// DGAAnalysis DGA分析
type DGAAnalysis struct {
	IsDGA           bool        `json:"is_dga"`
	DGAScore        float64     `json:"dga_score"`
	EntropyScore    float64     `json:"entropy_score"`
	PatternAnalysis []DGAPattern `json:"pattern_analysis,omitempty"`
	PossibleFamily  string      `json:"possible_family,omitempty"`
}

// DGAPattern DGA模式
type DGAPattern struct {
	Pattern     string  `json:"pattern"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// TimelineAnalysis 时间线分析
type TimelineAnalysis struct {
	IssuancePattern    string              `json:"issuance_pattern"`    // normal, burst, suspicious
	ValidityPattern    string              `json:"validity_pattern"`    // normal, short, long
	TimeAnomalies      []TimeAnomaly       `json:"time_anomalies,omitempty"`
	RelatedCertificates []RelatedCertificate `json:"related_certificates,omitempty"`
}

// TimeAnomaly 时间异常
type TimeAnomaly struct {
	Type        string    `json:"type"`        // future_date, weekend_issuance, bulk_issuance
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"`
}

// RelatedCertificate 相关证书
type RelatedCertificate struct {
	Fingerprint   string    `json:"fingerprint"`
	CommonName    string    `json:"common_name"`
	Issuer        string    `json:"issuer"`
	IssuedAt      time.Time `json:"issued_at"`
	Relationship  string    `json:"relationship"` // same_batch, same_requester, suspicious_timing
}

// AnomalyDetection 异常检测
type AnomalyDetection struct {
	AnomalyScore    float64   `json:"anomaly_score"`
	Anomalies       []Anomaly `json:"anomalies,omitempty"`
	BaselineDeviation float64 `json:"baseline_deviation"`
}

// Anomaly 异常项
type Anomaly struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	Evidence    string  `json:"evidence"`
}

// NewAdvancedAnalyzer 创建高级分析器
func NewAdvancedAnalyzer(config *AdvancedAnalysisConfig) *AdvancedAnalyzer {
	if config == nil {
		config = &AdvancedAnalysisConfig{
			EnableThreatIntel:     true,
			EnablePhishingDetect:  true,
			EnableDGADetection:    true,
			EnableTimelineAnalysis: true,
			MinValidityDays:       30,
			MaxValidityDays:       825, // CA/Browser Forum规定的最大有效期
			SuspiciousKeywords: []string{
				"temp", "test", "dev", "staging", "backup", "admin", "root",
				"secure", "bank", "pay", "login", "account", "verify",
			},
		}
	}
	return &AdvancedAnalyzer{config: config}
}

// AnalyzeCertificate 执行高级证书分析
func (aa *AdvancedAnalyzer) AnalyzeCertificate(cert *x509.Certificate, relatedSites *RelatedSites) *AdvancedAnalysisResult {
	result := &AdvancedAnalysisResult{
		RiskFactors:     make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// 威胁情报分析
	if aa.config.EnableThreatIntel {
		result.ThreatIntelligence = aa.analyzeThreatIntelligence(cert, relatedSites)
	}

	// 钓鱼检测
	if aa.config.EnablePhishingDetect {
		result.PhishingAnalysis = aa.analyzePhishing(cert)
	}

	// DGA检测
	if aa.config.EnableDGADetection {
		result.DGAAnalysis = aa.analyzeDGA(cert)
	}

	// 时间线分析
	if aa.config.EnableTimelineAnalysis {
		result.TimelineAnalysis = aa.analyzeTimeline(cert)
	}

	// 异常检测
	result.AnomalyDetection = aa.detectAnomalies(cert)

	// 计算综合风险评分
	result.RiskScore = aa.calculateRiskScore(result)

	// 生成建议
	result.Recommendations = aa.generateRecommendations(result)

	return result
}

// analyzeThreatIntelligence 威胁情报分析
func (aa *AdvancedAnalyzer) analyzeThreatIntelligence(cert *x509.Certificate, relatedSites *RelatedSites) *ThreatIntelResult {
	result := &ThreatIntelResult{
		IOCMatches:          make([]IOCMatch, 0),
		SimilarThreats:      make([]SimilarThreat, 0),
		InfrastructureLinks: make([]InfrastructureLink, 0),
	}

	// 检查证书指纹是否在威胁情报中
	fingerprint := fmt.Sprintf("%X", cert.Raw)
	if aa.isKnownMaliciousFingerprint(fingerprint) {
		result.IsMalicious = true
		result.ThreatType = "known_malicious_certificate"
		result.IOCMatches = append(result.IOCMatches, IOCMatch{
			Type:        "fingerprint",
			Value:       fingerprint,
			Source:      "threat_intelligence",
			Confidence:  0.95,
			Description: "Certificate fingerprint matches known malicious IOC",
		})
	}

	// 检查域名是否在黑名单中
	for _, domain := range cert.DNSNames {
		if aa.isKnownMaliciousDomain(domain) {
			result.IsMalicious = true
			result.IOCMatches = append(result.IOCMatches, IOCMatch{
				Type:        "domain",
				Value:       domain,
				Source:      "domain_blacklist",
				Confidence:  0.9,
				Description: "Domain matches known malicious domain",
			})
		}
	}

	// 分析基础设施关联
	if relatedSites != nil {
		for _, site := range relatedSites.Sites {
			if aa.isKnownMaliciousDomain(site.Domain) {
				result.InfrastructureLinks = append(result.InfrastructureLinks, InfrastructureLink{
					LinkedDomain: site.Domain,
					LinkType:     "same_cert",
					Confidence:   0.8,
					LastSeen:     site.LastSeen,
				})
			}
		}
	}

	return result
}

// analyzePhishing 钓鱼分析
func (aa *AdvancedAnalyzer) analyzePhishing(cert *x509.Certificate) *PhishingAnalysis {
	result := &PhishingAnalysis{
		PhishingIndicators: make([]PhishingIndicator, 0),
	}

	commonName := cert.Subject.CommonName
	if commonName == "" && len(cert.DNSNames) > 0 {
		commonName = cert.DNSNames[0]
	}

	// 检查是否为知名品牌的拼写错误
	for _, brand := range aa.getKnownBrands() {
		similarity := aa.calculateStringSimilarity(commonName, brand)
		if similarity > 0.7 && similarity < 1.0 {
			result.IsPhishing = true
			result.TargetBrand = brand
			result.SimilarityScore = similarity
			
			result.PhishingIndicators = append(result.PhishingIndicators, PhishingIndicator{
				Type:        "typosquatting",
				Description: fmt.Sprintf("Domain '%s' is similar to legitimate brand '%s'", commonName, brand),
				Severity:    aa.getSeverityFromSimilarity(similarity),
				Confidence:  similarity,
			})
		}
	}

	// 检查同形异义字符攻击
	if aa.containsHomographs(commonName) {
		result.PhishingIndicators = append(result.PhishingIndicators, PhishingIndicator{
			Type:        "homograph",
			Description: "Domain contains homograph characters that may confuse users",
			Severity:    "high",
			Confidence:  0.8,
		})
	}

	// 检查可疑关键词
	for _, keyword := range aa.config.SuspiciousKeywords {
		if strings.Contains(strings.ToLower(commonName), keyword) {
			result.PhishingIndicators = append(result.PhishingIndicators, PhishingIndicator{
				Type:        "suspicious_keyword",
				Description: fmt.Sprintf("Domain contains suspicious keyword: %s", keyword),
				Severity:    "medium",
				Confidence:  0.6,
			})
		}
	}

	return result
}

// analyzeDGA DGA分析
func (aa *AdvancedAnalyzer) analyzeDGA(cert *x509.Certificate) *DGAAnalysis {
	result := &DGAAnalysis{
		PatternAnalysis: make([]DGAPattern, 0),
	}

	commonName := cert.Subject.CommonName
	if commonName == "" && len(cert.DNSNames) > 0 {
		commonName = cert.DNSNames[0]
	}

	// 计算域名熵值
	entropy := aa.calculateEntropy(commonName)
	result.EntropyScore = entropy

	// 高熵值可能表示DGA
	if entropy > 4.0 {
		result.IsDGA = true
		result.DGAScore = entropy / 5.0 // 归一化到0-1
		
		result.PatternAnalysis = append(result.PatternAnalysis, DGAPattern{
			Pattern:     "high_entropy",
			Confidence:  entropy / 5.0,
			Description: fmt.Sprintf("High entropy score: %.2f", entropy),
		})
	}

	// 检查字符模式
	if aa.hasRandomPattern(commonName) {
		result.IsDGA = true
		result.PatternAnalysis = append(result.PatternAnalysis, DGAPattern{
			Pattern:     "random_characters",
			Confidence:  0.7,
			Description: "Domain appears to contain random character sequences",
		})
	}

	// 检查长度异常
	if len(commonName) > 20 && entropy > 3.5 {
		result.PatternAnalysis = append(result.PatternAnalysis, DGAPattern{
			Pattern:     "unusual_length",
			Confidence:  0.6,
			Description: "Unusually long domain name with high entropy",
		})
	}

	return result
}

// analyzeTimeline 时间线分析
func (aa *AdvancedAnalyzer) analyzeTimeline(cert *x509.Certificate) *TimelineAnalysis {
	result := &TimelineAnalysis{
		TimeAnomalies: make([]TimeAnomaly, 0),
	}

	now := time.Now()
	
	// 检查证书有效期
	validityDays := int(cert.NotAfter.Sub(cert.NotBefore).Hours() / 24)
	
	if validityDays < aa.config.MinValidityDays {
		result.ValidityPattern = "short"
		result.TimeAnomalies = append(result.TimeAnomalies, TimeAnomaly{
			Type:        "short_validity",
			Description: fmt.Sprintf("Certificate validity period is unusually short: %d days", validityDays),
			Timestamp:   cert.NotBefore,
			Severity:    "medium",
		})
	} else if validityDays > aa.config.MaxValidityDays {
		result.ValidityPattern = "long"
		result.TimeAnomalies = append(result.TimeAnomalies, TimeAnomaly{
			Type:        "long_validity",
			Description: fmt.Sprintf("Certificate validity period exceeds recommended maximum: %d days", validityDays),
			Timestamp:   cert.NotBefore,
			Severity:    "low",
		})
	} else {
		result.ValidityPattern = "normal"
	}

	// 检查未来时间
	if cert.NotBefore.After(now) {
		result.TimeAnomalies = append(result.TimeAnomalies, TimeAnomaly{
			Type:        "future_date",
			Description: "Certificate NotBefore date is in the future",
			Timestamp:   cert.NotBefore,
			Severity:    "high",
		})
	}

	// 检查周末签发（可能表示自动化或可疑活动）
	if cert.NotBefore.Weekday() == time.Saturday || cert.NotBefore.Weekday() == time.Sunday {
		result.TimeAnomalies = append(result.TimeAnomalies, TimeAnomaly{
			Type:        "weekend_issuance",
			Description: "Certificate was issued during weekend",
			Timestamp:   cert.NotBefore,
			Severity:    "low",
		})
	}

	return result
}

// detectAnomalies 异常检测
func (aa *AdvancedAnalyzer) detectAnomalies(cert *x509.Certificate) *AnomalyDetection {
	result := &AnomalyDetection{
		Anomalies: make([]Anomaly, 0),
	}

	var anomalyScore float64

	// 检查公钥大小异常
	if pubKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
		keySize := pubKey.N.BitLen()
		if keySize < 2048 {
			anomalyScore += 0.3
			result.Anomalies = append(result.Anomalies, Anomaly{
				Type:        "weak_key_size",
				Description: fmt.Sprintf("RSA key size is below recommended minimum: %d bits", keySize),
				Score:       0.3,
				Evidence:    fmt.Sprintf("Key size: %d bits", keySize),
			})
		} else if keySize > 4096 {
			anomalyScore += 0.1
			result.Anomalies = append(result.Anomalies, Anomaly{
				Type:        "unusual_key_size",
				Description: fmt.Sprintf("RSA key size is unusually large: %d bits", keySize),
				Score:       0.1,
				Evidence:    fmt.Sprintf("Key size: %d bits", keySize),
			})
		}
	}

	// 检查SAN域名数量异常
	sanCount := len(cert.DNSNames)
	if sanCount > 100 {
		anomalyScore += 0.2
		result.Anomalies = append(result.Anomalies, Anomaly{
			Type:        "excessive_san_domains",
			Description: fmt.Sprintf("Certificate contains unusually many SAN domains: %d", sanCount),
			Score:       0.2,
			Evidence:    fmt.Sprintf("SAN count: %d", sanCount),
		})
	}

	// 检查组织信息异常
	if cert.Subject.Organization != nil && len(cert.Subject.Organization) > 0 {
		org := cert.Subject.Organization[0]
		if aa.containsSuspiciousOrgName(org) {
			anomalyScore += 0.25
			result.Anomalies = append(result.Anomalies, Anomaly{
				Type:        "suspicious_organization",
				Description: "Organization name contains suspicious patterns",
				Score:       0.25,
				Evidence:    fmt.Sprintf("Organization: %s", org),
			})
		}
	}

	result.AnomalyScore = anomalyScore
	return result
}

// calculateRiskScore 计算综合风险评分
func (aa *AdvancedAnalyzer) calculateRiskScore(result *AdvancedAnalysisResult) int {
	var score float64

	// 威胁情报权重最高
	if result.ThreatIntelligence != nil && result.ThreatIntelligence.IsMalicious {
		score += 50
	}

	// 钓鱼检测
	if result.PhishingAnalysis != nil && result.PhishingAnalysis.IsPhishing {
		score += 30
	}

	// DGA检测
	if result.DGAAnalysis != nil && result.DGAAnalysis.IsDGA {
		score += 25
	}

	// 异常检测
	if result.AnomalyDetection != nil {
		score += result.AnomalyDetection.AnomalyScore * 20
	}

	// 时间线异常
	if result.TimelineAnalysis != nil {
		for _, anomaly := range result.TimelineAnalysis.TimeAnomalies {
			switch anomaly.Severity {
			case "critical":
				score += 15
			case "high":
				score += 10
			case "medium":
				score += 5
			case "low":
				score += 2
			}
		}
	}

	// 确保分数在0-100范围内
	if score > 100 {
		score = 100
	}

	return int(score)
}

// generateRecommendations 生成安全建议
func (aa *AdvancedAnalyzer) generateRecommendations(result *AdvancedAnalysisResult) []string {
	recommendations := make([]string, 0)

	if result.ThreatIntelligence != nil && result.ThreatIntelligence.IsMalicious {
		recommendations = append(recommendations, "CRITICAL: Block this certificate immediately - matches known malicious IOCs")
		recommendations = append(recommendations, "Investigate all systems that may have interacted with this certificate")
	}

	if result.PhishingAnalysis != nil && result.PhishingAnalysis.IsPhishing {
		recommendations = append(recommendations, "WARNING: Potential phishing certificate detected")
		recommendations = append(recommendations, "Verify the legitimacy of the certificate issuer and domain owner")
		if result.PhishingAnalysis.TargetBrand != "" {
			recommendations = append(recommendations, fmt.Sprintf("Contact %s security team to report potential brand abuse", result.PhishingAnalysis.TargetBrand))
		}
	}

	if result.DGAAnalysis != nil && result.DGAAnalysis.IsDGA {
		recommendations = append(recommendations, "Suspicious domain generation pattern detected - investigate for malware activity")
		recommendations = append(recommendations, "Monitor network traffic to this domain for malicious activity")
	}

	if result.AnomalyDetection != nil && result.AnomalyDetection.AnomalyScore > 0.5 {
		recommendations = append(recommendations, "Multiple anomalies detected - conduct thorough security review")
	}

	if result.RiskScore > 70 {
		recommendations = append(recommendations, "HIGH RISK: Consider blocking or closely monitoring this certificate")
	} else if result.RiskScore > 40 {
		recommendations = append(recommendations, "MEDIUM RISK: Additional verification recommended")
	}

	return recommendations
}

// 辅助函数
func (aa *AdvancedAnalyzer) isKnownMaliciousFingerprint(fingerprint string) bool {
	// 这里应该查询威胁情报数据库
	// 为演示目的，返回false
	return false
}

func (aa *AdvancedAnalyzer) isKnownMaliciousDomain(domain string) bool {
	// 这里应该查询域名黑名单
	// 为演示目的，返回false
	return false
}

func (aa *AdvancedAnalyzer) getKnownBrands() []string {
	return []string{
		"google.com", "facebook.com", "amazon.com", "microsoft.com",
		"apple.com", "netflix.com", "paypal.com", "ebay.com",
		"twitter.com", "instagram.com", "linkedin.com", "github.com",
	}
}

func (aa *AdvancedAnalyzer) calculateStringSimilarity(s1, s2 string) float64 {
	// 简单的Levenshtein距离相似度计算
	if len(s1) == 0 {
		return float64(len(s2))
	}
	if len(s2) == 0 {
		return float64(len(s1))
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = minInt(
				matrix[i-1][j]+1,      // deletion
				minInt(matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost)) // substitution
		}
	}

	maxLen := maxInt(len(s1), len(s2))
	return 1.0 - float64(matrix[len(s1)][len(s2)])/float64(maxLen)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (aa *AdvancedAnalyzer) containsHomographs(domain string) bool {
	// 检查常见的同形异义字符
	homographs := map[rune][]rune{
		'a': {'а', 'ɑ'},  // Cyrillic а, Latin ɑ
		'o': {'о', 'ο'},  // Cyrillic о, Greek ο
		'p': {'р'},       // Cyrillic р
		'e': {'е'},       // Cyrillic е
	}

	for _, char := range domain {
		if alternatives, exists := homographs[char]; exists {
			for _, alt := range alternatives {
				if strings.ContainsRune(domain, alt) {
					return true
				}
			}
		}
	}
	return false
}

func (aa *AdvancedAnalyzer) calculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	freq := make(map[rune]int)
	for _, char := range s {
		freq[char]++
	}

	var entropy float64
	length := float64(len(s))
	for _, count := range freq {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}

	return entropy
}

func (aa *AdvancedAnalyzer) hasRandomPattern(domain string) bool {
	// 检查连续的随机字符模式
	randomPattern := regexp.MustCompile(`[a-z]{8,}`)
	matches := randomPattern.FindAllString(strings.ToLower(domain), -1)
	
	for _, match := range matches {
		if aa.calculateEntropy(match) > 3.5 {
			return true
		}
	}
	return false
}

func (aa *AdvancedAnalyzer) getSeverityFromSimilarity(similarity float64) string {
	if similarity > 0.9 {
		return "critical"
	} else if similarity > 0.8 {
		return "high"
	} else if similarity > 0.7 {
		return "medium"
	}
	return "low"
}

func (aa *AdvancedAnalyzer) containsSuspiciousOrgName(org string) bool {
	suspiciousPatterns := []string{
		"temp", "test", "fake", "phish", "scam", "fraud",
		"security", "verify", "account", "update", "confirm",
	}
	
	orgLower := strings.ToLower(org)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(orgLower, pattern) {
			return true
		}
	}
	return false
}