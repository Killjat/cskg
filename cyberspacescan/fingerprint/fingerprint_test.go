package fingerprint

import (
	"encoding/base64"
	"testing"
)

func TestIdentifyNginx(t *testing.T) {
	banner := "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0\r\nDate: Wed, 07 Jan 2026 10:00:00 GMT\r\n"
	
	fps := Identify(banner, nil)
	
	if len(fps) == 0 {
		t.Fatal("未识别出任何指纹")
	}
	
	found := false
	for _, fp := range fps {
		if fp.Product == "Nginx" {
			found = true
			if fp.Version != "1.18.0" {
				t.Errorf("版本号错误，期望: 1.18.0, 实际: %s", fp.Version)
			}
			if fp.Category != "Web服务器" {
				t.Errorf("类别错误，期望: Web服务器, 实际: %s", fp.Category)
			}
		}
	}
	
	if !found {
		t.Error("未识别出Nginx")
	}
}

func TestIdentifyApache(t *testing.T) {
	banner := "HTTP/1.1 200 OK\r\nServer: Apache/2.4.41 (Ubuntu)\r\n"
	
	fps := Identify(banner, nil)
	
	found := false
	for _, fp := range fps {
		if fp.Product == "Apache" {
			found = true
			if fp.Version != "2.4.41" {
				t.Errorf("版本号错误，期望: 2.4.41, 实际: %s", fp.Version)
			}
			if fp.OS != "Linux/Ubuntu" {
				t.Errorf("操作系统错误，期望: Linux/Ubuntu, 实际: %s", fp.OS)
			}
		}
	}
	
	if !found {
		t.Error("未识别出Apache")
	}
}

func TestIdentifyPHP(t *testing.T) {
	banner := "HTTP/1.1 200 OK\r\nX-Powered-By: PHP/7.4.3\r\nServer: Apache\r\n"
	
	fps := Identify(banner, nil)
	
	foundPHP := false
	foundApache := false
	
	for _, fp := range fps {
		if fp.Product == "PHP" {
			foundPHP = true
			if fp.Version != "7.4.3" {
				t.Errorf("PHP版本号错误，期望: 7.4.3, 实际: %s", fp.Version)
			}
		}
		if fp.Product == "Apache" {
			foundApache = true
		}
	}
	
	if !foundPHP {
		t.Error("未识别出PHP")
	}
	if !foundApache {
		t.Error("未识别出Apache")
	}
}

func TestIdentifyWithResponse(t *testing.T) {
	banner := "HTTP/1.1 200 OK\r\nServer: nginx\r\n"
	response := base64.StdEncoding.EncodeToString([]byte("<!DOCTYPE html>\n<html>\n<head>\n<meta name=\"generator\" content=\"WordPress 5.8\">\n</head>\n<body></body>\n</html>"))
	
	fps := Identify(banner, []byte(response))
	
	foundNginx := false
	foundWordPress := false
	
	for _, fp := range fps {
		if fp.Product == "Nginx" {
			foundNginx = true
		}
		if fp.Product == "WordPress" {
			foundWordPress = true
		}
	}
	
	if !foundNginx {
		t.Error("未识别出Nginx")
	}
	if !foundWordPress {
		t.Error("未识别出WordPress")
	}
}

func TestGetTopFingerprint(t *testing.T) {
	banner := "HTTP/1.1 200 OK\r\nServer: nginx/1.20.1\r\nX-Powered-By: PHP/8.0.0\r\n"
	
	top := GetTopFingerprint(banner, nil)
	
	if top == nil {
		t.Fatal("未获取到顶级指纹")
	}
	
	// Nginx和PHP都应该被识别，但置信度应该选择最高的
	if top.Confidence < 80 {
		t.Errorf("置信度过低: %d", top.Confidence)
	}
}

func TestHasTag(t *testing.T) {
	banner := "HTTP/1.1 200 OK\r\nServer: nginx\r\n"
	
	if !HasTag(banner, nil, "web") {
		t.Error("应该包含 'web' 标签")
	}
	
	if HasTag(banner, nil, "database") {
		t.Error("不应该包含 'database' 标签")
	}
}

func TestGetCategories(t *testing.T) {
	banner := "HTTP/1.1 200 OK\r\nServer: Apache/2.4.41\r\nX-Powered-By: PHP/7.4.3\r\n"
	
	categories := GetCategories(banner, nil)
	
	if len(categories) < 2 {
		t.Errorf("应该识别出至少2个类别，实际: %d", len(categories))
	}
	
	hasWeb := false
	hasLanguage := false
	
	for _, cat := range categories {
		if cat == "Web服务器" {
			hasWeb = true
		}
		if cat == "编程语言" {
			hasLanguage = true
		}
	}
	
	if !hasWeb {
		t.Error("应该包含 'Web服务器' 类别")
	}
	if !hasLanguage {
		t.Error("应该包含 '编程语言' 类别")
	}
}

func TestIdentifyRedis(t *testing.T) {
	banner := "$5\r\nredis_version:6.2.5\r\n"
	
	fps := Identify(banner, nil)
	
	found := false
	for _, fp := range fps {
		if fp.Product == "Redis" {
			found = true
			if fp.Version != "6.2.5" {
				t.Errorf("Redis版本号错误，期望: 6.2.5, 实际: %s", fp.Version)
			}
		}
	}
	
	if !found {
		t.Error("未识别出Redis")
	}
}

func TestIdentifySSH(t *testing.T) {
	banner := "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.3"
	
	fps := Identify(banner, nil)
	
	found := false
	for _, fp := range fps {
		if fp.Product == "OpenSSH" {
			found = true
			if fp.Version != "8.2" {
				t.Errorf("OpenSSH版本号错误，期望: 8.2, 实际: %s", fp.Version)
			}
			if fp.OS != "Linux/Ubuntu" {
				t.Errorf("操作系统错误，期望: Linux/Ubuntu, 实际: %s", fp.OS)
			}
		}
	}
	
	if !found {
		t.Error("未识别出OpenSSH")
	}
}

func TestGenerateCPE(t *testing.T) {
	fp := &Fingerprint{
		Product: "Nginx",
		Vendor:  "Nginx Inc.",
		Version: "1.18.0",
	}
	
	cpe := generateCPE(fp)
	expected := "cpe:/a:nginx_inc.:nginx:1.18.0"
	
	if cpe != expected {
		t.Errorf("CPE生成错误，期望: %s, 实际: %s", expected, cpe)
	}
}
