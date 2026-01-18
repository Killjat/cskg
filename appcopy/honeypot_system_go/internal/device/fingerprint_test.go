package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFingerprintManager(t *testing.T) {
	// 测试创建指纹管理器
	fpManager := NewFingerprintManager(
		"./test_fingerprints.db",
		true,
		true,
		true,
		true,
	)
	assert.NotNil(t, fpManager, "Fingerprint manager should not be nil")

	// 测试关闭功能
	fpManager.Close()
}

func TestFingerprintGeneration(t *testing.T) {
	// 测试指纹生成功能
	fpManager := NewFingerprintManager(
		"./test_fingerprints.db",
		true,
		true,
		true,
		true,
	)
	defer fpManager.Close()

	// 测试生成设备指纹
	clientIP := "192.168.1.100"
	httpHeaders := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	}
	tcpOptions := []byte{0x02, 0x04, 0x05, 0xb4} // MSS选项
	tlsExtensions := []byte{0x00, 0x00, 0x00, 0x00} // TLS扩展示例

	fingerprint := fpManager.GenerateFingerprint(clientIP, httpHeaders, tcpOptions, tlsExtensions)
	assert.NotEmpty(t, fingerprint, "Generated fingerprint should not be empty")
	assert.Len(t, fingerprint, 64, "Fingerprint should be 64 characters long (SHA-256)")
}

func TestFingerprintStorage(t *testing.T) {
	// 测试指纹存储功能
	fpManager := NewFingerprintManager(
		"./test_fingerprints.db",
		true,
		true,
		true,
		true,
	)
	defer fpManager.Close()

	// 测试存储和检索指纹
	clientIP := "192.168.1.100"
	httpHeaders := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}
	tcpOptions := []byte{0x02, 0x04, 0x05, 0xb4}
	tlsExtensions := []byte{0x00, 0x00, 0x00, 0x00}

	fingerprint := fpManager.GenerateFingerprint(clientIP, httpHeaders, tcpOptions, tlsExtensions)

	// 存储指纹
	fpManager.StoreFingerprint(fingerprint, clientIP, "Test Device", time.Now())

	// 检索指纹信息
	fpInfo, err := fpManager.GetFingerprintInfo(fingerprint)
	assert.NoError(t, err, "Should not error when retrieving fingerprint info")
	assert.Equal(t, clientIP, fpInfo.IPAddress, "IP address should match")
	assert.Equal(t, "Test Device", fpInfo.DeviceType, "Device type should match")
}
