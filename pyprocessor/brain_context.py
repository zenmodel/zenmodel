import sqlite3
import json
import sys
import hashlib
import os
from typing import Any
from abc import ABC

class BrainContextReader(ABC):
    def __init__(self, db_path: str):
        self.db_path = db_path
        self.conn = self.init_db()
        self.current_neuron_id = ""

    def init_db(self):
        if not os.path.exists(self.db_path):
            print(f"错误：数据库文件 '{self.db_path}' 不存在", file=sys.stderr)
            sys.exit(1)
        
        try:
            conn = sqlite3.connect(self.db_path)
            cursor = conn.cursor()
            # 检查 memory 表是否存在
            cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='memory'")
            if cursor.fetchone() is None:
                print(f"错误：数据库文件 '{self.db_path}' 中不存在 'memory' 表", file=sys.stderr)
                sys.exit(1)
            return conn
        except sqlite3.Error as e:
            print(f"数据库连接错误: {e}", file=sys.stderr)
            sys.exit(1)

    def get_memory(self, key: Any) -> Any:
        try:
            cursor = self.conn.cursor()
            hashed_key = self.hash_key(key)
            cursor.execute("SELECT value, type FROM memory WHERE key = ?", (str(hashed_key),))
            result = cursor.fetchone()
            
            if result is None:
                raise KeyError(f"未找到键 '{key}'")
            
            value_json, value_type = result
            
            if value_type == 'int':
                return int(value_json)
            elif value_type == 'float':
                return float(value_json)
            elif value_type == 'bool':
                return json.loads(value_json)
            else:
                return json.loads(value_json)
        except (sqlite3.Error, json.JSONDecodeError, KeyError) as e:
            print(f"获取内存错误 ({key}): {e}", file=sys.stderr)
            return None

    def exist_memory(self, key: Any) -> bool:
        cursor = self.conn.cursor()
        hashed_key = self.hash_key(key)
        cursor.execute("SELECT 1 FROM memory WHERE key = ?", (str(hashed_key),))
        return cursor.fetchone() is not None

    def get_current_neuron_id(self) -> str:
        return self.current_neuron_id

    @staticmethod
    def hash_key(key):
        if not isinstance(key, (int, float, str, bytes, bytearray)):
            raise TypeError("不支持的 key 类型")
        
        key_bytes = str(key).encode('utf-8')
        hash_bytes = hashlib.sha256(key_bytes).digest()
        hash_value = int.from_bytes(hash_bytes[:8], 'big', signed=True)
        
        return hash_value

class BrainContext(BrainContextReader):
    def __init__(self, db_path: str):
        super().__init__(db_path)

    def set_memory(self, *keys_and_values: Any) -> None:
        if len(keys_and_values) % 2 != 0:
            raise ValueError("键值对数量必须是偶数")
        
        for i in range(0, len(keys_and_values), 2):
            key, value = keys_and_values[i], keys_and_values[i+1]
            self._set_single_memory(key, value)

    def _set_single_memory(self, key: Any, value: Any) -> None:
        try:
            cursor = self.conn.cursor()
            if isinstance(value, str):
                value_type = "string"
            elif isinstance(value, bool):
                value_type = "bool"
            elif isinstance(value, int):
                value_type = "int"
            elif isinstance(value, float):
                value_type = "float"
            else:
                value_type = "json"
            
            value_json = json.dumps(value)
            
            hashed_key = self.hash_key(key)
            cursor.execute("INSERT OR REPLACE INTO memory (key, value, type) VALUES (?, ?, ?)",
                           (str(hashed_key), value_json, value_type))
            self.conn.commit()
        except (sqlite3.Error, json.JSONDecodeError) as e:
            print(f"设置内存错误 ({key}): {e}", file=sys.stderr)

    def delete_memory(self, key: Any) -> None:
        cursor = self.conn.cursor()
        hashed_key = self.hash_key(key)
        cursor.execute("DELETE FROM memory WHERE key = ?", (str(hashed_key),))
        self.conn.commit()

    def clear_memory(self) -> None:
        cursor = self.conn.cursor()
        cursor.execute("DELETE FROM memory")
        self.conn.commit()

    def continue_cast(self) -> None:
        # 这里需要实现继续处理的逻辑
        pass
