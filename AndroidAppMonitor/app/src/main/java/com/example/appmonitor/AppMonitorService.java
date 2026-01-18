package com.example.appmonitor;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.app.usage.UsageEvents;
import android.app.usage.UsageStats;
import android.app.usage.UsageStatsManager;
import android.content.Context;
import android.content.Intent;
import android.content.pm.ApplicationInfo;
import android.content.pm.PackageManager;
import android.os.Handler;
import android.os.IBinder;
import android.util.Log;

import androidx.core.app.NotificationCompat;

import java.util.ArrayList;
import java.util.List;
import java.util.SortedMap;
import java.util.TreeMap;

/**
 * 应用监控服务
 */
public class AppMonitorService extends Service {

    private static final String TAG = "AppMonitorService";
    private static final int NOTIFICATION_ID = 1;
    private static final String CHANNEL_ID = "app_monitor_channel";
    
    private ContentMonitor mContentMonitor;
    private FileMonitor mFileMonitor;
    private Handler mHandler;

    @Override
    public void onCreate() {
        super.onCreate();
        Log.i(TAG, "监控服务已创建");
        
        mHandler = new Handler();
        
        // 初始化监控器
        mContentMonitor = new ContentMonitor(mHandler, this);
        mFileMonitor = new FileMonitor(this);
        
        // 启动监控
        startMonitoring();
        
        // 创建通知渠道
        createNotificationChannel();
        
        // 启动前台服务
        startForeground(NOTIFICATION_ID, buildNotification());
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        Log.i(TAG, "监控服务已启动");
        return START_STICKY;
    }

    @Override
    public void onDestroy() {
        super.onDestroy();
        Log.i(TAG, "监控服务已销毁");
        
        // 停止监控
        stopMonitoring();
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    /**
     * 启动监控
     */
    private void startMonitoring() {
        // 注册内容观察者
        mContentMonitor.register();
        
        // 启动文件监控
        mFileMonitor.startMonitoring();
        
        Log.i(TAG, "监控已启动");
    }

    /**
     * 停止监控
     */
    private void stopMonitoring() {
        // 取消注册内容观察者
        mContentMonitor.unregister();
        
        // 停止文件监控
        mFileMonitor.stopMonitoring();
        
        Log.i(TAG, "监控已停止");
    }

    /**
     * 创建通知渠道
     */
    private void createNotificationChannel() {
        NotificationChannel channel = new NotificationChannel(
                CHANNEL_ID,
                "应用监控通知",
                NotificationManager.IMPORTANCE_DEFAULT
        );
        channel.setDescription("监控应用访问通讯录和相册的通知");
        
        NotificationManager notificationManager = getSystemService(NotificationManager.class);
        if (notificationManager != null) {
            notificationManager.createNotificationChannel(channel);
        }
    }

    /**
     * 构建通知
     */
    private Notification buildNotification() {
        Intent notificationIntent = new Intent(this, MainActivity.class);
        PendingIntent pendingIntent = PendingIntent.getActivity(
                this,
                0,
                notificationIntent,
                PendingIntent.FLAG_IMMUTABLE
        );
        
        return new NotificationCompat.Builder(this, CHANNEL_ID)
                .setContentTitle("应用监控服务")
                .setContentText("正在监控应用访问通讯录和相册")
                .setSmallIcon(android.R.drawable.stat_sys_download_done)
                .setContentIntent(pendingIntent)
                .setPriority(NotificationCompat.PRIORITY_DEFAULT)
                .build();
    }

    /**
     * 获取当前前台运行的APP
     */
    public static AppInfo getCurrentApp(Context context) {
        try {
            UsageStatsManager usageStatsManager = (UsageStatsManager) context.getSystemService(Context.USAGE_STATS_SERVICE);
            if (usageStatsManager == null) {
                return null;
            }
            
            long currentTime = System.currentTimeMillis();
            // 获取最近10秒的使用情况
            List<UsageStats> usageStatsList = usageStatsManager.queryUsageStats(
                    UsageStatsManager.INTERVAL_BEST,
                    currentTime - 10000,
                    currentTime
            );
            
            if (usageStatsList != null && !usageStatsList.isEmpty()) {
                SortedMap<Long, UsageStats> sortedMap = new TreeMap<>();
                for (UsageStats usageStats : usageStatsList) {
                    sortedMap.put(usageStats.getLastTimeUsed(), usageStats);
                }
                
                if (!sortedMap.isEmpty()) {
                    UsageStats currentStats = sortedMap.get(sortedMap.lastKey());
                    String packageName = currentStats.getPackageName();
                    
                    // 获取应用名称
                    PackageManager packageManager = context.getPackageManager();
                    ApplicationInfo appInfo = packageManager.getApplicationInfo(packageName, 0);
                    String appName = packageManager.getApplicationLabel(appInfo).toString();
                    
                    return new AppInfo(packageName, appName);
                }
            }
        } catch (Exception e) {
            Log.e(TAG, "获取当前应用失败: " + e.getMessage());
        }
        
        return null;
    }

    /**
     * 发送通知
     */
    public static void sendNotification(Context context, AccessRecord record) {
        try {
            Intent notificationIntent = new Intent(context, MainActivity.class);
            PendingIntent pendingIntent = PendingIntent.getActivity(
                    context,
                    0,
                    notificationIntent,
                    PendingIntent.FLAG_IMMUTABLE
            );
            
            NotificationCompat.Builder builder = new NotificationCompat.Builder(context, CHANNEL_ID)
                    .setSmallIcon(android.R.drawable.stat_sys_warning)
                    .setContentTitle("应用访问提醒")
                    .setContentText(record.getAppName() + " 访问了" + record.getAccessType())
                    .setStyle(new NotificationCompat.BigTextStyle()
                            .bigText(record.getAppName() + " 在 " + record.getTimestamp().toString() + " 访问了" + record.getAccessType() + "\n" + record.getAccessDetails()))
                    .setPriority(NotificationCompat.PRIORITY_HIGH)
                    .setContentIntent(pendingIntent)
                    .setAutoCancel(true);
            
            NotificationManager notificationManager = (NotificationManager) context.getSystemService(Context.NOTIFICATION_SERVICE);
            if (notificationManager != null) {
                notificationManager.notify((int) System.currentTimeMillis(), builder.build());
            }
        } catch (Exception e) {
            Log.e(TAG, "发送通知失败: " + e.getMessage());
        }
    }

    /**
     * 检查是否有使用情况访问权限
     */
    public static boolean hasUsageStatsPermission(Context context) {
        UsageStatsManager usageStatsManager = (UsageStatsManager) context.getSystemService(Context.USAGE_STATS_SERVICE);
        if (usageStatsManager == null) {
            return false;
        }
        
        long currentTime = System.currentTimeMillis();
        List<UsageStats> usageStatsList = usageStatsManager.queryUsageStats(
                UsageStatsManager.INTERVAL_DAILY,
                currentTime - 1000 * 3600 * 24,
                currentTime
        );
        
        return usageStatsList != null && !usageStatsList.isEmpty();
    }
}
