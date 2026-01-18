package com.example.appmonitor;

import java.util.Date;

/**
 * 访问记录数据模型
 */
public class AccessRecord {
    private long id;
    private String packageName;
    private String appName;
    private String accessType;
    private String accessDetails;
    private Date timestamp;

    public AccessRecord() {
    }

    public AccessRecord(String packageName, String appName, String accessType, String accessDetails, Date timestamp) {
        this.packageName = packageName;
        this.appName = appName;
        this.accessType = accessType;
        this.accessDetails = accessDetails;
        this.timestamp = timestamp;
    }

    public long getId() {
        return id;
    }

    public void setId(long id) {
        this.id = id;
    }

    public String getPackageName() {
        return packageName;
    }

    public void setPackageName(String packageName) {
        this.packageName = packageName;
    }

    public String getAppName() {
        return appName;
    }

    public void setAppName(String appName) {
        this.appName = appName;
    }

    public String getAccessType() {
        return accessType;
    }

    public void setAccessType(String accessType) {
        this.accessType = accessType;
    }

    public String getAccessDetails() {
        return accessDetails;
    }

    public void setAccessDetails(String accessDetails) {
        this.accessDetails = accessDetails;
    }

    public Date getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(Date timestamp) {
        this.timestamp = timestamp;
    }

    @Override
    public String toString() {
        return "AccessRecord{" +
                "id=" + id +
                ", packageName='" + packageName + '\'' +
                ", appName='" + appName + '\'' +
                ", accessType='" + accessType + '\'' +
                ", accessDetails='" + accessDetails + '\'' +
                ", timestamp=" + timestamp +
                '}';
    }
}
