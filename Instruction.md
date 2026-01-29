1. Start esr, orchestrator
2. Start Prosys
3. Update uaclient/systemconfig:
```
...
"services": [
    {
        ...
    },
    {
        "definition": "access",
        "subpath": "access",
        "details": {
            "node": [
                "counter"
            ]
        },
        "registrationPeriod": 30,
        "costUnit": ""
    }
],
"traits": [
    {
        "serverAddress": "opc.tcp://127.0.0.1:53530/OPCUA/SimulationServer",
        "NodeList": {
            "Node_Id": [
                "ns=3;i=1001"
            ]
        },
        ...
    }
]
```
4. Start uaclient
5. Start mosquitto
6. Update telegrapher/systemconfig:
```
...
 "unit_assets": [
    {
        "name": "counter/access",
        "details": {
            "mqtt": [
                "prosys"
            ]
        },
        "services": [],
        "traits": [{
            "broker": "tcp://localhost:1883",
            "pattern": ["node"],
            "period": 5
        }]
    },
    {
        "name": "counter/access", 
        "services": [
            {
                "definition": "access",
                "subpath": "access",
                "registrationPeriod": 30,
                "costUnit": ""
            }
        ],
        "traits": [{
            "broker": "tcp://localhost:1883",
            "pattern": ["mqtt"],
            "period": -1
        }]
    }
],
...
```
7. Start telegrapher
8. Start modbus simulator
9. Update modboss/systemconfig:
```
...
"traits": [
    {
        "serverAddress": "127.0.0.1:5020",
        "register_map": {
            "holdingRegister": [
                "00000,CounterValue,rw,16-bit INT"
            ]
        },
        "period": 5,
        "details": {
            "mqtt": [
                "counter"
            ]
        }
    }
]
...
```
10. Start modboss