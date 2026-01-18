# GeoLite2 数据库获取与使用指南

## 1. 获取 GeoLite2 数据库

GeoLite2 是 MaxMind 提供的免费 IP 地理位置数据库。要获取最新的 GeoLite2 数据库，您需要按照以下步骤操作：

### 1.1 注册 MaxMind 账号
1. 访问 [MaxMind 官方网站](https://www.maxmind.com/)
2. 点击 "Sign Up" 注册新账号
3. 验证邮箱并登录

### 1.2 创建 License Key
1. 登录后，进入 [Account Settings](https://www.maxmind.com/en/account)
2. 点击左侧菜单的 "License Keys"
3. 点击 "Generate new license key"
4. 选择 "No" 表示不使用 GeoIP Update 客户端
5. 点击 "Confirm"
6. 复制生成的 License Key，妥善保存

### 1.3 下载 GeoLite2 数据库
1. 访问 [GeoLite2 下载页面](https://www.maxmind.com/en/accounts/current/geoip/downloads)
2. 下载以下两个数据库文件：
   - GeoLite2-City.mmdb (包含城市级地理位置数据)
   - GeoLite2-ASN.mmdb (包含 ASN 和 ISP 信息)

## 2. 在程序中使用 GeoLite2 数据库

### 2.1 放置数据库文件
将下载的 `.mmdb` 文件放置在以下任一位置：

- 程序当前目录
- `/usr/share/GeoIP/`
- `/usr/local/share/GeoIP/`
- `/var/lib/GeoIP/`

### 2.2 配置程序使用 GeoLite2

#### 使用命令行参数指定数据库路径
```bash
./taiwan-ip-scan -geoip-db /path/to/GeoLite2-City.mmdb
```

#### 使用配置文件指定数据库路径
编辑 `config.yaml` 文件，添加以下配置：
```yaml
goip:
  db_path: "/path/to/GeoLite2-City.mmdb"
  use_mock_data: false
```

## 3. 验证 GeoLite2 数据库是否正常工作

运行程序后，查看日志输出。如果看到以下日志，表示 GeoLite2 数据库已成功加载：
```
2026/01/18 00:00:00 成功加载GeoLite2数据库: /path/to/GeoLite2-City.mmdb
```

## 4. 自动更新 GeoLite2 数据库

GeoLite2 数据库每周更新一次。您可以使用 `geoipupdate` 工具自动更新数据库：

### 4.1 安装 geoipupdate
```bash
# Debian/Ubuntu
sudo apt-get install geoipupdate

# CentOS/RHEL
sudo yum install geoipupdate

# macOS
brew install geoipupdate
```

### 4.2 配置 geoipupdate
创建或编辑配置文件 `/etc/GeoIP.conf`：
```
AccountID YOUR_ACCOUNT_ID
LicenseKey YOUR_LICENSE_KEY
EditionIDs GeoLite2-City GeoLite2-ASN
DatabaseDirectory /usr/share/GeoIP
```

### 4.3 运行更新
```bash
sudo geoipupdate
```

### 4.4 设置定时任务
```bash
# 每周一凌晨1点更新
sudo crontab -e
0 1 * * 1 /usr/bin/geoipupdate
```

## 5. 注意事项

1. GeoLite2 数据库是免费的，但数据准确性略低于付费版本的 GeoIP2 数据库
2. 请遵守 MaxMind 的 [EULA](https://www.maxmind.com/en/geolite2/eula)
3. 定期更新数据库以确保数据准确性
4. 程序会自动检测并使用可用的 GeoLite2 数据库
5. 如果没有找到 GeoLite2 数据库，程序会跳过 IP 位置查询，但仍会继续执行扫描操作

## 6. 示例输出

当 GeoLite2 数据库正常工作时，程序输出示例：
```
2026/01/18 00:00:00 成功加载GeoLite2数据库: ./GeoLite2-City.mmdb
2026/01/18 00:00:00 开始处理IP段: 1.1.1.0/24
2026/01/18 00:00:05 IP段 1.1.1.0/24 扫描完成，发现 5 个活跃IP
2026/01/18 00:00:05 从缓存获取IP 1.1.1.1 的位置信息
2026/01/18 00:00:05 成功查询 5 个活跃IP的位置信息
2026/01/18 00:00:05 已将IP段 1.1.1.0/24 的扫描结果写入InfluxDB
```