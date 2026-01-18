package com.example.appmonitor;

import android.content.ContentValues;
import android.content.Context;
import android.database.Cursor;
import android.database.sqlite.SQLiteDatabase;
import android.database.sqlite.SQLiteOpenHelper;

import java.util.ArrayList;
import java.util.Date;
import java.util.List;

/**
 * 数据库辅助类，用于存储和管理访问记录
 */
public class DBHelper extends SQLiteOpenHelper {

    private static final String DATABASE_NAME = "app_monitor.db";
    private static final int DATABASE_VERSION = 1;

    // 访问记录表
    private static final String TABLE_ACCESS_RECORDS = "access_records";
    private static final String COLUMN_ID = "id";
    private static final String COLUMN_PACKAGE_NAME = "package_name";
    private static final String COLUMN_APP_NAME = "app_name";
    private static final String COLUMN_ACCESS_TYPE = "access_type";
    private static final String COLUMN_ACCESS_DETAILS = "access_details";
    private static final String COLUMN_TIMESTAMP = "timestamp";

    // 创建表的SQL语句
    private static final String CREATE_TABLE_ACCESS_RECORDS = "CREATE TABLE " + TABLE_ACCESS_RECORDS + "(" +
            COLUMN_ID + " INTEGER PRIMARY KEY AUTOINCREMENT, " +
            COLUMN_PACKAGE_NAME + " TEXT NOT NULL, " +
            COLUMN_APP_NAME + " TEXT NOT NULL, " +
            COLUMN_ACCESS_TYPE + " TEXT NOT NULL, " +
            COLUMN_ACCESS_DETAILS + " TEXT, " +
            COLUMN_TIMESTAMP + " INTEGER NOT NULL" +
            ");";

    public DBHelper(Context context) {
        super(context, DATABASE_NAME, null, DATABASE_VERSION);
    }

    @Override
    public void onCreate(SQLiteDatabase db) {
        db.execSQL(CREATE_TABLE_ACCESS_RECORDS);
    }

    @Override
    public void onUpgrade(SQLiteDatabase db, int oldVersion, int newVersion) {
        // 如果需要升级数据库，可以在这里处理
        db.execSQL("DROP TABLE IF EXISTS " + TABLE_ACCESS_RECORDS);
        onCreate(db);
    }

    /**
     * 添加访问记录
     */
    public long addAccessRecord(AccessRecord record) {
        SQLiteDatabase db = this.getWritableDatabase();
        ContentValues values = new ContentValues();
        values.put(COLUMN_PACKAGE_NAME, record.getPackageName());
        values.put(COLUMN_APP_NAME, record.getAppName());
        values.put(COLUMN_ACCESS_TYPE, record.getAccessType());
        values.put(COLUMN_ACCESS_DETAILS, record.getAccessDetails());
        values.put(COLUMN_TIMESTAMP, record.getTimestamp().getTime());

        long id = db.insert(TABLE_ACCESS_RECORDS, null, values);
        db.close();
        return id;
    }

    /**
     * 获取所有访问记录
     */
    public List<AccessRecord> getAllAccessRecords() {
        List<AccessRecord> records = new ArrayList<>();
        String selectQuery = "SELECT * FROM " + TABLE_ACCESS_RECORDS + " ORDER BY " + COLUMN_TIMESTAMP + " DESC";

        SQLiteDatabase db = this.getReadableDatabase();
        Cursor cursor = db.rawQuery(selectQuery, null);

        if (cursor.moveToFirst()) {
            do {
                AccessRecord record = new AccessRecord();
                record.setId(cursor.getLong(cursor.getColumnIndex(COLUMN_ID)));
                record.setPackageName(cursor.getString(cursor.getColumnIndex(COLUMN_PACKAGE_NAME)));
                record.setAppName(cursor.getString(cursor.getColumnIndex(COLUMN_APP_NAME)));
                record.setAccessType(cursor.getString(cursor.getColumnIndex(COLUMN_ACCESS_TYPE)));
                record.setAccessDetails(cursor.getString(cursor.getColumnIndex(COLUMN_ACCESS_DETAILS)));
                record.setTimestamp(new Date(cursor.getLong(cursor.getColumnIndex(COLUMN_TIMESTAMP))));
                records.add(record);
            } while (cursor.moveToNext());
        }

        cursor.close();
        db.close();
        return records;
    }

    /**
     * 获取指定类型的访问记录
     */
    public List<AccessRecord> getAccessRecordsByType(String accessType) {
        List<AccessRecord> records = new ArrayList<>();
        String selectQuery = "SELECT * FROM " + TABLE_ACCESS_RECORDS + " WHERE " + COLUMN_ACCESS_TYPE + " = ? ORDER BY " + COLUMN_TIMESTAMP + " DESC";

        SQLiteDatabase db = this.getReadableDatabase();
        Cursor cursor = db.rawQuery(selectQuery, new String[]{accessType});

        if (cursor.moveToFirst()) {
            do {
                AccessRecord record = new AccessRecord();
                record.setId(cursor.getLong(cursor.getColumnIndex(COLUMN_ID)));
                record.setPackageName(cursor.getString(cursor.getColumnIndex(COLUMN_PACKAGE_NAME)));
                record.setAppName(cursor.getString(cursor.getColumnIndex(COLUMN_APP_NAME)));
                record.setAccessType(cursor.getString(cursor.getColumnIndex(COLUMN_ACCESS_TYPE)));
                record.setAccessDetails(cursor.getString(cursor.getColumnIndex(COLUMN_ACCESS_DETAILS)));
                record.setTimestamp(new Date(cursor.getLong(cursor.getColumnIndex(COLUMN_TIMESTAMP))));
                records.add(record);
            } while (cursor.moveToNext());
        }

        cursor.close();
        db.close();
        return records;
    }

    /**
     * 清空所有访问记录
     */
    public void clearAllRecords() {
        SQLiteDatabase db = this.getWritableDatabase();
        db.delete(TABLE_ACCESS_RECORDS, null, null);
        db.close();
    }

    /**
     * 获取访问记录总数
     */
    public int getRecordsCount() {
        String countQuery = "SELECT * FROM " + TABLE_ACCESS_RECORDS;
        SQLiteDatabase db = this.getReadableDatabase();
        Cursor cursor = db.rawQuery(countQuery, null);
        int count = cursor.getCount();
        cursor.close();
        db.close();
        return count;
    }
}