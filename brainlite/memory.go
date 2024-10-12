package brainlite

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/zenmodel/zenmodel/internal/errors"

	_ "github.com/mattn/go-sqlite3"
)

type BrainMemory struct {
	db             *sql.DB
	datasourceName string
	// 是否在 brain Shutdown 时保留数据库文件
	keepMemory     bool
}


func (m *BrainMemory)Init() error {
	db, err := sql.Open("sqlite3", m.datasourceName)
	if err != nil {
		return errors.Wrapf(err, "init memory failed")
	}
	m.db = db

	// 创建 memory 表
	_, err = m.db.Exec(`CREATE TABLE IF NOT EXISTS memory (
		key INTEGER PRIMARY KEY,
		value JSON,
		type TEXT
	)`)
	if err != nil {
		return errors.Wrapf(err, "init memory table failed")
	}

	return nil
}

func (m *BrainMemory)Set(key, value any) error {
	var valueType string
	var valueJSON []byte
	var err error

	hashedKey, err := hashKey(key)
	if err != nil {
		return fmt.Errorf("无法哈希键: %v", err)
	}

	switch value.(type) {
	case string:
		valueType = "string"
	case int, int32, int64, uint32:
		valueType = "int"
	case float64:
		valueType = "float"
	case bool:
		valueType = "bool"
	default:
		valueType = "json"
	}

	valueJSON, err = json.Marshal(value)
	if err != nil {
		return fmt.Errorf("无法序列化值: %v", err)
	}

	_, err = m.db.Exec("INSERT OR REPLACE INTO memory (key, value, type) VALUES (?, ?, ?)",
		hashedKey, valueJSON, valueType)
	if err != nil {
		return fmt.Errorf("存储数据时出错: %v", err)
	}

	return nil
}

func (m *BrainMemory)Get(key any) (any, error) {
	hashedKey, err 	:= hashKey(key)
	if err != nil {
		return nil, fmt.Errorf("无法哈希键: %v", err)
	}

	var valueJSON []byte
	var valueType string
	err = m.db.QueryRow("SELECT value, type FROM memory WHERE key = ?", hashedKey).Scan(&valueJSON, &valueType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到键 '%v'", key)
		}
		return nil, fmt.Errorf("查询数据时出错: %v", err)
	}

	var value any
	switch valueType {
	case "string":
		err = json.Unmarshal(valueJSON, &value)
	case "int":
		var intValue int64
		err = json.Unmarshal(valueJSON, &intValue)
		if err != nil {
			return nil, err
		}
		// 根据数值范围选择合适的类型
		switch {
		case intValue >= int64(math.MinInt32) && intValue <= int64(math.MaxInt32):
			value = int(intValue)
		default:
			value = intValue
		}
	case "float":
		var floatValue float64
		err = json.Unmarshal(valueJSON, &floatValue)
		value = floatValue
	case "bool":
		var boolValue bool
		err = json.Unmarshal(valueJSON, &boolValue)
		value = boolValue
	case "json":
		err = json.Unmarshal(valueJSON, &value)
	}

	if err != nil {
		return nil, fmt.Errorf("解析数据时出错: %v", err)
	}

	return value, nil
}

func (m *BrainMemory)Del(key any) error {
	hashedKey, err := hashKey(key)
	if err != nil {
		return fmt.Errorf("无法哈希键: %v", err)
	}

	_, err = m.db.Exec("DELETE FROM memory WHERE key = ?", hashedKey)
	if err != nil {
		return fmt.Errorf("删除数据时出错: %v", err)
	}

	return nil
}

func (m *BrainMemory)Clear() error {
	_, err := m.db.Exec("DELETE FROM memory")
	if err != nil {
		return fmt.Errorf("清空数据时出错: %v", err)
	}

	return nil
}

func (m *BrainMemory)Close() error {
	if err := m.db.Close(); err != nil {
		return err
	}
	m.db = nil

	if !m.keepMemory {
		if err := os.Remove(m.datasourceName); err != nil {
			return fmt.Errorf("删除数据库文件时出错: %v", err)
		}
	}


	return nil
}	

// hashKey 将任意类型的 key 转换为 int64
func hashKey(key any) (int64, error) {
    switch key.(type) {
    case int, int32, int64, uint32, uint64, float64, string, []byte, byte:
        // 继续处理
    default:
        return 0, fmt.Errorf("unsupported key type %T", key)
    }

    h := sha256.New()
    h.Write([]byte(fmt.Sprintf("%v", key)))
    hashBytes := h.Sum(nil)
    // 取前8个字节并转换为 int64
    return int64(binary.BigEndian.Uint64(hashBytes[:8])), nil
}