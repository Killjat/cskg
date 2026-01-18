package com.example.appmonitor;

import android.content.Context;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import java.text.SimpleDateFormat;
import java.util.List;
import java.util.Locale;

/**
 * 访问记录适配器
 */
public class AccessRecordAdapter extends RecyclerView.Adapter<AccessRecordAdapter.ViewHolder> {

    private Context mContext;
    private List<AccessRecord> mRecords;
    private SimpleDateFormat mDateFormat;

    /**
     * 构造函数
     * @param context 上下文
     */
    public AccessRecordAdapter(Context context) {
        this.mContext = context;
        this.mDateFormat = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.getDefault());
    }

    /**
     * 设置数据
     * @param records 访问记录列表
     */
    public void setRecords(List<AccessRecord> records) {
        this.mRecords = records;
        notifyDataSetChanged();
    }

    @NonNull
    @Override
    public ViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(mContext).inflate(R.layout.item_access_record, parent, false);
        return new ViewHolder(view);
    }

    @Override
    public void onBindViewHolder(@NonNull ViewHolder holder, int position) {
        if (mRecords == null || mRecords.isEmpty()) {
            return;
        }

        AccessRecord record = mRecords.get(position);
        
        // 设置应用名称
        holder.appNameTv.setText(record.getAppName());
        
        // 设置访问类型
        holder.accessTypeTv.setText(record.getAccessType());
        
        // 设置访问详情
        holder.accessDetailsTv.setText(record.getAccessDetails());
        
        // 设置时间戳
        String timeStr = mDateFormat.format(record.getTimestamp());
        holder.timestampTv.setText(timeStr);
        
        // 设置包名
        holder.packageNameTv.setText(record.getPackageName());
    }

    @Override
    public int getItemCount() {
        return mRecords == null ? 0 : mRecords.size();
    }

    /**
     * 视图持有者
     */
    static class ViewHolder extends RecyclerView.ViewHolder {
        TextView appNameTv;
        TextView accessTypeTv;
        TextView accessDetailsTv;
        TextView timestampTv;
        TextView packageNameTv;

        ViewHolder(@NonNull View itemView) {
            super(itemView);
            appNameTv = itemView.findViewById(R.id.app_name_tv);
            accessTypeTv = itemView.findViewById(R.id.access_type_tv);
            accessDetailsTv = itemView.findViewById(R.id.access_details_tv);
            timestampTv = itemView.findViewById(R.id.timestamp_tv);
            packageNameTv = itemView.findViewById(R.id.package_name_tv);
        }
    }
}
