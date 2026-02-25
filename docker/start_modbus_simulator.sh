#!/usr/bin/env python3
"""Modbus TCP server - pymodbus 3.11.x"""
from pymodbus.datastore import ModbusSequentialDataBlock, ModbusDeviceContext, ModbusServerContext
from pymodbus.server import StartTcpServer

store = ModbusDeviceContext(
    di=ModbusSequentialDataBlock(0, [0] * 100),  # Discrete Inputs
    co=ModbusSequentialDataBlock(0, [0] * 100),  # Coils
    hr=ModbusSequentialDataBlock(0, [0] * 100),  # Holding Registers
    ir=ModbusSequentialDataBlock(0, [0] * 100)   # Input Registers
)
context = ModbusServerContext(devices=store, single=True)

print("Modbus TCP server running on 0.0.0.0:5020", flush=True)
StartTcpServer(context=context, address=("0.0.0.0", 5020))
