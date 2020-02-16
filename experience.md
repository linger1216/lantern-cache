

Local cache设计
得与失公理


得:
很多好处

失:
1. 可以把某些key删除掉, 以换来更多好处




### expire
设计概念如下:

用一个map来保存过期信息

key的生成策略
storageBucketFunc := func (t int64) int64 {
    return (t / 5) + 1
}

是一个uint64, 具体算法为time/5+1, 换句话说每5s会生成一个key
这样好处是利用时间来自然分桶, 能把key通过时间维度, 划分到下一层去处理, 也就是v中处理

v也是一个map, 其中保存了 key hash -> key conflict (第二种hash)
cleanupBucketFunc := func (t int64) int64 {
    return storageBucketFunc(t) - 1
}
清除策略, 每隔N时间, 取合适的桶, 取的策略是cleanupBucketFunc, 算法差不多, 意思就是取前面的一些桶进行处理,
但这个算法有个局限性, 那就是必须清除的时间不能太短, 这样保证每次都能抓到旧的桶, 但如果时间比较长,才执行前面的桶会丢失

time:0s storageBucket:1 cleanupBucket:0
time:1s storageBucket:1 cleanupBucket:0
time:2s storageBucket:1 cleanupBucket:0
time:3s storageBucket:1 cleanupBucket:0
time:4s storageBucket:1 cleanupBucket:0
time:5s storageBucket:2 cleanupBucket:1
time:6s storageBucket:2 cleanupBucket:1
time:7s storageBucket:2 cleanupBucket:1
time:8s storageBucket:2 cleanupBucket:1
time:9s storageBucket:2 cleanupBucket:1
time:10s storageBucket:3 cleanupBucket:2
time:11s storageBucket:3 cleanupBucket:2
time:12s storageBucket:3 cleanupBucket:2
time:13s storageBucket:3 cleanupBucket:2
time:14s storageBucket:3 cleanupBucket:2
time:15s storageBucket:4 cleanupBucket:3

当0-4s的时候, 清除的是0桶, 5s清除的1桶, 正好1桶里面有数据, 这样5s一次清理, 会一直下去, 挺好
但时间如果长一点进行清理, 比如12s的时候清理的是2桶, 那么1桶的呢? 根据算法再也不会找到1桶了, 所以这个k生成策略确定了, 清理周期不能
太长超过10s, 这就是局限性.



基于这个考虑, 我们要去除这方面的依赖, 需要重新设计expire的策略.
过期的需求就是把那些本该删除的key, 从系统中删除掉

┌──────┬─────┬──────┐
│ 64k  │ 64k │ 64k  │
│      │     │      │
└──────┴─────┴──────┘

每个64k都是如下的结构[]uint64
┌──────┬──────┬──────┬──────┬──────┬──────┬──────┬──────┐
│  3   │  99  │  12  │  7   │  4   │  55  │  6   │      │
│      │      │      │      │      │      │      │      │
└──────┴──────┴──────┴──────┴──────┴──────┴──────┴──────┘
用2个变量记录读写位置

当清理时, 从读的地方开始读, 如果发现当前的key过期了, 那么将写位置(末尾)的元素替换到读位置, 再次重新判断
如果没有过期r++, 直到rw重合, 这样遍历下来有效的数据总是在头部, 利于下一次遍历

但有个问题, 当前已经有了10个chunk, chunk1清理完毕, 还有一些元素, 如果新key到来还写到chunk1去的话, 此时发生了清理, chunk1, 清除掉大部分
元素, 接着又写入了大部分数据, 又清理..这样chunk2可能有很少的机会进行清理, 如果每次我们不止清理一个元素的话, 其实也很容易热度只发生在前面几个

为此我觉得应该这样设计
1. 新数据始终往最新的chunk里面写
2. 清理只发生在旧的chunk, 具体几个在说
3. 当一个chunk全部清除后, 可以作为可用内存直接利用下次分配64K, 循环利用

但2又引入了一个问题, 旧的chunk, 每次都清理不完, 因为还有数据, 不能被再利用, 但他的时间又真的很长, 比如1年之类的,
这种情况下, 我们可以通过
11. 将数量很少的chunk直接丢弃掉.
22. 将数量很少的chunk, copy到下一个chunk去.
33. 或者记录清理次数, 如果超过多少次就直接丢弃, 全部删除, 谁叫你运气不好, 投错了胎

但也可能有多提, 如果用22, 会带来频繁的copy, 可能, 可能,



好处就是成功避免了map还是双重map, 只使用数组节省了gc, 省了膨胀.


### store