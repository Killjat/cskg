package com.example.appmonitor;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.util.Log;

/**
 * 开机自启动广播接收器
 */
public class BootReceiver extends BroadcastReceiver {

    private static final String TAG = "BootReceiver";

    @Override
    public void onReceive(Context context, Intent intent) {
        if (intent != null && intent.getAction() != null) {
            Log.i(TAG, "接收到广播: " + intent.getAction());
            
            // 检查是否是开机完成广播
            if (Intent.ACTION_BOOT_COMPLETED.equals(intent.getAction()) ||
                "android.intent.action.QUICKBOOT_POWERON".equals(intent.getAction()) ||
                "com.htc.intent.action.QUICKBOOT_POWERON".equals(intent.getAction())) {
                
                Log.i(TAG, "开机完成，启动监控服务");
                
                // 启动监控服务
                Intent serviceIntent = new Intent(context, AppMonitorService.class);
                context.startService(serviceIntent);
            }
        }
    }
}
