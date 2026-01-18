# Android App Monitor

一个监控手机上哪些APP读取通讯录以及相册的Android应用。

## 功能特点

- 监控APP读取通讯录的行为
- 监控APP读取相册的行为
- 显示详细的访问记录
- 实时通知
- 访问统计分析

## 技术实现

- 使用Android权限管理API
- 使用ContentObserver监控数据变化
- 使用FileObserver监控文件访问
- 使用JobScheduler定期检查

## 项目结构

```
AndroidAppMonitor/
├── app/
│   ├── src/
│   │   ├── main/
│   │   │   ├── AndroidManifest.xml
│   │   │   ├── java/com/example/appmonitor/
│   │   │   │   ├── MainActivity.java
│   │   │   │   ├── AppMonitorService.java
│   │   │   │   ├── ContentMonitor.java
│   │   │   │   ├── FileMonitor.java
│   │   │   │   ├── AccessRecord.java
│   │   │   │   └── AccessRecordAdapter.java
│   │   │   └── res/
│   │   │       ├── layout/
│   │   │       │   └── activity_main.xml
│   │   │       └── layout/
│   │   │           └── item_access_record.xml
│   │   └── test/
│   └── build.gradle
├── build.gradle
└── settings.gradle
```

## 如何使用

1. 将项目导入Android Studio
2. 构建并运行应用
3. 授予必要的权限
4. 查看监控记录

## 必要权限

- READ_CONTACTS
- WRITE_EXTERNAL_STORAGE (用于监控相册)
- READ_EXTERNAL_STORAGE
- PACKAGE_USAGE_STATS
- RECEIVE_BOOT_COMPLETED
- FOREGROUND_SERVICE

## 实现原理

1. **通讯录监控**：使用ContentObserver监听ContactsContract.Contacts.CONTENT_URI的变化
2. **相册监控**：使用FileObserver监控DCIM和Pictures目录的访问
3. **APP识别**：通过UsageStatsManager获取当前前台运行的APP
4. **记录存储**：使用SQLite数据库存储访问记录
5. **实时通知**：当检测到访问时发送通知

## 开发说明

- 最低支持Android 6.0 (API 23)
- 推荐使用Android Studio Arctic Fox或更高版本
- 需要在设置中授予"使用情况访问权限"

## 后续优化

- 添加数据可视化图表
- 支持导出监控报告
- 添加白名单功能
- 支持实时监控模式
- 添加访问频率统计
