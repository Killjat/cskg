// MongoDB 初始化脚本
// 创建数据库和用户

// 切换到 cyberstroll 数据库
db = db.getSiblingDB('cyberstroll');

// 创建用户 (如果不存在)
try {
    db.createUser({
        user: "cyberstroll",
        pwd: "cyberstroll123",
        roles: [
            {
                role: "readWrite",
                db: "cyberstroll"
            }
        ]
    });
    print("✅ 用户 cyberstroll 创建成功");
} catch (e) {
    if (e.code === 51003) {
        print("ℹ️  用户 cyberstroll 已存在");
    } else {
        print("❌ 创建用户失败: " + e.message);
    }
}

// 创建任务集合
db.createCollection("tasks");
print("✅ 任务集合创建成功");

// 创建任务索引
db.tasks.createIndex({ "task_id": 1 }, { unique: true });
db.tasks.createIndex({ "task_initiator": 1 });
db.tasks.createIndex({ "task_status": 1 });
db.tasks.createIndex({ "created_time": -1 });
print("✅ 任务索引创建成功");

// 创建扫描结果集合
db.createCollection("scan_results");
print("✅ 扫描结果集合创建成功");

// 创建扫描结果索引
db.scan_results.createIndex({ "task_id": 1 });
db.scan_results.createIndex({ "ip": 1 });
db.scan_results.createIndex({ "scan_time": -1 });
print("✅ 扫描结果索引创建成功");

// 创建系统配置集合
db.createCollection("system_config");
print("✅ 系统配置集合创建成功");

// 插入默认系统配置
db.system_config.insertOne({
    _id: "default",
    system_ip_pools: [
        "8.8.8.8",
        "1.1.1.1", 
        "114.114.114.114",
        "223.5.5.5"
    ],
    scan_intervals: {
        system_task: 300,
        health_check: 60
    },
    created_time: new Date(),
    updated_time: new Date()
});
print("✅ 默认系统配置插入成功");

print("🎉 MongoDB 初始化完成!");