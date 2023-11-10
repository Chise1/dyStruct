# 动态结构体

## 子对象访问指南

### 切片
切片分两种，1中是带sliceKey的，另外一种是不带sliceKey的。
如果带sliceKey，则访问子对象的时候，路径后面必须跟上子对象的id，就算是写入也必须带。
如果不带sliceKey，在后面追加则子对象id设置为*，否则设置为数字，如果是数字，能找到则覆盖，不能找到则创建新对象。