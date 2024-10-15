#ifndef PYPROCESSOR_WRAPPER_H
#define PYPROCESSOR_WRAPPER_H

#include <Python.h>

// 初始化 Python 解释器
static inline void initPython() {
    Py_Initialize();
}

// 检查 Python 是否已初始化
static inline int pyIsInitialized() {
    return Py_IsInitialized();
}

// 获取 GIL
static inline PyGILState_STATE pyGILStateEnsure() {
    return PyGILState_Ensure();
}

// 释放 GIL
static inline void pyGILStateRelease(PyGILState_STATE state) {
    PyGILState_Release(state);
}

// 获取属性
static inline PyObject* pyObjectGetAttrString(PyObject *o, const char *attr_name) {
    return PyObject_GetAttrString(o, attr_name);
}

// 检查是否为 Unicode 对象
static inline int pyUnicodeCheck(PyObject *o) {
    return PyUnicode_Check(o);
}

// 将 Unicode 对象转换为 UTF-8 字符串
static inline const char* pyUnicodeAsUTF8(PyObject *unicode) {
    return PyUnicode_AsUTF8(unicode);
}

// 导入模块
static inline PyObject* importModule(const char* moduleName) {
    return PyImport_ImportModule(moduleName);
}

// 获取类
static inline PyObject* getClass(PyObject* module, const char* className) {
    return PyObject_GetAttrString(module, className);
}

// 实例化 Python 类
static inline PyObject* createInstance(PyObject* class) {
    return PyObject_CallObject(class, NULL);
}

// 创建新的元组
static inline PyObject* pyTupleNew(Py_ssize_t size) {
    return PyTuple_New(size);
}

// 设置元组项
static inline int pyTupleSetItem(PyObject *p, Py_ssize_t pos, PyObject *o) {
    return PyTuple_SetItem(p, pos, o);
}

// 调用对象
static inline PyObject* pyObjectCallObject(PyObject *callable, PyObject *args) {
    return PyObject_CallObject(callable, args);
}

#endif // PYPROCESSOR_WRAPPER_H
