import mysql.connector
from mysql.connector import Error

class Database:
    def __init__(self, host='localhost', user='root', password='password', database='lan_scan'):
        self.host = host
        self.user = user
        self.password = password
        self.database = database
        self.connection = None
        self.cursor = None

    def connect(self):
        """连接到MySQL数据库"""
        try:
            self.connection = mysql.connector.connect(
                host=self.host,
                user=self.user,
                password=self.password,
                database=self.database
            )
            if self.connection.is_connected():
                self.cursor = self.connection.cursor()
                print(f"成功连接到数据库: {self.database}")
                return True
        except Error as e:
            print(f"数据库连接错误: {e}")
            return False

    def create_tables(self):
        """创建数据库表"""
        if not self.connection or not self.connection.is_connected():
            self.connect()

        # 创建设备表
        device_table = """
        CREATE TABLE IF NOT EXISTS devices (
            id INT AUTO_INCREMENT PRIMARY KEY,
            ip VARCHAR(15) NOT NULL,
            mac VARCHAR(17),
            hostname VARCHAR(255),
            status VARCHAR(20),
            scan_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE KEY unique_ip (ip)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
        """

        # 创建端口表
        port_table = """
        CREATE TABLE IF NOT EXISTS ports (
            id INT AUTO_INCREMENT PRIMARY KEY,
            device_id INT,
            port INT NOT NULL,
            protocol VARCHAR(10),
            status VARCHAR(20),
            service VARCHAR(255),
            application VARCHAR(255),
            scan_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
            UNIQUE KEY unique_device_port (device_id, port, protocol)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
        """

        # 创建流量表
        traffic_table = """
        CREATE TABLE IF NOT EXISTS traffic (
            id INT AUTO_INCREMENT PRIMARY KEY,
            source_ip VARCHAR(15),
            destination_ip VARCHAR(15),
            source_port INT,
            destination_port INT,
            protocol VARCHAR(10),
            length INT,
            timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
        """

        try:
            tables = [device_table, port_table, traffic_table]
            for table in tables:
                self.cursor.execute(table)
            self.connection.commit()
            print("数据库表创建成功")
            return True
        except Error as e:
            print(f"创建表错误: {e}")
            return False

    def insert_device(self, ip, mac=None, hostname=None, status='up'):
        """插入设备信息"""
        try:
            # 先尝试更新，不存在则插入
            update_query = """
            INSERT INTO devices (ip, mac, hostname, status)
            VALUES (%s, %s, %s, %s)
            ON DUPLICATE KEY UPDATE 
                mac = VALUES(mac),
                hostname = VALUES(hostname),
                status = VALUES(status),
                scan_time = CURRENT_TIMESTAMP
            """
            self.cursor.execute(update_query, (ip, mac, hostname, status))
            self.connection.commit()
            # 获取设备ID
            self.cursor.execute("SELECT id FROM devices WHERE ip = %s", (ip,))
            result = self.cursor.fetchone()
            return result[0] if result else None
        except Error as e:
            print(f"插入设备错误: {e}")
            return None

    def insert_port(self, device_id, port, protocol, status, service=None, application=None):
        """插入端口信息"""
        try:
            insert_query = """
            INSERT INTO ports (device_id, port, protocol, status, service, application)
            VALUES (%s, %s, %s, %s, %s, %s)
            ON DUPLICATE KEY UPDATE 
                status = VALUES(status),
                service = VALUES(service),
                application = VALUES(application),
                scan_time = CURRENT_TIMESTAMP
            """
            self.cursor.execute(insert_query, (device_id, port, protocol, status, service, application))
            self.connection.commit()
            return True
        except Error as e:
            print(f"插入端口错误: {e}")
            return False

    def insert_traffic(self, source_ip, destination_ip, source_port, destination_port, protocol, length):
        """插入流量信息"""
        try:
            insert_query = """
            INSERT INTO traffic (source_ip, destination_ip, source_port, destination_port, protocol, length)
            VALUES (%s, %s, %s, %s, %s, %s)
            """
            self.cursor.execute(insert_query, (source_ip, destination_ip, source_port, destination_port, protocol, length))
            self.connection.commit()
            return True
        except Error as e:
            print(f"插入流量错误: {e}")
            return False

    def get_all_devices(self):
        """获取所有设备信息"""
        try:
            query = "SELECT * FROM devices ORDER BY scan_time DESC"
            self.cursor.execute(query)
            return self.cursor.fetchall()
        except Error as e:
            print(f"查询设备错误: {e}")
            return []

    def get_device_ports(self, device_id):
        """获取设备的所有端口信息"""
        try:
            query = "SELECT * FROM ports WHERE device_id = %s ORDER BY port"
            self.cursor.execute(query, (device_id,))
            return self.cursor.fetchall()
        except Error as e:
            print(f"查询端口错误: {e}")
            return []

    def get_recent_traffic(self, limit=100):
        """获取最近的流量信息"""
        try:
            query = "SELECT * FROM traffic ORDER BY timestamp DESC LIMIT %s"
            self.cursor.execute(query, (limit,))
            return self.cursor.fetchall()
        except Error as e:
            print(f"查询流量错误: {e}")
            return []

    def close(self):
        """关闭数据库连接"""
        if self.cursor:
            self.cursor.close()
        if self.connection and self.connection.is_connected():
            self.connection.close()
            print("数据库连接已关闭")
