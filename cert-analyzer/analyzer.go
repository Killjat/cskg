package main

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"hash"
	"net"
	"net/url"
	"strings"
	"time"
)

// CertificateAnalyzer 证书分析器
type CertificateAnalyzer struct {
	config   *Config
	searcher *CertificateSearcher
}

// NewCertificateAnalyzer 创建新的证书分析器
func NewCertificateAnalyzer(config *Config) *CertificateAnalyzer {
	analyzer := &CertificateAnalyzer{
		config: config,
	}
	
	// 如果配置了搜索功能，创建搜索器
	if config.SearchConfig != nil && config.SearchConfig.EnableSearch {
		analyzer.searcher = NewCertificateSearcher(config.SearchConfig)
	}
	
	return analyzer
}

// AnalyzeURL 分析指定URL的证书
func (ca *CertificateAnalyzer) AnalyzeURL(targetURL string) *CertificateResult {
	result := &CertificateResult{
		URL:       targetURL,
		Timestamp: time.Now(),
		Status:    "success",
	}

	startTime := time.Now()

	// 解析URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		result.Status = "error"
		result.Error = fmt.Sprintf("Invalid URL: %v", err)
		return result
	}

	// 确保是HTTPS
	if parsedURL.Scheme != "https" {
		result.Status = "error"
		result.Error = "URL must use HTTPS scheme"
		return result
	}

	// 获取主机和端口
	host := parsedURL.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}

	// 建立TLS连接
	tlsConfig := &tls.Config{
		InsecureSkipVerify: ca.config.SkipVerify,
		ServerName:         parsedURL.Hostname(),
	}

	conn, err := tls.DialWithDialer(&net.Dialer{
		Timeout: ca.config.Timeout,
	}, "tcp", host, tlsConfig)

	if err != nil {
		result.Status = "error"
		result.Error = fmt.Sprintf("TLS connection failed: %v", err)
		return result
	}
	defer conn.Close()

	connectTime := time.Since(startTime).Milliseconds()

	// 获取连接状态
	state := conn.ConnectionState()

	// 分析证书
	if len(state.PeerCertificates) == 0 {
		result.Status = "error"
		result.Error = "No certificates found"
		return result
	}

	// 主证书
	cert := state.PeerCertificates[0]
	result.Certificate = ca.analyzeCertificate(cert)

	// 证书链
	result.CertificateChain = make([]CertificateInfo, len(state.PeerCertificates))
	for i, c := range state.PeerCertificates {
		result.CertificateChain[i] = *ca.analyzeCertificate(c)
	}

	// 连接信息
	result.ConnectionInfo = &ConnectionInfo{
		TLSVersion:    ca.getTLSVersion(state.Version),
		CipherSuite:   tls.CipherSuiteName(state.CipherSuite),
		ServerName:    state.ServerName,
		PeerCerts:     len(state.PeerCertificates),
		ConnectTime:   connectTime,
	}

	// 设置协议信息
	if state.NegotiatedProtocol != "" {
		result.ConnectionInfo.Protocols = []string{state.NegotiatedProtocol}
	}

	// 安全分析
	result.SecurityAnalysis = ca.performSecurityAnalysis(cert, state.PeerCertificates)

	return result
}

// analyzeCertificate 分析单个证书
func (ca *CertificateAnalyzer) analyzeCertificate(cert *x509.Certificate) *CertificateInfo {
	info := &CertificateInfo{
		Subject:            ca.extractSubjectInfo(cert.Subject),
		Issuer:             ca.extractIssuerInfo(cert.Issuer),
		Validity:           ca.extractValidityInfo(cert),
		SANDomains:         cert.DNSNames,
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKey:          ca.extractPublicKeyInfo(cert),
		SerialNumber:       fmt.Sprintf("%X", cert.SerialNumber),
		FingerprintSHA1:    ca.calculateFingerprint(cert.Raw, sha1.New()),
		FingerprintSHA256:  ca.calculateFingerprint(cert.Raw, sha256.New()),
		Version:            cert.Version,
	}

	// 提取扩展信息
	for _, ext := range cert.Extensions {
		extInfo := ExtensionInfo{
			OID:      ext.Id.String(),
			Critical: ext.Critical,
		}
		info.Extensions = append(info.Extensions, extInfo)
	}

	// 搜索相关网站
	if ca.searcher != nil {
		info.RelatedSites = ca.searcher.SearchRelatedSites(cert)
	}

	return info
}

// extractSubjectInfo 提取主题信息
func (ca *CertificateAnalyzer) extractSubjectInfo(subject pkix.Name) SubjectInfo {
	return SubjectInfo{
		CommonName:         subject.CommonName,
		Organization:       strings.Join(subject.Organization, ", "),
		OrganizationalUnit: strings.Join(subject.OrganizationalUnit, ", "),
		Country:            strings.Join(subject.Country, ", "),
		Province:           strings.Join(subject.Province, ", "),
		Locality:           strings.Join(subject.Locality, ", "),
	}
}

// extractIssuerInfo 提取颁发者信息
func (ca *CertificateAnalyzer) extractIssuerInfo(issuer pkix.Name) IssuerInfo {
	return IssuerInfo{
		CommonName:         issuer.CommonName,
		Organization:       strings.Join(issuer.Organization, ", "),
		OrganizationalUnit: strings.Join(issuer.OrganizationalUnit, ", "),
		Country:            strings.Join(issuer.Country, ", "),
	}
}

