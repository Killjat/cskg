from pymodbus import __version__ as pymodbus_version
from pymodbus.server import StartTcpServer
from pymodbus.datastore import (
    ModbusSequentialDataBlock,
    ModbusDeviceContext,
    ModbusServerContext
)
from pymodbus.framer import FramerRTU

print(f"Using pymodbus version: {pymodbus_version}")

# 初始化数据存储
# 创建一个设备上下文，包含不同类型的寄存器
def setup_datastore():
    # 离散输入（只读）
    di = ModbusSequentialDataBlock(0, [17] * 100)
    # 线圈（读写）
    co = ModbusSequentialDataBlock(0, [0] * 100)
    # 保持寄存器（读写）
    hr = ModbusSequentialDataBlock(0, [100, 200, 300, 400, 500] * 20)
    # 输入寄存器（只读）
    ir = ModbusSequentialDataBlock(0, [1000, 2000, 3000, 4000, 5000] * 20)

    # 创建设备上下文
    device_context = ModbusDeviceContext(
        di=di,  # 离散输入
        co=co,  # 线圈
        hr=hr,  # 保持寄存器
        ir=ir,  # 输入寄存器
    )

    # 创建服务器上下文
    context = ModbusServerContext(devices=device_context, single=True)
    return context

if __name__ == "__main__":
    try:
        # 设置数据存储
        store = setup_datastore()
        
        # 启动TCP服务器
        print("Starting Modbus TCP Server on 0.0.0.0:502...")
        StartTcpServer(
            context=store,
            address=('0.0.0.0', 502)
        )
    except KeyboardInterrupt:
        print("\nModbus TCP Server stopped.")
    except Exception as e:
        print(f"Error: {e}")
