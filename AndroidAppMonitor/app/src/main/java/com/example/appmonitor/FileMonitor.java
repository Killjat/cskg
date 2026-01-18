package com.example.appmonitor;

import android.content.Context;
import android.os.Environment;
import android.os.FileObserver;
import android.util.Log;

import java.io.File;
import java.util.Date;

/**
 * 文件观察者，用于监控相册访问
 */
public class FileMonitor {

    private static final String TAG = "FileMonitor";
    private static final String ACCESS_TYPE_GALLERY = "相册";
    
    private Context mContext;
    private DBHelper mDbHelper;
    private MediaFileObserver mDcimObserver;
    private MediaFileObserver mPicturesObserver;

    public FileMonitor(Context context) {
        this.mContext = context;
        this.mDbHelper = new DBHelper(context);
    }

    /**
     * 启动监控
     */
    public void startMonitoring() {
        // 监控DCIM目录
        String dcimPath = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DCIM).getPath();
        mDcimObserver = new MediaFileObserver(dcimPath);
        mDcimObserver.startWatching();
        Log.i(TAG, "开始监控DCIM目录: " + dcimPath);

        // 监控Pictures目录
        String picturesPath = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_PICTURES).getPath();
        mPicturesObserver = new MediaFileObserver(picturesPath);
        mPicturesObserver.startWatching();
        Log.i(TAG, "开始监控Pictures目录: " + picturesPath);
    }

    /**
     * 停止监控
     */
    public void stopMonitoring() {
        if (mDcimObserver != null) {
            mDcimObserver.stopWatching();
            mDcimObserver = null;
        }
        if (mPicturesObserver != null) {
            mPicturesObserver.stopWatching();
            mPicturesObserver = null;
        }
        Log.i(TAG, "停止监控相册目录");
    }

    /**
     * 媒体文件观察者
     */
    private class MediaFileObserver extends FileObserver {

        /**
         * Creates a new file observer for a certain file or directory.
         * Monitoring does not start on creation!  You must call {@link #startWatching()}
         * before you will receive events.
         *
         * @param path The file or directory to monitor
         */
        public MediaFileObserver(String path) {
            super(path, ALL_EVENTS);
        }

        @Override
        public void onEvent(int event, String path) {
            if (path == null) {
                return;
            }

            // 过滤掉一些不需要的事件
            if ((event & (ACCESS | MODIFY | OPEN | CLOSE_NOWRITE | CLOSE_WRITE)) != 0) {
                Log.i(TAG, "相册文件事件: " + event + ", 文件: " + path);
                recordFileAccess(path);
            }
        }
    }

    /**
     * 记录文件访问
     */
    private void recordFileAccess(String filePath) {
        try {
            // 获取当前前台运行的APP
            AppInfo currentApp = AppMonitorService.getCurrentApp(mContext);
            if (currentApp != null) {
                // 创建访问记录
                AccessRecord record = new AccessRecord(
                        currentApp.getPackageName(),
                        currentApp.getAppName(),
                        ACCESS_TYPE_GALLERY,
                        "访问了相册文件: " + filePath,
                        new Date()
                );
                // 保存到数据库
                mDbHelper.addAccessRecord(record);
                Log.i(TAG, "记录相册访问: " + currentApp.getAppName() + " -> " + filePath);
                
                // 发送通知
                AppMonitorService.sendNotification(mContext, record);
            }
        } catch (Exception e) {
            Log.e(TAG, "记录文件访问失败: " + e.getMessage());
        }
    }
}
