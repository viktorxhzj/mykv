# ViktorXHZJ's KV - an imitation of Redis

## Data structures

### ZipList

压缩列表旨在使用连续内存空间尽可能的按需求大小存储字符串（字节数组）和整数。

优点：
- 可以存储不定长的字节数组和整数，极为节约内存

缺点：
- 取出指定位置元素的时间复杂度为O(n)
- 判断是否存在某字节数组的时间复杂度为O(nm)

### IntSet

整数集合旨在使用连续内存空间存储排序好的整数，并根据最大最小元素的值动态调整内存范围。

优点：
- 按最大最小元素的值调整内存范围，极为节约内存
- 元素按序存储，查找时间复杂度为O(nlogn)
- 提供交集等集合操作

缺点：
- 添加、删除时间复杂度为O(n)，因为可能涉及内存调整

### QuickList

快链表底层为非连续内存空间的双向链表，而每个链表节点为压缩列表。平衡了存储效率和查询效率。

优点/缺点：
- 相同于传统链表

### Dictionary

字典为传统哈希表的实现，类似于Java的HashMap和Go的Map。区别在于字典的重哈希并不是在扩容操作中直接完成的，而是分散在增删改查的操作中逐渐完成的。

优点/缺点：
- 相同于传统哈希表，但是重哈希的过程分散在了增删改查各个操作之中，降低了传统扩容操作的耗时

### SkipList

跳跃表，是一种可以用来代替平衡树的数据结构，它采取了概率平衡而不是严格强制的平衡。

优点：
- 插入和删除的算法比平衡树的等效算法简单得多
- 范围查找容易
- 节约内存

缺点：
 - O(logn)查询时间复杂度，弱于哈希表