// extractValidityInfo 提取有效期信息
func (ca *CertificateAnalyzer) extractValidityInfo(cert *x509.Certificate) ValidityInfo {
	now := time.Now()
	daysRemaining := int(cert.NotAfter.Sub(now).Hours() / 24)
	
	return ValidityInfo{
		NotBefore:     cert.NotBefore,
		NotAfter:      cert.NotAfter,
		DaysRemaining: daysRemaining,
		IsExpired:     now.After(cert.NotAfter),
		ExpiresSoon:   daysRemaining <= 30 && daysRemaining > 0,
	}
}

// extractPublicKeyInfo 提取公钥信息
func (ca *CertificateAnalyzer) extractPublicKeyInfo(cert *x509.Certificate) PublicKeyInfo {
	info := PublicKeyInfo{}
	
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		info.Algorithm = "RSA"
		info.Size = pub.N.BitLen()
	case *ecdsa.PublicKey:
		info.Algorithm = "ECDSA"
		info.Size = pub.Curve.Params().BitSize
	case *dsa.PublicKey:
		info.Algorithm = "DSA"
		info.Size = pub.P.BitLen()
	default:
		info.Algorithm = "Unknown"
		info.Size = 0
	}
	
	return info
}

// calculateFingerprint 计算证书指纹
func (ca *CertificateAnalyzer) calculateFingerprint(certRaw []byte, hasher hash.Hash) string {
	hasher.Write(certRaw)
	fingerprint := hasher.Sum(nil)
	
	var result []string
	for _, b := range fingerprint {
		result = append(result, fmt.Sprintf("%02X", b))
	}
	
	return strings.Join(result, ":")
}

// getTLSVersion 获取TLS版本字符串
func (ca *CertificateAnalyzer) getTLSVersion(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04X)", version)
	}
}

// performSecurityAnalysis 执行安全分析
func (ca *CertificateAnalyzer) performSecurityAnalysis(cert *x509.Certificate, chain []*x509.Certificate) *SecurityAnalysis {
	analysis := &SecurityAnalysis{
		CertificateChainValid: true,
		SecurityScore:         100,
	}

	now := time.Now()
	
	// 检查过期状态
	analysis.IsExpired = now.After(cert.NotAfter)
	analysis.ExpiresSoon = cert.NotAfter.Sub(now).Hours() <= 30*24 && !analysis.IsExpired

	// 检查自签名
	analysis.IsSelfSigned = cert.Subject.String() == cert.Issuer.String()

	// 检查弱签名算法
	weakAlgorithms := []string{"MD5", "SHA1"}
	sigAlg := cert.SignatureAlgorithm.String()
	for _, weak := range weakAlgorithms {
		if strings.Contains(strings.ToUpper(sigAlg), weak) {
			analysis.WeakSignature = true
			break
		}
	}

	// 验证证书链
	roots := x509.NewCertPool()
	intermediates := x509.NewCertPool()
	
	for i, c := range chain {
		if i == 0 {
			continue // 跳过叶子证书
		}
		if i == len(chain)-1 {
			roots.AddCert(c) // 根证书
		} else {
			intermediates.AddCert(c) // 中间证书
		}
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
	}

	_, err := cert.Verify(opts)
	if err != nil {
		analysis.CertificateChainValid = false
	}

	// 计算安全评分和建议
	ca.calculateSecurityScore(analysis, cert)
	ca.generateRecommendations(analysis, cert)

	return analysis
}

// calculateSecurityScore 计算安全评分
func (ca *CertificateAnalyzer) calculateSecurityScore(analysis *SecurityAnalysis, cert *x509.Certificate) {
	score := 100

	if analysis.IsExpired {
		score -= 50
		analysis.Warnings = append(analysis.Warnings, "Certificate has expired")
	}

	if analysis.ExpiresSoon {
		score -= 20
		analysis.Warnings = append(analysis.Warnings, "Certificate expires within 30 days")
	}

	if analysis.IsSelfSigned {
		score -= 30
		analysis.Warnings = append(analysis.Warnings, "Certificate is self-signed")
	}

	if analysis.WeakSignature {
		score -= 25
		analysis.Warnings = append(analysis.Warnings, "Certificate uses weak signature algorithm")
	}

	if !analysis.CertificateChainValid {
		score -= 20
		analysis.Warnings = append(analysis.Warnings, "Certificate chain validation failed")
	}

	// 检查公钥强度
	if pubKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
		if pubKey.N.BitLen() < 2048 {
			score -= 15
			analysis.Warnings = append(analysis.Warnings, "RSA key size is less than 2048 bits")
		}
	}

	if score < 0 {
		score = 0
	}

	analysis.SecurityScore = score
}

// generateRecommendations 生成安全建议
func (ca *CertificateAnalyzer) generateRecommendations(analysis *SecurityAnalysis, cert *x509.Certificate) {
	if analysis.IsExpired {
		analysis.Recommendations = append(analysis.Recommendations, "Renew the expired certificate immediately")
	}

	if analysis.ExpiresSoon {
		analysis.Recommendations = append(analysis.Recommendations, "Plan certificate renewal before expiration")
	}

	if analysis.IsSelfSigned {
		analysis.Recommendations = append(analysis.Recommendations, "Consider using a certificate from a trusted CA")
	}

	if analysis.WeakSignature {
		analysis.Recommendations = append(analysis.Recommendations, "Upgrade to SHA-256 or stronger signature algorithm")
	}

	if pubKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
		if pubKey.N.BitLen() < 2048 {
			analysis.Recommendations = append(analysis.Recommendations, "Upgrade to at least 2048-bit RSA key")
		}
	}

	if len(cert.DNSNames) == 0 {
		analysis.Recommendations = append(analysis.Recommendations, "Add Subject Alternative Names (SAN) for better compatibility")
	}
}