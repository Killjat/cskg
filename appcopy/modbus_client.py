from pymodbus import __version__ as pymodbus_version
from pymodbus.client import ModbusTcpClient
from pymodbus.exceptions import ModbusException

print(f"Using pymodbus version: {pymodbus_version}")

def modbus_client_example():
    # 创建Modbus TCP客户端
    client = ModbusTcpClient('localhost', port=502)
    
    try:
        # 连接到服务器
        client.connect()
        print("Connected to Modbus TCP Server")
        
        # 1. 读取保持寄存器（地址0-4）
        print("\n1. Reading Holding Registers (addresses 0-4):")
        response = client.read_holding_registers(address=0, count=5, device_id=1)
        if response.isError():
            print(f"   Error: {response}")
        else:
            print(f"   Values: {response.registers}")
        
        # 2. 读取输入寄存器（地址0-4）
        print("\n2. Reading Input Registers (addresses 0-4):")
        response = client.read_input_registers(address=0, count=5, device_id=1)
        if response.isError():
            print(f"   Error: {response}")
        else:
            print(f"   Values: {response.registers}")
        
        # 3. 读取线圈（地址0-4）
        print("\n3. Reading Coils (addresses 0-4):")
        response = client.read_coils(address=0, count=5, device_id=1)
        if response.isError():
            print(f"   Error: {response}")
        else:
            print(f"   Values: {response.bits[:5]}")
        
        # 4. 读取离散输入（地址0-4）
        print("\n4. Reading Discrete Inputs (addresses 0-4):")
        response = client.read_discrete_inputs(address=0, count=5, device_id=1)
        if response.isError():
            print(f"   Error: {response}")
        else:
            print(f"   Values: {response.bits[:5]}")
        
        # 5. 写入单个保持寄存器
        print("\n5. Writing to Holding Register (address 0):")
        new_value = 999
        response = client.write_register(address=0, value=new_value, device_id=1)
        if response.isError():
            print(f"   Error: {response}")
        else:
            print(f"   Success: Written {new_value} to address 0")
            # 验证写入结果
            response = client.read_holding_registers(address=0, count=1, device_id=1)
            if not response.isError():
                print(f"   Verification: Address 0 now has value {response.registers[0]}")
        
        # 6. 写入多个保持寄存器
        print("\n6. Writing multiple Holding Registers (addresses 1-3):")
        new_values = [888, 777, 666]
        response = client.write_registers(address=1, values=new_values, device_id=1)
        if response.isError():
            print(f"   Error: {response}")
        else:
            print(f"   Success: Written {new_values} to addresses 1-3")
            # 验证写入结果
            response = client.read_holding_registers(address=1, count=3, device_id=1)
            if not response.isError():
                print(f"   Verification: Addresses 1-3 now have values {response.registers}")
        
        # 7. 写入单个线圈
        print("\n7. Writing to Coil (address 0):")
        coil_value = True
        response = client.write_coil(address=0, value=coil_value, device_id=1)
        if response.isError():
            print(f"   Error: {response}")
        else:
            print(f"   Success: Written {coil_value} to coil 0")
            # 验证写入结果
            response = client.read_coils(address=0, count=1, device_id=1)
            if not response.isError():
                print(f"   Verification: Coil 0 now has value {response.bits[0]}")
        
        # 8. 写入多个线圈
        print("\n8. Writing multiple Coils (addresses 1-4):")
        coil_values = [True, False, True, False]
        response = client.write_coils(address=1, values=coil_values, device_id=1)
        if response.isError():
            print(f"   Error: {response}")
        else:
            print(f"   Success: Written {coil_values} to coils 1-4")
            # 验证写入结果
            response = client.read_coils(address=1, count=4, device_id=1)
            if not response.isError():
                print(f"   Verification: Coils 1-4 now have values {response.bits[:4]}")
            
    except ModbusException as e:
        print(f"Modbus Error: {e}")
    except Exception as e:
        print(f"Error: {e}")
    finally:
        # 关闭连接
        client.close()
        print("\nDisconnected from Modbus TCP Server")

if __name__ == "__main__":
    modbus_client_example()
