package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// Installer 远程安装器
type Installer struct {
	// 客户端二进制文件路径
	ClientBinaryPath string
}

// NewInstaller 创建安装器
func NewInstaller(clientBinaryPath string) *Installer {
	return &Installer{
		ClientBinaryPath: clientBinaryPath,
	}
}

// Install 安装客户端
func (i *Installer) Install(host string, port int, user, password, nodeID string) error {
	// 建立SSH连接
	conn, err := i.establishSSHConnection(host, port, user, password)
	if err != nil {
		return fmt.Errorf("无法建立SSH连接: %v", err)
	}
	defer conn.Close()

	// 上传客户端二进制文件
	err = i.uploadClientBinary(conn, nodeID)
	if err != nil {
		return fmt.Errorf("无法上传客户端二进制文件: %v", err)
	}

	// 创建systemd服务文件
	err = i.createSystemdService(conn, nodeID)
	if err != nil {
		return fmt.Errorf("无法创建systemd服务文件: %v", err)
	}

	// 启动服务
	err = i.startService(conn, nodeID)
	if err != nil {
		return fmt.Errorf("无法启动服务: %v", err)
	}

	return nil
}

// Uninstall 卸载客户端
func (i *Installer) Uninstall(host string, port int, user, password, nodeID string) error {
	// 建立SSH连接
	conn, err := i.establishSSHConnection(host, port, user, password)
	if err != nil {
		return fmt.Errorf("无法建立SSH连接: %v", err)
	}
	defer conn.Close()

	// 停止服务
	err = i.stopService(conn, nodeID)
	if err != nil {
		return fmt.Errorf("无法停止服务: %v", err)
	}

	// 删除客户端文件和服务
	err = i.removeClientFiles(conn, nodeID)
	if err != nil {
		return fmt.Errorf("无法删除客户端文件: %v", err)
	}

	return nil
}

// establishSSHConnection 建立SSH连接
func (i *Installer) establishSSHConnection(host string, port int, user, password string) (*ssh.Client, error) {
	// 配置SSH客户端
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 连接到远程主机
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// uploadClientBinary 上传客户端二进制文件
func (i *Installer) uploadClientBinary(conn *ssh.Client, nodeID string) error {
	// 创建SSH会话
	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	// 读取本地二进制文件
	localPath := i.ClientBinaryPath
	data, err := os.ReadFile(localPath)
	if err != nil {
		return err
	}

	// 确保目标目录存在
	targetDir := fmt.Sprintf("/opt/nodemanage/%s", nodeID)
	cmd := fmt.Sprintf("mkdir -p %s", targetDir)
	if err := sess.Run(cmd); err != nil {
		return err
	}

	// 上传文件
	targetPath := fmt.Sprintf("%s/client", targetDir)
	cmd = fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", targetPath, string(data))
	if err := sess.Run(cmd); err != nil {
		return err
	}

	// 设置可执行权限
	cmd = fmt.Sprintf("chmod +x %s", targetPath)
	if err := sess.Run(cmd); err != nil {
		return err
	}

	return nil
}

// createSystemdService 创建systemd服务文件
func (i *Installer) createSystemdService(conn *ssh.Client, nodeID string) error {
	// 创建SSH会话
	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	// 服务文件内容
	serviceContent := fmt.Sprintf(`[Unit]
Description=Nodemanage Client
After=network.target

[Service]
Type=simple
ExecStart=/opt/nodemanage/%s/client -node-id %s
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, nodeID, nodeID)

	// 创建服务文件
	cmd := fmt.Sprintf("cat > /etc/systemd/system/nodemanage-%s.service << 'EOF'\n%s\nEOF", nodeID, serviceContent)
	if err := sess.Run(cmd); err != nil {
		return err
	}

	// 重新加载systemd
	cmd = "systemctl daemon-reload"
	if err := sess.Run(cmd); err != nil {
		return err
	}

	return nil
}

// startService 启动服务
func (i *Installer) startService(conn *ssh.Client, nodeID string) error {
	// 创建SSH会话
	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	// 启动服务
	cmd := fmt.Sprintf("systemctl start nodemanage-%s.service", nodeID)
	if err := sess.Run(cmd); err != nil {
		return err
	}

	// 启用服务
	cmd = fmt.Sprintf("systemctl enable nodemanage-%s.service", nodeID)
	if err := sess.Run(cmd); err != nil {
		return err
	}

	return nil
}

// stopService 停止服务
func (i *Installer) stopService(conn *ssh.Client, nodeID string) error {
	// 创建SSH会话
	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	// 停止服务
	cmd := fmt.Sprintf("systemctl stop nodemanage-%s.service 2>/dev/null || true", nodeID)
	if err := sess.Run(cmd); err != nil {
		return err
	}

	// 禁用服务
	cmd = fmt.Sprintf("systemctl disable nodemanage-%s.service 2>/dev/null || true", nodeID)
	if err := sess.Run(cmd); err != nil {
		return err
	}

	return nil
}

// removeClientFiles 删除客户端文件
func (i *Installer) removeClientFiles(conn *ssh.Client, nodeID string) error {
	// 创建SSH会话
	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	// 删除服务文件
	cmd := fmt.Sprintf("rm -f /etc/systemd/system/nodemanage-%s.service", nodeID)
	if err := sess.Run(cmd); err != nil {
		return err
	}

	// 删除客户端目录
	cmd = fmt.Sprintf("rm -rf /opt/nodemanage/%s", nodeID)
	if err := sess.Run(cmd); err != nil {
		return err
	}

	// 重新加载systemd
	cmd = "systemctl daemon-reload"
	if err := sess.Run(cmd); err != nil {
		return err
	}

	return nil
}
