package com.example.appmonitor;

import android.content.Context;
import android.database.ContentObserver;
import android.net.Uri;
import android.os.Handler;
import android.provider.ContactsContract;
import android.util.Log;

import java.util.Date;

/**
 * 内容观察者，用于监控通讯录的访问
 */
public class ContentMonitor extends ContentObserver {

    private static final String TAG = "ContentMonitor";
    private static final String ACCESS_TYPE_CONTACTS = "通讯录";
    
    private Context mContext;
    private DBHelper mDbHelper;

    /**
     * 创建一个新的内容观察者
     *
     * @param handler The handler to run {@link #onChange} on, or null if none.
     */
    public ContentMonitor(Handler handler, Context context) {
        super(handler);
        this.mContext = context;
        this.mDbHelper = new DBHelper(context);
    }

    @Override
    public void onChange(boolean selfChange) {
        super.onChange(selfChange);
        Log.i(TAG, "通讯录数据发生变化");
        // 当通讯录数据发生变化时，记录访问行为
        recordAccess();
    }

    @Override
    public void onChange(boolean selfChange, Uri uri) {
        super.onChange(selfChange, uri);
        Log.i(TAG, "通讯录数据发生变化: " + uri.toString());
        // 当通讯录数据发生变化时，记录访问行为
        recordAccess();
    }

    /**
     * 记录访问行为
     */
    private void recordAccess() {
        try {
            // 获取当前前台运行的APP
            AppInfo currentApp = AppMonitorService.getCurrentApp(mContext);
            if (currentApp != null) {
                // 创建访问记录
                AccessRecord record = new AccessRecord(
                        currentApp.getPackageName(),
                        currentApp.getAppName(),
                        ACCESS_TYPE_CONTACTS,
                        "访问了通讯录数据",
                        new Date()
                );
                // 保存到数据库
                mDbHelper.addAccessRecord(record);
                Log.i(TAG, "记录通讯录访问: " + currentApp.getAppName());
                
                // 发送通知
                AppMonitorService.sendNotification(mContext, record);
            }
        } catch (Exception e) {
            Log.e(TAG, "记录访问行为失败: " + e.getMessage());
        }
    }

    /**
     * 注册观察者
     */
    public void register() {
        // 监听通讯录数据变化
        mContext.getContentResolver().registerContentObserver(
                ContactsContract.Contacts.CONTENT_URI,
                true, // 监听子目录
                this
        );
        Log.i(TAG, "已注册通讯录观察者");
    }

    /**
     * 取消注册观察者
     */
    public void unregister() {
        mContext.getContentResolver().unregisterContentObserver(this);
        Log.i(TAG, "已取消注册通讯录观察者");
    }
}
