package com.example.appmonitor;

import android.app.AlertDialog;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.net.Uri;
import android.os.Bundle;
import android.provider.Settings;
import android.view.Menu;
import android.view.MenuItem;
import android.view.View;
import android.widget.Button;
import android.widget.ProgressBar;
import android.widget.TextView;
import android.widget.Toast;

import androidx.annotation.NonNull;
import androidx.appcompat.app.AppCompatActivity;
import androidx.core.app.ActivityCompat;
import androidx.core.content.ContextCompat;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import java.util.List;

public class MainActivity extends AppCompatActivity {

    private static final int PERMISSION_REQUEST_CODE = 100;
    private static final String[] REQUIRED_PERMISSIONS = {
            android.Manifest.permission.READ_CONTACTS,
            android.Manifest.permission.READ_EXTERNAL_STORAGE
    };

    private RecyclerView mRecyclerView;
    private AccessRecordAdapter mAdapter;
    private ProgressBar mProgressBar;
    private TextView mEmptyText;
    private Button mStartServiceBtn;
    private Button mStopServiceBtn;
    private DBHelper mDbHelper;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        mDbHelper = new DBHelper(this);
        
        // 初始化UI组件
        initUI();
        
        // 检查权限
        checkPermissions();
        
        // 加载访问记录
        loadAccessRecords();
        
        // 检查并启动服务
        checkAndStartService();
    }

    @Override
    protected void onResume() {
        super.onResume();
        // 重新加载数据
        loadAccessRecords();
    }

    /**
     * 初始化UI组件
     */
    private void initUI() {
        mRecyclerView = findViewById(R.id.recycler_view);
        mProgressBar = findViewById(R.id.progress_bar);
        mEmptyText = findViewById(R.id.empty_text);
        mStartServiceBtn = findViewById(R.id.start_service_btn);
        mStopServiceBtn = findViewById(R.id.stop_service_btn);

        // 设置RecyclerView
        mRecyclerView.setLayoutManager(new LinearLayoutManager(this));
        mAdapter = new AccessRecordAdapter(this);
        mRecyclerView.setAdapter(mAdapter);

        // 设置按钮点击事件
        mStartServiceBtn.setOnClickListener(v -> startMonitoringService());
        mStopServiceBtn.setOnClickListener(v -> stopMonitoringService());
    }

    /**
     * 检查权限
     */
    private void checkPermissions() {
        boolean allPermissionsGranted = true;
        
        for (String permission : REQUIRED_PERMISSIONS) {
            if (ContextCompat.checkSelfPermission(this, permission) != PackageManager.PERMISSION_GRANTED) {
                allPermissionsGranted = false;
                break;
            }
        }
        
        if (!allPermissionsGranted) {
            ActivityCompat.requestPermissions(this, REQUIRED_PERMISSIONS, PERMISSION_REQUEST_CODE);
        }
        
        // 检查使用情况访问权限
        if (!AppMonitorService.hasUsageStatsPermission(this)) {
            showUsageStatsPermissionDialog();
        }
    }

    /**
     * 显示使用情况访问权限对话框
     */
    private void showUsageStatsPermissionDialog() {
        new AlertDialog.Builder(this)
                .setTitle("需要使用情况访问权限")
                .setMessage("为了监控应用访问行为，需要您授予使用情况访问权限")
                .setPositiveButton("去设置", (dialog, which) -> {
                    Intent intent = new Intent(Settings.ACTION_USAGE_ACCESS_SETTINGS);
                    startActivity(intent);
                })
                .setNegativeButton("取消", null)
                .show();
    }

    /**
     * 检查并启动服务
     */
    private void checkAndStartService() {
        // 启动监控服务
        startMonitoringService();
    }

    /**
     * 启动监控服务
     */
    private void startMonitoringService() {
        Intent serviceIntent = new Intent(this, AppMonitorService.class);
        ContextCompat.startForegroundService(this, serviceIntent);
        Toast.makeText(this, "监控服务已启动", Toast.LENGTH_SHORT).show();
        updateServiceButtons(true);
    }

    /**
     * 停止监控服务
     */
    private void stopMonitoringService() {
        Intent serviceIntent = new Intent(this, AppMonitorService.class);
        stopService(serviceIntent);
        Toast.makeText(this, "监控服务已停止", Toast.LENGTH_SHORT).show();
        updateServiceButtons(false);
    }

    /**
     * 更新服务按钮状态
     */
    private void updateServiceButtons(boolean isRunning) {
        mStartServiceBtn.setEnabled(!isRunning);
        mStopServiceBtn.setEnabled(isRunning);
    }

    /**
     * 加载访问记录
     */
    private void loadAccessRecords() {
        mProgressBar.setVisibility(View.VISIBLE);
        mEmptyText.setVisibility(View.GONE);
        
        // 在后台线程加载数据
        new Thread(() -> {
            List<AccessRecord> records = mDbHelper.getAllAccessRecords();
            
            // 在主线程更新UI
            runOnUiThread(() -> {
                mProgressBar.setVisibility(View.GONE);
                
                if (records.isEmpty()) {
                    mEmptyText.setVisibility(View.VISIBLE);
                } else {
                    mEmptyText.setVisibility(View.GONE);
                    mAdapter.setRecords(records);
                }
            });
        }).start();
    }

    /**
     * 清空所有记录
     */
    private void clearAllRecords() {
        new AlertDialog.Builder(this)
                .setTitle("确认清空")
                .setMessage("确定要清空所有访问记录吗？")
                .setPositiveButton("确定", (dialog, which) -> {
                    mDbHelper.clearAllRecords();
                    loadAccessRecords();
                    Toast.makeText(this, "记录已清空", Toast.LENGTH_SHORT).show();
                })
                .setNegativeButton("取消", null)
                .show();
    }

    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        getMenuInflater().inflate(R.menu.main_menu, menu);
        return true;
    }

    @Override
    public boolean onOptionsItemSelected(@NonNull MenuItem item) {
        int id = item.getItemId();
        
        if (id == R.id.action_clear) {
            clearAllRecords();
            return true;
        } else if (id == R.id.action_settings) {
            // 打开设置页面
            Toast.makeText(this, "设置功能开发中", Toast.LENGTH_SHORT).show();
            return true;
        } else if (id == R.id.action_about) {
            // 显示关于页面
            showAboutDialog();
            return true;
        }
        
        return super.onOptionsItemSelected(item);
    }

    /**
     * 显示关于对话框
     */
    private void showAboutDialog() {
        new AlertDialog.Builder(this)
                .setTitle("关于")
                .setMessage("Android App Monitor v1.0\n\n监控手机上哪些APP读取通讯录以及相册的应用")
                .setPositiveButton("确定", null)
                .show();
    }

    @Override
    public void onRequestPermissionsResult(int requestCode, @NonNull String[] permissions, @NonNull int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        
        if (requestCode == PERMISSION_REQUEST_CODE) {
            boolean allGranted = true;
            for (int result : grantResults) {
                if (result != PackageManager.PERMISSION_GRANTED) {
                    allGranted = false;
                    break;
                }
            }
            
            if (allGranted) {
                Toast.makeText(this, "权限已授予", Toast.LENGTH_SHORT).show();
            } else {
                Toast.makeText(this, "部分权限未授予，应用可能无法正常工作", Toast.LENGTH_SHORT).show();
            }
        }
    }
}
